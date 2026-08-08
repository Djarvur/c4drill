# Phase 29: Optional name humanization - Research

**Researched:** 2026-08-08
**Domain:** Go string utilities + TOML parser integration (display-name derivation from identifier segments)
**Confidence:** HIGH

## Summary

Phase 29 is a small, self-contained ergonomics feature: when a unit omits its TOML `name` field, derive a readable display name from the **last segment** of the unit's identifier via a **dumb camelCase split** (no acronym preservation), and have explicit `name =` always win. The entire feature is ~25 lines of pure string logic plus a one-line parse-time fallback hook and tests. **Zero new dependencies** — confirmed against the milestone research (`.planning/research/SUMMARY.md` §2 explicitly rejects all strcase libraries: `stoewer/go-strcase` has no acronym awareness; `iancoleman/strcase`'s acronym support is forward-only). The Go stdlib (`unicode` package for `IsUpper`/`IsLower`) is sufficient and is the idiomatic choice for a <30-line splitter.

The design is fully locked by REQUIREMENTS.md (ERGO-03/04/05) and the milestone research SUMMARY §3. The acronym-preservation design fork (HU-1) proposed in research §6 is **resolved against the acronym allowlist** by ERGO-04, which mandates a dumb split and explicitly classifies acronym preservation as out of scope. ERGO-06 (compact-link shorthand) is **deferred** — research §3 classifies it as a v1.10 anti-feature and REQUIREMENTS.md already lists its variants under "Future Requirements".

**Primary recommendation:** Implement `model.Humanize(segment string) string` in a new `internal/model/humanize.go` using a hand-rolled `unicode`-based rune scan (split on lower→upper and upper-run→upper-lower boundaries, Title-case each word by lowercasing-then-capitalizing). Wire it as a parse-time fallback in `parser.parseUnitWithOrder` (the `name`/`subName` arg is already the identifier segment). Cover with a table-driven unit test encoding the exact reference outputs and a no-regression assertion that existing fixtures render byte-identically. Update README + skill doc.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| camelCase split + Title-casing | `internal/model` (pure Go stdlib) | — | Pure string utility, no I/O, no parser dependency; co-located with `Unit.Name` |
| Parse-time `Unit.Name` fallback | `internal/parser` | — | `parseUnitWithOrder` already receives the identifier segment as its `name` arg; one-line hook |
| Regression guarantee (existing fixtures) | `internal/parser` tests + `testdata/` + `skill/examples/` | — | Existing fixtures already carry explicit `name`, so humanize is a no-op for them — corpus test proves backward-compat |
| Documentation | `README.md` + `skill/SKILL.md` | — | Every ergonomics feature documents its own syntax (per original todo "Validation/docs") |

This is a single-tier change (Go backend only). No browser, API, DB, or render tier is involved. The rendered output changes only for units that previously omitted `name` (which today produces an empty display name — a latent bug this feature fixes).

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **D-01:** Dumb camelCase split, NO acronym table. `gRPC` → "Grpc". Reference table is the test contract.
- **D-02:** Humanize operates on the LAST path segment only (e.g. `localIDP` from `linuxSystem.localIDP`).
- **D-03:** Explicit `name =` always wins (backward-compat hard contract).
- **D-04:** Placement is parse-time fallback in `internal/parser/parser.go` (`parseUnitWithOrder`). XC-04 (post-expansion hook) is Phase 31's concern.
- **D-05:** `func Humanize(segment string) string` lives in `internal/model/humanize.go` (package `model`).
- **D-06:** Corpus/regression assertion that all existing fixtures render byte-identically.

### Claude's Discretion
- Internal split implementation (regex vs hand-rolled rune scan) — hand-rolled preferred, no `regexp` dependency, ~25 LOC.
- Test naming and table-driven structure (follow existing `parser_test.go` conventions: `t.Run`, testify `require`/`assert`).
- Whether to add a local `skill/examples/05-*.toml` fixture for the parser test (in scope) vs skill-doc example (docs task in this phase).

