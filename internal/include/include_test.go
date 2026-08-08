// Package include_test exercises the recursive include resolver (Plan 32-02).
//
// These tests cover INC-01 through INC-10, D-10 (cross-file subunit merge),
// D-11 (same-file diamond dedup vs cross-file collision), and XC-02 (templates
// in included files flow through the merge). They are pure-Go tests of the
// resolver against *parser.Model — they do NOT touch the go-graphviz WASM
// engine, so t.Parallel() is safe.
package include_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Djarvur/c4drill/internal/include"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeFiles writes the given name→content map into a fresh t.TempDir() and
// returns the directory path. Used for multi-file fixtures whose include graph
// is self-contained in the test source (cycle/diamond/once/transitive/dup/
// props/missing/templates) so the test intent is co-located with the test.
func writeFiles(t *testing.T, files map[string]string) string {
	t.Helper()

	dir := t.TempDir()

	for name, content := range files {
		path := filepath.Join(dir, name)
		require.NoError(t, os.WriteFile(path, []byte(content), 0o600), "write %s", name)
	}

	return dir
}

// parseAndResolve is the common harness: ParseFile the entry, then Resolve.
// entryDir is filepath.Dir(entryPath) so [[include]] paths resolve relative to
// the including file (INC-02).
func parseAndResolve(t *testing.T, entryPath string) (*parser.Model, error) {
	t.Helper()

	entry, err := parser.ParseFile(entryPath)
	require.NoError(t, err, "ParseFile(%s) should not error", entryPath)

	return include.Resolve(entry, filepath.Dir(entryPath))
}

// requireParseError asserts err is a *parser.ParseError and returns it.
func requireParseError(t *testing.T, err error) *parser.ParseError {
	t.Helper()

	require.Error(t, err, "expected an error")

	var pe *parser.ParseError
	require.ErrorAs(t, err, &pe, "error must be *parser.ParseError")

	return pe
}

// --- INC-01 / INC-09: two-file merge + UnitOrder append ordering ---

// TestResolveTwoFilesMerge (INC-01, INC-09): main.toml includes auth.toml; the
// merged Units has both files' top-level units and UnitOrder is main's units
// followed by auth's units in include-directive order (D-09 append).
func TestResolveTwoFilesMerge(t *testing.T) {
	t.Parallel()

	merged, err := parseAndResolve(t, "testdata/main.toml")
	require.NoError(t, err, "Resolve should merge without error")

	// Both files' top-level units present.
	assert.Contains(t, merged.Units, "user", "entry unit 'user' preserved")
	assert.Contains(t, merged.Units, "authService", "included unit 'authService' merged in")
	assert.Contains(t, merged.Units, "tokenStore", "included unit 'tokenStore' merged in")
	assert.Len(t, merged.Units, 3, "merged Units has entry + 2 included")

	// UnitOrder: entry's units first, then included file's units in include-
	// directive order (auth.toml declares authService then tokenStore).
	require.Len(t, merged.UnitOrder, 3, "UnitOrder has all three units")
	assert.Equal(t, "user", merged.UnitOrder[0], "entry unit comes first (D-09 append)")
	assert.Equal(t, "authService", merged.UnitOrder[1], "first included unit (auth.toml order)")
	assert.Equal(t, "tokenStore", merged.UnitOrder[2], "second included unit (auth.toml order)")

	// Includes drained to nil after resolution.
	assert.Empty(t, merged.Includes, "Includes drained to nil after Resolve")
}

// TestMergeUnitOrderAppend (INC-09): explicit standalone assertion that entry
// units come before included units in UnitOrder (the append import model, D-09).
func TestMergeUnitOrderAppend(t *testing.T) {
	t.Parallel()

	dir := writeFiles(t, map[string]string{
		"entry.toml": `
[properties]
name = "Entry"

[entryFirst]
type = "system"
name = "Entry First"

[entrySecond]
type = "system"
name = "Entry Second"

[[include]]
path = "lib.toml"
`,
		"lib.toml": `
[libOne]
type = "db"
name = "Lib One"

[libTwo]
type = "queue"
name = "Lib Two"
`,
	})

	merged, err := parseAndResolve(t, filepath.Join(dir, "entry.toml"))
	require.NoError(t, err, "Resolve should merge without error")

	// Entry's two units first, then lib's two units — NOT alphabetical, NOT
	// splice-in-place.
	require.Equal(t, []string{"entryFirst", "entrySecond", "libOne", "libTwo"},
		merged.UnitOrder, "UnitOrder: entry units first, then included units (D-09 append)")
}

// --- INC-02: relative-to-including-file path resolution (cd-proof) ---

