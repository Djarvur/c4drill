package dev.djarvur.c4drill.core

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

/** Wire-parameter construction for c4drill/renderDiagram (mirrors renderParams.test.ts from #27). */
class RenderParamsTest {
    @Test
    fun `initial state is the c1 context view with model defaults`() {
        val params = buildRenderParams("file:///m.toml", initialViewState)

        assertEquals("file:///m.toml", params.textDocumentUri)
        assertEquals("", params.target)
        assertNull(params.allExpanded)
        assertNull(params.expanded) // omitted = model default
        assertNull(params.legend)
        assertEquals("svg", params.format)
    }

    @Test
    fun `all expanded overrides the target`() {
        val params = buildRenderParams(
            "file:///m.toml",
            PreviewViewState(target = "cloud.ui", allExpanded = true),
        )

        assertTrue(params.allExpanded!!)
        assertNull(params.target)
    }

    @Test
    fun `expanded text maps to an explicit expanded set`() {
        val params = buildRenderParams("u", PreviewViewState(expandedText = "cloud, db.io"))

        assertEquals(listOf("cloud", "db.io"), params.expanded)
    }

    @Test
    fun `blank expanded text keeps the model default`() {
        assertNull(buildRenderParams("u", initialViewState).expanded)
        assertNull(buildRenderParams("u", PreviewViewState(expandedText = "  ")).expanded)
        assertNull(buildRenderParams("u", PreviewViewState(expandedText = ",, ")).expanded)
    }

    @Test
    fun `collapse all sends the empty array override`() {
        val params = buildRenderParams("u", PreviewViewState(collapseAll = true, expandedText = "ignored"))

        assertEquals(emptyList<String>(), params.expanded)
    }

    @Test
    fun `legend overrides map to booleans`() {
        assertTrue(buildRenderParams("u", PreviewViewState(legend = LegendChoice.ON)).legend!!)
        assertFalse(buildRenderParams("u", PreviewViewState(legend = LegendChoice.OFF)).legend!!)
        assertNull(buildRenderParams("u", PreviewViewState(legend = LegendChoice.DEFAULT)).legend)
    }

    @Test
    fun `expanded set parsing splits on commas and whitespace`() {
        assertEquals(listOf("a", "b", "c"), parseExpandedSet("a, b\n\tc"))
        assertEquals(listOf("single"), parseExpandedSet("single"))
        assertNull(parseExpandedSet(""))
    }

    @Test
    fun `breadcrumb trail grows one entry per dotted segment`() {
        val trail = breadcrumbTrail(PreviewViewState(target = "cloud.ui"))

        assertEquals(listOf("C1", "cloud", "ui"), trail.map { it.label })
        assertEquals(listOf("", "cloud", "cloud.ui"), trail.map { it.target })
        assertFalse(trail[0].current)
        assertFalse(trail[1].current)
        assertTrue(trail[2].current)
    }

    @Test
    fun `breadcrumb trail for c1 and all-expanded`() {
        val c1 = breadcrumbTrail(initialViewState)

        assertEquals(1, c1.size)
        assertTrue(c1[0].current)

        val expanded = breadcrumbTrail(PreviewViewState(allExpanded = true))

        assertEquals(listOf("C1", "expanded"), expanded.map { it.label })
        assertTrue(expanded[1].current)
    }
}
