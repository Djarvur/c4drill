// main.ts assembles the three-area layout: file tree + tabs + CodeMirror
// editor (left), live preview with the render-controls toolbar (center), and
// (in P1) the AI chat panel (right).

import "./style.css";
import { backend, call, isDesktop, wailsRuntime } from "./rpc";
import { EditorArea } from "./editor";
import { Preview } from "./preview";
import type { AppInfo, DiagnosticsEvent, ExportResult, FileContent, ProjectInfo, RenderOptions } from "./types";

interface OpenTab {
  path: string;
  saved: string; // last saved/loaded text (dirty tracking)
}

const appEl = document.querySelector<HTMLDivElement>("#app")!;

appEl.innerHTML = `
  <div class="titlebar">
    <span class="logo">c4drill</span>
    <div class="project-controls">
      <input id="project-dir" type="text" placeholder="path/to/project" spellcheck="false" />
      <button id="btn-open" title="Open project directory">Open</button>
      <button id="btn-browse" title="Browse…" hidden>Browse…</button>
    </div>
    <span id="project-label" class="project-label"></span>
  </div>
  <div class="layout">
    <div class="sidebar">
      <div class="sidebar-header">FILES</div>
      <div id="file-tree" class="file-tree"></div>
    </div>
    <div class="editor-pane">
      <div id="tabs" class="tabs"></div>
      <div id="editor-host" class="editor-host"></div>
      <div class="editor-status">
        <button id="btn-format" title="Format document (LSP formatting)">Format</button>
        <button id="btn-save" title="Save (Mod-S)">Save</button>
        <button id="btn-definition" title="Go to definition">Definition</button>
        <span id="editor-hint" class="hint">Mod-Space completion · hover for peer info</span>
      </div>
    </div>
    <div class="preview-pane">
      <div class="toolbar" id="toolbar">
        <label>Level
          <select id="level-select">
            <option value="">C1 — context</option>
          </select>
        </label>
        <label class="check"><input type="checkbox" id="chk-expanded-all" /> Expand all</label>
        <label class="check"><input type="checkbox" id="chk-legend" /> Legend</label>
        <label class="check"><input type="checkbox" id="chk-pause" /> Pause auto-render</label>
        <span class="spacer"></span>
        <span id="expanded-chips" class="chips"></span>
        <span class="spacer"></span>
        <button id="btn-zoom-in" title="Zoom in">+</button>
        <button id="btn-zoom-out" title="Zoom out">−</button>
        <button id="btn-zoom-reset" title="Reset view">Fit</button>
        <select id="export-select" title="Export current diagram">
          <option value="">Export…</option>
          <option value="svg">SVG</option>
          <option value="html">HTML</option>
          <option value="dot">DOT</option>
          <option value="png">PNG</option>
          <option value="plantuml">PlantUML</option>
        </select>
      </div>
      <div class="breadcrumbs" id="breadcrumbs"></div>
      <div class="preview-container" id="preview-container">
        <div class="preview-stage" id="preview-stage"></div>
      </div>
      <div class="preview-error" id="preview-error" hidden></div>
      <div class="preview-status-bar">
        <span id="preview-status"></span>
        <span class="hint">click a node to drill down · drag to pan · wheel to zoom</span>
      </div>
    </div>
  </div>
  <div id="toast" class="toast" hidden></div>
`;

const editorHost = appEl.querySelector<HTMLElement>("#editor-host")!;
const fileTreeEl = appEl.querySelector<HTMLElement>("#file-tree")!;
const tabsEl = appEl.querySelector<HTMLElement>("#tabs")!;
const projectInput = appEl.querySelector<HTMLInputElement>("#project-dir")!;
const projectLabel = appEl.querySelector<HTMLElement>("#project-label")!;
const levelSelect = appEl.querySelector<HTMLSelectElement>("#level-select")!;
const expandedChips = appEl.querySelector<HTMLElement>("#expanded-chips")!;
const exportSelect = appEl.querySelector<HTMLSelectElement>("#export-select")!;
const toastEl = appEl.querySelector<HTMLElement>("#toast")!;

const openTabs: OpenTab[] = [];
let activeTab: string | null = null;
let projectDir = "";

// --- editor + preview wiring ------------------------------------------------

