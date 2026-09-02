// chat.ts is the P1 AI chat panel: provider settings (per-provider base URL
// / model / API key stored in local app config only), streaming conversation,
// a stop control for in-flight requests (issue #36), and edit proposals that
// apply ONLY on explicit confirmation.

import { backend, call } from "./rpc";
import type { ChatConfigResult, ChatEvent, ChatMessage, ChatProvider, Diag, ProposedEdit, PublicChatSettings } from "./types";

export interface ChatHooks {
  activePath(): string;
  activeText(): string;
  activeSelection(): string;
  activeDiagnostics(): Diag[];
  onEditsApplied(paths: string[]): void;
}

interface ProposalView {
  proposal: ProposedEdit;
  diff: DiffLine[];
  applied: boolean;
}

interface DiffLine {
  kind: "ctx" | "add" | "del";
  text: string;
}

/** Per-provider base URL hint (issue #36). */
const PROVIDER_BASE_URL_PLACEHOLDER: Record<ChatProvider, string> = {
  "openai-compatible": "http://localhost:11434/v1",
  anthropic: "https://api.anthropic.com",
};

/** FieldValues are the editable settings fields for one provider slot. */
interface FieldValues {
  baseURL: string;
  model: string;
  apiKey: string;
  systemPrompt: string;
}

const FIELD_SELECTORS: Record<keyof FieldValues, string> = {
  baseURL: "#chat-base-url",
  model: "#chat-model",
  apiKey: "#chat-api-key",
  systemPrompt: "#chat-prompt",
};

export class ChatPanel {
  private root: HTMLElement;

  private messagesEl: HTMLElement;

  private formEl: HTMLElement;

  private settingsBtn: HTMLButtonElement;

  private input: HTMLTextAreaElement;

  private sendBtn: HTMLButtonElement;

  private stopBtn: HTMLButtonElement;

  private transcript: ChatMessage[] = [];

  private pending: Map<string, HTMLElement> = new Map();

  private proposalViews: ProposalView[] = [];

  private streaming = false;

  /** currentRequestID is the in-flight request — the Stop button's target. */
  private currentRequestID: string | null = null;

  /** activeProvider is the provider whose fields the form currently shows. */
  private activeProvider: ChatProvider = "openai-compatible";

  /** fieldCache holds per-provider form values: saved slots from the backend
   * plus unsaved edits, so switching providers never clobbers the other
   * slot's settings (issue #36). */
  private fieldCache: Partial<Record<ChatProvider, FieldValues>> = {};

  constructor(parent: HTMLElement, private hooks: ChatHooks) {
    parent.insertAdjacentHTML("beforeend", `
      <div class="chat-pane" id="chat-pane">
        <div class="chat-header">
          <span>AI ASSISTANT</span>
          <button id="chat-settings-btn" title="Provider settings">⚙</button>
        </div>
        <div class="chat-settings" id="chat-settings" hidden>
          <label>Provider
            <select id="chat-provider">
              <option value="openai-compatible">OpenAI-compatible</option>
              <option value="anthropic">Anthropic</option>
            </select>
          </label>
          <label>Base URL <input id="chat-base-url" placeholder="http://localhost:11434/v1" /></label>
          <label>Model <input id="chat-model" placeholder="llama3.1 / gpt-4o-mini / claude-sonnet" /></label>
          <label>API key <input id="chat-api-key" type="password" placeholder="(stored in local app config)" /></label>
          <label>Extra system prompt
            <textarea id="chat-prompt" rows="3" placeholder="optional additional instructions"></textarea>
          </label>
          <button id="chat-save-settings">Save settings</button>
          <div id="chat-settings-msg" class="chat-settings-msg"></div>
        </div>
        <div class="chat-messages" id="chat-messages"></div>
        <div class="chat-input-row">
          <textarea id="chat-input" rows="2" placeholder="Ask about the architecture… (Enter to send)"></textarea>
          <button id="chat-stop" title="Stop generating" class="chat-stop" hidden>■ Stop</button>
          <button id="chat-send" title="Send">Send</button>
        </div>
      </div>
    `);

    this.root = parent.querySelector<HTMLElement>("#chat-pane")!;
    this.messagesEl = this.root.querySelector<HTMLElement>("#chat-messages")!;
    this.formEl = this.root.querySelector<HTMLElement>("#chat-settings")!;
    this.settingsBtn = this.root.querySelector<HTMLButtonElement>("#chat-settings-btn")!;
    this.input = this.root.querySelector<HTMLTextAreaElement>("#chat-input")!;
    this.sendBtn = this.root.querySelector<HTMLButtonElement>("#chat-send")!;
    this.stopBtn = this.root.querySelector<HTMLButtonElement>("#chat-stop")!;

    this.settingsBtn.addEventListener("click", () => this.toggleSettings());
    this.root.querySelector<HTMLButtonElement>("#chat-save-settings")!
      .addEventListener("click", () => void this.saveSettings());
    this.sendBtn.addEventListener("click", () => void this.send());
    this.stopBtn.addEventListener("click", () => void this.stop());
    this.root.querySelector<HTMLSelectElement>("#chat-provider")!
      .addEventListener("change", (ev) => this.switchProvider((ev.target as HTMLSelectElement).value as ChatProvider));
    this.input.addEventListener("keydown", (ev) => {
      if (ev.key === "Enter" && !ev.shiftKey) {
        ev.preventDefault();
        void this.send();
      }
    });

    backend.on("chat", (frame: ChatEvent) => this.onChatFrame(frame));

    void this.loadSettings();
  }

