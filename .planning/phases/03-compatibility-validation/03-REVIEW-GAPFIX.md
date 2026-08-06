---
phase: 03-compatibility-validation
plan: 04
reviewed: 2026-08-06T00:00:00Z
depth: standard
files_reviewed: 7
files_reviewed_list:
  - internal/graph/path.go
  - internal/graph/path_test.go
  - internal/render/converter.go
  - internal/render/navigation.go
  - internal/render/navigation_test.go
  - internal/render/integration_test.go
  - cmd/c4drill/root_test.go
findings:
  critical: 0
  blocker: 0
  warning: 3
  info: 2
  total: 5
status: issues_found
---

# Phase 03 (plan 03-04): Code Review Report — Gap-Closure Delta

**Reviewed:** 2026-08-06
**Depth:** standard
**Files Reviewed:** 7
**Status:** issues_found

## Summary

The 03-04 delta closes the three UAT-diagnosed C2/C3 navigation gaps (ComputeExploreURL bidirectional relative paths + self-link guard; GraphViz HTML TABLE/TD HREF idiom; forced `.svg` in nav URLs). All targeted bugs are genuinely fixed: the new `ComputeExploreURL` algorithm is correct across ancestor/sibling/descendant/cross-branch cases (traced and verified), the unchanged C2→C3 descendant case still resolves against the real file tree, the GraphViz `<TD HREF>` idiom renders as `<a xlink:href>` anchors in SVG, and the end-to-end `TestCompat02_NavigationLinksResolve` actually `os.Stat`s the resolved paths (good test quality). The full suite is green under `-race` and `go vet` is clean.

The review surfaced one substantive robustness defect (WR-01) that is reachable today for unit names containing `&`, plus two lower-severity issues. None of the targeted gaps regressed.

## Verification performed

- Traced `ComputeExploreURL` across 13 edge cases (self-link, ancestor-up x1/x2, sibling, descendant, cross-branch, disjoint, C1 variants). All file-tree-resolving cases produce correct relative paths; ancestor cases yield `../parent.svg` / `../../grandparent.svg` as required.
- Confirmed empirically (in-repo render harness) that the `<TD HREF="url">` idiom renders `<a xlink:href="url">` anchors in SVG and that display names are HTML-escaped (`html.EscapeString`).
- Confirmed the unchanged descendant case (`cur="mainSystem" tgt="mainSystem.sshAuth"` → `"mainSystem/sshAuth.svg"`) resolves to `{basename}/mainSystem/sshAuth.svg` from the C2 file's directory, and `nested_path_with_multiple_levels` stays green.
- Reproduced WR-01 empirically: a navigation URL containing a raw `&` causes GraphViz to silently drop the entire navigation TD (no `<a xlink:href>` emitted in SVG).
- `go test -race ./...` all packages green; `golangci-lint run` flags `TestComputeExploreURL` funlen (WR-03) plus pre-existing issues outside this delta.

## Narrative Findings (AI reviewer)

## Warnings

### WR-01: Navigation `<TD HREF>` URLs are not HTML-attribute-escaped; a `&` in a unit name silently drops the navigation bar

**File:** `internal/render/navigation.go:52`, `internal/render/navigation.go:91`
**Issue:**
The clickable navigation TDs embed the URL value raw into the HTML-like label attribute:

```go
tds = append(tds, fmt.Sprintf(`<TD HREF="%s">%s</TD>`, nav.BackLink.URL, label))   // line 52
return fmt.Sprintf(`<TD HREF="%s">%s</TD>`, item.URL, escaped)                      // line 91
```

The display label is correctly HTML-escaped (`html.EscapeString`), but the URL is not. The URLs are produced by `URLEncodePath` (`url.PathEscape`), which **does not escape `&`** (`url.PathEscape("a&b") == "a&b"`, verified). GraphViz's HTML-like label parser rejects a raw `&` that is not part of a valid entity, and on failure it silently drops the offending TD — and in practice drops the whole navigation anchor.

Empirical reproduction (in-repo render harness, navigation with `URL: "../a&b.svg"`):
- DOT label still contains `<TD HREF="../a&b.svg">` (parsed leniently for DOT emission) but `lheight=0.00` indicates the label failed to lay out.
- SVG output contains **no** `<a xlink:href>` anchor for that navigation item — the link disappears entirely.

By contrast, node explore URLs go through `cn.SetURL(...)` (a GraphViz attribute, not an HTML label) and are correctly escaped to `xlink:href="explore&amp;a.svg"` in SVG. So the bug is specific to the navigation TD HREF path introduced by this delta.

