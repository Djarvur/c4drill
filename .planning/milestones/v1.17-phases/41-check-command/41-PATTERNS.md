# Phase 41: Check Command - Pattern Map

**Mapped:** 2026-09-03
**Files analyzed:** 6
**Analogs found:** 6 / 6

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `cmd/c4drill/check.go` (NEW) | controller (CLI verb) | file-I/O (read-only) | `cmd/c4drill/fmt.go` | exact |
| `cmd/c4drill/root.go` (MODIFY — extract shared front-half helper + register subcommand) | controller (pipeline) | file-I/O → transform | itself (stages 1→2 at lines 176-223) | exact |
| `cmd/c4drill/check_test.go` (NEW) | test | request-response (cobra Execute) | `cmd/c4drill/fmt_test.go` (+ `root_test.go` validation test) | exact |
| `cmd/c4drill/testdata/check_orphan.toml` (NEW fixture) | test fixture | file-I/O | `cmd/c4drill/testdata/valid.toml` shape + `root_test.go:227-244` inline invalid model | exact |
| `cmd/c4drill/testdata/check_composed_main.toml` + `check_composed_auth.toml` (NEW fixtures) | test fixture | file-I/O (multi-file) | `internal/include/testdata/main.toml` + `auth.toml` | exact |
| `README.adoc` + `skill/SKILL.md` (MODIFY) | docs | — | README.adoc `=== fmt` section (1452-1475); SKILL.md usage list (150-166) | exact |

## Pattern Assignments

### `cmd/c4drill/check.go` (controller, file-I/O read-only)

**Analog:** `cmd/c4drill/fmt.go` — the single-purpose-verb subcommand precedent (registration comment convention, own error sentinels, `SilenceUsage`, exact-args gate).

**Command construction pattern** (fmt.go:46-71):
```go
func newFMTCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fmt [--check] <file|dir>...",
		Short: "Format TOML and C4D diagram files in place (gofmt-style)",
		Long: `...multi-line help with usage + semantics...`,
		Args:         cobra.MinimumNArgs(1),
		RunE:         runFMT,
		SilenceUsage: true,
	}
	return cmd
}
```
check variant: `Use: "check <file.toml|file.c4d>"`, `Args: cobra.ExactArgs(1)` (convert.go:103 precedent for ExactArgs), `RunE: runCheck`, `SilenceUsage: true`, no flags of its own (D-05).

**Registration pattern** (root.go:138-146):
```go
	// Subcommands (Plan 35-07): convert between TOML and C4D formats.
	cmd.AddCommand(newConvertCmd())
	// Subcommands (Plan 35-08): format both authoring formats in place.
	cmd.AddCommand(newFMTCmd())
	// Subcommand (issue #32): the language server ...
	cmd.AddCommand(newServeCmd())
```
check appends: `// Subcommand (issue #41): render-free model validation.` + `cmd.AddCommand(newCheckCmd())`.

**Error sentinel pattern** (fmt.go:32-36, root.go:27-34):
```go
var (
	errFmtNoTargets   = errors.New("no .toml or .c4d files found")
	errFmtCheckNeeded = errors.New("files need formatting")
)
```
check needs NO new sentinel — it reuses root.go's `errValidationFailed` (root.go:30) via the shared helper.

### `cmd/c4drill/root.go` — shared front-half helper (controller, file-I/O → transform)

**Analog:** its own stages 1→2 (root.go:176-223) — the code to extract; cross-check `cmd/c4drill/convert.go:175-201` (`validateSourceForConvert`) which is a live copy of the same sequence.

**Stage pipeline to extract** (root.go:176-223):
```go
	m, err := parseInput(inputPath)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}
	if m, err = include.Resolve(m, filepath.Dir(inputPath), inputPath); err != nil {
		return fmt.Errorf("include: %w", err)
	}
	m, err = template.Expand(m)
	if err != nil {
		return fmt.Errorf("expand: %w", err)
	}
	if err := peer.Resolve(m); err != nil {
		return fmt.Errorf("resolve peers: %w", err)
	}
	valErrors := validator.Validate(m)
	if len(valErrors) > 0 {
		validator.ReportErrors(valErrors, cmd.OutOrStderr())
		return errValidationFailed
	}
```
Boundary rules: helper starts at `parseInput` (NOT at flag validation root.go:164, NOT at `render.LabelRatio = getLabelRatio()` root.go:169 — both are render-only); helper keeps `cmd.OutOrStderr()` (render-identical, D-03; convert.go:195 uses `ErrOrStderr()` — do NOT copy that variant); helper returns `(*parser.Model, error)` so runRoot continues into stages 3-6 unchanged.

**parseInput reuse** (root.go:249-265): already shared, fail-closed extension dispatch (.toml/.c4d else `errUnsupportedExt`) — check gets it free through the helper.

### `cmd/c4drill/check_test.go` (test, request-response)

