// The lsp4j wire surface for the c4drill server (issue #32 contract). Only
// the pieces the platform client does not already provide live here: the
// custom `c4drill/renderDiagram` request used by the live preview. The class
// is installed via C4drillLspServerDescriptor.lsp4jServerClass, which makes
// the client's lsp4j proxy implement it (SDK docs: "To send custom
// (undocumented) requests ... override lsp4jServerClass").
//
// Field names mirror internal/lsp/protocol.go exactly (RenderDiagramParams /
// RenderDiagramResult / Diagnostic). `expanded` has no default-on-the-wire:
// an EMPTY list means "collapse all" and must reach the server, while null
// means "model default" (lsp4j's Gson omits nulls).

package dev.djarvur.c4drill.lsp

import org.eclipse.lsp4j.jsonrpc.services.JsonRequest
import org.eclipse.lsp4j.services.LanguageServer
import java.util.concurrent.CompletableFuture

class C4drillWirePosition(
    @JvmField var line: Int = 0,
    @JvmField var character: Int = 0,
)

class C4drillWireRange(
    @JvmField var start: C4drillWirePosition = C4drillWirePosition(),
    @JvmField var end: C4drillWirePosition = C4drillWirePosition(),
)

class C4drillWireDiagnostic(
    @JvmField var range: C4drillWireRange = C4drillWireRange(),
    @JvmField var severity: Int = 0,
    @JvmField var source: String? = null,
    @JvmField var message: String = "",
)

class C4drillWireTextDocumentIdentifier(
    @JvmField var uri: String = "",
)

class C4drillRenderDiagramParams(
    @JvmField var textDocument: C4drillWireTextDocumentIdentifier = C4drillWireTextDocumentIdentifier(),
    @JvmField var target: String? = null,
    @JvmField var allExpanded: Boolean = false,
    @JvmField var expanded: List<String>? = null,
    @JvmField var legend: Boolean? = null,
    @JvmField var format: String? = null,
)

class C4drillRenderDiagramResult(
    @JvmField var svg: String = "",
    @JvmField var diagnostics: List<C4drillWireDiagnostic> = emptyList(),
)

interface C4drillLsp4jServer : LanguageServer {
    @JsonRequest("c4drill/renderDiagram")
    fun renderDiagram(params: C4drillRenderDiagramParams): CompletableFuture<C4drillRenderDiagramResult>
}
