// The core-side seam between always-loaded plugin code (preview tool window,
// actions) and the LSP integration that lives behind the optional
// `com.intellij.modules.lsp` dependency (lsp.xml). Core code references ONLY
// this interface; the LSP-backed implementation is registered as a project
// extension of dev.djarvur.c4drill.gateway in lsp.xml and is instantiated
// only when the platform LSP client is present.

package dev.djarvur.c4drill.lsp

import com.intellij.openapi.extensions.ExtensionPointName
import com.intellij.openapi.project.Project
import com.intellij.openapi.vfs.VirtualFile

/** One pipeline diagnostic, wire-shaped like the LSP Diagnostic the server publishes. */
data class GatewayDiagnostic(
    val line: Int,
    val message: String,
)

/** One c4drill/renderDiagram result: the SVG text (empty on failure) plus the pipeline diagnostics. */
data class GatewayRenderResult(
    val svg: String,
    val diagnostics: List<GatewayDiagnostic>,
)

/**
 * C4drillGateway is the preview/actions' view of the running c4drill language
 * server. All methods are safe to call from any thread; implementations must
 * hop threads internally as required.
 */
interface C4drillGateway {
    /** A short human-readable reason when the gateway cannot serve requests (no server, no binary...), null when healthy. */
    fun unavailableReason(project: Project): String?

    /** True when the document currently receives language-server traffic (i.e. was opened while managed). */
    fun isAttached(project: Project, file: VirtualFile): Boolean

    /**
     * Runs the CLI-identical render pipeline for the document. Returns null
     * when the server is unavailable (see [unavailableReason]) or the request
     * timed out; validation failures yield an empty svg plus diagnostics.
     */
    fun renderDiagram(
        project: Project,
        file: VirtualFile,
        target: String?,
        allExpanded: Boolean,
        expanded: List<String>?,
        legend: Boolean?,
    ): GatewayRenderResult?

    /** Sends the current buffer to the server so it re-runs validation and republishes diagnostics. */
    fun nudgeValidation(project: Project, file: VirtualFile)

    /** Starts the language server if not running (mirrors the provider's fileOpened decision for [file]). */
    fun ensureStarted(project: Project, file: VirtualFile)

    /** Restarts the language server (settings changed, or a file was activated/deactivated). */
    fun restart(project: Project)
}

/** lookup resolves the LSP-backed gateway; null when the LSP module is absent in this IDE. */
object C4drillGatewayLookup {
    val EP_NAME: ExtensionPointName<C4drillGateway> = ExtensionPointName("dev.djarvur.c4drill.gateway")

    fun gateway(project: Project): C4drillGateway? = EP_NAME.getExtensionList(project).firstOrNull()
}
