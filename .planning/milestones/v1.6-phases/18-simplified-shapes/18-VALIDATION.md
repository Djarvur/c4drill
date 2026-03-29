# Phase 18: Simplified Shapes - Validation Strategy

**Phase:** 18
**Slug:** simplified-shapes
**Created:** 2026-03-24

## Validation Architecture

### Test Framework

- **Unit tests:** Go testing package with `go test`
- **Integration tests:** Existing testdata in `testdata/` directory
- **Build verification:** `go build ./...`

### Test Mapping by Requirement

| Requirement | Test Type | Test Location | Validation Method |
|-------------|-----------|---------------|-------------------|
| ICON-01 | File deletion | N/A | `test ! -d internal/render/icons/` |
| ICON-02 | File deletion | N/A | `test ! -f internal/render/icon_extractor.go` |
| ICON-03 | Build | N/A | `go build ./...` (no icon dir refs) |
| ICON-04 | File deletion | N/A | `test ! -f internal/render/svg_icons.go` |
| DB-01 | Unit | labels_test.go | `TestDbHTMLLabel` checks shape=cylinder |
| DB-02 | Unit | labels_test.go | `TestDbExternalHTMLLabel` checks shape=cylinder |
| DB-03 | Unit | labels_test.go | Label contains 3 rows, no icon column |
| QUEUE-01 | Unit | labels_test.go | `TestQueueHTMLLabel` checks shape=cylinder + orientation |
| QUEUE-02 | Unit | labels_test.go | `TestQueueExternalHTMLLabel` checks rotated cylinder |
| QUEUE-03 | Unit | labels_test.go | Label contains 3 rows, no icon column |
| PERSON-01 | Unit | labels_test.go | `TestPersonHTMLLabel` checks 2-column table |
| PERSON-02 | Unit | labels_test.go | First column contains `&#x1F464;` |
| PERSON-03 | Unit | labels_test.go | Second column contains name/description |
| PERSON-04 | Unit | labels_test.go | `TestPersonExternalHTMLLabel` same format |
| LABEL-01 | Unit | labels_test.go | `TestSystemHTMLLabel` 3-row table |
| LABEL-02 | Unit | labels_test.go | `TestBoxHTMLLabel` 3-row table |
| LABEL-03 | Unit | labels_test.go | Container/Component 3-row table |
| LABEL-04 | Unit | labels_test.go | No `<img` in non-Person labels |
| WRAP-01 | Unit | wrap_test.go | Existing wrap tests pass |
| WRAP-02 | CLI | root.go | `--label-ratio` flag works |

### Automated Verification Commands

**Task 1 verification:**
```bash
go build ./... 2>&1 | head -50
```

**Task 2 verification:**
```bash
go build ./... 2>&1 | head -50
```

**Task 3 verification:**
```bash
go test ./internal/render/... -v -short 2>&1 | tail -30
```

### Key Validation Points

1. **Icon system removal** — Files must not exist after deletion
2. **Cylinder shapes** — Converter must set `CylinderShape` for DB/Queue types
3. **Queue rotation** — Converter must call `SetOrientation(90.0)` for Queue types
4. **Person emoji** — Labels must contain `&#x1F464;` (not `<img`)
5. **Simplified labels** — No `<img` tags in non-Person labels
6. **Word-wrap preserved** — Existing wrap tests continue to pass

### Regression Prevention

- All existing tests in `labels_test.go`, `wrap_test.go` must pass
- Build must succeed without icon-related imports
- No references to `iconReserve`, `IconExtractor`, `icons.` package

---

*Phase: 18-simplified-shapes*
*Validation strategy created: 2026-03-24*
