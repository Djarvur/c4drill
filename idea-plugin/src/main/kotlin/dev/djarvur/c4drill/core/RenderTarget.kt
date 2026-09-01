// Drill-down link resolution for the live preview (issue #29, mirrors the
// VS Code extension's src/core/renderTarget.ts from #27).
//
// The rendered SVG's internal links are RELATIVE .svg URLs computed by
// internal/graph/path.go (ComputeExploreURL / ComputeBackLinkURL), which
// mirror the CLI's on-disk output layout:
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
// Pure logic, no IntelliJ imports, so it is unit-testable with plain JUnit.

package dev.djarvur.c4drill.core

/** basenameOfUri returns the diagram basename the CLI would use: the document file name without its extension. */
fun basenameOfUri(uri: String): String {
    val withoutScheme = uri.substringAfterLast('/', missingDelimiterValue = uri)
    val file = decodeUriComponent(withoutScheme)
    val dot = file.lastIndexOf('.')

    return if (dot > 0) file.substring(0, dot) else file
}

/** virtualFilePath is the outDir-root-relative location of the current view's diagram file, mirroring the CLI output layout. */
fun virtualFilePath(target: String, basename: String): String =
    if (target.isEmpty()) {
        "$basename.svg"
    } else {
        "$basename/${target.split('.').joinToString("/")}.svg"
    }

/**
 * resolveRenderTarget maps a clicked link href (raw attribute value, possibly
 * URL-encoded, possibly relative) to the next render target ("" = C1), or
 * null when the link is not an internal drill-down (external http(s)
 * reference links, mailto:, malformed, or outside the diagram tree).
 */
fun resolveRenderTarget(currentTarget: String, basename: String, clickedHref: String): String? {
    val href = clickedHref.trim()
    if (href.isEmpty()) {
        return null
    }

    if (href.matches(Regex("^[a-zA-Z][a-zA-Z0-9+.-]*:.*"))) {
        return null // http(s)://, mailto:, etc. — the caller treats these as external
    }

    // Strip query/fragment (reference URLs may carry #anchor; internal
    // diagram links do not, but be tolerant).
    val path = href.substringBefore('#').substringBefore('?')
    if (path.isEmpty()) {
        return null
    }

    val currentFile = virtualFilePath(currentTarget, basename)
    val currentDir = if ('/' in currentFile) currentFile.substringBeforeLast('/') + "/" else ""

    // Resolve the href against the current view's directory (URL semantics),
    // decoding each segment (ComputeExploreURL PathEscapes every segment).
    val baseSegments = currentDir.split('/').filter { it.isNotEmpty() }.toMutableList()

    for (raw in path.split('/')) {
        val seg = decodeUriComponent(raw)

        if (seg.isEmpty() || seg == ".") {
            continue
        }

        if (seg == "..") {
            if (baseSegments.isEmpty()) {
                return null // escapes the diagram tree — not an internal link
            }
            baseSegments.removeAt(baseSegments.size - 1)

            continue
        }

        baseSegments.add(seg)
    }

    val resolved = baseSegments.joinToString("/")

    if (resolved == "$basename.svg") {
        return ""
    }

    val prefix = "$basename/"
    if (resolved.startsWith(prefix) && resolved.endsWith(".svg")) {
        return resolved.removePrefix(prefix).removeSuffix(".svg").split('/').joinToString(".")
    }

    return null
}

/**
 * Minimal percent-decoder for URL path segments. Unlike java.net.URLDecoder
 * it treats '+' as a literal plus (path semantics, not form semantics) and
 * leaves malformed escapes untouched instead of throwing.
 */
fun decodeUriComponent(value: String): String {
    if ('%' !in value) {
        return value
    }

    val out = StringBuilder(value.length)
    var i = 0

    while (i < value.length) {
        val c = value[i]

        if (c == '%' && i + 3 <= value.length) {
            val code = value.substring(i + 1, i + 3).toIntOrNull(16)

            if (code != null) {
                out.append(code.toChar())
                i += 3

                continue
            }
        }

        out.append(c)
        i++
    }

    return out.toString()
}