// TestResolveRelativePathIndependentOfCwd (INC-02): resolving the same entry
// file from two different working directories yields identical merged
// Units/UnitOrder, because paths resolve relative to the including file's dir
// (not CLI cwd).
func TestResolveRelativePathIndependentOfCwd(t *testing.T) {
	t.Parallel()

	dir := writeFiles(t, map[string]string{
		"entry.toml": `
[properties]
name = "Cwd-Proof"

[host]
type = "system"
name = "Host"

[[include]]
path = "inc.toml"
`,
		"inc.toml": `
[guest]
type = "db"
name = "Guest"
`,
	})

	originalCwd, err := os.Getwd()
	require.NoError(t, err, "Getwd")

	t.Cleanup(func() {
		_ = os.Chdir(originalCwd)
	})

	entryPath := filepath.Join(dir, "entry.toml")

	// First resolve from the original cwd.
	require.NoError(t, os.Chdir(os.TempDir()), "chdir to TempDir for first resolve")

	mergedFromTempDir, err := parseAndResolve(t, entryPath)
	require.NoError(t, err, "Resolve from TempDir")

	// Second resolve from a completely different cwd.
	otherDir := t.TempDir()
	require.NoError(t, os.Chdir(otherDir), "chdir to a different temp dir")

	mergedFromOtherDir, err := parseAndResolve(t, entryPath)
	require.NoError(t, err, "Resolve from a different cwd")

	// Identical regardless of cwd (INC-02).
	assert.Equal(t, mergedFromTempDir.UnitOrder, mergedFromOtherDir.UnitOrder,
		"UnitOrder identical regardless of cwd (INC-02)")
	assert.Len(t, mergedFromTempDir.Units, 2, "both files' units present")
	assert.Contains(t, mergedFromTempDir.Units, "host")
	assert.Contains(t, mergedFromTempDir.Units, "guest")
}

// --- INC-03: transitive includes resolve recursively ---

// TestResolveTransitive (INC-03): top→mid→leaf chain. All three files' units
// merge into one model.
func TestResolveTransitive(t *testing.T) {
	t.Parallel()

	dir := writeFiles(t, map[string]string{
		"top.toml": `
[properties]
name = "Transitive"

[topUnit]
type = "system"
name = "Top"

[[include]]
path = "mid.toml"
`,
		"mid.toml": `
[midUnit]
type = "container"
name = "Mid"

[[include]]
path = "leaf.toml"
`,
		"leaf.toml": `
[leafUnit]
type = "component"
name = "Leaf"
`,
	})

	merged, err := parseAndResolve(t, filepath.Join(dir, "top.toml"))
	require.NoError(t, err, "transitive Resolve should not error")

	assert.Contains(t, merged.Units, "topUnit", "top unit present")
	assert.Contains(t, merged.Units, "midUnit", "transitively-included mid unit present")
	assert.Contains(t, merged.Units, "leafUnit", "transitively-included leaf unit present (INC-03)")

	// Append order: top, then mid, then leaf.
	require.Equal(t, []string{"topUnit", "midUnit", "leafUnit"}, merged.UnitOrder,
		"transitive UnitOrder: entry, then each include's units in directive order")
}

// --- INC-04: cycle detection (self + mutual) ---

// TestResolveCycleFatal (INC-04): mutual cycle (a→b→a) and self-cycle (s→s)
// both fatal with a *parser.ParseError naming the cycle.
func TestResolveCycleFatal(t *testing.T) {
	t.Parallel()

	t.Run("mutual cycle", func(t *testing.T) {
		t.Parallel()

		dir := writeFiles(t, map[string]string{
			"cycle_a.toml": `
[properties]
name = "Cycle A"
[[include]]
path = "cycle_b.toml"
`,
			"cycle_b.toml": `
[[include]]
path = "cycle_a.toml"
`,
		})

		_, err := parseAndResolve(t, filepath.Join(dir, "cycle_a.toml"))
		pe := requireParseError(t, err)
		assert.Contains(t, pe.Error(), "cycle", "error message names the cycle (INC-04)")
	})

	t.Run("self cycle", func(t *testing.T) {
		t.Parallel()

		dir := writeFiles(t, map[string]string{
			"self_cycle.toml": `
[properties]
name = "Self Cycle"
[[include]]
path = "self_cycle.toml"
`,
		})

		_, err := parseAndResolve(t, filepath.Join(dir, "self_cycle.toml"))
		pe := requireParseError(t, err)
		assert.Contains(t, pe.Error(), "cycle", "self-include is a cycle (INC-04)")
	})
}

// --- INC-05 / D-11: diamond not a cycle; same-file diamond auto-dedup ---

