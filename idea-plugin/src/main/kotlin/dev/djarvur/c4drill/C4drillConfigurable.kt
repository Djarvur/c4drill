// Settings UI: Settings | Tools | C4Drill (project level). Exposes the
// server binary override, the preview debounce, and the c4drill TOML scoping
// patterns; applies restart the language server when the path changes.

package dev.djarvur.c4drill

import com.intellij.openapi.options.BoundConfigurable
import com.intellij.openapi.project.Project
import com.intellij.openapi.ui.DialogPanel
import com.intellij.ui.components.JBTextArea
import com.intellij.ui.components.JBTextField
import com.intellij.ui.dsl.builder.Align
import com.intellij.ui.dsl.builder.panel
import com.intellij.openapi.application.ApplicationManager
import dev.djarvur.c4drill.lsp.C4drillGatewayLookup

class C4drillConfigurable(private val project: Project) : BoundConfigurable("C4Drill") {
    private val global get() = C4drillGlobalSettings.getInstance()
    private val projectSettings get() = C4drillProjectSettings.getInstance(project)

    private lateinit var serverPathField: JBTextField
    private lateinit var debounceField: JBTextField
    private lateinit var patternsArea: JBTextArea

    override fun createPanel(): DialogPanel {
        serverPathField = JBTextField(global.serverPath, 40)
        debounceField = JBTextField(global.previewDebounceMs.toString(), 8)
        patternsArea = JBTextArea(projectSettings.tomlPatterns.joinToString("\n"), 5, 60)

        return panel {
            group("Language server") {
                row("c4drill server path:") {
                    cell(serverPathField)
                        .align(Align.FILL)
                        .comment("Leave blank to discover the <code>c4drill</code> binary on PATH. The server runs as <code>c4drill serve --lsp</code>.")
                }
            }
            group("Preview") {
                row("Debounce (ms):") {
                    cell(debounceField).comment("Re-render delay after edits (50-5000, default 200).")
                }
            }
            group("c4drill TOML scoping") {
                row("c4drill.toml.patterns:") {
                    cell(patternsArea)
                        .align(Align.FILL)
                        .comment(
                            "One glob per line (relative to the project root or absolute). " +
                                "Only matching .toml files get c4drill features; plain TOML keeps the built-in TOML plugin. " +
                                "Use \"C4Drill: Activate for This File\" for per-file opt-in.",
                        )
                }
            }
        }
    }

    override fun isModified(): Boolean {
        val patterns = patternsArea.text.lines().map { it.trim() }.filter { it.isNotEmpty() }

        return serverPathField.text.trim() != global.serverPath ||
            (debounceField.text.trim().toIntOrNull() ?: global.previewDebounceMs) != global.previewDebounceMs ||
            patterns != projectSettings.tomlPatterns
    }

    override fun apply() {
        val serverPathChanged = serverPathField.text.trim() != global.serverPath

        global.serverPath = serverPathField.text.trim()
        global.previewDebounceMs = debounceField.text.trim().toIntOrNull() ?: global.previewDebounceMs

        val patterns = patternsArea.text.lines().map { it.trim() }.filter { it.isNotEmpty() }
        val patternsChanged = patterns != projectSettings.tomlPatterns
        projectSettings.tomlPatterns = patterns

        if (serverPathChanged || patternsChanged) {
            ApplicationManager.getApplication().executeOnPooledThread {
                C4drillGatewayLookup.gateway(project)?.restart(project)
            }
        }
    }
}
