// View state and c4drill/renderDiagram parameter construction for the live
// preview tool window (issue #29 scope update; mirrors the VS Code
// extension's src/core/renderParams.ts from #27). The wire contract is
// documented on internal/lsp RenderDiagramParams / RenderDiagramResult:
//
//	{ textDocument: { uri }, target, allExpanded, expanded, legend, format }
//
//	target       "" = C1 context, one segment = that unit's C2, deeper = C3
//	allExpanded  the single all-nested diagram (--expanded); overrides target
//	expanded     replacement for [properties].expanded C1 drill-down set;
//	             an EMPTY array is meaningful ("collapse all") — the field is
//	             deliberately sent even when empty, while "model default" is
//	             expressed by OMITTING the field (null here)
//	legend       true/false overrides properties.legend; null = default
//	format       only "svg" in v1
//
// Pure logic, no IntelliJ imports, so it is unit-testable with plain JUnit.

package dev.djarvur.c4drill.core

enum class LegendChoice { DEFAULT, ON, OFF }

data class PreviewViewState(
    /** Render target path; "" is the C1 context diagram. */
    val target: String = "",
    /** All-expanded diagram mode (the CLI --expanded mode); overrides target. */
    val allExpanded: Boolean = false,
    /** Comma/whitespace-separated unit paths replacing [properties].expanded; empty = model default. */
    val expandedText: String = "",
    /**
     * Explicit "collapse all" override: sends expanded: [] (an EMPTY array is
     * meaningful on the wire — the server treats it as "collapse everything").
     * Takes precedence over expandedText; cleared as soon as the user edits
     * the expanded-set input.
     */
    val collapseAll: Boolean = false,
    val legend: LegendChoice = LegendChoice.DEFAULT,
)

data class RenderDiagramWireParams(
    val textDocumentUri: String,
    val target: String? = null,
    val allExpanded: Boolean? = null,
    /** null = model default; non-null (possibly empty list) = explicit override. */
    val expanded: List<String>? = null,
    val legend: Boolean? = null,
    val format: String? = null,
)

const val renderDiagramFormat: String = "svg"

val initialViewState: PreviewViewState = PreviewViewState()

/** buildRenderParams maps view state + document URI to the renderDiagram request params. */
fun buildRenderParams(uri: String, state: PreviewViewState): RenderDiagramWireParams {
    var params = RenderDiagramWireParams(textDocumentUri = uri, format = renderDiagramFormat)

    params = if (state.allExpanded) {
        params.copy(allExpanded = true)
    } else {
        params.copy(target = state.target)
    }

    params = if (state.collapseAll) {
        // The wire contract: an empty array IS the "collapse all" override.
        params.copy(expanded = emptyList())
    } else {
        val expanded = parseExpandedSet(state.expandedText)

        if (expanded != null) params.copy(expanded = expanded) else params
    }

    return when (state.legend) {
        LegendChoice.ON -> params.copy(legend = true)
        LegendChoice.OFF -> params.copy(legend = false)
        LegendChoice.DEFAULT -> params
    }
}

/** parseExpandedSet splits the expanded-set input on commas and whitespace; blank input yields null (model default). */
fun parseExpandedSet(text: String): List<String>? {
    val entries = text.split(Regex("[\\s,]+"))
        .map { it.trim() }
        .filter { it.isNotEmpty() }

    return if (entries.isEmpty()) null else entries
}

data class BreadcrumbEntry(
    val label: String,
    val target: String,
    val current: Boolean,
)

/** breadcrumbTrail returns the clickable ancestor chain for the current target: the C1 root plus one entry per segment. */
fun breadcrumbTrail(state: PreviewViewState): List<BreadcrumbEntry> {
    val trail = mutableListOf(BreadcrumbEntry("C1", "", !state.allExpanded && state.target.isEmpty()))

    if (state.allExpanded) {
        trail.add(BreadcrumbEntry("expanded", "", current = true))

        return trail
    }

    if (state.target.isEmpty()) {
        return trail
    }

    val segments = state.target.split('.')

    for (i in segments.indices) {
        val t = segments.subList(0, i + 1).joinToString(".")

        trail.add(BreadcrumbEntry(segments[i], t, current = i == segments.size - 1))
    }

    return trail
}
