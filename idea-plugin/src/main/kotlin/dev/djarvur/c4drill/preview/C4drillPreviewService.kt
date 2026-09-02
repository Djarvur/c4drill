// The "C4Drill Preview" tool window controller (issue #29 scope update).
//
// Renders the active managed document through the c4drill/renderDiagram LSP
// request, debounced (~200 ms) on document change, and displays the SVG in an
// embedded browser (JCEF). Clicks on the SVG's internal drill-down links
// (relative .svg URLs) are intercepted and turned into a new renderDiagram
// call for that target (C1 -> C2 -> C3 with breadcrumbs); external http(s)
// links open in the system browser. Parse/validate failures show the
// CLI-identical diagnostics instead of a stale diagram.

package dev.djarvur.c4drill.preview

import com.intellij.openapi.Disposable
import com.intellij.openapi.application.ApplicationManager
import com.intellij.openapi.components.Service
import com.intellij.openapi.components.service
import com.intellij.openapi.editor.event.DocumentEvent
import com.intellij.openapi.editor.event.DocumentListener
import com.intellij.openapi.editor.EditorFactory
import com.intellij.openapi.fileEditor.FileDocumentManager
import com.intellij.openapi.project.Project
import com.intellij.openapi.util.Disposer
import com.intellij.openapi.vfs.VirtualFile
import com.intellij.openapi.wm.ToolWindow
import com.intellij.openapi.wm.ToolWindowManager
import com.intellij.ide.BrowserUtil
import dev.djarvur.c4drill.C4drillGlobalSettings
import dev.djarvur.c4drill.core.BreadcrumbEntry
import dev.djarvur.c4drill.core.Debouncer
import dev.djarvur.c4drill.core.basenameOfUri
import dev.djarvur.c4drill.core.resolveRenderTarget
import dev.djarvur.c4drill.isManagedFile
import dev.djarvur.c4drill.lsp.C4drillGatewayLookup
import dev.djarvur.c4drill.lsp.GatewayRenderResult
import java.util.concurrent.TimeUnit
import java.util.concurrent.atomic.AtomicLong

@Service(Service.Level.PROJECT)
class C4drillPreviewService(private val project: Project) : Disposable {
    /** The document the preview follows; null until Show Preview runs. */
    private var previewedFile: VirtualFile? = null

    private val state = PreviewViewStateHolder()
    private val renderGeneration = AtomicLong(0)
    private val debounceExecutor = com.intellij.util.concurrency.AppExecutorUtil.getAppScheduledExecutorService()

    @Volatile
    private var pendingRender: Debouncer.Cancelable? = null
    /** The last successfully rendered result (SVG kept for export). */
    @Volatile
    var lastRender: GatewayRenderResult? = null
        private set

    private var panel: PreviewPanel? = null

    init {
        EditorFactory.getInstance().eventMulticaster.addDocumentListener(
            object : DocumentListener {
                override fun documentChanged(event: DocumentEvent) {
                    val file = previewedFile ?: return

                    if (FileDocumentManager.getInstance().getFile(event.document) !== file) {
                        return
                    }

                    scheduleRender()
                }
            },
            this,
        )
    }

    /** attachToolWindow installs the preview panel into the tool window (idempotent). */
    fun attachToolWindow(toolWindow: ToolWindow) {
        ApplicationManager.getApplication().invokeLater {
            val cm = toolWindow.contentManager
            val p = panel ?: PreviewPanel(project, state, ::onBreadcrumb, ::onLinkClicked, ::renderNow, ::exportSvg).also {
                panel = it
                Disposer.register(this@C4drillPreviewService, it)
            }

            val content = cm.getContent(0) ?: cm.factory.createContent(p.component, "", false).also { cm.addContent(it) }
            content.setComponent(p.component)
            content.displayName = "C1"
        }
    }

    /** showFor switches the preview to [file] (a managed document) and reveals the tool window. */
    fun showFor(file: VirtualFile) {
        previewedFile = file
        state.reset()
        lastRender = null

        val toolWindow = ToolWindowManager.getInstance(project).getToolWindow(C4drillPreviewToolWindowFactory.ID)
        if (toolWindow == null) {
            return
        }

        attachToolWindow(toolWindow)

        val gateway = C4drillGatewayLookup.gateway(project)

        if (gateway != null) {
            ApplicationManager.getApplication().executeOnPooledThread { gateway.ensureStarted(project, file) }
        }

        toolWindow.activate {
            renderNow()
        }
    }