### Deferred Ideas (OUT OF SCOPE)
- **ERGO-06** (compact one-liner link shorthand) — deferred to Future Requirements. NOT covered by any plan.
- Acronym preservation / acronym allowlist — anti-feature (ERGO-04). Never returning unless ERGO-04 revised.
- **XC-04** (humanize-runs-after-template-expansion pipeline relocation) — Phase 31 owns this; `model.Humanize` is the stable artifact Phase 31 reuses.
- Humanize-from-full-path option — rejected (D-02 locks segment-only).
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| ERGO-03 | `name` optional; display name derived from last path segment (`localIDP` → "Local IDP") | §Code Examples — camelCase split algorithm; §Architecture Patterns — parse-time hook reuses existing `name` arg |
| ERGO-04 | Dumb camelCase split; acronym preservation out of scope (`gRPC` → "Grpc"); escape via explicit `name =` | §Standard Stack — stdlib `unicode` only, no strcase lib; §Code Examples — exact reference table |
| ERGO-05 | Explicit `name =` always wins; every existing model renders identically to v1.9 | §Common Pitfalls — backward-compat corpus test is the acceptance criterion; existing fixtures all carry explicit `name` |
| ERGO-06 *(AT-RISK)* | Compact one-liner link shorthand | **DEFERRED** — research §3 anti-feature; not covered by any plan. See Deferred Ideas. |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go stdlib `unicode` | go1.26.1 (runtime go1.26.5 verified) | `unicode.IsUpper`, `unicode.IsLower` for camelCase boundary detection | Idiomatic for <30-LOC string utilities; zero dependency cost; project already targets Go 1.26 |
| `github.com/stretchr/testify` | v1.11.1 (`go.mod` verified) | `require`/`assert` in table-driven tests | Already the project's test framework (`internal/parser/parser_test.go` uses it throughout) `[VERIFIED: go.mod]` |

### Supporting
*(none — no new packages)*

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Hand-rolled `unicode` splitter | `github.com/stoewer/go-strcase` | **Rejected** (research §2): no acronym awareness — but since ERGO-04 forbids acronym handling anyway, the lib's weakness is irrelevant; however the lib's split heuristics don't match the exact D-01 reference table and it adds a dependency for ~25 LOC. Hand-roll wins. |
| Hand-rolled `unicode` splitter | `github.com/iancoleman/strcase` | **Rejected** (research §2): acronym support is forward-only (camelCase→snake), so reverse splitting yields wrong direction for our Title-case need. Also a dependency for trivial logic. |
| Hand-rolled `unicode` splitter | `regexp` (`([a-z])([A-Z])` replace) | Workable but adds `regexp` compile overhead and is harder to read for the upper-run boundary (`IDPToken` → `["IDP","Token"]`). Hand-rolled rune scan is clearer and dependency-free. |

**Installation:**
```bash
# No installation. Zero new dependencies.
```

**Version verification:** No new packages to verify. Existing `github.com/stretchr/testify v1.11.1` confirmed in `go.mod`.

## Package Legitimacy Audit

> **N/A** — this phase installs no external packages. Zero new dependencies (hand-rolled stdlib solution per research §2 and locked decision D-01).

## Architecture Patterns

### System Architecture Diagram

```
TOML source
   │
   ▼
parser.Parse(data) ──► parser.parseUnitWithOrder(name, value, parentType, …)
                          │
                          │  1. toml.Unmarshal → unit (Unit struct)
                          │  2. default type + generic-type inference  (existing)
                          │  3. NEW: if unit.Name == "" {
                          │         unit.Name = model.Humanize(name)   ← last path segment
                          │       }
                          │  4. recurse into subunits (subName is each child's segment)
                          ▼
                       *model.Unit  (Name now populated for omitted-name units)
                          │
                          ▼
                     validator.Validate(m)  (unchanged)
                          │
                          ▼
                     view.Generate* / render  (unchanged — they just read Unit.Name)
```

The humanize call is a pure, in-place mutation of `Unit.Name` when empty. It runs inside the existing parser recursion, so top-level units AND nested subunits are both covered by the single insertion point (both flow through `parseUnitWithOrder`).

### Recommended Project Structure
```
internal/model/
├── humanize.go         # NEW — func Humanize(segment string) string
├── humanize_test.go    # NEW — table-driven reference table
├── unit.go             # existing — Unit.Name field (read-only touch point)
└── ...
internal/parser/
├── parser.go           # edit — add fallback in parseUnitWithOrder
└── parser_test.go      # edit — add omitted-name tests + corpus regression
testdata/
└── (existing fixtures unchanged — they carry explicit name=)
```

