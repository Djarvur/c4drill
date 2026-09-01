// rpc.ts is the frontend's single channel to the Go backend. The same
// Dispatch(method, params) JSON surface is served by both transports —
// the Wails binding in the desktop app and POST /api/dispatch in the
// browser/dev fallback — so no UI code knows which one is running.

export type DispatchFn = (method: string, params?: unknown) => Promise<any>;
export type EventListener = (payload: any) => void;

interface WailsWindow {
  go?: {
    main?: { App?: { Dispatch?: (method: string, params: string) => Promise<string> } };
    backend?: { App?: { Dispatch?: (method: string, params: string) => Promise<string> } };
  };
  runtime?: {
    EventsOn?: (event: string, cb: (...data: any[]) => void) => void;
    EventsOff?: (event: string) => void;
    OpenDirectoryDialog?: (options: { title?: string; defaultDirectory?: string }) => Promise<string | "">;
  };
}

function wailsBinding(): ((method: string, params: string) => Promise<string>) | null {
  const w = window as unknown as WailsWindow;
  return w.go?.main?.App?.Dispatch ?? w.go?.backend?.App?.Dispatch ?? null;
}

export function wailsRuntime(): WailsWindow["runtime"] {
  return (window as unknown as WailsWindow).runtime;
}

export function isDesktop(): boolean {
  return wailsBinding() !== null;
}

class Backend {
  private listeners = new Map<string, Set<EventListener>>();

  readonly dispatch: DispatchFn = async (method, params) => {
    const binding = wailsBinding();
    if (binding) {
      const raw = await binding(method, JSON.stringify(params ?? {}));
      return raw === "" || raw === "null" ? null : JSON.parse(raw);
    }

    const resp = await fetch("/api/dispatch", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ method, params: params ?? {} }),
    });

    const body = await resp.json().catch(() => ({ error: resp.statusText }));
    if (!resp.ok || body.error) {
      throw new Error(typeof body.error === "string" ? body.error : `request failed: ${resp.status}`);
    }

    return body.result;
  };

  /** Subscribe to a backend event; returns an unsubscribe function. */
  on(event: string, cb: EventListener): () => void {
    let set = this.listeners.get(event);
    if (!set) {
      set = new Set();
      this.listeners.set(event, set);
    }
    set.add(cb);

    return () => set!.delete(cb);
  }

  private emit(event: string, payload: any): void {
    this.listeners.get(event)?.forEach((cb) => cb(payload));
  }

  /** Start listening for backend events on whichever transport is live. */
  startEvents(): void {
    const rt = wailsRuntime();
    if (rt?.EventsOn) {
      rt.EventsOn("backend", (payload: any) => this.emit(payload?.event ?? "", payload?.payload));
      return;
    }

    // Browser fallback: Server-Sent Events. Each SSE event name is the
    // backend event; the payload arrives as JSON in data.
    const source = new EventSource("/api/events");
    source.onmessage = (msg) => {
      try {
        const parsed = JSON.parse(msg.data);
        this.emit(parsed.event, parsed.payload);
      } catch {
        // ignore malformed frames
      }
    };
    source.onerror = () => {
      // EventSource reconnects on its own; nothing to do.
    };
  }
}

export const backend = new Backend();

/** Typed helper: dispatch and cast. */
export function call<T>(method: string, params?: unknown): Promise<T> {
  return backend.dispatch(method, params) as Promise<T>;
}
