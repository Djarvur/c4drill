// Registers the "C4Drill Preview" tool window (plugin.xml <toolWindow>).

package dev.djarvur.c4drill.preview

import com.intellij.openapi.project.DumbAware
import com.intellij.openapi.project.Project
import com.intellij.openapi.wm.ToolWindow
import com.intellij.openapi.wm.ToolWindowFactory

class C4drillPreviewToolWindowFactory : ToolWindowFactory, DumbAware {
    override fun createToolWindowContent(project: Project, toolWindow: ToolWindow) {
        C4drillPreviewService.getInstance(project).attachToolWindow(toolWindow)
    }

    companion object {
        const val ID: String = "C4Drill Preview"
    }
}