Reachability: a TOML unit key or display name containing `&` (e.g. `[r&d]`, `name = "R&D"`) flows through `ComputeBackLinkURL`/`computeBreadcrumbURL` → `URLEncodePath(parentName)` / encoded target segments, producing a URL with a raw `&`. The `basename` (output filename stem) is also placed verbatim into C2 back-links and C1 explore URLs, so an output basename containing `&` triggers it too. Names with `<` are escaped by `url.PathEscape` (`%3C`), so only `&` is the practical trigger.

This defeats the Gap 2 fix for the `&`-in-name case: the navigation bar that the delta makes clickable vanishes instead of rendering.

**Severity rationale:** Treated as WARNING rather than BLOCKER because the common case (alphanumeric unit names) works correctly and the full UAT/COMPAT-02 fixtures pass; the defect manifests only for names containing `&`. It is, however, a user-facing silent failure (missing nav bar) on a realistic input, so it should be fixed before the delta ships broadly.

**Fix:**
HTML-escape the URL when placing it into the `HREF="..."` attribute (GraphViz HTML labels require entity-escaped attribute values), mirroring the escaping already applied to the display label and to the title in `converter.go:204`:

```go
// navigation.go, navigationTDs (back-link):
if nav.BackLink != nil && nav.BackLink.URL != "" {
    label := html.EscapeString("Back to " + nav.BackLink.Name)
    href := html.EscapeString(nav.BackLink.URL)
    tds = append(tds, fmt.Sprintf(`<TD HREF="%s">%s</TD>`, href, label))
}

// breadcrumbItemTD:
func breadcrumbItemTD(item graph.BreadcrumbItem) string {
    escaped := html.EscapeString(item.Name)
    if item.URL == "" {
        return plainTD(escaped)
    }
    return fmt.Sprintf(`<TD HREF="%s">%s</TD>`, html.EscapeString(item.URL), escaped)
}
```

`html.EscapeString` escapes `&`, `<`, `>`, `'`, `"` — exactly the set that is unsafe in an HTML attribute. `url.PathEscape`-encoded segments (`%20`, `%28`, etc.) contain no characters that `html.EscapeString` alters, so legitimate encoded URLs are unaffected. Add a unit test with `URL: "../a&b.svg"` asserting the emitted label contains `HREF="../a&amp;b.svg"` and that the rendered SVG still contains `<a xlink:href="../a&amp;b.svg">`.

---

### WR-02: `ComputeExploreURL` self-link guard is asymmetric and the empty-target cases emit broken `".svg"` (latent robustness gap)

**File:** `internal/graph/path.go:35`, `internal/graph/path.go:46-74`
**Issue:**
The self-link guard is `currentPath != "" && targetPath == currentPath` (line 35). This is asymmetric: it only fires when `currentPath` is non-empty. Tracing the uncovered cases:

| `currentPath` | `targetPath` | Result | Problem |
|---------------|--------------|--------|---------|
| `"a"` (C2) | `""` (C1 root) | `".svg"` | broken empty filename |
| `""` (C1) | `""` | `"basename/.svg"` | broken path; guard did not fire because `currentPath == ""` |

The downstream `tgtParts := strings.Split("", ".") == []string{""}` yields `lastTarget := URLEncodePath("") == ""`, so the URL collapses to `".svg"` (or `"basename/.svg"` for the C1 branch). These are exactly the "broken href=`.svg`" class of bug that Gap 1 symptom B fixed for the non-empty self-link case — just in a different input region.

Reachability today: **not reachable through current call sites.** `ComputeExploreURL` is called from (a) `builder.go:596/605` with `targetPath = node.ID` (unit keys are never empty), and (b) `computeBreadcrumbURL` with `ancestorPath = parts[:ancestorIndex+1]` joined, which is never empty for valid `parts`. So this is a latent gap, not an active bug. It is, however, a trap: the function's exported contract makes no statement rejecting empty targets, the guard looks comprehensive but isn't, and a future caller (e.g. a breadcrumb-to-root feature, or a node with an empty ID from a parser change) would reintroduce the exact symptom B bug.

**Severity rationale:** WARNING (robustness/defensive-correctness), not BLOCKER, because no current input triggers it and all existing tests pass.

**Fix:**
Make the guard symmetric and reject empty targets explicitly, so the function cannot emit a degenerate `".svg"`:

```go
func ComputeExploreURL(currentPath, targetPath, basename, _ string) string {
    const linkFormat = "svg"

    // Self-link / empty-target guard (Gap 1 symptom B, generalized): a target
    // equal to the current path, or an empty target, cannot yield a valid
    // exploration URL — return empty so the caller omits the link.
    if targetPath == "" || targetPath == currentPath {
        return ""
    }
    // ... rest unchanged
}
```