  /** loadSettings fills the form from the backend (keys masked). */
  private async loadSettings(): Promise<void> {
    try {
      const res = await call<ChatConfigResult>("chatConfig");

      this.fieldCache = {};
      for (const [provider, cfg] of Object.entries(res.providers ?? {}) as [ChatProvider, PublicChatSettings][]) {
        this.fieldCache[provider] = {
          baseURL: cfg.baseURL,
          model: cfg.model,
          apiKey: "",
          systemPrompt: cfg.systemPrompt,
        };
      }

      this.setProvider(res.provider ?? "openai-compatible", false);

      const apiKeyInput = this.root.querySelector<HTMLInputElement>("#chat-api-key")!;
      apiKeyInput.placeholder = res.config.hasAPIKey ? "(saved — type to replace)" : "(stored in local app config)";
    } catch {
      // settings not configured yet: the form starts empty
    }
  }

  /** setProvider shows one provider slot's fields in the form. */
  private setProvider(provider: ChatProvider, keepCached = true): void {
    if (keepCached && this.activeProvider !== provider) {
      this.fieldCache[this.activeProvider] = this.readFields();
    }

    this.activeProvider = provider;
    (this.root.querySelector<HTMLSelectElement>("#chat-provider")!).value = provider;

    const values = this.fieldCache[provider] ?? { baseURL: "", model: "", apiKey: "", systemPrompt: "" };

    for (const field of Object.keys(FIELD_SELECTORS) as (keyof FieldValues)[]) {
      const el = this.root.querySelector<HTMLInputElement | HTMLTextAreaElement>(FIELD_SELECTORS[field])!;
      el.value = values[field];
    }

    (this.root.querySelector<HTMLInputElement>("#chat-base-url")!).placeholder =
      PROVIDER_BASE_URL_PLACEHOLDER[provider];
  }

  /** switchProvider stashes the outgoing slot's edits and swaps fields. */
  private switchProvider(provider: ChatProvider): void {
    this.setProvider(provider);
  }

  private readFields(): FieldValues {
    const values = {} as FieldValues;
    for (const field of Object.keys(FIELD_SELECTORS) as (keyof FieldValues)[]) {
      values[field] = this.root.querySelector<HTMLInputElement | HTMLTextAreaElement>(FIELD_SELECTORS[field])!.value;
    }

    return values;
  }

  private toggleSettings(): void {
    this.formEl.hidden = !this.formEl.hidden;
  }