const editor = new EditorArea(editorHost, {
  onChange(_path, _text, _version) {
    // Push every buffer update to the LSP (in-process, cheap); the preview
    // re-render is debounced inside Preview.
    void call("didChange", { path: _path, text: _text, version: _version }).catch(() => undefined);
    preview.invalidate();
    markDirty(_path);
  },
  onSave(path, text) {
    void saveFile(path, text);
  },
});

const preview = new Preview(
  appEl.querySelector<HTMLElement>("#preview-container")!,
  appEl.querySelector<HTMLElement>("#breadcrumbs")!,
  appEl.querySelector<HTMLElement>("#preview-error")!,
  appEl.querySelector<HTMLElement>("#preview-status")!,
  {
    onTargetChange(target) {
      syncLevelSelect(target);
      renderExpandedChips();
    },
    onDiagnostics(diags) {
      if (editor.currentPath() && diags) {
        editor.applyDiagnostics(diags);
      }
    },
  },
);

// --- diagnostics events ------------------------------------------------------

backend.on("diagnostics", (payload: DiagnosticsEvent) => {
  if (payload.path === editor.currentPath()) {
    editor.applyDiagnostics(payload.diagnostics ?? []);
  }
});

// --- project + files ----------------------------------------------------------

async function openProject(dir: string): Promise<void> {
  if (!dir.trim()) return;

  try {
    const info = await call<ProjectInfo>("openProject", { dir: dir.trim() });
    projectDir = info.dir;
    projectInput.value = info.dir;
    projectLabel.textContent = info.dir;
    await refreshTree();
    await toast(`project opened: ${info.files.length} model file(s)`);
  } catch (err) {
    await toast(String(err), true);
  }
}

async function refreshTree(): Promise<void> {
  const info = await call<ProjectInfo>("listFiles");
  fileTreeEl.innerHTML = "";

  if (!info.files.length) {
    const empty = document.createElement("div");
    empty.className = "tree-empty";
    empty.textContent = "no .toml / .c4d files";
    fileTreeEl.appendChild(empty);
    return;
  }

  for (const file of info.files) {
    const item = document.createElement("button");
    item.className = "tree-item";
    item.textContent = file.path;
    item.title = file.path;
    item.addEventListener("click", () => void openFile(file.path));
    fileTreeEl.appendChild(item);
  }
}

async function openFile(path: string): Promise<void> {
  try {
    let text: string;

    const existing = openTabs.find((t) => t.path === path);
    if (existing) {
      text = existing.saved;
    } else {
      const content = await call<FileContent>("readFile", { path });
      text = content.text;
      openTabs.push({ path, saved: text });
      void call("didOpen", { path, text }).catch(() => undefined);
    }

    activateTab(path);
    editor.openDoc(path, text);
    preview.setFile(path);
    await preview.renderNow();
    syncLevelSelect("");
    renderExpandedChips();
  } catch (err) {
    await toast(String(err), true);
  }
}

function activateTab(path: string): void {
  activeTab = path;
  tabsEl.innerHTML = "";

  for (const tab of openTabs) {
    const el = document.createElement("button");
    el.className = `tab${tab.path === activeTab ? " active" : ""}`;
    el.textContent = fileName(tab.path);
    el.title = tab.path;
    el.addEventListener("click", () => void openFile(tab.path));

    const close = document.createElement("span");
    close.className = "tab-close";
    close.textContent = "×";
    close.addEventListener("click", (ev) => {
      ev.stopPropagation();
      void closeTab(tab.path);
    });
    el.appendChild(close);
    tabsEl.appendChild(el);
  }
}

async function closeTab(path: string): Promise<void> {
  const idx = openTabs.findIndex((t) => t.path === path);
  if (idx < 0) return;

  openTabs.splice(idx, 1);
  await call("didClose", { path }).catch(() => undefined);

  if (activeTab === path) {
    activeTab = null;
    if (openTabs.length) {
      await openFile(openTabs[Math.max(0, idx - 1)].path);
    } else {
      tabsEl.innerHTML = "";
      editor.openDoc("", "");
    }
  } else {
    activateTab(activeTab!);
  }
}

function markDirty(path: string): void {
  const tab = openTabs.find((t) => t.path === path);
  if (tab) {
    (tab as OpenTab & { dirty?: boolean }).dirty = true;
  }
  [...tabsEl.children].forEach((el) => {
    if (el.textContent === fileName(path)) el.classList.add("dirty");
  });
}

