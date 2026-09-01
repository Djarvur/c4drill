// preview.ts is the centerpiece: the live diagram panel. It owns the
// debounced auto-render loop, drill-down click handling, breadcrumbs, the
// error state panel (CLI-identical messages, no stale diagram), and
// zoom/pan.

import type { Breadcrumb, Diag, RenderOptions, RenderResult } from "./types";
import { call } from "./rpc";

export interface PreviewState {
  target: string;
  allExpanded: boolean;
  expanded?: string[]; // undefined = model default
  legend: boolean | null; // null = model default
}

export interface PreviewHooks {
  onTargetChange(target: string, breadcrumbs: Breadcrumb[]): void;
  onDiagnostics(diags: Diag[]): void;
}

export class Preview {
  private container: HTMLElement;

  private stage: HTMLElement;

  private breadcrumbsEl: HTMLElement;

  private errorEl: HTMLElement;

  private statusEl: HTMLElement;

  private hooks: PreviewHooks;

  private state: PreviewState = { target: "", allExpanded: false, legend: null };

  private path = "";

  private renderTimer: number | null = null;

  private paused = false;

  private scale = 1;

  private panX = 0;

  private panY = 0;

  private dragging: { x: number; y: number } | null = null;

  private renderSeq = 0;

  constructor(
    container: HTMLElement,
    breadcrumbsEl: HTMLElement,
    errorEl: HTMLElement,
    statusEl: HTMLElement,
    hooks: PreviewHooks,
  ) {
    this.container = container;
    this.stage = container.querySelector<HTMLElement>(".preview-stage")!;
    this.breadcrumbsEl = breadcrumbsEl;
    this.errorEl = errorEl;
    this.statusEl = statusEl;
    this.hooks = hooks;

    this.installPointerHandlers();
  }

  /** setFile switches the rendered document (resets navigation). */
  setFile(path: string): void {
    this.path = path;
    this.state = { target: "", allExpanded: false, legend: this.state.legend };
    this.invalidate();
  }

  setFileKeepView(path: string): void {
    this.path = path;
  }

  getState(): PreviewState {
    return { ...this.state };
  }

  setExpanded(expanded?: string[]): void {
    this.state.expanded = expanded;
    this.scheduleRender(0);
  }

  setAllExpanded(all: boolean): void {
    this.state.allExpanded = all;
    this.scheduleRender(0);
  }

  setLegend(legend: boolean | null): void {
    this.state.legend = legend;
    this.scheduleRender(0);
  }

  setPaused(paused: boolean): void {
    this.paused = paused;
    this.updateStatus(paused ? "auto-render paused" : "");
  }

  isPaused(): boolean {
    return this.paused;
  }

  /** invalidate schedules the debounced auto re-render (~200ms). */
  invalidate(): void {
    this.scheduleRender(200);
  }

  /** navigateTo renders a specific target immediately (breadcrumb clicks). */
  navigateTo(target: string): void {
    this.state.target = target;
    this.scheduleRender(0);
  }

  private scheduleRender(delayMs: number): void {
    if (this.renderTimer !== null) {
      window.clearTimeout(this.renderTimer);
    }

    if (this.paused) {
      this.updateStatus("auto-render paused — dirty");
      return;
    }

    this.renderTimer = window.setTimeout(() => {
      this.renderTimer = null;
      void this.renderNow();
    }, delayMs);
  }

  async renderNow(): Promise<void> {
    if (!this.path) return;

    const seq = ++this.renderSeq;

    let res: RenderResult;
    try {
      res = await call<RenderResult>("render", {
        path: this.path,
        opts: {
          target: this.state.target,
          allExpanded: this.state.allExpanded,
          expanded: this.state.expanded,
          legend: this.state.legend,
        } satisfies RenderOptions,
      });
    } catch (err) {
      if (seq === this.renderSeq) this.showError([{ message: String(err) }]);
      return;
    }

    if (seq !== this.renderSeq) return; // a newer render superseded us

    this.hooks.onDiagnostics(res.diagnostics ?? []);

    if (!res.svg) {
      this.showError(res.diagnostics ?? []);
      return;
    }

    this.showSVG(res.svg);
    this.renderBreadcrumbs(res.breadcrumbs ?? []);
    this.state.target = res.target;
    this.hooks.onTargetChange(res.target, res.breadcrumbs ?? []);
    this.updateStatus("");
  }