Add unit subtests: `ComputeExploreURL("a", "", "x", "svg") == ""` and `ComputeExploreURL("", "", "x", "svg") == ""`.

---

### WR-03: `TestComputeExploreURL` exceeds funlen limit (103 > 60) without the repo's `//nolint:funlen` directive

**File:** `internal/graph/path_test.go:13`
**Issue:**
The 03-04 delta grew `TestComputeExploreURL` from ~61 lines to 103 lines by appending the Gap 1 subtests. `funlen` is enabled in `.golangci.yml` (default limit 60) and `golangci-lint run` now flags it:

```
internal/graph/path_test.go:13:6: Function 'TestComputeExploreURL' is too long (103 > 60) (funlen)
```

The established repo convention for long test functions is to suppress with a directive — `internal/graph/builder_test.go` carries `//nolint:funlen // Test functions with model setup are naturally longer` on 9 functions, and `internal/render/navigation_test.go:15` already uses `//nolint:paralleltest,funlen`. The new delta omitted the directive on the one function it pushed over the limit, so this file regresses the lint gate that the plan's `<done>` criteria ("`golangci-lint run --new-from-rev=<last-commit> ./...` 0 issues") requires.

**Fix:** Add the directive above the function declaration, consistent with the sibling tests:

```go
//nolint:funlen // Test functions with many subtests are naturally longer
func TestComputeExploreURL(t *testing.T) {
```

(Pre-existing funlen/gocognit/errcheck findings in `converter.go:175-176`, `builder.go`, `wrap.go`, and other root_test.go goconst hits are out of scope for this delta — they predate 03-04 and are not called out here.)

---

## Info

### IN-01: `BuildNavigationLabel` is exported but has no production callers after the delta; near-duplicate TABLE wrapping

**File:** `internal/render/navigation.go:24`
**Issue:**
The Gap 2 fix in `converter.go:191` calls the unexported `navigationTDs(nav)` directly and assembles its own `<TABLE>` wrapper (so it can merge navigation + title into one multi-row table). As a result `BuildNavigationLabel` — which wraps `navigationTDs` in its own single-row `<TABLE>` — is now referenced only by `navigation_test.go`. It is therefore an exported symbol with no production callers, and the two TABLE-assembly sites (`converter.go:207` and `navigation.go:34`) are near-duplicate constructions of the same `<TABLE BORDER="0" CELLBORDER="0" CELLSPACING="0">...</TABLE>` shell.

This is not a bug (the function is unit-tested and the export may be intentional API surface), but it is an unnecessary public export and a small DRY smell. Options: (a) unexport `BuildNavigationLabel` → `buildNavigationLabel` and have the converter reuse it when there is no title to merge, or (b) have `BuildNavigationLabel` accept an optional title row so the converter stops duplicating the TABLE shell. Either keeps one canonical TABLE builder. No action required if the export is deliberate.

---

### IN-02: C1 diagrams now render their title inside an HTML TABLE label (behavioral change broader than the plan states)

**File:** `internal/render/converter.go:191-216`
**Issue:**
The plan frames Gap 2 as a C2/C3 navigation fix, but the implemented `configureGraphSettings` unconditionally builds an HTML `<TABLE>` label whenever `g.Title != ""`, even for C1 diagrams that have no navigation. Previously C1 titles were plain-text labels (`label="System Context"`); now they are HTML labels (`label=<<TABLE...><TR><TD>System Context</TD></TR></TABLE>>`, verified empirically). The delta's prose ("C1 for `""` → `{basename}.svg`") and the plan's scope ("only enter the HTML-label path when there IS navigation ... OR the title needs HTML") suggest the simpler/safer approach of keeping C1 on the plain-text path.

This is harmless — the C1 title still renders, `TestIntegrationFullPipelineDOT` and `TestIntegration_SVG_C1NoNavigation` still pass (substring/not-contains assertions), and no golden test locks the C1 label format. It is, however, an unintended broadening: every diagram's title is now an HTML label rather than just the C2/C3 navigation bar. If C1 title rendering is ever golden-locked in future, or if a title legitimately contains characters that need plain-text (not HTML) semantics, this will surface. Worth a one-line decision: either keep the unified HTML path (current) and document it, or short-circuit to plain `cg.SetLabel(g.Title)` when `len(navTDs) == 0`.

---

_Reviewed: 2026-08-06_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