async function saveFile(path: string, text: string): Promise<void> {
  try {
    await call("writeFile", { path, text });
    const tab = openTabs.find((t) => t.path === path);
    if (tab) tab.saved = text;
    activateTab(activeTab ?? path);
    await toast(`saved ${path}`);
  } catch (err) {
    await toast(String(err), true);
  }
}

function fileName(path: string): string {
  const i = path.lastIndexOf("/");
  return i < 0 ? path : path.slice(i + 1);
}

// --- toolbar -------------------------------------------------------------------

levelSelect.addEventListener("change", () => {
  preview.navigateTo(levelSelect.value);
});

/** syncLevelSelect reflects the current target in the level dropdown. */
function syncLevelSelect(target: string): void {
  levelSelect.value = target;
}

async function rebuildLevelSelect(): Promise<void> {
  const path = editor.currentPath();
  if (!path) return;

  let symbols: any[] = [];
  try {
    symbols = (await call<any[]>("symbols", { path })) ?? [];
  } catch {
    return;
  }

  // Keep the current value; rebuild options from the outline tree. Symbol
  // names nest (children), so the full dotted unit path is the name chain.
  const current = preview.getState().target;
  levelSelect.innerHTML = "";

  const c1 = document.createElement("option");
  c1.value = "";
  c1.textContent = "C1 — context";
  levelSelect.appendChild(c1);

  const walk = (syms: any[], prefix: string) => {
    for (const s of syms ?? []) {
      const fullPath = prefix ? `${prefix}.${s.name}` : s.name;
      const opt = document.createElement("option");
      opt.value = fullPath;
      opt.textContent = `${fullPath}${s.detail ? ` — ${s.detail}` : ""}`;
      levelSelect.appendChild(opt);
      walk(s.children ?? [], fullPath);
    }
  };
  walk(symbols, "");

  levelSelect.value = current;
}

appEl.querySelector<HTMLButtonElement>("#btn-zoom-in")!.addEventListener("click", () => preview.zoomIn());
appEl.querySelector<HTMLButtonElement>("#btn-zoom-out")!.addEventListener("click", () => preview.zoomOut());
appEl.querySelector<HTMLButtonElement>("#btn-zoom-reset")!.addEventListener("click", () => preview.resetView());

appEl.querySelector<HTMLInputElement>("#chk-expanded-all")!.addEventListener("change", (ev) => {
  preview.setAllExpanded((ev.target as HTMLInputElement).checked);
});

const legendChk = appEl.querySelector<HTMLInputElement>("#chk-legend")!;
legendChk.checked = true;
legendChk.addEventListener("change", () => {
  preview.setLegend(legendChk.checked);
});

appEl.querySelector<HTMLInputElement>("#chk-pause")!.addEventListener("change", (ev) => {
  preview.setPaused((ev.target as HTMLInputElement).checked);
});

appEl.querySelector<HTMLButtonElement>("#btn-format")!.addEventListener("click", () => void formatActive());
appEl.querySelector<HTMLButtonElement>("#btn-save")!.addEventListener("click", () => {
  if (activeTab) void saveFile(activeTab, editor.currentText());
});
appEl.querySelector<HTMLButtonElement>("#btn-definition")!.addEventListener("click", () => void editor.goDefinition());

exportSelect.addEventListener("change", async () => {
  const format = exportSelect.value;
  exportSelect.value = "";
  if (!format || !activeTab) return;

  try {
    const opts = preview.getState();
    const res = await call<ExportResult>("export", {
      path: activeTab,
      opts: {
        target: opts.target,
        allExpanded: opts.allExpanded,
        expanded: opts.expanded,
        legend: opts.legend,
      } satisfies RenderOptions,
      format,
      outDir: "", // default: <project>/diagrams
    });
    await toast(`exported ${res.format}: ${res.files.join(", ")} (in ${projectDir}/diagrams)`);
  } catch (err) {
    await toast(String(err), true);
  }
});

async function formatActive(): Promise<void> {
  if (!activeTab) return;

  try {
    const res = await call<{ text: string }>("format", { path: activeTab });
    if (res?.text) {
      editor.view.dispatch({
        changes: { from: 0, to: editor.view.state.doc.length, insert: res.text },
      });
      await toast("formatted");
    } else {
      await toast("already formatted");
    }
  } catch (err) {
    await toast(String(err), true);
  }
}