  private showSVG(svg: string): void {
    this.errorEl.hidden = true;
    this.errorEl.innerHTML = "";
    this.stage.innerHTML = svg;

    const svgEl = this.stage.querySelector("svg");
    if (svgEl) {
      svgEl.removeAttribute("width");
      svgEl.removeAttribute("height");
      svgEl.style.maxWidth = "none";
      // Clickable navigation must stay in-app, never navigate the webview.
      svgEl.querySelectorAll("a").forEach((a) => a.setAttribute("target", "_self"));
    }

    this.resetView();
  }

  private showError(diags: Diag[]): void {
    this.stage.innerHTML = "";
    this.errorEl.hidden = false;
    this.errorEl.innerHTML = "";

    const title = document.createElement("div");
    title.className = "preview-error-title";
    title.textContent = "Cannot render — fix the errors below (same messages the CLI reports):";
    this.errorEl.appendChild(title);

    for (const d of diags) {
      const line = document.createElement("div");
      line.className = "preview-error-line";
      line.textContent = d.message;
      this.errorEl.appendChild(line);
    }
  }

  private renderBreadcrumbs(crumbs: Breadcrumb[]): void {
    this.breadcrumbsEl.innerHTML = "";

    crumbs.forEach((crumb, i) => {
      if (i > 0) {
        const sep = document.createElement("span");
        sep.className = "crumb-sep";
        sep.textContent = "›";
        this.breadcrumbsEl.appendChild(sep);
      }

      const el = document.createElement(i < crumbs.length - 1 ? "button" : "span");
      el.className = i < crumbs.length - 1 ? "crumb crumb-link" : "crumb crumb-current";
      el.textContent = crumb.name;
      if (i < crumbs.length - 1) {
        el.addEventListener("click", () => this.navigateTo(crumb.target));
      }

      this.breadcrumbsEl.appendChild(el);
    });
  }

  private updateStatus(text: string): void {
    this.statusEl.textContent = text;
  }

  /** installPointerHandlers wires drill-down clicks and zoom/pan gestures. */
  private installPointerHandlers(): void {
    // Drill-down: any internal *.svg anchor resolves through the backend's
    // layout algebra; external links open normally.
    this.container.addEventListener("click", (ev) => {
      const anchor = (ev.target as Element | null)?.closest("a");
      if (!anchor) return;

      const href = anchor.getAttribute("href") ?? anchor.getAttribute("xlink:href") ?? "";
      if (!href) return;

      if (/^[a-z]+:\/\//i.test(href) || href.startsWith("mailto:")) {
        return; // external reference: let the runtime decide
      }

      ev.preventDefault();

      void call<{ target: string }>("resolveDrill", {
        path: this.path,
        target: this.state.target,
        href,
      })
        .then((res) => {
          if (res?.target !== undefined) this.navigateTo(res.target);
        })
        .catch(() => undefined);
    });

    // Zoom (wheel) and pan (drag).
    this.container.addEventListener(
      "wheel",
      (ev) => {
        ev.preventDefault();
        const factor = ev.deltaY < 0 ? 1.1 : 1 / 1.1;
        this.scale = Math.min(8, Math.max(0.1, this.scale * factor));
        this.applyTransform();
      },
      { passive: false },
    );

    this.container.addEventListener("mousedown", (ev) => {
      this.dragging = { x: ev.clientX - this.panX, y: ev.clientY - this.panY };
    });

    window.addEventListener("mousemove", (ev) => {
      if (!this.dragging) return;
      this.panX = ev.clientX - this.dragging.x;
      this.panY = ev.clientY - this.dragging.y;
      this.applyTransform();
    });

    window.addEventListener("mouseup", () => {
      this.dragging = null;
    });
  }

  zoomIn(): void {
    this.scale = Math.min(8, this.scale * 1.25);
    this.applyTransform();
  }

  zoomOut(): void {
    this.scale = Math.max(0.1, this.scale / 1.25);
    this.applyTransform();
  }

  resetView(): void {
    this.scale = 1;
    this.panX = 0;
    this.panY = 0;
    this.applyTransform();
  }

  private applyTransform(): void {
    this.stage.style.transform = `translate(${this.panX}px, ${this.panY}px) scale(${this.scale})`;
    this.stage.style.transformOrigin = "center top";
  }
}
