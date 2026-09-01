// Editor actions (issue #29): Validate file (re-runs the server's
// CLI-identical validation), Format document (delegates to the standard
// Reformat Code action, which the platform LSP client backs with
// textDocument/formatting), Show preview, and the explicit TOML
// activate/deactivate scoping actions.

package dev.djarvur.c4drill.actions

import com.intellij.openapi.actionSystem.ActionUpdateThread
import com.intellij.openapi.actionSystem.AnAction
import com.intellij.openapi.actionSystem.AnActionEvent
import com.intellij.openapi.actionSystem.ActionManager
import com.intellij.openapi.actionSystem.CommonDataKeys
import com.intellij.openapi.actionSystem.IdeActions
import com.intellij.openapi.actionSystem.ex.ActionUtil
import com.intellij.openapi.application.ApplicationManager
import com.intellij.notification.NotificationGroupManager
import com.intellij.notification.NotificationType
import com.intellij.openapi.project.Project
import com.intellij.openapi.util.io.FileUtil
import com.intellij.openapi.vfs.VirtualFile
import dev.djarvur.c4drill.C4drillProjectSettings
import dev.djarvur.c4drill.core.NOT_HANDLED_MESSAGE
import dev.djarvur.c4drill.isManagedFile
import dev.djarvur.c4drill.lsp.C4drillGatewayLookup
import dev.djarvur.c4drill.preview.C4drillPreviewService

/** localFile returns the file under the caret when it is a local file, or null. */
private fun localFile(e: AnActionEvent): VirtualFile? {
    val file = e.getData(CommonDataKeys.VIRTUAL_FILE) ?: return null

    return if (file.isInLocalFileSystem) file else null
}

private fun notifyNotHandled(project: Project?) {
    if (project != null) {
        NotificationGroupManager.getInstance()
            .getNotificationGroup("C4Drill")
            .createNotification("C4Drill", NOT_HANDLED_MESSAGE, NotificationType.INFORMATION)
            .notify(project)
    }
}

/** C4drillValidateAction re-sends the buffer so the server reruns validation and republishes diagnostics. */
class C4drillValidateAction : AnAction() {
    override fun getActionUpdateThread(): ActionUpdateThread = ActionUpdateThread.BGT

    override fun update(e: AnActionEvent) {
        val project = e.project
        val file = localFile(e)

        e.presentation.isEnabled = project != null && file != null && isManagedFile(project, file)
    }

    override fun actionPerformed(e: AnActionEvent) {
        val project = e.project ?: return
        val file = localFile(e) ?: return

        if (!isManagedFile(project, file)) {
            notifyNotHandled(project)

            return
        }

        val gateway = C4drillGatewayLookup.gateway(project)

        gateway?.let { g ->
            ApplicationManager.getApplication().executeOnPooledThread {
                g.ensureStarted(project, file)
                g.nudgeValidation(project, file)
            }
        }
    }
}

/** C4drillFormatAction runs the standard reformat; the platform LSP client supplies the formatter. */
class C4drillFormatAction : AnAction() {
    override fun getActionUpdateThread(): ActionUpdateThread = ActionUpdateThread.BGT

    override fun update(e: AnActionEvent) {
        val project = e.project
        val editor = e.getData(CommonDataKeys.EDITOR)
        val file = localFile(e)

        e.presentation.isEnabled = project != null && editor != null && file != null && isManagedFile(project, file)
    }

    override fun actionPerformed(e: AnActionEvent) {
        val project = e.project ?: return
        val file = localFile(e) ?: return

        if (!isManagedFile(project, file)) {
            notifyNotHandled(project)

            return
        }

        // Delegate to the platform's reformat action: for LSP-backed files it
        // routes to textDocument/formatting on the c4drill server.
        // ActionUtil.performActionDumbAwareWithCallbacks is the supported way
        // to invoke another action programmatically —
        // AnAction.actionPerformed is @ApiStatus.OverrideOnly and must not be
        // called directly (flagged by the Plugin Verifier, issue #35).
        val reformat = ActionManager.getInstance().getAction(IdeActions.ACTION_EDITOR_REFORMAT)

        if (reformat != null && isManagedFile(project, file)) {
            ActionUtil.performActionDumbAwareWithCallbacks(reformat, e)
        }
    }
}

/** C4drillShowPreviewAction opens the live diagram tool window for the active managed document. */
class C4drillShowPreviewAction : AnAction() {
    override fun getActionUpdateThread(): ActionUpdateThread = ActionUpdateThread.BGT

    override fun update(e: AnActionEvent) {
        val project = e.project
        val file = localFile(e)

        e.presentation.isEnabled = project != null && file != null && isManagedFile(project, file)
    }

    override fun actionPerformed(e: AnActionEvent) {
        val project = e.project ?: return
        val file = localFile(e) ?: return

        if (!isManagedFile(project, file)) {
            notifyNotHandled(project)

            return
        }

        C4drillPreviewService.getInstance(project).showFor(file)
    }
}

/**
 * C4drillActivateForFileAction explicitly opts the current .toml file into
 * c4drill handling (persisted per project); the language server restarts so
 * the file attaches immediately.
 */
class C4drillActivateForFileAction : AnAction() {
    override fun getActionUpdateThread(): ActionUpdateThread = ActionUpdateThread.BGT

    override fun update(e: AnActionEvent) {
        val project = e.project ?: return
        val file = localFile(e) ?: return

        e.presentation.isEnabled = file.extension == "toml" && !isManagedFile(project, file)
    }

    override fun actionPerformed(e: AnActionEvent) {
        val project = e.project ?: return
        val file = localFile(e) ?: return

        val changed = C4drillProjectSettings.getInstance(project).activate(normalizePath(file))

        if (changed) {
            C4drillGatewayLookup.gateway(project)?.restart(project)
        }
    }
}

/** C4drillDeactivateForFileAction removes the explicit opt-in and restarts the server. */
class C4drillDeactivateForFileAction : AnAction() {
    override fun getActionUpdateThread(): ActionUpdateThread = ActionUpdateThread.BGT

    override fun update(e: AnActionEvent) {
        val project = e.project ?: return
        val file = localFile(e) ?: return

        e.presentation.isEnabled = file.extension != "c4d" &&
            normalizePath(file) in C4drillProjectSettings.getInstance(project).activatedPaths
    }

    override fun actionPerformed(e: AnActionEvent) {
        val project = e.project ?: return
        val file = localFile(e) ?: return

        val changed = C4drillProjectSettings.getInstance(project).deactivate(normalizePath(file))

        if (changed) {
            C4drillGatewayLookup.gateway(project)?.restart(project)
        }
    }
}

/** normalizePath produces the canonical '/'-separated absolute path used as the activation key. */
internal fun normalizePath(file: VirtualFile): String = FileUtil.toCanonicalPath(file.path).replace('\\', '/')
