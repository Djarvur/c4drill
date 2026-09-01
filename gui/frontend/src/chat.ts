// chat.ts is the P1 AI chat panel: provider settings (base URL / model /
// API key stored in local app config only), streaming conversation, and
// edit proposals that apply ONLY on explicit confirmation.

import { backend, call } from "./rpc";
import type { ChatEvent, ChatMessage, Diag, ProposedEdit } from "./types";

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

export class ChatPanel {
  private root: HTMLElement;

  private messagesEl: HTMLElement;

  private formEl: HTMLElement;

  private settingsBtn: HTMLButtonElement;

  private input: HTMLTextAreaElement;

  private sendBtn: HTMLButtonElement;

  private transcript: ChatMessage[] = [];

  private pending: Map<string, HTMLElement> = new Map();

  private proposalViews: ProposalView[] = [];

  private streaming = false;

  constructor(parent: HTMLElement, private hooks: ChatHooks) {
    parent.insertAdjacentHTML("beforeend", `
      <div class="chat-pane" id="chat-pane">
        <div class="chat-header">
          <span>AI ASSISTANT</span>
          <button id="chat-settings-btn" title="Provider settings">⚙</button>
        </div>
        <div class="chat-settings" id="chat-settings" hidden>
          <label>Base URL <input id="chat-base-url" placeholder="http://localhost:11434/v1" /></label>
          <label>Model <input id="chat-model" placeholder="llama3.1 / gpt-4o-mini" /></label>
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

    this.settingsBtn.addEventListener("click", () => this.toggleSettings());
    this.root.querySelector<HTMLButtonElement>("#chat-save-settings")!
      .addEventListener("click", () => void this.saveSettings());
    this.sendBtn.addEventListener("click", () => void this.send());
    this.input.addEventListener("keydown", (ev) => {
      if (ev.key === "Enter" && !ev.shiftKey) {
        ev.preventDefault();
        void this.send();
      }
    });

    backend.on("chat", (frame: ChatEvent) => this.onChatFrame(frame));

    void this.loadSettings();
  }

  /** loadSettings fills the form from the backend (key masked). */
  private async loadSettings(): Promise<void> {
    try {
      const res = await call<{ config: { baseURL: string; model: string; hasAPIKey: boolean; systemPrompt: string } }>("chatConfig");
      this.setFormValue("#chat-base-url", res.config.baseURL);
      this.setFormValue("#chat-model", res.config.model);
      this.setFormValue("#chat-api-key", "");
      (this.root.querySelector<HTMLInputElement>("#chat-api-key") as HTMLInputElement).placeholder =
        res.config.hasAPIKey ? "(saved — type to replace)" : "(stored in local app config)";
      this.setFormValue("#chat-prompt", res.config.systemPrompt);
    } catch {
      // settings not configured yet: the form starts empty
    }
  }

  private setFormValue(selector: string, value: string): void {
    const el = this.root.querySelector<HTMLInputElement | HTMLTextAreaElement>(selector);
    if (el) el.value = value;
  }

  private formValue(selector: string): string {
    return this.root.querySelector<HTMLInputElement | HTMLTextAreaElement>(selector)?.value ?? "";
  }

  private toggleSettings(): void {
    this.formEl.hidden = !this.formEl.hidden;
  }

  private async saveSettings(): Promise<void> {
    const msg = this.root.querySelector<HTMLElement>("#chat-settings-msg")!;

    try {
      await call("saveChatConfig", {
        baseURL: this.formValue("#chat-base-url").trim(),
        apiKey: this.formValue("#chat-api-key"),
        model: this.formValue("#chat-model").trim(),
        systemPrompt: this.formValue("#chat-prompt"),
      });
      msg.textContent = "saved";
      await this.loadSettings();
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

      this.appendStreamTarget(res.requestID);
    } catch (err) {
      this.streaming = false;
      this.sendBtn.disabled = false;
      this.appendBubble("assistant", `⚠ ${err}`);
    }
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
    this.streaming = false;
    this.sendBtn.disabled = false;

    if (frame.error) {
      bubble.textContent = bubble.textContent
        ? `${bubble.textContent}\n\n⚠ ${frame.error}`
        : `⚠ ${frame.error}`;
    }

    const answer = frame.answer ?? bubble.textContent;
    this.transcript.push({ role: "assistant", content: answer });

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
