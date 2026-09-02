# Plugin Integration Landscape: distribution to SCM, docs, and chat platforms

**Domain:** C4Drill plugin feasibility — what a "c4drill plugin" concretely is for GitHub, GitLab, Bitbucket, Confluence, Jira, Slack, Mattermost, Notion, Azure DevOps
**Researched:** 2026-09-01
**Scope:** integration surfaces, effort/value, auth & secret model, artifact-format fit (SVG link survival, PNG dependency), distribution channels, maintenance burden, build/defer/skip decision
**Confidence:** HIGH for platform rendering/sanitization verdicts (each verified against current platform docs or tracker issues, September 2026 — see citations); MEDIUM for two explicitly-flagged items that need a spike (Bitbucket PR SVG display, clickability of links inside Confluence's inline SVG embed)

> This document answers issue #28: it classifies each candidate integration as **build / defer / skip** and proposes an implementation shortlist. The platform-behavior claims in §1 were verified against current documentation, not assumed — two of the issue's working assumptions turned out to be **stale** and are corrected in §1.4 and §1.5 (Confluence Cloud now displays SVG attachments; this materially changes the Confluence verdict).

## C4Drill's load-bearing constraints (the filter for every verdict)

1. **Static, dependency-free binary.** `release.yml` already builds `CGO_ENABLED=0` binaries for linux/darwin/windows × amd64/arm64. A container image needs no package install — the image can be `FROM scratch` + one binary. This is what makes the shared-image strategy (§4) cheap.
2. **No native Graphviz.** The WASM go-graphviz engine renders wherever the binary runs; every CI container accepts the image with zero setup beyond `image:`. No `apt install graphviz` step exists to break.
3. **No hosting.** C4Drill is a single-shot CLI with no server component. Any integration whose value depends on c4drill operating a public endpoint is out of scope by construction (§5 confirms which ones those are).
4. **Two artifact formats ship today** (`-f dot|svg|html`; PNG is #26, in flight). SVG carries clickable `<a>` anchors via the GraphViz `URL` attribute; `-f html` is a self-contained file with a JS nav shim. §1 tracks which platforms destroy or preserve those properties.

---

## 1. Verified platform behavior — SVG, sanitization, and image embedding

This section is the evidence base. Every verdict carries the source it was checked against.

### 1.1 The browser-level constraint everything else derives from

An SVG loaded via `<img>` (or CSS background) is rendered in **static image mode**: no scripts, no interactivity, **no clickable `<a>` anchors inside the SVG**. Anchors survive only when the SVG is opened as a top-level document or embedded via `<object>`/`<iframe>`/inline. ([MDN — SVG overview / img embedding](https://developer.mozilla.org/en-US/docs/Web/SVG); [img vs object comparison](https://stackoverflow.com/questions/4476526/do-i-use-img-object-or-embed-for-svg-files); [drawio #935 documents the same behavior](https://github.com/jgraph/drawio/issues/935))

**Consequence:** any platform that displays an SVG through an `<img>` tag (which is all of them, for security) shows the diagram but kills c4drill's clickable `reference` links *in that embedding context*. The links are not lost — they survive in the file itself and work when the file is opened directly (download, Pages direct navigation, release-asset download). The platform question is therefore always: *does this surface render SVG at all, and is "open directly" one click away?*

### 1.2 GitHub

| Surface | Verdict | Evidence |
|---|---|---|
| Inline `<svg>` markup in Markdown (README, issues, PRs, comments) | **Stripped by the sanitizer**; does not render | [Community discussion #151372](https://github.com/orgs/community/discussions/151372) (sanitizer removes inline SVG for XSS/CSS-attack reasons); [isaacs/github #316](https://github.com/isaacs/github/issues/316) ("by design") |
| SVG *file* referenced via `![](x.svg)` / `<img>` in Markdown | **Renders** (proxied through Camo) but as static image — **links inside are dead**; clicking the image navigates to the raw file | [alexwlchan — SVGs only render via img tag](https://alexwlchan.net/notes/2024/how-to-render-svgs-on-github/); [SO: GitHub Markdown SVG links not working](https://stackoverflow.com/questions/70545385/github-markdown-svg-file-links-not-working); [Shields.io discussion #5593](https://github.com/badges/shields/discussions/5593) ("links embedded in SVG simply will not work… camo") |
| Release assets as image source | **Does not render** — release assets are served as `application/octet-stream`, which Camo refuses | [Community discussion #59781](https://github.com/orgs/community/discussions/59781) |
| Downloaded artifact / release asset opened locally | Full fidelity; links clickable | File itself is unmodified; browser opens SVG as document |
| GitHub Pages, `.svg` served by extension | Served as `image/svg+xml`; **direct navigation preserves clickable links** | [Pages/MIME discussion](https://github.com/orgs/community/discussions/59781) (contrast case: Pages serves correct MIME where release assets don't); [SO #13808020](https://stackoverflow.com/questions/13808020/include-an-svg-hosted-on-github-in-markdown) |

### 1.3 GitLab

| Surface | Verdict | Evidence |
|---|---|---|
| Uploaded SVG files | **Sanitized on serve** — active content/scripts stripped (Loofah-based SVG 1.1-conformant scrubber); plain `<a>` anchors are not the target of the sanitizer, but no surface renders the SVG interactively anyway | [gitlab-ce #27471 (XSS rationale)](https://gitlab.com/gitlab-org/gitlab-ce/-/issues/27471); [MR !3401 (sanitizer implementation)](https://gitlab.com/gitlab-org/gitlab-ce/-/merge_requests/3401); [GitLab security docs — user file uploads](https://docs.gitlab.com/security/user_file_uploads/) |
| **MR diff view for a changed `.svg`** | **Renders as text (source diff), not as an image** — still an open feature request, state `opened`, updated 2026-08-14 | [gitlab #15284 — "Render SVG files as images in MR diffs" (verified open via GitLab API)](https://gitlab.com/gitlab-org/gitlab/-/issues/15284); [MR changes docs (image commenting implies raster images only)](https://docs.gitlab.com/user/project/merge_requests/changes/) |
| Repo README / blob view | SVG referenced from Markdown renders (sanitized, static) — links dead in the embed | [gitlab #26104 (Markdown vs AsciiDoc SVG embedding)](https://gitlab.com/gitlab-org/gitlab/-/issues/26104); [GLFM docs](https://docs.gitlab.com/user/markdown/) |
| GitLab Pages | Same as GitHub Pages: correct MIME, direct navigation preserves links | Pages serves files by extension (same mechanism as §1.2) |

**GitLab-specific consequence:** MR diffs — the natural review surface — show an SVG change as a wall of XML text. A PNG render diff (#26) is what makes MR review *visual* on GitLab.

### 1.4 Confluence Cloud — the issue's assumption is stale

- **SVG attachments now display.** [CONFCLOUD-1762 "Embedding of SVG"](https://jira.atlassian.com/browse/CONFCLOUD-1762) (open since 2004, 367 votes) was **closed as Fixed on 2024-09-10**: a new SVG rendering engine (shipped Oct 2023, hardened through 2024) enables inline display and previews of SVG attachments. The issue #28 table's premise — "storage-format sanitization makes SVG unusable in Confluence" — no longer holds for *display*.
- **Scripts are still stripped.** SVG active content is removed for XSS protection ([community: Confluence does not retain uploaded SVG script content](https://community.atlassian.com/forums/Confluence-questions/Confluence-does-not-retain-use-uploaded-SVG-images/qaq-p/1418344); [CONFCLOUD-25488 (historical XSS via SVG attachment)](https://jira.atlassian.com/browse/CONFCLOUD-25488)). **This does not affect c4drill:** GraphViz SVG output contains only shapes, text, and `<a xlink:href>` anchors — no scripts, no JS shim (the JS shim lives in `-f html`, which must *never* be uploaded to Confluence).
- **Embedding path is documented and stable:** the storage format supports `<ac:image><ri:attachment ri:filename="d.svg"/></ac:image>` and `<ac:image><ri:url ri:value="..."/></ac:image>` ([Confluence Storage Format](https://confluence.atlassian.com/doc/confluence-storage-format-790796544.html)). REST upload is `POST /wiki/rest/api/content/{id}/child/attachment` ([Confluence REST — attachments](https://developer.atlassian.com/cloud/confluence/rest/v1/intro/)).
- **Unverified (spike item):** whether links inside an SVG are clickable in Confluence's inline embed. The embed renders through Atlassian's sanitized preview, likely in image mode (§1.1), so plan for "not clickable inline; PNG fallback via #26 as parity option; full fidelity survives on the attachment itself." This is a presentation detail, not a feasibility blocker.

### 1.5 Jira Cloud

- SVG attachments **upload fine but do not preview** — no thumbnail, no inline render. Long-standing open request: [JRACLOUD-47728 — "Jira doesn't support viewing SVG image"](https://jira.atlassian.com/browse/JRACLOUD-47728); [community: SVG attachments show generic icon](https://community.atlassian.com/forums/Jira-questions/Can-you-show-thumbnails-for-SVG-attachments-in-Jira-Server/qaq-p/1131913).
- Inline images in descriptions/comments require **ADF `media` nodes referencing an uploaded attachment** (REST v3; wiki markup is legacy) — and inline rendering works for raster images only: [community — inline images via ADF](https://community.atlassian.com/forums/Jira-questions/How-to-inline-images-in-Issue-body-or-issue-comment-body/qaq-p/1658456); [KB — inline image display](https://support.atlassian.com/jira/kb/image-attachments-are-not-displayed-inline-in-wiki-renderer-fields/).
- **Verdict: `publish jira` is hard-gated on #26 (PNG).** Without PNG, Jira gets dead attachments; with PNG it gets a first-class inline diagram.

### 1.6 Chat platforms (Slack, Mattermost)

- **Slack:** SVG files upload but are **not previewed** (generic file icon); raster images preview normally ([community/Custodian #5214 — "Slack does not support SVG images"](https://github.com/cloud-custodian/cloud-custodian/issues/5214); [Slack help — file types](https://slack.com/help/articles/201330736-Add-files-to-Slack)). Image blocks require a **publicly hosted URL** or a `slack_file` reference to an uploaded file ([image block reference](https://docs.slack.dev/reference/block-kit/blocks/image-block); [Slack blog — private files in image blocks](https://slack.com/blog/developers/uploading-private-images-blockkit)).
- **Mattermost:** SVG is not rendered as an inline image preview; raster formats are ([mattermost-mobile #1481 — inline SVG not supported](https://github.com/mattermost/mattermost-mobile/issues/1481)); SVG upload is permitted but treated as a download. The security posture is deliberate ([CVE-2023-1776 — stored XSS via SVG in Boards, fixed 7.1.6/7.7.2](https://nvd.nist.gov/vuln/detail/cve-2023-1776)).
- **Verdict: both are PNG-only for visible output** (gated on #26). File upload APIs are simple (Slack `files.getUploadURLExternal` flow; Mattermost `POST /api/v4/files` with a PAT).

### 1.7 Notion — SVG is explicitly supported

The Notion API's supported-types table for file uploads lists **`.svg` / `image/svg+xml` as a first-class Image type** ([Notion — Working with files and media](https://developers.notion.com/guides/data-apis/working-with-files-and-media)). Upload flow: create file upload → send (multipart, ≤20 MB single-part) → attach via a `file_upload` file object in an `image` block ([File upload object](https://developers.notion.com/reference/file-upload)). Limits: 5 MiB/file on free workspaces, 5 GiB on paid — diagrams are KB–MB scale, irrelevant. Auth is a plain integration token; the target page must be shared with the integration. **Notion is the only API-publish target where SVG survives end-to-end.** (Minor unverified: Notion's *client-side* rendering fidelity of SVG image blocks — spike check, low risk.)

### 1.8 Azure DevOps

Wiki SVG embedding is **unreliable/problematic** — open Developer Community reports of referenced SVGs failing in wiki ([SVG images don't work in ADO Wiki](https://developercommunity.visualstudio.com/t/referenced-svg-images-dont-work-in-azure-devops-wi/619280); [SVG Embedding in ADO Wiki](https://developercommunity.azure.com/t/SVG-Embedding-in-Azure-DevOps-Wiki/10979800)); Microsoft's own guidance steers diagrams to **native Mermaid** in wiki ([markdown guidance](https://learn.microsoft.com/en-us/azure/devops/project/wiki/markdown-guidance?view=azure-devops)). Work-item attachments: 100 files × 60 MB each ([manage attachments](https://learn.microsoft.com/en-us/azure/devops/boards/work-items/manage-attachments?view=azure-devops)) — but that is plain attachment storage, not rendering.

### 1.9 The `-f html` format

`-f html` (self-contained, JS nav shim) is safe and useful **only** where the HTML file is downloaded or hosted as its own document: CI artifacts, release assets, Pages, local files. It must never be pasted into a sanitizing editor (Confluence storage, Notion blocks, Jira ADF) — the sanitizer will strip the script and leave a broken fragment. Upload targets take SVG/PNG only; HTML stays in the CI-artifact lane.

---

## 2. Per-tool dossiers

Effort: **S** ≤ a few days, one-time + light upkeep; **M** = a focused milestone item; **L** = ongoing product surface. "Value" names the concrete beneficiary.

### 2.1 GitHub — Action + Pages recipe

- **What the plugin is:** (a) a **composite Action** (`runs.using: composite`) that downloads the pinned release binary for the runner OS/arch, runs `c4drill` over a glob of `.toml` models, and uploads artifacts / commits rendered SVGs back / writes a job summary; (b) a documented **Pages workflow recipe** publishing `-f html` + SVGs. Not a Docker action by default — composite runs natively on macOS/Windows runners too, where no Docker exists.
- **Effort:** **S** (action.yml + one workflow; the binary pipeline already exists in `release.yml`).
- **Value:** the largest SCM population; author-time feedback on PRs (renders valid, artifact inspectable) and zero-effort doc sites (Pages).
- **Auth/secret model:** `GITHUB_TOKEN` only, scoped `contents: write` for commit-back; **no user secrets at all** for the artifacts/Pages path. Pages via the standard `actions/deploy-pages` OIDC flow.
- **Artifact fit:** SVG commits render in the repo UI but links are dead *there* (§1.2); links live in artifacts, release assets, and Pages direct navigation. `-f html` artifact is the drill-down experience.
- **Distribution:** GitHub Marketplace (public repo + version tag; no review gate — [publishing actions](https://docs.github.com/en/actions/creating-actions/publishing-actions-in-github-marketplace)). Users pin by tag or SHA.
- **Maintenance:** low — release-download-by-tag keeps the Action decoupled from binary internals.

### 2.2 GitLab — CI template + registry image

- **What the plugin is:** a public repo shipping (a) a reusable `.gitlab-ci.yml` job consumed via `include: project: 'djarvur/c4drill'` (or listed in the **CI/CD Catalog** — [docs](https://docs.gitlab.com/ci/ci_catalog/)), and (b) the shared container image (§4) for `image:`. Optional Pages recipe mirroring GitHub's.
- **Effort:** **S** — the template is ~20 lines once the image exists.
- **Value:** MR review workflows. Capped by §1.3: SVG diffs render as **text**, so the review surface is artifacts + Pages, or PNG diff once #26 lands (the one genuinely GitLab-motivating use of PNG).
- **Auth/secret model:** `CI_JOB_TOKEN` for registry pull; none for render. Commit-back or Pages deploys use a project token — same story as GitHub.
- **Artifact fit:** same as GitHub. Public project templates avoid the authenticated-raw-URL problem when referencing images from MRs of *other* repos.
- **Distribution:** public repo + `include:project`; Catalog listing is the discoverable channel.
- **Maintenance:** low (image tag bump).

### 2.3 Bitbucket — Pipelines step / Pipe

- **What the plugin is:** (a) a documented Pipelines step using the shared image (`image: djarvur/c4drill:latest` + `artifacts:`), and (b) optionally a formal **Bitbucket Pipe** — a Docker image plus a small `pipe.yml` descriptor listed in the Pipes directory ([Pipes docs](https://support.atlassian.com/bitbucket-cloud/docs/pipes/)). The pipe is a thin wrapper over the same image.
- **Effort:** **S** for the pipeline example; **S–M** for a versioned pipe (descriptor + semver discipline).
- **Value:** smaller than GitHub/GitLab (population skews private/enterprise), but marginal cost is near zero once the image exists — this is the third wrapper on one asset.
- **Auth/secret model:** none for render; repository variables for anything publish-like.
- **Artifact fit:** **unverified — spike item.** How Bitbucket Cloud PRs display committed SVGs and whether `![](x.svg)` renders in PR/README Markdown was not verifiable from official docs in this study. The pipeline artifact path (download, full fidelity) is safe regardless. Plan PNG fallback behind #26.
- **Maintenance:** low; a pipe adds semver/tag upkeep.

### 2.4 Azure DevOps — pipeline template

- **What the plugin is:** a YAML template + the shared image via the `containers:` resource — the same pattern as GitLab, one more wrapper.
- **Effort:** **S** (deferred for *demand* reasons, not difficulty).
- **Value:** weakest of the CI trio: SVG in wiki is unreliable (§1.8) and Microsoft steers diagrams to Mermaid; the render-in-CI value exists but the display surfaces are poor. Work-item attachment publishing would duplicate the Jira publisher for a smaller audience.
- **Auth/secret model:** `System.AccessToken`/service connections — standard.
- **Artifact fit:** PNG-leaning given wiki behavior.
- **Maintenance:** low.
- **Verdict: defer** — build on evidence of demand (issue or user signal) rather than speculatively.

### 2.5 Confluence — `c4drill publish confluence`

- **What the plugin is:** a CLI subcommand: upload rendered diagrams as page attachments (`POST /wiki/rest/api/content/{id}/child/attachment`, version-tolerant update on re-run) and embed them in page storage format via `<ac:image><ri:attachment/></ac:image>` ([storage format](https://confluence.atlassian.com/doc/confluence-storage-format-790796544.html)). Likely flags: `--space --page --ancestor`, idempotent re-upload, generated-marker comment so it can find/update its own attachments.
- **Effort:** **M** — REST is straightforward; the cost is auth variety and attachment lifecycle (update vs duplicate, page lookup by title, error UX).
- **Value:** architects and reviewers living in Confluence get diagrams *inside the doc*, auto-refreshed from the model. Unblocked by the 2024 SVG fix (§1.4); PNG option (#26) as parity if inline link behavior disappoints.
- **Auth/secret model:** API token (email + token) for user-owned automation, or OAuth 2.0 (3LO) for org deployments. Site-scoped secret; no callback infra needed for token mode.
- **Artifact fit:** SVG now displays inline; scripts stripped (c4drill emits none); inline link clickability unverified (§1.4 spike). **`-f html` must be rejected by the publisher.**
- **Distribution:** CLI subcommand in the normal release binaries — no marketplace.
- **Maintenance:** medium — REST surface stability is good; sanitization policy could tighten again (Atlassian proven willing, §1.4).

### 2.6 Jira — `c4drill publish jira`

- **What the plugin is:** attach rendered diagrams to issues and embed them inline in a comment/description via ADF `mediaSingle` (§1.5). Plus a free rider: the dev-panel linkage (issue ↔ PR) comes for free once teams adopt the GitHub Action — no Jira work needed.
- **Effort:** **S–M**, **hard-gated on #26 (PNG)** — without raster output the command produces invisible attachments (§1.5).
- **Value:** lower than Confluence (diagrams are docs-shaped, not ticket-shaped); worth building only after #26 and after `publish confluence` proves the Atlassian REST/auth layer — Jira reuses it.
- **Auth/secret model:** same API-token/OAuth family as Confluence.
- **Distribution:** CLI subcommand.
- **Maintenance:** medium (ADF shapes are verbose; API v3 stable).

### 2.7 Slack — `c4drill publish slack`

- **What the plugin is:** post rendered diagrams (PNG, §1.6) to a channel: file upload via `files.getUploadURLExternal`/complete, then `chat.postMessage` with `files:` — or a plain incoming-webhook mode for zero-setup use.
- **Effort:** **S** (PNG-gated on #26).
- **Value:** CI-completion notifications with the diagram visible in-thread; useful for pipelines that rebuild a model on change.
- **Auth/secret model:** bot token (`chat:write`, `files:write`) or webhook URL — both plain secrets.
- **Artifact fit:** PNG only for preview; SVG uploads as inert file. Unfurls **out of scope** (§5).
- **Distribution:** CLI subcommand.
- **Maintenance:** low.

### 2.8 Mattermost — `c4drill publish mattermost`

- **What the plugin is:** same as Slack: `POST /api/v4/files` + post with `file_ids`, PAT or bot token ([Mattermost API](https://docs.mattermost.com/deployment-guide/server/image-proxy.html) docs family).
- **Effort:** **S** (PNG-gated).
- **Value:** Slack's story for the self-hosted population — Mattermost is common where Confluence/Bitbucket already are (Atlassian-adjacent enterprises); self-hosted means no marketplace gatekeeping.
- **Auth/secret model:** PAT; site URL + token.
- **Artifact fit:** PNG only (§1.6).
- **Maintenance:** low; API v4 is stable and self-hosted versions lag harmlessly.

### 2.9 Notion — `c4drill publish notion`

- **What the plugin is:** upload rendered SVGs via the file-upload API and insert/replace `image` blocks on a target page ([working with files and media](https://developers.notion.com/guides/data-apis/working-with-files-and-media)).
- **Effort:** **M** — three-step upload dance, block insertion/replacement logic, page discovery.
- **Value:** Notion is where many startups keep architecture docs; **SVG support end-to-end (§1.7) makes this the cleanest publish target** — vector output, no #26 dependency.
- **Auth/secret model:** internal integration token; user must share the page with the integration (a docs step, not code).
- **Artifact fit:** SVG survives upload; size limits irrelevant.
- **Distribution:** CLI subcommand.
- **Maintenance:** medium (API versioning is active; behavior changes are announced).

---

## 3. Cross-cutting question 1: one shared container image, or bespoke artifacts?

**Verdict: one OCI image + thin per-CI wrappers covers all four CI targets. No bespoke artifacts needed.** Evidence:

- All four CI systems accept `image:`/container-job references to any registry-hosted OCI image: GitHub (container jobs — and composite actions that skip Docker entirely), GitLab (`image:`), Bitbucket Pipelines (`image:`; Pipes are literally docker images + metadata — [Pipes](https://support.atlassian.com/bitbucket-cloud/docs/pipes/)), ADO (`containers:` resource).
- Because c4drill is `CGO_ENABLED=0` with no native deps (WASM graphviz), the image is `FROM scratch` + one ~10–20 MB binary — no distro, no graphviz packages, nothing to CVE-scan or rot. Building it is one extra job in the existing `release.yml` matrix (linux/amd64 + linux/arm64 in a buildx manifest; amd64 alone suffices for all managed runners, arm64 covers self-hosted).
- The wrappers are genuinely thin: GitHub composite (~50 lines YAML, downloads release binary — no image pull, works on macOS/Windows runners), GitLab template (~20 lines, `include:project`), Bitbucket step (~10 lines), ADO template (~15 lines). Each pins a version tag.
- **One nuance:** the GitHub Action should be *composite-by-default* (binary download) rather than Docker, for runner breadth; the image serves GitLab/Bitbucket/ADO and anyone wanting one artifact.

Cost of the shared strategy: one image-tag convention + one "how to pin" doc. No platform needs c4drill to ship anything bespoke.

## 4. Cross-cutting question 2: is a hosted diagram viewer required anywhere?

**Verdict: no. The no-hosting boundary holds for every target in scope.** Confirmed per feature:

| Feature that "needs" hosting | Why it's hosting-bound | Verdict |
|---|---|---|
| Slack unfurls | Unfurling requires a fully-qualified public `https` URL on a domain **registered with the app**; Slack's crawler fetches it ([unfurling docs](https://docs.slack.dev/messaging/unfurling-links-in-messages)) | Out of scope. Slack value is delivered by file upload instead (§2.7). |
| Atlassian Smart Links | Rich previews render from public URL metadata fetched by Atlassian; no public c4drill-hosted page → no rich preview; attachment paths unaffected | Out of scope. Confluence/Jira value delivered by attachments/embeds. |
| Forge/Connect Confluence macro | A hosted app by definition: app lifecycle, storage, auth flows, Atlassian reviews — an L-sized ongoing product | **Skip.** The 2024 native SVG support (§1.4) removed the main reason a macro existed (inline display without sanitization fears). |
| "Diagram viewer/registry" service | Explicitly out of scope in #28 | Skip — and unnecessary: **`-f html` *is* the viewer, distributed as a file** (artifacts, release assets, Pages). Pages substitutes for hosting where a URL is wanted. |

Everything valuable in §2 — CI rendering, artifact delivery, commit-back, Pages, and all four `publish` commands — operates with zero c4drill-operated infrastructure.

## 5. Recommendation matrix

| Tool | Verdict | Effort | Wave | Rationale (one line) |
|---|---|---|---|---|
| Shared container image | **Build** | S | 1 | One `FROM scratch` asset from the existing release matrix; prerequisite for every CI wrapper. |
| GitHub (Action + Pages recipe) | **Build** | S | 1 | Biggest audience, lowest cost (GITHUB_TOKEN-only, composite action), de-risks the whole CI family via the spike. |
| GitLab (CI template + Catalog) | **Build** | S | 1 | Third wrapper on the same asset; PNG diff (#26) unlocks MR-visual review later. |
| Bitbucket (Pipeline step / Pipe) | **Build** | S–M | 1 | Fourth wrapper; Pipe form only if the pipeline example proves demand. |
| Confluence (`publish confluence`) | **Build** | M | 2 | SVG display fixed 2024 (§1.4); highest docs-workflow value; first publish target to prove the REST/auth layer. |
| Slack (`publish slack`) | **Build** | S | 2 | Cheapest publish command; PNG-gated on #26. |
| Mattermost (`publish mattermost`) | **Build** | S | 2 | Same code shape as Slack; self-hosted population overlaps Atlassian shops. |
| Notion (`publish notion`) | **Build** | M | 2 | Only publish target with end-to-end SVG (§1.7); no #26 dependency. |
| Jira (`publish jira`) | **Build** (gated on #26) | S–M | 2/3 | Invisible without PNG (§1.5); value below Confluence; reuses Atlassian layer. Sequence after Confluence + #26. |
| Azure DevOps (template) | **Defer** | S | — | Effort-trivial but display surfaces are poor (§1.8) and Mermaid is the platform's answer; build on demand signal. |
| GitLab/Bitbucket marketplace *apps* (beyond template/pipe) | **Defer** | M–L | — | No product surface beyond what the template/image already gives. |
| Confluence Forge/Connect macro | **Skip** | L | — | Hosted app with ongoing lifecycle; native SVG support removed its raison d'être. |
| Smart Links / Slack unfurls / hosted viewer-registry | **Skip** | L | — | All require c4drill-operated public infrastructure (§4); `-f html` + Pages already fill the role. |

**Ordering rationale.** Wave 1 is one asset and three-to-four YAML wrappers sharing a single release pipeline — the highest value-per-line in the whole plan, and the GitHub spike (below) validates the family before any wrapper is polished. Wave 2 is the `publish` family, sequenced so each command inherits the previous one's solved problems: Confluence proves Atlassian REST + auth + attachment lifecycle; Slack/Mattermost prove file-upload APIs (and are trivial); Notion is independent (token-only) but benefits from the publisher scaffolding; Jira is last because it is double-gated (#26 PNG + Atlassian layer) with the least doc-shaped value. Hosted apps are skipped outright, not deferred, because their cost is structural (hosting + lifecycle), not sequencing.

**Dependency notes.** `publish slack|mattermost|jira` are gated on #26 (PNG). #25 (PlantUML) and #26 do not block Wave 1. The VS Code extension (#27) is orthogonal and untouched, per #28's scope.

## 6. Recommended spike (before finalizing Wave 1)

A proof-of-concept composite GitHub Action in this repo, run on PRs touching `examples/`:

1. Render every `examples/*.toml` with a pinned release binary (or `go run ./cmd/c4drill` for bootstrap simplicity).
2. Upload SVG + `-f html` as workflow artifacts; write a job-summary table linking to them.
3. (Optional stretch) commit-back rendered SVGs to a branch and post a comment with repo-embedded images — deliberately exposing the §1.2 link-death so the UX verdict is empirical.

This de-risks exactly the claims the dossiers lean on: artifact UX on PRs, version pinning, and the observable GitHub rendering behavior. Per #28, the spike is the only permitted implementation, and follow-up issues get filed per selected target.

## 7. Summary

- **Build (Wave 1):** shared multi-arch container image + GitHub Action + GitLab CI template + Bitbucket pipeline (Pipe optional). One asset, thin wrappers, no secrets, no hosting.
- **Build (Wave 2):** `c4drill publish confluence` → `publish slack` / `publish mattermost` → `publish notion` → `publish jira` (last, PNG-gated on #26).
- **Defer:** Azure DevOps template; deeper marketplace-app forms for GitLab/Bitbucket.
- **Skip:** Forge/Connect macro, Smart Links, Slack unfurls, any hosted viewer — the no-hosting boundary is confirmed by the platforms' own documentation, and `-f html`-as-artifact + Pages covers the "someone needs to view this" job.
- **Corrected assumptions vs #28:** Confluence Cloud *does* display SVG attachments since 2024-09 (CONFCLOUD-1762 fixed); GitLab MR diffs render SVG as *text*, making PNG-diff (#26) the GitLab review unlock; Jira's SVG ban makes it the only publish target that cannot ship before #26; Notion, conversely, accepts SVG natively.

---

## Sources

### GitHub
- [Community discussion #151372 — why GitHub Markdown disallows SVG embedding](https://github.com/orgs/community/discussions/151372)
- [isaacs/github #316 — SVG in READMEs by design](https://github.com/isaacs/github/issues/316)
- [alexwlchan — SVGs only render on GitHub via `<img>`](https://alexwlchan.net/notes/2024/how-to-render-svgs-on-github/)
- [SO — GitHub Markdown SVG file links not working](https://stackoverflow.com/questions/70545385/github-markdown-svg-file-links-not-working)
- [Shields.io discussion #5593 — links in SVG stripped by camo/img](https://github.com/badges/shields/discussions/5593)
- [Community discussion #59781 — release assets served as octet-stream; Pages MIME contrast](https://github.com/orgs/community/discussions/59781)
- [SO #13808020 — embedding SVG hosted on GitHub](https://stackoverflow.com/questions/13808020/include-an-svg-hosted-on-github-in-markdown)
- [GitHub Docs — publishing actions in GitHub Marketplace](https://docs.github.com/en/actions/creating-actions/publishing-actions-in-github-marketplace)

### GitLab
- [gitlab #15284 — "Render SVG files as images in MR diffs" (state `opened`, verified 2026-09 via GitLab API)](https://gitlab.com/gitlab-org/gitlab/-/issues/15284)
- [gitlab-ce #27471 — SVG XSS rationale](https://gitlab.com/gitlab-org/gitlab-ce/-/issues/27471)
- [gitlab-ce MR !3401 — Loofah-based SVG sanitizer](https://gitlab.com/gitlab-org/gitlab-ce/-/merge_requests/3401)
- [GitLab security docs — user file uploads](https://docs.gitlab.com/security/user_file_uploads/)
- [GitLab docs — GLFM](https://docs.gitlab.com/user/markdown/); [MR changes](https://docs.gitlab.com/user/project/merge_requests/changes/); [CI/CD Catalog](https://docs.gitlab.com/ci/ci_catalog/)
- [gitlab #26104 — SVG from repo into README.md](https://gitlab.com/gitlab-org/gitlab/-/issues/26104)

### Confluence
- [CONFCLOUD-1762 — Embedding of SVG (closed Fixed 2024-09-10)](https://jira.atlassian.com/browse/CONFCLOUD-1762)
- [Community — Confluence strips script content from uploaded SVGs](https://community.atlassian.com/forums/Confluence-questions/Confluence-does-not-retain-use-uploaded-SVG-images/qaq-p/1418344)
- [CONFCLOUD-25488 — XSS via SVG attachment (history)](https://jira.atlassian.com/browse/CONFCLOUD-25488)
- [Confluence Storage Format — `ac:image`, `ri:attachment`, `ri:url`](https://confluence.atlassian.com/doc/confluence-storage-format-790796544.html)
- [Confluence REST API — attachments](https://developer.atlassian.com/cloud/confluence/rest/v1/intro/)

### Jira
- [JRACLOUD-47728 — Jira doesn't support viewing SVG images](https://jira.atlassian.com/browse/JRACLOUD-47728)
- [Community — SVG thumbnails show generic icon](https://community.atlassian.com/forums/Jira-questions/Can-you-show-thumbnails-for-SVG-attachments-in-Jira-Server/qaq-p/1131913)
- [Community — inline images via ADF media nodes (REST v3)](https://community.atlassian.com/forums/Jira-questions/How-to-inline-images-in-Issue-body-or-issue-comment-body/qaq-p/1658456)
- [Atlassian KB — inline image display in wiki renderer fields](https://support.atlassian.com/jira/kb/image-attachments-are-not-displayed-inline-in-wiki-renderer-fields/)

### Slack
- [Slack docs — unfurling links (registered public https domains)](https://docs.slack.dev/messaging/unfurling-links-in-messages)
- [Slack docs — image block (public URL / slack_file)](https://docs.slack.dev/reference/block-kit/blocks/image-block)
- [Slack blog — private files in image blocks](https://slack.com/blog/developers/uploading-private-images-blockkit)
- [cloud-custodian #5214 — Slack does not support SVG images](https://github.com/cloud-custodian/cloud-custodian/issues/5214)
- [Slack help — file type restrictions](https://slack.com/help/articles/201330736-Add-files-to-Slack)

### Mattermost
- [mattermost-mobile #1481 — inline SVG not supported](https://github.com/mattermost/mattermost-mobile/issues/1481)
- [CVE-2023-1776 — stored XSS via SVG (sanitization posture)](https://nvd.nist.gov/vuln/detail/cve-2023-1776)
- [Mattermost docs — image proxy / external image handling](https://docs.mattermost.com/deployment-guide/server/image-proxy.html)

### Notion
- [Notion docs — Working with files and media (supported types incl. `image/svg+xml`)](https://developers.notion.com/guides/data-apis/working-with-files-and-media)
- [Notion docs — File upload object](https://developers.notion.com/reference/file-upload)

### Azure DevOps
- [Microsoft Learn — Markdown syntax for ADO](https://learn.microsoft.com/en-us/azure/devops/project/wiki/markdown-guidance?view=azure-devops)
- [Microsoft Learn — manage work item attachments (100 × 60 MB)](https://learn.microsoft.com/en-us/azure/devops/boards/work-items/manage-attachments?view=azure-devops)
- [Developer Community — SVG images don't work in ADO wiki](https://developercommunity.visualstudio.com/t/referenced-svg-images-dont-work-in-azure-devops-wi/619280)
- [Developer Community — SVG embedding in ADO wiki](https://developercommunity.azure.com/t/SVG-Embedding-in-Azure-DevOps-Wiki/10979800)

### Bitbucket
- [Atlassian — Bitbucket Pipes (docker image + descriptor)](https://support.atlassian.com/bitbucket-cloud/docs/pipes/)

### Browser/SVG fundamentals
- [MDN — SVG overview (img = static image mode)](https://developer.mozilla.org/en-US/docs/Web/SVG)
- [SO — `<img>` vs `<object>` for SVG](https://stackoverflow.com/questions/4476526/do-i-use-img-object-or-embed-for-svg-files)
- [jgraph/drawio #935 — links work opened directly, not via `<img>`](https://github.com/jgraph/drawio/issues/935)
