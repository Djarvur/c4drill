# Phase 42 — Pattern Map

**Phase:** 42 - Desktop GUI Binding Fix
**Generated:** 2026-09-03
**Sources:** `42-CONTEXT.md`, `42-RESEARCH.md`

Files to create/modify extracted from CONTEXT.md + RESEARCH.md, each mapped to its closest existing analog with concrete excerpts. The executor should mirror these patterns exactly.

---

## File Inventory

| # | File | Action | Role | Data Flow |
|---|------|--------|------|-----------|
| 1 | `internal/gui/frontend/src/rpc.ts` | modify | Frontend transport resolver (client tier) | UI modules → `Backend.dispatch` → resolver picks Wails binding or HTTP fetch → Go `Dispatch` |
| 2 | `internal/gui/frontend/src/rpc.test.ts` | create | Frontend unit test (vitest) | Stubs `window`, asserts resolver contract (D-03) |

No other files are created or modified. The Go side (`cmd/c4drill-gui/main.go`) is read-only context (source of truth for the namespace).

---

## File 1: `internal/gui/frontend/src/rpc.ts` (modify)

**Closest analog:** itself — this is the only transport file in the frontend (roadmap-verified: sole reference to the phantom namespaces). The change is confined to the `WailsWindow.go` declaration (lines 9-19) and the `wailsBinding()` resolver (lines 21-24). Everything below those lines (`isDesktop`, `Backend`, `call`) is untouched.

**Current state (excerpt — the exact lines being replaced):**

```typescript
interface WailsWindow {
  go?: {
    main?: { App?: { Dispatch?: (method: string, params: string) => Promise<string> } };
    backend?: { App?: { Dispatch?: (method: string, params: string) => Promise<string> } };
  };
  runtime?: { /* EventsOn / EventsOff / OpenDirectoryDialog — unchanged, already correct */ };
}

function wailsBinding(): ((method: string, params: string) => Promise<string>) | null {
  const w = window as unknown as WailsWindow;
  return w.go?.main?.App?.Dispatch ?? w.go?.backend?.App?.Dispatch ?? null;
}
```

**Target shape (per D-01/D-02 — mirrors the optional-chaining + null-fallback pattern already used):**

```typescript
interface WailsWindow {
  go?: {
    main?: { desktop?: { Dispatch?: (method: string, params: string) => Promise<string> } };
  };
  runtime?: { /* unchanged */ };
}

function wailsBinding(): ((method: string, params: string) => Promise<string>) | null {
  const w = window as unknown as WailsWindow;
  return w.go?.main?.desktop?.Dispatch ?? null;
}
```

**Read-only source of truth (Go namespace, do not modify):** `cmd/c4drill-gui/main.go` — `package main` (L15), `type desktop struct {` (L74), `func (g *desktop) Dispatch(method, params string) (string, error)` (L86), `Bind: []interface{}{g}` (L124).

---

## File 2: `internal/gui/frontend/src/rpc.test.ts` (create)

**Closest analog:** `internal/gui/frontend/src/language/c4d.test.ts` — the project's only frontend test file. Its conventions to replicate:

**Analog excerpt — header comment + imports (c4d.test.ts:1-10):**

```typescript
// c4d.test.ts — token-level tests for the C4D Lezer grammar (issue #36):
// the hard cases from c4d.peg (header forms, external, arrows, option
// blocks, `;` one-liners, triple-quoted strings, ${param}, comments) plus
// whole-file parses of the .c4d example fixtures, highlighting spans, and
// the folding/indentation node props wired into the editor.

import { describe, expect, it } from "vitest";
```

**Analog excerpt — domain-grouped describe/it naming with behavior-focused assertions (c4d.test.ts:64-81):**

```typescript
describe("c4d grammar — unit header forms", () => {
  it("parses the id-led header with type, external and display name", () => {
    ...
    expect(childNames(header)).toEqual(["Ident", "Ident", "external", "String"]);
  });
```

**Patterns to replicate in rpc.test.ts:**
- Leading file-header comment stating purpose + issue reference (`// rpc.test.ts — binding-shape tests for the desktop transport resolver (issue #38 / GUI-01): ...`).
- Named imports from `vitest` (`describe, expect, it` plus `vi, afterEach` for stubbing — an addition, but same style).
- Domain-grouped `describe` blocks with full-sentence `it` names, e.g. `describe("rpc transport resolution")` / `it("resolves window.go.main.desktop.Dispatch and dispatches through it")`.
- Small typed helpers above the tests (analog's `parse`/`countNodes` pattern), e.g. a `stubDesktopWindow(dispatch)` helper returning the stubbed `window`.

**Harness facts (from RESEARCH.md Validation Architecture):**
- Runner: `vitest ^4.1.11`; command `cd internal/gui/frontend && npx vitest run src/rpc.test.ts`; `npm test` = `vitest run`.
- No vitest.config.ts — default `node` environment; stub `window` with `vi.stubGlobal("window", {...})` and restore via `vi.unstubAllGlobals()` in `afterEach`.
- The unit under test is the module's exported surface: `isDesktop()` and `backend.dispatch` / `call<T>` (rpc.ts exports `backend` and `call`). The mock binding: `vi.fn().mockResolvedValue(JSON.stringify({...}))` placed at `window.go.main.desktop.Dispatch`, asserting `Dispatch` receives `(method, JSON.stringify(params))` and that `"null"`/`""` raw results normalize to `null` (rpc.ts:41 normalization is existing behavior worth pinning while the file is open — optional, executor judgment).
- TDD ordering (D-03): this test is the RED test — write and run it against the unmodified rpc.ts first (resolver never finds `main.desktop`, so the resolve test must fail), then apply the rpc.ts edit and watch it go GREEN.

---

## Anti-Pattern Reminders (from RESEARCH.md)

- No dead fallbacks: do not keep `main?.App` / `backend?.App` anywhere (D-02).
- Do not touch the fetch fallback, `startEvents()`, `style.css`, `index.html`, or DOM-building code (GUI-02 + UI-SPEC zero-delta budget).
- Do not import generated `wailsjs/go/...` modules; probe the runtime `window.go` object (existing approach).