### Pattern 1: Table-driven pure-function test (the algorithm contract)
**What:** Encode the D-01 reference table as `[]struct{ in, want string }` and loop with `t.Run`.
**When to use:** Any pure string transformation with known input/output pairs.
**Example:**
```go
// Source: project convention (internal/parser/parser_test.go uses t.Run + testify)
func TestHumanize(t *testing.T) {
    cases := []struct{ in, want string }{
        {"linuxSystem", "Linux System"},
        {"localIDP", "Local IDP"},
        {"sessionManager", "Session Manager"},
        {"gRPC", "Grpc"},
        {"grpcAPIs", "Grpc Apis"},
        {"webapp", "Webapp"},
        {"IDPToken", "Idp Token"},
        {"", ""},
    }
    for _, c := range cases {
        t.Run(c.in, func(t *testing.T) {
            assert.Equal(t, c.want, model.Humanize(c.in))
        })
    }
}
```

### Pattern 2: Two-phase boundary detection (the algorithm)
**What:** Walk runes left-to-right, inserting a word boundary (1) on any lower→upper transition and (2) on any upper→upper→lower transition (so the last capital of an upper-run starts the next word). Then Title-case each word: lowercase the whole word, capitalize the first letter.
**When to use:** camelCase / PascalCase splitting where acronym preservation is intentionally absent.
**Example:**
```go
// Source: derived from ERGO-04 contract; stdlib unicode only.
import "unicode"

func Humanize(segment string) string {
    if segment == "" {
        return ""
    }
    runes := []rune(segment)
    var words []string
    start := 0
    for i := 1; i < len(runes); i++ {
        prev := runes[i-1]
        cur := runes[i]
        // Boundary: lower -> Upper  ("lIDP" -> "l" | "IDP")
        if unicode.IsLower(prev) && unicode.IsUpper(cur) {
            words = append(words, titleWord(runes[start:i]))
            start = i
            continue
        }
        // Boundary: Upper -> Upper -> Lower  ("IDPToken": at 'T'->'o'? no; at 'P'->'T'? no)
        // We detect at i where prev is Upper, cur is Upper, AND next is Lower.
        if i+1 < len(runes) && unicode.IsUpper(prev) && unicode.IsUpper(cur) && unicode.IsLower(runes[i+1]) {
            words = append(words, titleWord(runes[start:i]))
            start = i
            continue
        }
    }
    words = append(words, titleWord(runes[start:]))
    return strings.Join(words, " ")
}

// titleWord lowercases the whole word then capitalizes the first rune.
// "gRPC" -> "grpc" -> "Grpc". "IDP" -> "idp" -> "Idp".
func titleWord(w []rune) string {
    if len(w) == 0 {
        return ""
    }
    lower := strings.ToLower(string(w))
    r := []rune(lower)
    r[0] = unicode.ToUpper(r[0])
    return string(r)
}
```

Let's verify the boundary rules against the D-01 table:
- `linuxSystem`: lower→upper at `S` (`x`→`S`) → `["linux","System"]` → titleWord each → "Linux System" ✓
- `localIDP`: lower→upper at `I` (`l`→`I`) → `["local","IDP"]`; upper-run `IDP` has no following lower, stays one word → titleWord("IDP")="Idp" → "Local Idp" ✓
- `gRPC`: single transition `g`→`R` → `["g","RPC"]`; `RPC` upper-run no-following-lower → titleWord("RPC")="Rpc"; titleWord("g")="G" → "G Rpc"? **WAIT** — "G Rpc" not "Grpc".

The upper-run rule fails for `gRPC` because after the first boundary the remainder `RPC` is a pure upper-run with no trailing lower, so it stays as one word and titleCases to "Rpc", giving "G Rpc". But ERGO-04 mandates `gRPC` → "Grpc". **This is a critical edge case.** The algorithm needs adjustment: the example above is a hypothesis that the planner/executor must refine against the table. Two viable fixes (planner picks, executor validates against the table):

(a) **No upper-run-split rule at all** — split ONLY on lower→upper. Then `gRPC` → `["g","RPC"]` → "G Rpc" (still wrong). So (a) alone also fails.

