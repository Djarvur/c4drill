// HTML shell for the JCEF preview panel (issue #29). The server's SVG is
// embedded inline; a small script intercepts clicks on the SVG's internal
// drill-down links and forwards the href to the IDE bridge (a JBCefJSQuery
// named __BRIDGE_NAME__ in the document). Pure string work — unit-tested
// without the platform.

package dev.djarvur.c4drill.preview

object PreviewHtml {
    const val BRIDGE_PLACEHOLDER: String = "__BRIDGE_NAME__"

    /** statusHtml renders the CLI-identical error/state panel shown instead of a stale diagram. */
    fun statusHtml(title: String, lines: List<String>): String {
        val body = lines.joinToString("\n") { line ->
            val html = escape(line)

            if (html.startsWith("Error:") || html.contains("error", ignoreCase = true)) {
                "<div class=\"line error\">$html</div>"
            } else {
                "<div class=\"line\">$html</div>"
            }
        }

        return """
<!doctype html>
<html>
<head>
<meta charset="utf-8">
<style>
  :host, html, body { margin: 0; padding: 0; background: #ffffff; }
  body { font-family: system-ui, sans-serif; color: #1f2328; padding: 24px; }
  h2 { font-size: 15px; margin: 0 0 12px; }
  .line { font-family: ui-monospace, monospace; font-size: 12px; padding: 2px 0; }
  .error { color: #cf222e; }
</style>
</head>
<body>
<h2>${escape(title)}</h2>
$body
</body>
</html>
""".trimIndent()
    }

    /**
     * diagramHtml wraps the rendered SVG with the click-interception script.
     * [bridgeName] is the JBCefJSQuery function name; every click on an
     * anchor is preventDefault()-ed and reported to the IDE as the raw href
     * value via the bridge call.
     */
    fun diagramHtml(svg: String, bridgeName: String): String {
        require(BRIDGE_PLACEHOLDER !in bridgeName)

        val script = """
(function() {
  function report(href) {
    window['$bridgeName'] && window['$bridgeName'](href);
  }
  document.addEventListener('click', function(e) {
    var t = e.target;
    while (t && t.tagName !== 'A' && t !== document.body) t = t.parentNode;
    if (!t || t.tagName !== 'A') return;
    var href = t.getAttribute('href');
    if (!href) return;
    e.preventDefault();
    report(href);
  }, true);
})();
""".trimIndent()

        return """
<!doctype html>
<html>
<head>
<meta charset="utf-8">
<style>
  html, body { margin: 0; padding: 12px; background: #ffffff; }
  svg { max-width: 100%; height: auto; }
  a { cursor: pointer; }
</style>
<script>
$script
</script>
</head>
<body>
$svg
</body>
</html>
""".trimIndent()
    }

    fun escape(s: String): String = s
        .replace("&", "&amp;")
        .replace("<", "&lt;")
        .replace(">", "&gt;")
        .replace("\"", "&quot;")
}