    /** scheduleRender coalesces bursts of document changes into one render after the debounce window. */
    fun scheduleRender() {
        if (previewedFile == null) {
            return
        }

        val delayMs = C4drillGlobalSettings.getInstance().previewDebounceMs.toLong().coerceIn(50, 5_000)
        renderGeneration.incrementAndGet()

        val debouncer = Debouncer(delayMs) { task, delay ->
            val future = debounceExecutor.schedule(task, delay, TimeUnit.MILLISECONDS)
            Debouncer.Cancelable { future.cancel(false) }
        }
        pendingRender = debouncer.trigger {
            ApplicationManager.getApplication().invokeLater {
                renderNow()
            }
        }
    }

    /** renderNow runs the CLI-identical render pipeline off the EDT and updates the panel. */
    fun renderNow() {
        val file = previewedFile ?: return

        if (!isManagedFile(project, file)) {
            panel?.showStatus(
                "Not a c4drill model",
                listOf(
                    "C4Drill does not handle this file. For .toml models use",
                    "\"C4Drill: Activate for This File\" or configure c4drill.toml.patterns.",
                ),
            )

            return
        }

        val gateway = C4drillGatewayLookup.gateway(project)

        if (gateway == null) {
            panel?.showStatus(
                "Language server unavailable",
                listOf(
                    "This IDE does not provide the IntelliJ Platform LSP client.",
                    "Use a 2025.3+ commercial IDE (IntelliJ IDEA, GoLand, PyCharm, WebStorm, ...) for the preview.",
                ),
            )

            return
        }

        val generation = renderGeneration.incrementAndGet()
        val snapshot = state.snapshot()

        ApplicationManager.getApplication().executeOnPooledThread {
            val unavailable = gateway.unavailableReason(project)

            val result: GatewayRenderResult? = if (unavailable == null) {
                gateway.renderDiagram(
                    project = project,
                    file = file,
                    target = snapshot.target.ifEmpty { null },
                    allExpanded = snapshot.allExpanded,
                    expanded = snapshot.expandedOverride,
                    legend = snapshot.legendOverride,
                )
            } else {
                null
            }

            ApplicationManager.getApplication().invokeLater {
                if (renderGeneration.get() != generation) {
                    return@invokeLater // a newer render superseded this one
                }

                val p = panel ?: return@invokeLater

                when {
                    unavailable != null -> p.showStatus("Language server unavailable", listOf(unavailable))

                    result == null -> p.showStatus(
                        "Render failed",
                        listOf("The c4drill language server did not respond (timeout or request error)."),
                    )

                    result.svg.isEmpty() -> p.showStatus("Model has errors", result.diagnostics.map { it.message })

                    else -> {
                        lastRender = result
                        p.showDiagram(result.svg)
                        updateToolWindowTitle()
                    }
                }
            }
        }
    }

    private fun updateToolWindowTitle() {
        val toolWindow = ToolWindowManager.getInstance(project).getToolWindow(C4drillPreviewToolWindowFactory.ID) ?: return
        val snapshot = state.snapshot()
        val leaf = if (snapshot.allExpanded) "expanded" else snapshot.target.ifEmpty { "C1" }

        toolWindow.title = leaf
    }

    /** onBreadcrumb is the click callback for breadcrumb entries. */
    private fun onBreadcrumb(entry: BreadcrumbEntry) {
        state.setTarget(entry.target)
        renderNow()
    }

    /** onLinkClicked routes a clicked href from the SVG: internal .svg links drill down, http(s) opens the browser. */
    private fun onLinkClicked(href: String) {
        val trimmed = href.trim()

        if (trimmed.startsWith("http://") || trimmed.startsWith("https://")) {
            BrowserUtil.browse(trimmed)

            return
        }

        val file = previewedFile ?: return
        val basename = basenameOfUri(file.presentableName)
        val target = resolveRenderTarget(state.snapshot().target, basename, trimmed)

        when (target) {
            null -> Unit // not an internal link (mailto:, out-of-tree, malformed) — ignore

            else -> {
                state.setTarget(target)
                renderNow()
            }
        }
    }

    private fun exportSvg(): Boolean {
        val svg = lastRender?.svg ?: return false

        return PreviewExport.exportSvg(project, svg)
    }

    override fun dispose() {
        pendingRender?.cancel()
    }

    companion object {
        @JvmStatic
        fun getInstance(project: Project): C4drillPreviewService = project.service()
    }
}