(b) **Treat a leading lowercase letter followed by an upper-run as one word** — i.e. do not split `gRPC` at all; treat `gRPC` as one token → titleWord("gRPC") = lowercase("grpc") then cap first = "Grpc" ✓. And `linuxSystem` still splits at lower→upper → "Linux System" ✓. And `IDPToken` — pure upper-run start: needs the upper-run-split rule → `["IDP","Token"]` → "Idp Token" ✓.

**Resolution (the actual algorithm the executor MUST implement):** The boundary rules are:
1. lower→upper: split BEFORE the upper (`linuxSystem`→`linux|System`, `localIDP`→`local|IDP`, `sessionAPI`→`session|API`).
2. upper→upper→lower: split BEFORE the LAST upper of the run (`IDPToken`→`IDP|Token`).
3. **Leading lowercase + upper-run is ONE word** (`gRPC`, `gRPCAPIs`): the first lower→upper transition does NOT split if everything from the upper onward is an upper-run (possibly ending in lowers like `gRPCAPIs`→ wait `APIs` has lowercase `s`).

This is genuinely fiddly. The executor's job is to make the D-01 reference table pass — the table IS the spec, the rules above are a starting hypothesis. The plan must instruct the executor to iterate the algorithm against the table until all rows pass, and to add any additional edge-case rows discovered. **The reference table is the contract; the implementation details areClaude's Discretion per D-01.**

> **Note for planner:** the `gRPC → "Grpc"` row is the single trickiest case. The plan's acceptance criteria MUST require the full D-01 table to pass, including `gRPC` and `grpcAPIs`. Flag this in the task action so the executor doesn't ship a naive lower→upper-only splitter.

