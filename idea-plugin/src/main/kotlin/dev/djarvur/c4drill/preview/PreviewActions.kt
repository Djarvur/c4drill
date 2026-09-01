// Preview toolbar actions: all-expanded mode, legend override, expanded-set
// editing, collapse-all, and SVG export. They mutate the shared
// PreviewViewStateHolder and re-render through the provided callback.

package dev.djarvur.c4drill.preview

import com.intellij.openapi.actionSystem.ActionUpdateThread
import com.intellij.openapi.actionSystem.AnAction
import com.intellij.openapi.actionSystem.AnActionEvent
import com.intellij.openapi.actionSystem.DefaultActionGroup
import com.intellij.openapi.actionSystem.ToggleAction
import com.intellij.openapi.project.Project
import com.intellij.openapi.ui.DialogWrapper
import dev.djarvur.c4drill.core.LegendChoice
import javax.swing.JComponent
import javax.swing.JTextField

object PreviewActions {
    fun addAll(group: DefaultActionGroup, state: PreviewViewStateHolder, onRender: () -> Unit, onExport: () -> Boolean) {
        group.add(ToggleAllExpandedAction(state, onRender))
        group.addSeparator()
        group.add(legendGroup(state, onRender))
        group.add(SetExpandedSetAction(state, onRender))
        group.add(ToggleCollapseAllAction(state, onRender))
        group.addSeparator()
        group.add(ExportSvgAction(onExport))
    }

    private fun legendGroup(state: PreviewViewStateHolder, onRender: () -> Unit): DefaultActionGroup {
        val group = DefaultActionGroup("C4DrillLegend", true)

        group.add(LegendChoiceAction("Legend: Model Default", LegendChoice.DEFAULT, state, onRender))
        group.add(LegendChoiceAction("Legend: On", LegendChoice.ON, state, onRender))
        group.add(LegendChoiceAction("Legend: Off", LegendChoice.OFF, state, onRender))

        return group
    }
}

private class ToggleAllExpandedAction(
    private val state: PreviewViewStateHolder,
    private val onRender: () -> Unit,
) : ToggleAction("All Expanded", "Render the single all-nested diagram (the CLI --expanded mode)", null) {
    override fun isSelected(e: AnActionEvent): Boolean = state.current().allExpanded

    override fun setSelected(e: AnActionEvent, sel: Boolean) {
        state.setAllExpanded(sel)
        onRender()
    }
}

private class LegendChoiceAction(
    private val label: String,
    private val choice: LegendChoice,
    private val state: PreviewViewStateHolder,
    private val onRender: () -> Unit,
) : ToggleAction(label) {
    override fun isSelected(e: AnActionEvent): Boolean = state.current().legend == choice

    override fun setSelected(e: AnActionEvent, sel: Boolean) {
        if (sel) {
            state.setLegend(choice)
            onRender()
        }
    }
}

private class ToggleCollapseAllAction(
    private val state: PreviewViewStateHolder,
    private val onRender: () -> Unit,
) : ToggleAction("Collapse All", "Send an empty expanded set — overrides [properties].expanded with \"collapse everything\"", null) {
    override fun isSelected(e: AnActionEvent): Boolean = state.current().collapseAll

    override fun setSelected(e: AnActionEvent, sel: Boolean) {
        state.setCollapseAll(sel)
        onRender()
    }
}

private class SetExpandedSetAction(
    private val state: PreviewViewStateHolder,
    private val onRender: () -> Unit,
) : AnAction("Expanded Set...", "Comma-separated unit paths replacing [properties].expanded; blank = model default", null) {
    override fun getActionUpdateThread(): ActionUpdateThread = ActionUpdateThread.BGT

    override fun actionPerformed(e: AnActionEvent) {
        val project: Project = e.project ?: return
        val field = JTextField(state.current().expandedText, 32)

        val dialog = object : DialogWrapper(project, true) {
            init {
                title = "Expanded Set"
                init()
            }

            override fun createCenterPanel(): JComponent = field
        }

        if (dialog.showAndGet()) {
            state.setExpandedText(field.text)
            onRender()
        }
    }
}

private class ExportSvgAction(
    private val onExport: () -> Boolean,
) : AnAction("Export SVG...", "Save the currently rendered diagram as an .svg file", null) {
    override fun getActionUpdateThread(): ActionUpdateThread = ActionUpdateThread.BGT

    override fun update(e: AnActionEvent) {
        e.presentation.isEnabled = C4drillPreviewService.getInstance(e.project ?: return).lastRender != null
    }

    override fun actionPerformed(e: AnActionEvent) {
        onExport()
    }
}
