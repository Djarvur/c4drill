# Phase 37 Deferred Items

## Stale Milestone-v1.10 reference artifacts (out of scope, pre-existing)

Discovered during 37-05 Task 2 (golden re-baseline audit). The committed files
`cmd/c4drill/testdata/expanded.dot` and `cmd/c4drill/testdata/expanded/mainsystem.dot`
(last touched in commit 2f21325 "Milestone v1.10: Model Composition (#1)") no longer
match current output of `cmd/c4drill/testdata/expanded.toml` — they predate breadcrumb
nav labels, kind-derived edge attributes, and other rendering changes from earlier
milestones.

- NOT consumed by any test: `root_test.go` renders into tmp dirs and only stats
  generated SVGs; the only test-compared goldens are `multilevel.expanded.dot`
  (internal/graph, byte-identical) and `plain.dot` / `plain.expanded.dot`
  (cmd/c4drill, canonical-compared, identical under `--plain`).
- The drift accumulated over multiple milestones before phase 37 — not caused by
  37-01..37-04, so it cannot be attributed to documented CTX deltas and was NOT
  re-baselined (per plan Task 2 step 3: only documented CTX deltas are absorbable).
- Suggested future work: either regenerate these reference artifacts or delete them
  (they duplicate what root_test.go's tmp-dir E2E assertions already cover).