  private async saveSettings(): Promise<void> {
    const msg = this.root.querySelector<HTMLElement>("#chat-settings-msg")!;
    const provider = (this.root.querySelector<HTMLSelectElement>("#chat-provider")!).value as ChatProvider;
    const fields = this.readFields();

    try {
      const saved = await call<{ provider: ChatProvider }>("saveChatConfig", {
        provider,
        baseURL: fields.baseURL.trim(),
        apiKey: fields.apiKey,
        model: fields.model.trim(),
        systemPrompt: fields.systemPrompt,
      });

      this.fieldCache[saved.provider] = { ...fields, apiKey: "" };
      this.setProvider(saved.provider, false);

      const apiKeyInput = this.root.querySelector<HTMLInputElement>("#chat-api-key")!;
      apiKeyInput.placeholder = "(saved — type to replace)";
      msg.textContent = "saved";
    } catch (err) {
      msg.textContent = String(err);
    }
  }

  /** send posts the user message with the current authoring context. */
  private async send(): Promise<void> {
    const text = this.input.value.trim();
    if (!text || this.streaming) return;

    this.input.value = "";
    this.appendBubble("user", text);

    const diagnostics = this.hooks.activeDiagnostics()
      .map((d) => `${this.hooks.activePath()}: ${d.message}`);

    this.streaming = true;
    this.sendBtn.disabled = true;
    this.stopBtn.hidden = false;

    try {
      const res = await call<{ requestID: string }>("chat", {
        history: this.transcript,
        text,
        ctx: {
          activeFile: this.hooks.activePath(),
          activeContent: this.hooks.activeText(),
          selection: this.hooks.activeSelection(),
          diagnostics,
        },
      });

      this.currentRequestID = res.requestID;
      this.appendStreamTarget(res.requestID);
    } catch (err) {
      this.resetComposer();
      this.appendBubble("assistant", `⚠ ${err}`);
    }
  }

  /** stop aborts the in-flight request (issue #36). The backend cancels the
   * provider stream; the terminal frame keeps the partial answer and carries
   * the aborted marker that finishes the composer state. */
  private async stop(): Promise<void> {
    if (!this.currentRequestID) return;

    const requestID = this.currentRequestID;
    this.currentRequestID = null;
    this.stopBtn.disabled = true;

    try {
      await call("chatAbort", { requestID });
    } catch {
      // the request may already have finished — the terminal frame resolves
      // the composer state either way
    }
  }

  /** resetComposer restores the input row after a request ends or fails. */
  private resetComposer(): void {
    this.streaming = false;
    this.currentRequestID = null;
    this.sendBtn.disabled = false;
    this.stopBtn.hidden = true;
    this.stopBtn.disabled = false;
    this.input.focus();
  }

  /** appendStreamTarget creates the assistant bubble that deltas append to. */
  private appendStreamTarget(requestID: string): void {
    const bubble = this.appendBubble("assistant", "");
    bubble.dataset.requestId = requestID;
    this.pending.set(requestID, bubble);
  }

  /** onChatFrame applies one streamed frame (delta or terminal). */
  private onChatFrame(frame: ChatEvent): void {
    const bubble = this.pending.get(frame.requestID);
    if (!bubble) return;

    if (frame.delta) {
      bubble.textContent += frame.delta;
      this.messagesEl.scrollTop = this.messagesEl.scrollHeight;
    }

    if (!frame.done) return;

    this.pending.delete(frame.requestID);
    this.resetComposer();

    if (frame.error) {
      bubble.textContent = bubble.textContent
        ? `${bubble.textContent}\n\n⚠ ${frame.error}`
        : `⚠ ${frame.error}`;
    }

    // Capture the transcript text before any aborted marker is appended.
    const answer = frame.answer ?? bubble.textContent;

    if (frame.aborted) {
      bubble.classList.add("aborted");
      bubble.insertAdjacentHTML("beforeend", '<div class="aborted-marker">⏹ stopped — answer is partial</div>');
    }

    this.transcript.push({ role: "assistant", content: answer });

    if (frame.aborted) return; // partial answers never carry proposals

    this.proposalViews = (frame.proposals ?? []).map((proposal) => ({
      proposal,
      diff: lineDiff(proposal.oldContent, proposal.newContent),
      applied: false,
    }));

    for (const view of this.proposalViews) {
      bubble.appendChild(this.renderProposal(view));
    }
  }