**Analog:** `cmd/c4drill/fmt_test.go` — exec helper + exit-code/no-write assertions; `root_test.go:227-247` — validation-stage failure through the root command.

**Exec helper pattern** (fmt_test.go:293-301):
```go
func execFMT(t *testing.T, args ...string) (*bytes.Buffer, error) {
	t.Helper()
	cmd := NewRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(append([]string{"fmt"}, args...))
	return buf, cmd.Execute() //nolint:wrapcheck // test returns the command error verbatim
}
```
check variant: `execCheck(t, args ...string)` prepending `"check"`. Tests run through `NewRootCmd()` for argv parity.

**Exit-code + silence assertions** (fmt_test.go:135-158):
```go
	buf, runErr := execFMT(t, "--check", path)
	require.Error(t, runErr, "--check exits 1 on a misformatted file (D-31)")
	assert.Contains(t, buf.String(), path)
	assert.Equal(t, before, readFixture(t, path), "--check writes NOTHING (zero byte change)")
	// ...and:
	buf, runErr := execFMT(t, "--check", path)
	require.NoError(t, runErr, "--check exits 0 on a formatted file")
	assert.Empty(t, buf.String(), "--check is silent when clean")
```
Same shape pins CHECK-01 (no writes: compare dir listing/bytes before+after) and D-04 (silent success: empty out+err buffers).

**Validation-failure fixture technique** (root_test.go:227-247) — the in-repo way to reach the VALIDATOR stage rather than peer.Resolve:
```go
	// Uses a DOTTED peer so it skips the relative-peer resolver (D-16 step 1:
	// peers containing "." are absolute) and reaches the validator's
	// undefined-unit check ... A bare peer would now be caught by peer.Resolve
	// before validation (Phase 30).
	[[user.link]]
	peer = "no.such.unit"
```
For CHECK-02's named orphan-unit rule, the alternative is a unit with zero links/subunits (`internal/validator/rules.go:125-142` fires `unit "X" has no incoming or outgoing links`). Prefer the orphan fixture (CHECK-02 names orphan rules explicitly); the dotted-peer trick is the fallback if orphan detection needs subunit coverage.

**Parallelism directive:** every cmd/c4drill test carries `//nolint:paralleltest // go-graphviz WASM engine has concurrency issues` (root_test.go:21) or `//nolint:paralleltest // cobra flags bind package-level vars; serial execution only` (fmt_test.go:148) — check tests must too.

### `cmd/c4drill/testdata/check_*.toml` fixtures (test fixture, file-I/O)

**Analog:** `cmd/c4drill/testdata/valid.toml` (valid-model shape: `[properties]`, `[user]` person with `[[user.link]]`, `[app]`/`[app.api]` with `[[app.api.linkFrom]]` — every unit linked, so it validates clean) and `internal/include/testdata/main.toml`+`auth.toml` (include directive shape):

```toml
# internal/include/testdata/main.toml — entry with one [[include]]
[properties]
name = "Main System"
[[include]]
path = "auth.toml"
```
Composed fixture: entry file with `[[include]] path = "check_composed_auth.toml"` plus a link crossing the file boundary (`peer = "authService"` — resolves only when include ran, proving CHECK-03 ordering).

### `README.adoc` + `skill/SKILL.md` (docs)

**Analog:** README.adoc `=== fmt` section (1452-1475):
```adoc
=== fmt

[source,text]
----
c4drill fmt [--check] <file|dir>...
----

* Formats `.c4d` and `.toml` files *in place*, ...
* `--check` reports misformatted files one per line and exits 1
  without writing anything — the CI gate:

[source,bash]
----
c4drill fmt --check .   # exits 1 listing offending files
----
```
`=== check` mirrors this: usage line `c4drill check <file.toml|file.c4d>`, bullets for both formats + what it runs (same front-half as render), exit codes (0 valid / 1 invalid), no-output guarantee, a bash example. SKILL.md adds a `c4drill check` line to the "Converting and Formatting" bash block (skill/SKILL.md:152-166).

## Shared Patterns

- **Stage-prefixed error wrapping** (`fmt.Errorf("<stage>: %w", err)`): parse / include / expand / resolve peers — identical strings across root.go:182-214; the shared helper makes them single-sourced.
- **Sentinel errors + `%w` wrapping** (root.go:27-34): no new sentinels needed for check.
- **`//nolint:` directives**: `wrapcheck` on RunE returns and test Execute returns; `paralleltest` on every test (see above); `gochecknoglobals` for cobra flag vars (root.go:56) if check ever needed one (it does not — D-05).
- **Silent success convention**: root.go:246 `return nil // Success - silent per spec` — check's happy path is the same silence (D-04).
- **Fixtures under `cmd/c4drill/testdata/`**: flat TOML files, lowercase snake_case names prefixed by feature (`edges_override.toml`, `peer_walkup.toml`) → prefix `check_`.
