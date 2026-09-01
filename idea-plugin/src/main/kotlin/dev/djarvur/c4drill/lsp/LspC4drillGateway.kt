// The LSP-backed C4drillGateway: talks to the running `c4drill serve --lsp`
// client through the platform LSP API. Registered in lsp.xml; only loaded
// when the optional com.intellij.modules.lsp dependency is present.
//
// Threading: renderDiagram/nudgeValidation may block up to the request
// timeout — callers must invoke them off the EDT (the preview does).

package dev.djarvur.c4drill.lsp

import com.intellij.openapi.diagnostic.Logger
import com.intellij.openapi.editor.Document
import com.intellij.openapi.fileEditor.FileDocumentManager
import com.intellij.openapi.progress.ProcessCanceledException
import com.intellij.openapi.project.Project
import com.intellij.openapi.util.registry.Registry
import com.intellij.openapi.vfs.VirtualFile
import com.intellij.platform.lsp.api.LspServer
import com.intellij.platform.lsp.api.LspServerManager
import com.intellij.platform.lsp.api.LspServerState
import dev.djarvur.c4drill.lsp.C4drillGateway
import dev.djarvur.c4drill.lsp.GatewayDiagnostic
import dev.djarvur.c4drill.lsp.GatewayRenderResult
import org.eclipse.lsp4j.DidChangeTextDocumentParams
import org.eclipse.lsp4j.TextDocumentContentChangeEvent
import org.eclipse.lsp4j.VersionedTextDocumentIdentifier

class LspC4drillGateway : C4drillGateway {
    override fun unavailableReason(project: Project): String? {
        try {
            resolveServerCommand()
        } catch (e: Exception) {
            return e.message
        }

        val server = currentServer(project) ?: return "The c4drill language server has not started yet — open a c4drill model file first."

        return when (server.state) {
            LspServerState.Running -> null
            LspServerState.Initializing -> "The c4drill language server is still starting."
            LspServerState.ShutdownNormally -> "The c4drill language server is not running."
            LspServerState.ShutdownUnexpectedly -> "The c4drill language server stopped unexpectedly."
        }
    }

    override fun isAttached(project: Project, file: VirtualFile): Boolean = currentServer(project) != null

    override fun renderDiagram(
        project: Project,
        file: VirtualFile,
        target: String?,
        allExpanded: Boolean,
        expanded: List<String>?,
        legend: Boolean?,
    ): GatewayRenderResult? {
        val server = currentServer(project) ?: return null

        val params = C4drillRenderDiagramParams(
            textDocument = C4drillWireTextDocumentIdentifier(uri = server.getDocumentIdentifier(file).uri),
            target = target ?: "",
            allExpanded = allExpanded,
            expanded = expanded,
            legend = legend,
            format = "svg",
        )

        val result = try {
            server.sendRequestSync(renderTimeoutMs()) { lsp4j ->
                (lsp4j as C4drillLsp4jServer).renderDiagram(params)
            }
        } catch (e: ProcessCanceledException) {
            throw e
        } catch (e: Exception) {
            LOG.warn("c4drill/renderDiagram request failed", e)

            return null
        } ?: return null

        return GatewayRenderResult(
            svg = result.svg,
            diagnostics = result.diagnostics.map { d ->
                GatewayDiagnostic(line = d.range?.start?.line ?: 0, message = d.message)
            },
        )
    }

    override fun nudgeValidation(project: Project, file: VirtualFile) {
        val server = currentServer(project) ?: return

        val document: Document = FileDocumentManager.getInstance().getDocument(file) ?: return
        val uri = server.getDocumentIdentifier(file).uri
        val version = server.getDocumentVersion(document)
        val text = document.text

        server.sendNotification { lsp4j ->
            lsp4j.textDocumentService.didChange(
                DidChangeTextDocumentParams(
                    VersionedTextDocumentIdentifier(uri, version),
                    listOf(TextDocumentContentChangeEvent(text)),
                ),
            )
        }
    }

    override fun restart(project: Project) {
        LspServerManager.getInstance(project).stopAndRestartIfNeeded(C4drillLspServerSupportProvider::class.java)
    }

    /** ensureStarted mirrors the provider's fileOpened decision for the currently edited managed file. */
    override fun ensureStarted(project: Project, file: VirtualFile) {
        LspServerManager.getInstance(project).ensureServerStarted(
            C4drillLspServerSupportProvider::class.java,
            C4drillLspServerDescriptor(project),
        )
    }

    private fun currentServer(project: Project): LspServer? =
        LspServerManager.getInstance(project)
            .getServersForProvider(C4drillLspServerSupportProvider::class.java)
            .firstOrNull()

    companion object {
        private val LOG = Logger.getInstance(LspC4drillGateway::class.java)

        /** Render request timeout override for large models: registry key `c4drill.render.timeout.ms`. */
        fun renderTimeoutMs(): Int = Registry.intValue("c4drill.render.timeout.ms", 10_000)
    }
}