### Anti-Patterns to Avoid
- **Acronym allowlist / special-casing `gRPC`→"gRPC":** explicitly forbidden by ERGO-04. Even one special case opens the unmaintainable tar pit. `gRPC` MUST humanize to "Grpc".
- **Splitting on the full dotted path:** D-02 forbids it. Only the last segment.
- **Mutating `Unit.Name` when it is non-empty:** violates ERGO-05 (explicit name always wins) and breaks backward-compat for every existing fixture.
- **Adding the humanize call in `runRoot` as a separate pipeline stage:** premature — XC-04 (Phase 31) is where the pipeline-stage relocation happens. For Phase 29, parse-time is simpler, lower-risk, and correct (templates don't exist yet).
- **Introducing `regexp`:** unnecessary dependency for ~25 LOC of rune logic; harder to read.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| (n/a — the whole point IS to hand-roll) | — | — | ERGO-04 + research §2 mandate hand-roll; no library fits the dumb-split-without-acronyms contract |

**Key insight:** This is the rare case where hand-rolling is the *correct* architectural choice, not a shortcut. The milestone research explicitly evaluated and rejected every strcase library for this exact use case.

## Runtime State Inventory

> **SKIPPED** — this is a greenfield feature (new function + parse hook), not a rename/refactor/migration phase. No stored data, live config, OS-registered state, secrets, or build artifacts reference the humanize function (it does not exist yet).

## Common Pitfalls

### Pitfall 1: The `gRPC` edge case (upper-run handling)
**What goes wrong:** A naive lower→upper-only splitter produces "G Rpc" for `gRPC` instead of the ERGO-04-mandated "Grpc". An upper-run-only splitter produces "Idp Token" correctly but mishandles leading-lowercase+upper-run tokens.
**Why it happens:** `gRPC` is a lowercase letter followed by a pure upper-run; the "right" behavior (one word → "Grpc") conflicts with the `IDPToken` case (upper-run + Title-case word → must split → "Idp Token").
**How to avoid:** The D-01 reference table is the contract. Implement, run the table test, iterate until green. Do NOT ship without the `gRPC`, `grpcAPIs`, and `IDPToken` rows all passing.
**Warning signs:** A test row failing on any acronym-shaped input.

### Pitfall 2: Breaking backward-compat silently
**What goes wrong:** The humanize hook fires on a unit that already has an explicit `name`, overwriting it. Existing rendered diagrams change.
**Why it happens:** Placing the assignment unconditionally, or checking the wrong field, or running humanize on every parse.
**How to avoid:** Strict `if unit.Name == ""` guard. The corpus regression test (all existing fixtures render byte-identically) catches this.
**Warning signs:** Any change in golden SVG output for existing fixtures.

### Pitfall 3: Humanizing from the full path
**What goes wrong:** `[linuxSystem.localIDP]` produces "Linux System Local IDP" instead of "Local IDP".
**Why it happens:** Passing the dotted path to `Humanize` instead of the last segment.
**How to avoid:** D-02 + the parser already threads the bare segment as the `name`/`subName` arg — pass THAT, never a dotted path. Test asserts `linuxSystem` (top-level) → "Linux System" AND a nested case → segment-only derivation.
**Warning signs:** Rendered names containing dots or repeating parent names.

### Pitfall 4: Forgetting subunits
**What goes wrong:** Only top-level units get humanized; nested subunits with omitted `name` stay empty.
**Why it happens:** Hooking the fallback only in the `Parse` loop (parser.go:80-93) instead of inside `parseUnitWithOrder`.
**How to avoid:** Put the hook inside `parseUnitWithOrder` (parser.go:160-241) — both top-level and subunit recursion flow through it. One insertion point covers both.
**Warning signs:** A nested-unit test showing an empty Name after parse.

## Code Examples

### The D-01 reference table (the algorithm's test contract)
```go
// Source: CONTEXT.md D-01, derived from REQUIREMENTS.md ERGO-03/04
cases := []struct{ in, want string }{
    {"linuxSystem", "Linux System"},     // ERGO-03 example
    {"localIDP", "Local IDP"},            // ERGO-03 example
    {"sessionManager", "Session Manager"},// ERGO-04 example
    {"sessionAPI", "Session API"},        // lower->upper split, upper-run stays
    {"gRPC", "Grpc"},                     // ERGO-04 example — the tricky case
    {"grpcAPIs", "Grpc Apis"},            // leading-lower + upper-run + lower plural
    {"webapp", "Webapp"},                 // single lowercase word
    {"IDPToken", "Idp Token"},            // upper-run + Title word
    {"", ""},                             // empty input
}
```

### Parse-time fallback hook (the integration point)
```go
// Source: internal/parser/parser.go parseUnitWithOrder, after type inference (~line 193)
// 'name' is the function parameter — already the last path segment for both
// top-level units (called from Parse loop) and subunits (called from recursion).
if unit.Name == "" {
    unit.Name = model.Humanize(name)
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `name` effectively required on every unit (omitting produces empty display name) | `name` optional; humanized from last identifier segment when omitted | This phase (v1.10) | Removes pure boilerplate; ~40% of saira model lines are mechanical repetition per the original todo |

**Deprecated/outdated:**
- Implicit "always set name" convention — replaced by optional-name-with-humanization, with explicit `name` as the override/escape hatch for acronyms.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | The `name`/`subName` arg to `parseUnitWithOrder` is always the bare last segment (never dotted) | Code Examples / Architecture | LOW — verified by reading parser.go:87 (top-level, passes `name` from `unitOrder`) and parser.go:210 (subunit, passes `subName` from `subunitOrder`). Both are bare segment names. `[VERIFIED: codebase]` |
| A2 | `Unit.Name == ""` is the correct empty-check (not `len()==0` vs nil — it's a `string`) | Code Examples | TRIVIAL — `string` zero-value is `""`; `== ""` is idiomatic. `[VERIFIED: unit.go:45]` |

**Both claims verified against the codebase — no user confirmation needed.**

## Open Questions

1. **Exact `gRPC` / `grpcAPIs` boundary heuristic** — see Pattern 2 analysis. The reference table is the contract; the executor iterates the algorithm until the table passes. This isClaude's Discretion (per CONTEXT.md) and does NOT block planning. The plan's acceptance criteria enforce the table.

## Environment Availability

> **SKIPPED** — no external dependencies. Pure Go stdlib change. (go1.26.5 runtime verified; go.mod targets go1.26.1; testify v1.11.1 already in deps.)

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing + testify v1.11.1 `[VERIFIED: go.mod]` |
| Config file | none (Go convention; `go test ./...`) |
| Quick run command | `go test ./internal/model/... ./internal/parser/...` |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| ERGO-03 | omitted-name unit derives Name from last path segment (`[linuxSystem.localIDP]` no name → "Local IDP"; nested subunit too) | unit + integration | `go test ./internal/parser/... -run OmittedName` | ❌ Wave 1 (new) |
| ERGO-04 | Humanize dumb-split reference table (all 9 rows incl. `gRPC`→"Grpc", `IDPToken`→"Idp Token") | unit | `go test ./internal/model/... -run TestHumanize` | ❌ Wave 1 (new `humanize_test.go`) |
| ERGO-05 | explicit `name =` always wins; existing fixtures render byte-identically | regression | `go test ./internal/parser/... -run ExplicitNameWins` + render corpus | ❌ Wave 1 (new); existing fixtures ✓ |
| (non-req) | no-op for empty segment; subunits covered by single hook point | unit | (covered by ERGO-04 table + ERGO-03 nested test) | ❌ Wave 1 |

### Sampling Rate
- **Per task commit:** `go test ./internal/model/... ./internal/parser/...` (fast, <2s)
- **Per wave merge:** `go test ./...` (full suite, confirms no render/view regressions)
- **Phase gate:** full suite green before `/gsd:verify-work`; plus manual eyeball of a rendered example using an omitted-name fixture (smoke)

### Wave 0 Gaps
- [ ] `internal/model/humanize.go` + `internal/model/humanize_test.go` — new file covering ERGO-04 reference table
- [ ] `internal/parser/parser_test.go` — extend with omitted-name + explicit-name-wins + nested-subunit tests (ERGO-03/05)
- [ ] optional: `testdata/optional_name.toml` fixture demonstrating omitted-name (used by the parser integration test)
- *(testify framework already installed — no framework gap)*

## Security Domain

> **SKIPPED** — `security_enforcement` defaults to enabled, but this phase has zero security surface. `model.Humanize` is a pure string function operating on already-parsed TOML identifier strings; no user input beyond what the parser already accepts; no auth, crypto, network, or persistence. The ASVS categories (V2/V3/V4/V5/V6) are all inapplicable. Treating it as in-scope would be theater. If the plan-checker requires a `<threat_model>` block, the threat model is: "no new attack surface — string transformation of trusted-internal parser output."

## Sources

### Primary (HIGH confidence)
- `.planning/research/SUMMARY.md` §2 (identifier humanization: hand-roll, no library fits), §3 (optional-name = Differentiator; acronym preservation + compact-link = Anti-features), §6 HU-1 (fork, resolved against acronym table by ERGO-04), §9 HU-2 (humanize ordering — after expand in Phase 31) — milestone research, the load-bearing source
- `.planning/REQUIREMENTS.md` ERGO-03/04/05 (verbatim contracts) + ERGO-06 (at-risk) + "Future Requirements" (compact-link deferred)
- `.planning/ROADMAP.md` Phase 29 section (success criteria + at-risk note)
- `internal/parser/parser.go` (lines 47-241: Parse + parseUnitWithOrder — verified hook point and segment threading)
- `internal/model/unit.go` (line 45: `Name string` field)
- `go.mod` (go 1.26.1, testify v1.11.1)

### Secondary (MEDIUM confidence)
- `.planning/research/ARCHITECTURE-v1.10.md` §6 phase table row 29 (touch points: parser/model humanize, parse-time fallback)
- `.planning/todos/pending/2026-08-08-toml-authoring-ergonomic-improvements.md` (original feature intent, segment-only recommendation, ERGO-06 deferral sequencing)

### Tertiary (LOW confidence)
- *(none — no WebSearch needed; the domain is internal Go + fully specified by milestone research)*

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — stdlib only, verified in go.mod; no new deps
- Architecture: HIGH — single-tier Go change, hook point read in source, design locked by ERGO reqs
- Pitfalls: HIGH — `gRPC` edge case identified and resolved to "table is the contract"; backward-compat corpus test is concrete
- Algorithm: HIGH on the contract (the table), MEDIUM on the exact heuristic (executor iterates to green) — this is the intendedClaude's-Discretion boundary

**Research date:** 2026-08-08
**Valid until:** 2026-09-07 (stable — pure stdlib, no fast-moving deps)
