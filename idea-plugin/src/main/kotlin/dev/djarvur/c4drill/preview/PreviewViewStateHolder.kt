// Mutable holder for the preview's view state plus a read-only snapshot for
// render calls. The pure mapping from view state to wire params lives in
// core/RenderParams (unit-tested); this holder only guards it for the UI.

package dev.djarvur.c4drill.preview

import dev.djarvur.c4drill.core.LegendChoice
import dev.djarvur.c4drill.core.PreviewViewState
import dev.djarvur.c4drill.core.initialViewState

/** Snapshot is the read-only view passed to render calls. */
data class Snapshot(
    val target: String,
    val allExpanded: Boolean,
    /** Non-null when the user overrode the expanded set (possibly empty = collapse all). */
    val expandedOverride: List<String>?,
    /** Non-null when the user forced the legend on/off. */
    val legendOverride: Boolean?,
)

class PreviewViewStateHolder {
    @Volatile
    private var state: PreviewViewState = initialViewState

    @Synchronized
    fun snapshot(): Snapshot {
        val s = state

        val expandedOverride: List<String>? = when {
            s.collapseAll -> emptyList()
            else -> dev.djarvur.c4drill.core.parseExpandedSet(s.expandedText)
        }

        val legendOverride: Boolean? = when (s.legend) {
            LegendChoice.ON -> true
            LegendChoice.OFF -> false
            LegendChoice.DEFAULT -> null
        }

        return Snapshot(
            target = s.target,
            allExpanded = s.allExpanded,
            expandedOverride = expandedOverride,
            legendOverride = legendOverride,
        )
    }

    @Synchronized
    fun current(): PreviewViewState = state

    @Synchronized
    fun update(transform: (PreviewViewState) -> PreviewViewState) {
        state = transform(state)
    }

    @Synchronized
    fun setTarget(target: String) {
        state = if (state.allExpanded) {
            state.copy(allExpanded = false, target = target)
        } else {
            state.copy(target = target)
        }
    }

    @Synchronized
    fun setAllExpanded(allExpanded: Boolean) {
        state = state.copy(allExpanded = allExpanded)
    }

    @Synchronized
    fun setLegend(legend: LegendChoice) {
        state = state.copy(legend = legend)
    }

    @Synchronized
    fun setExpandedText(text: String) {
        state = state.copy(expandedText = text, collapseAll = false)
    }

    @Synchronized
    fun setCollapseAll(collapseAll: Boolean) {
        state = state.copy(collapseAll = collapseAll)
    }

    @Synchronized
    fun reset() {
        state = initialViewState
    }
}