/** renderExpandedChips shows the manual expanded-set override with reset. */
function renderExpandedChips(): void {
  expandedChips.innerHTML = "";

  const expanded = preview.getState().expanded;
  if (expanded === undefined) return; // model default in effect

  const label = document.createElement("span");
  label.className = "chips-label";
  label.textContent = "expanded override:";
  expandedChips.appendChild(label);

  if (!expanded.length) {
    const none = document.createElement("span");
    none.className = "chip chip-empty";
    none.textContent = "(collapsed all)";
    expandedChips.appendChild(none);
  }

  for (const unit of expanded) {
    const chip = document.createElement("button");
    chip.className = "chip";
    chip.textContent = `${unit} ×`;
    chip.title = `Remove ${unit} from the expanded set`;
    chip.addEventListener("click", () => {
      preview.setExpanded(expanded.filter((u) => u !== unit));
    });
    expandedChips.appendChild(chip);
  }

  const reset = document.createElement("button");
  reset.className = "chip chip-reset";
  reset.textContent = "model default";
  reset.title = "Discard the override and use the model's [properties].expanded";
  reset.addEventListener("click", () => preview.setExpanded(undefined));
  expandedChips.appendChild(reset);
}

// Alt-click a diagram node: toggle its in-place C1 expansion (manual override).
appEl.querySelector<HTMLElement>("#preview-container")!.addEventListener("click", (ev) => {
  if (!ev.altKey) return;
  const anchor = (ev.target as Element | null)?.closest("a");
  const href = anchor?.getAttribute("href") ?? "";
  if (!href.endsWith(".svg")) return;

  ev.preventDefault();

  void call<{ target: string }>("resolveDrill", {
    path: activeTab ?? "",
    target: preview.getState().target,
    href,
  }).then((res) => {
    const unit = res?.target;
    if (!unit) return;

    const cur = preview.getState().expanded ?? [];
    const next = cur.includes(unit) ? cur.filter((u) => u !== unit) : [...cur, unit];
    preview.setExpanded(next);
  });
});

// --- toast ----------------------------------------------------------------------

let toastTimer: number | null = null;

async function toast(text: string, isError = false): Promise<void> {
  toastEl.textContent = text;
  toastEl.classList.toggle("error", isError);
  toastEl.hidden = false;
  if (toastTimer !== null) window.clearTimeout(toastTimer);
  toastTimer = window.setTimeout(() => {
    toastEl.hidden = true;
  }, 4000);
}

// --- project controls --------------------------------------------------------------

appEl.querySelector<HTMLButtonElement>("#btn-open")!.addEventListener("click", () => void openProject(projectInput.value));

projectInput.addEventListener("keydown", (ev) => {
  if (ev.key === "Enter") void openProject(projectInput.value);
});

const browseBtn = appEl.querySelector<HTMLButtonElement>("#btn-browse")!;
if (isDesktop() && wailsRuntime()?.OpenDirectoryDialog) {
  browseBtn.hidden = false;
  browseBtn.addEventListener("click", async () => {
    const dir = await wailsRuntime()!.OpenDirectoryDialog!({ title: "Open project directory" });
    if (dir) await openProject(dir);
  });
}

// Go-to-definition requests from the editor open the target file.
window.addEventListener("c4drill:open", (ev) => {
  const path = (ev as CustomEvent).detail?.path;
  if (typeof path === "string" && path) void openFile(path);
});

// Also rebuild the level selector whenever the outline may have changed.
backend.on("diagnostics", () => void rebuildLevelSelect());

// --- boot -----------------------------------------------------------------------

backend.startEvents();

void (async () => {
  try {
    const info = await call<AppInfo>("appInfo");
    if (info?.initialDir) {
      await openProject(info.initialDir);
      // Auto-open the first model file so all three areas light up at once.
      const files = await call<ProjectInfo>("listFiles");
      const first = files.files.find((f) => f.path.endsWith(".toml") || f.path.endsWith(".c4d"));
      if (first) await openFile(first.path);
    }
  } catch {
    // Headless/dev without a project: the UI stays on the empty state.
  }
})();
