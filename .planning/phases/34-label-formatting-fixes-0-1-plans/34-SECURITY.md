---
phase: 34
slug: label-formatting-fixes-0-1-plans
status: verified
threats_open: 0
asvs_level: 1
created: 2026-08-10
---

# Phase 34 — Security

> Per-phase security contract: threat register, accepted risks, and audit trail.

---

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| author TOML → wrapText | user-authored label text (names, descriptions, technologies) crosses into DOT/SVG label output | HTML markup context (label strings) |

---

## Threat Register

| Threat ID | Category | Component | Disposition | Mitigation | Status |
|-----------|----------|-----------|-------------|------------|--------|
| T-34-01 | Tampering | wrapText / wrapAndEscape | mitigate | html.EscapeString applied per line (wrap.go:237-249) — escape path untouched by the word-splitting change | closed |
| T-34-02 | DoS | wrapText | accept | single-pass words loop, bounded by input size | closed |
| T-34-01 | Tampering | buildEdgeLabel rows | mitigate | all row helpers (writeNameRow/writeTechnologyRow/writeDescriptionRow) route text through wrapAndEscape; zero raw concatenations of label text (labels.go:256/266/276) | closed |
| T-34-02 | Spoofing | createEdge label emission | accept | cosmetic label rendering only; no auth/identity semantics | closed |
| T-34-03-1 | Tampering | tokenizer / wrapAndEscape | mitigate | tokenizer changes break opportunities only; escaping unchanged | closed |
| T-34-03-2 | DoS | tokenizeWrapText | accept | single-pass rune scan (wrap.go:150); token count bounded by input length | closed |
| T-34-04-1 | Tampering | label sizing | mitigate | only the maxChars source changed; escape path untouched | closed |
| T-34-04-2 | DoS | labelMaxCharsForText | accept | closed-form quadratic (math.Sqrt, wrap.go:294) — O(1), no loops | closed |
| T-34-SC (×4) | Tampering | package installs | accept | no go.mod/go.sum changes in any phase-34 commit (Go stdlib only: unicode, math, html) | closed |

*Status: open · closed*
*Disposition: mitigate (implementation required) · accept (documented risk) · transfer (third-party)*

---

## Accepted Risks Log

| Risk ID | Threat Ref | Rationale | Accepted By | Date |
|---------|------------|-----------|-------------|------|
| AR-01 | T-34-02 | Over-budget label words emit whole-word lines bounded by input size; no unbounded loop | user | 2026-08-10 |
| AR-02 | T-34-02 (Spoofing) | Edge labels are presentation-only; no trust semantics attached to rendered text | user | 2026-08-10 |
| AR-03 | T-34-03-2 | Tokenizer is a single linear scan; worst-case work is O(n) in label length | user | 2026-08-10 |
| AR-04 | T-34-04-2 | Quadratic sizing is a constant-time closed form; no iteration | user | 2026-08-10 |
| AR-05 | T-34-SC | No new dependencies introduced; stdlib only | user | 2026-08-10 |

---

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-08-10 | 9 | 9 | 0 | ZCode (inline verification — auditor subagent unavailable) |

---

## Sign-Off

- [x] All threats have a disposition (mitigate / accept / transfer)
- [x] Accepted risks documented in Accepted Risks Log
- [x] `threats_open: 0` confirmed
- [x] `status: verified` set in frontmatter

**Approval:** verified 2026-08-10