  /** renderProposal builds one confirmation card: diff + Apply/Discard. */
  private renderProposal(view: ProposalView): HTMLElement {
    const card = document.createElement("div");
    card.className = `proposal${view.proposal.valid ? "" : " invalid"}`;

    const head = document.createElement("div");
    head.className = "proposal-head";
    head.textContent = view.proposal.valid
      ? `Proposed edit: ${view.proposal.path}`
      : `Invalid proposal (${view.proposal.path}): ${view.proposal.error}`;
    card.appendChild(head);

    if (view.proposal.valid) {
      const diffEl = document.createElement("pre");
      diffEl.className = "proposal-diff";

      for (const line of view.diff) {
        if (line.kind === "ctx") continue; // context lines: keep the card small

        const row = document.createElement("div");
        row.className = `diff-${line.kind}`;
        row.textContent = `${line.kind === "add" ? "+" : "−"} ${line.text}`;
        diffEl.appendChild(row);
      }

      card.appendChild(diffEl);

      const actions = document.createElement("div");
      actions.className = "proposal-actions";

      const apply = document.createElement("button");
      apply.className = "proposal-apply";
      apply.textContent = "Apply";
      apply.addEventListener("click", () => void this.apply(view, card));

      const discard = document.createElement("button");
      discard.textContent = "Discard";
      discard.addEventListener("click", () => card.remove());

      actions.appendChild(apply);
      actions.appendChild(discard);
      card.appendChild(actions);
    }

    return card;
  }

  /** apply writes the proposal — the explicit confirmation gate. */
  private async apply(view: ProposalView, card: HTMLElement): Promise<void> {
    try {
      const res = await call<{ applied: string[]; errors: string[] }>("applyEdits", {
        proposals: [view.proposal],
      });

      if (res.applied.includes(view.proposal.path)) {
        view.applied = true;
        card.classList.add("applied");
        card.insertAdjacentHTML("beforeend", '<div class="proposal-applied">applied</div>');
        this.hooks.onEditsApplied([view.proposal.path]);
      } else {
        card.insertAdjacentHTML("beforeend",
          `<div class="proposal-error">${res.errors.join("; ")}</div>`);
      }
    } catch (err) {
      card.insertAdjacentHTML("beforeend", `<div class="proposal-error">${err}</div>`);
    }
  }

  /** appendBubble adds one transcript bubble and records user messages. */
  private appendBubble(role: "user" | "assistant", text: string): HTMLElement {
    const bubble = document.createElement("div");
    bubble.className = `bubble bubble-${role}`;
    bubble.textContent = text;
    this.messagesEl.appendChild(bubble);
    this.messagesEl.scrollTop = this.messagesEl.scrollHeight;

    if (role === "user") {
      this.transcript.push({ role: "user", content: text });
    }

    return bubble;
  }
}

/** lineDiff is the frontend's small LCS line diff for the confirmation UI. */
function lineDiff(oldText: string, newText: string): DiffLine[] {
  const a = oldText.replace(/\r\n/g, "\n").split("\n");
  const b = newText.replace(/\r\n/g, "\n").split("\n");

  // lcs[i][j] = LCS length of a[i:] vs b[j:].
  const lcs: number[][] = Array.from({ length: a.length + 1 }, () => new Array<number>(b.length + 1).fill(0));

  for (let i = a.length - 1; i >= 0; i--) {
    for (let j = b.length - 1; j >= 0; j--) {
      lcs[i][j] = a[i] === b[j] ? lcs[i + 1][j + 1] + 1 : Math.max(lcs[i + 1][j], lcs[i][j + 1]);
    }
  }

  const out: DiffLine[] = [];
  let i = 0;
  let j = 0;

  while (i < a.length && j < b.length) {
    if (a[i] === b[j]) {
      out.push({ kind: "ctx", text: a[i] });
      i++;
      j++;
    } else if (lcs[i + 1][j] >= lcs[i][j + 1]) {
      out.push({ kind: "del", text: a[i] });
      i++;
    } else {
      out.push({ kind: "add", text: b[j] });
      j++;
    }
  }

  while (i < a.length) out.push({ kind: "del", text: a[i++] });
  while (j < b.length) out.push({ kind: "add", text: b[j++] });

  return out;
}