// TestResolveDiamondNotCycle (INC-05, D-11): top→left+right, both left and right
// include shared.toml. This is NOT a cycle. shared's units appear exactly once
// (auto-dedup of the same file reached via two non-ancestral paths).
func TestResolveDiamondNotCycle(t *testing.T) {
	t.Parallel()

	dir := writeFiles(t, map[string]string{
		"top.toml": `
[properties]
name = "Diamond"
[topUnit]
type = "system"
name = "Top"
[[include]]
path = "left.toml"
[[include]]
path = "right.toml"
`,
		"left.toml": `
[leftUnit]
type = "container"
name = "Left"
[[include]]
path = "shared.toml"
`,
		"right.toml": `
[rightUnit]
type = "container"
name = "Right"
[[include]]
path = "shared.toml"
`,
		"shared.toml": `
[sharedUnit]
type = "db"
name = "Shared"
`,
	})

	merged, err := parseAndResolve(t, filepath.Join(dir, "top.toml"))
	require.NoError(t, err, "diamond is NOT a cycle (INC-05)")

	// shared's unit present exactly once (D-11 auto-dedup).
	assert.Contains(t, merged.Units, "sharedUnit", "shared unit present")
	assert.Contains(t, merged.Units, "leftUnit", "left unit present")
	assert.Contains(t, merged.Units, "rightUnit", "right unit present")

	// Count occurrences in UnitOrder: sharedUnit appears exactly once.
	count := 0

	for _, name := range merged.UnitOrder {
		if name == "sharedUnit" {
			count++
		}
	}

	assert.Equal(t, 1, count, "sharedUnit deduped to a single occurrence (D-11)")
	assert.Len(t, merged.Units, 4, "four distinct units: top, left, right, shared")
}

// --- INC-06: once=true skips re-inclusion ---

// TestResolveOnceDedup (INC-06): main includes lib twice (second with
// once=true); lib's units present once, no duplicate-definition error.
func TestResolveOnceDedup(t *testing.T) {
	t.Parallel()

	dir := writeFiles(t, map[string]string{
		"once_main.toml": `
[properties]
name = "Once"
[mainUnit]
type = "system"
name = "Main"
[[include]]
path = "once_lib.toml"
[[include]]
path = "once_lib.toml"
once = true
`,
		"once_lib.toml": `
[libUnit]
type = "db"
name = "Lib"
`,
	})

	merged, err := parseAndResolve(t, filepath.Join(dir, "once_main.toml"))
	require.NoError(t, err, "once=true should not error")

	assert.Contains(t, merged.Units, "libUnit", "lib unit present (once)")
	assert.Len(t, merged.Units, 2, "main + lib only (lib not duplicated)")
}

// --- INC-07 / D-11 cross-file: duplicate unit path hard-errors ---

// TestMergeDuplicateUnitPathError (INC-07, D-11 cross-file): two DIFFERENT files
// defining the same top-level unit path is a hard error naming both files.
func TestMergeDuplicateUnitPathError(t *testing.T) {
	t.Parallel()

	dir := writeFiles(t, map[string]string{
		"dup_main.toml": `
[properties]
name = "Dup"
[[include]]
path = "dup_other.toml"
[mailAdapter]
type = "system"
name = "Main Mail Adapter"
`,
		"dup_other.toml": `
[mailAdapter]
type = "system"
name = "Other Mail Adapter"
`,
	})

	_, err := parseAndResolve(t, filepath.Join(dir, "dup_main.toml"))
	pe := requireParseError(t, err)

	// Error names both files (D-11 cross-file attribution).
	msg := pe.Error()
	assert.Contains(t, msg, "mailAdapter", "error names the conflicting unit path (INC-07)")
	assert.Contains(t, msg, "dup_main.toml", "error names the entry file")
	assert.Contains(t, msg, "dup_other.toml", "error names the included file (D-11)")
}

// --- INC-08: properties root-wins / conflict hard-error ---

// TestMergePropertiesConflictError (INC-08): entry and included file both set
// properties.name with different values → fatal. When the included file's name
// is empty → root-wins, no error.
func TestMergePropertiesConflictError(t *testing.T) {
	t.Parallel()

	t.Run("conflict errors", func(t *testing.T) {
		t.Parallel()

		dir := writeFiles(t, map[string]string{
			"props_main.toml": `
[properties]
name = "Main"
[[include]]
path = "props_conflict.toml"
`,
			"props_conflict.toml": `
[properties]
name = "Conflict"
[extra]
type = "db"
name = "Extra"
`,
		})

		_, err := parseAndResolve(t, filepath.Join(dir, "props_main.toml"))
		pe := requireParseError(t, err)
		assert.Contains(t, pe.Error(), "properties", "error mentions properties (INC-08)")
		assert.Contains(t, pe.Error(), "props_main.toml", "error names the entry file")
		assert.Contains(t, pe.Error(), "props_conflict.toml", "error names the conflicting file")
	})

	t.Run("empty included name root-wins", func(t *testing.T) {
		t.Parallel()

		dir := writeFiles(t, map[string]string{
			"props_main.toml": `
[properties]
name = "Main"
[[include]]
path = "props_empty.toml"
`,
			"props_empty.toml": `
[extra]
type = "db"
name = "Extra"
`,
		})

		merged, err := parseAndResolve(t, filepath.Join(dir, "props_main.toml"))
		require.NoError(t, err, "empty included name should not conflict (root-wins)")
		assert.Equal(t, "Main", merged.Properties.Name, "entry's name wins (INC-08 root-wins)")
	})
}

