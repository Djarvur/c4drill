// Drill-down link resolution for the live preview (issue #27).
//
// The rendered SVG's internal links are RELATIVE .svg URLs computed by
// internal/graph/path.go (ComputeExploreURL / ComputeBackLinkURL), which
// mirror the CLI's on-disk output layout (internal/output/writer.go Write):
//
//	C1 (target "")       -> {basename}.svg                 (at the root)
//	target T (dotted)    -> {basename}/{T with '.'->'/'}.svg
//
//	e.g. basename "cloud-system", targets:
//	  ""            -> cloud-system.svg
//	  "cloud"       -> cloud-system/cloud.svg
//	  "amazon.rds"  -> cloud-system/amazon/rds.svg
//
// The preview keeps the current render target, and for a clicked href
// resolves it against the directory of the current view's virtual file —
// exactly how a browser resolves the CLI output tree. External http(s)
// reference links are NOT resolved here (the caller opens them in the
// system browser).
//
// Pure logic, no VS Code imports, so it is unit-testable with node:test.

/** basenameOfUri returns the diagram basename the CLI would use: the document file name without its extension. */
export function basenameOfUri(uri: string): string {
    const withoutScheme = uri.slice(uri.lastIndexOf('/') + 1);
    const file = decodeURIComponent(withoutScheme);
    const dot = file.lastIndexOf('.');

    return dot > 0 ? file.slice(0, dot) : file;
}

// virtualFilePath is the outDir-root-relative location of the current view's
// diagram file, mirroring internal/output/writer.go's layout.
export function virtualFilePath(target: string, basename: string): string {
    if (target === '') {
        return `${basename}.svg`;
    }

    return `${basename}/${target.split('.').join('/')}.svg`;
}

// resolveRenderTarget maps a clicked link href (raw attribute value, possibly
// URL-encoded, possibly relative) to the next render target ("" = C1), or
// null when the link is not an internal drill-down (external http(s)
// reference links, mailto:, malformed, or outside the diagram tree).
export function resolveRenderTarget(currentTarget: string, basename: string, clickedHref: string): string | null {
    const href = clickedHref.trim();
    if (href === '') {
        return null;
    }

    const schemeMatch = /^([a-zA-Z][a-zA-Z0-9+.-]*):/.exec(href);
    if (schemeMatch) {
        return null; // http(s)://, mailto:, etc. — the caller treats these as external
    }

    // Strip query/fragment (reference URLs may carry #anchor; internal
    // diagram links do not, but be tolerant).
    const path = href.split('#')[0].split('?')[0];
    if (path === '') {
        return null;
    }

    const currentFile = virtualFilePath(currentTarget, basename);
    const currentDir = currentFile.includes('/')
        ? currentFile.slice(0, currentFile.lastIndexOf('/') + 1)
        : '';

    // Resolve the href against the current view's directory (URL semantics),
    // decoding each segment (ComputeExploreURL PathEscapes every segment).
    const baseSegments = currentDir.split('/').filter((s) => s !== '');
    const hrefSegments = path.split('/');

    for (const raw of hrefSegments) {
        const seg = decodeURIComponent(raw);

        if (seg === '' || seg === '.') {
            continue;
        }

        if (seg === '..') {
            if (baseSegments.length === 0) {
                return null; // escapes the diagram tree — not an internal link
            }

            baseSegments.pop();

            continue;
        }

        baseSegments.push(seg);
    }

    const resolved = baseSegments.join('/');

    if (resolved === `${basename}.svg`) {
        return '';
    }

    const prefix = `${basename}/`;
    if (resolved.startsWith(prefix) && resolved.endsWith('.svg')) {
        return resolved.slice(prefix.length, -'.svg'.length).split('/').join('.');
    }

    return null;
}
