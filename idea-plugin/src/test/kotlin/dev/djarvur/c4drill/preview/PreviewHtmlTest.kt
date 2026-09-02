package dev.djarvur.c4drill.preview

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class PreviewHtmlTest {
    @Test
    fun `diagram html embeds the svg and the bridge function name`() {
        val html = PreviewHtml.diagramHtml("<svg><a href=\"cloud-system/cloud.svg\"><rect/></a></svg>", "javaBridge_1")

        assertTrue(html.contains("<svg>"))
        assertTrue(html.contains("cloud-system/cloud.svg"))
        assertTrue(html.contains("window['javaBridge_1']"))
        assertFalse(html.contains(PreviewHtml.BRIDGE_PLACEHOLDER))
        assertTrue(html.contains("preventDefault"))
    }

    @Test
    fun `status html escapes messages`() {
        val html = PreviewHtml.statusHtml("Model has errors", listOf("parse: <unexpected> \"token\""))

        assertTrue(html.contains("parse: &lt;unexpected&gt; &quot;token&quot;"))
        assertFalse(html.contains("<unexpected>"))
    }

    @Test
    fun `status html marks error lines`() {
        val html = PreviewHtml.statusHtml("Model has errors", listOf("Error: unit x", "note"))

        assertTrue(html.contains("class=\"line error\">Error: unit x</div>"))
        assertTrue(html.contains("class=\"line\">note</div>"))
    }

    @Test
    fun `escape handles ampersands first`() {
        assertEquals("&amp;lt;", PreviewHtml.escape("&lt;"))
    }
}