// --- INC-10 / D-12: missing include hard-errors naming path + including file ---

// TestResolveMissingIncludeError (INC-10, D-12): a missing include file is a
// fatal *parser.ParseError naming both the referenced path and the including
// file. No optional flag — unconditional hard error.
func TestResolveMissingIncludeError(t *testing.T) {
	t.Parallel()

	dir := writeFiles(t, map[string]string{
		"missing_main.toml": `
[properties]
name = "Missing"
[[include]]
path = "ghost.toml"
`,
	})

	_, err := parseAndResolve(t, filepath.Join(dir, "missing_main.toml"))
	pe := requireParseError(t, err)

	msg := pe.Error()
	assert.Contains(t, msg, "ghost.toml", "error names the referenced path (INC-10)")
	assert.Contains(t, msg, "missing_main.toml", "error names the including file (D-12)")
}

// --- D-10: cross-file subunit merge ---

// TestMergeCrossFileSubunits (D-10): entry defines [linuxSystem]; included file
// adds [linuxSystem.auth] and [linuxSystem.db] subunits. The subunits attach to
// the existing parent, append to SubunitOrder in include-file order, and do NOT
// appear as phantom top-level units.
func TestMergeCrossFileSubunits(t *testing.T) {
	t.Parallel()

	merged, err := parseAndResolve(t, "testdata/nested_subunits_main.toml")
	require.NoError(t, err, "cross-file subunit merge should not error")

	// Parent is a top-level unit with the two included subunits attached.
	parent, ok := merged.Units["linuxSystem"]
	require.True(t, ok, "parent 'linuxSystem' present as top-level unit")
	require.Contains(t, parent.Subunits, "auth", "auth subunit attached to parent (D-10)")
	require.Contains(t, parent.Subunits, "db", "db subunit attached to parent (D-10)")

	// SubunitOrder ends with auth, db in include-file order.
	require.Len(t, parent.SubunitOrder, 2, "two cross-file subunits in SubunitOrder")
	assert.Equal(t, "auth", parent.SubunitOrder[0], "first cross-file subunit")
	assert.Equal(t, "db", parent.SubunitOrder[1], "second cross-file subunit (include-file order)")

	// Neither subunit leaks as a phantom top-level unit.
	assert.NotContains(t, merged.Units, "auth", "subunit does not leak as top-level unit")
	assert.NotContains(t, merged.Units, "db", "subunit does not leak as top-level unit")
	assert.NotContains(t, merged.UnitOrder, "auth")
	assert.NotContains(t, merged.UnitOrder, "db")
}

// --- XC-02: templates in included files flow into merged Model.Templates ---

// TestMergeCarriesTemplates (XC-02): a template defined in an included file
// flows into merged Model.Templates so [[use]] in the entry file can
// instantiate it. Conditional on Phase 31 (which has landed).
func TestMergeCarriesTemplates(t *testing.T) {
	t.Parallel()

	dir := writeFiles(t, map[string]string{
		"templates_main.toml": `
[properties]
name = "Template Carry"
[[include]]
path = "templates_lib.toml"
[[use]]
template = "svc"
name = "billing"
tech = "Go"
`,
		"templates_lib.toml": `
[template.svc]
type = "container"
name = "${name} Service"
technology = "${tech}"
params = ["name", "tech"]
`,
	})

	merged, err := parseAndResolve(t, filepath.Join(dir, "templates_main.toml"))
	require.NoError(t, err, "template carry merge should not error")

	// The included file's template is present in the merged model.
	require.NotNil(t, merged.Templates, "Templates map non-nil after merge")
	assert.Contains(t, merged.Templates, "svc", "included template 'svc' carried into merged model (XC-02)")

	// The entry file's [[use]] instantiation is also present (carried through).
	require.Len(t, merged.Instantiations, 1, "entry's [[use]] carried through merge")
	assert.Equal(t, "svc", merged.Instantiations[0].Template, "instantiation references the carried template")
}
