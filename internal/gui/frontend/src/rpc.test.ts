// rpc.test.ts — binding-shape tests for the desktop transport resolver
// (issue #38 / GUI-01): the resolver must find the Wails-generated
// namespace window.go.main.desktop.Dispatch and route the whole
// Dispatch(method, params) JSON protocol through it, and must fail
// closed (null transport, isDesktop() false) when the binding is
// absent — the browser/--serve HTTP fallback stays the only other path.

import { afterEach, describe, expect, it, vi } from "vitest";
import { backend, isDesktop } from "./rpc";

/** DesktopDispatch mirrors the Wails-generated binding signature. */
type DesktopDispatch = (method: string, params: string) => Promise<string>;

/** stubGoWindow installs a fake `window` carrying the given go namespace. */
function stubGoWindow(go: unknown): void {
  vi.stubGlobal("window", { go });
}

/** stubDesktopWindow stubs the real Wails-generated desktop binding. */
function stubDesktopWindow(dispatch: DesktopDispatch): void {
  stubGoWindow({ main: { desktop: { Dispatch: dispatch } } });
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("rpc transport resolution", () => {
  it("resolves window.go.main.desktop.Dispatch, reports desktop mode, and dispatches through it", async () => {
    const dispatch = vi.fn<DesktopDispatch>().mockResolvedValue(JSON.stringify({ ok: true }));
    stubDesktopWindow(dispatch);

    expect(isDesktop()).toBe(true);

    const result = await backend.dispatch("someMethod", { a: 1 });

    expect(result).toEqual({ ok: true });
    expect(dispatch).toHaveBeenCalledTimes(1);
    expect(dispatch).toHaveBeenCalledWith("someMethod", JSON.stringify({ a: 1 }));
  });

  it("fails closed when the desktop binding is absent — no transport, not desktop mode", () => {
    stubGoWindow({});

    expect(isDesktop()).toBe(false);
  });
});

describe("rpc result normalization", () => {
  it("normalizes raw null and empty-string binding results to null", async () => {
    for (const raw of ["null", ""]) {
      const dispatch = vi.fn<DesktopDispatch>().mockResolvedValue(raw);
      stubDesktopWindow(dispatch);

      const result = await backend.dispatch("someMethod", { a: 1 });

      expect(result).toBeNull();
    }
  });
});
