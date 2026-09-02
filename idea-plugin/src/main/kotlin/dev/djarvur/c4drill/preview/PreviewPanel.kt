// The preview panel UI: a toolbar row (breadcrumb + view controls) above a
// card stack with either the JCEF browser (SVG) or a status panel (CLI-style
// errors / informational states). All UI work happens on the EDT; the
// service only calls showDiagram/showStatus/updateBreadcrumb from the EDT.

package dev.djarvur.c4drill.preview

import com.intellij.openapi.actionSystem.ActionManager
import com.intellij.openapi.actionSystem.ActionToolbar
import com.intellij.openapi.actionSystem.DefaultActionGroup
import com.intellij.openapi.project.Project
import com.intellij.openapi.Disposable
import com.intellij.openapi.ui.SimpleToolWindowPanel
import com.intellij.ui.components.ActionLink
import com.intellij.ui.components.JBLabel
import com.intellij.ui.components.JBScrollPane
import com.intellij.ui.jcef.JBCefApp
import com.intellij.ui.jcef.JBCefBrowser
import com.intellij.ui.jcef.JBCefBrowserBase
import com.intellij.ui.jcef.JBCefJSQuery
import com.intellij.util.ui.JBUI
import dev.djarvur.c4drill.core.BreadcrumbEntry
import dev.djarvur.c4drill.core.breadcrumbTrail
import java.awt.BorderLayout
import java.awt.CardLayout
import javax.swing.JPanel
import javax.swing.JComponent
import javax.swing.JEditorPane
import javax.swing.SwingConstants

class PreviewPanel(
    @Suppress("unused") private val project: Project,
    private val state: PreviewViewStateHolder,
    private val onBreadcrumb: (BreadcrumbEntry) -> Unit,
    private val onLink: (String) -> Unit,
    private val onRender: () -> Unit,
    private val onExport: () -> Boolean,
) : Disposable {
    val component: JComponent

    private val breadcrumbPanel = JPanel(BorderLayout())
    private val cardLayout = CardLayout()
    private val cardPanel = JPanel(cardLayout)

    private val browser: JBCefBrowserBase?
    private val jsQuery: JBCefJSQuery?
    private val statusPane = JEditorPane("text/html", "")

    init {
        // Toolbar: view controls on the left, breadcrumb row below it.
        val group = DefaultActionGroup().apply { PreviewActions.addAll(this, state, onRender, onExport) }

        val toolbar: ActionToolbar = ActionManager.getInstance()
            .createActionToolbar("C4DrillPreviewToolbar", group, true)

        val north = JPanel(BorderLayout()).apply {
            add(toolbar.component, BorderLayout.NORTH)
            add(breadcrumbPanel, BorderLayout.CENTER)
            border = JBUI.Borders.empty(2, 4)
        }

        // Content: browser or status card.
        if (JBCefApp.isSupported()) {
            val b: JBCefBrowserBase = JBCefBrowser()
            browser = b
            jsQuery = JBCefJSQuery.create(b)

            jsQuery.addHandler { request ->
                try {
                    onLink(request ?: "")
                } catch (e: Exception) {
                    com.intellij.openapi.diagnostic.Logger.getInstance(PreviewPanel::class.java)
                        .warn("preview link handling failed", e)
                }

                JBCefJSQuery.Response("ok")
            }

            cardPanel.add(b.component, CARD_DIAGRAM)
        } else {
            browser = null
            jsQuery = null

            cardPanel.add(statusCard("Preview requires JCEF", listOf("This IDE runtime does not embed Chromium (JCEF).", "Use a standard JetBrains Runtime to enable the live diagram.")), CARD_DIAGRAM)
        }

        statusPane.isEditable = false
        statusPane.background = JBUI.CurrentTheme.ToolWindow.background()
        cardPanel.add(JBScrollPane(statusPane), CARD_STATUS)

        cardLayout.show(cardPanel, CARD_STATUS)

        component = SimpleToolWindowPanel(true).apply {
            setToolbar(north)
            setContent(cardPanel)
        }
        toolbar.targetComponent = component
    }

    /** showDiagram renders the SVG (with the click-bridge injected) in the embedded browser. */
    fun showDiagram(svg: String) {
        val q = jsQuery
        val b = browser

        if (q == null || b == null) {
            return
        }

        updateBreadcrumb(state.snapshot())
        cardLayout.show(cardPanel, CARD_DIAGRAM)

        b.loadHTML(PreviewHtml.diagramHtml(svg, q.funcName).replace(PreviewHtml.BRIDGE_PLACEHOLDER, q.funcName))
        b.cefBrowser.executeJavaScript(q.inject(""), null, 0)
    }

    /** showStatus replaces the diagram with the CLI-style error/state panel — never leaves a stale diagram. */
    fun showStatus(title: String, lines: List<String>) {
        statusPane.text = PreviewHtml.statusHtml(title, lines)
        cardLayout.show(cardPanel, CARD_STATUS)
    }

    /** updateBreadcrumb rebuilds the breadcrumb row for the current target. */
    fun updateBreadcrumb(snapshot: Snapshot) {
        val entries = breadcrumbTrail(
            dev.djarvur.c4drill.core.PreviewViewState(
                target = snapshot.target,
                allExpanded = snapshot.allExpanded,
            ),
        )

        breadcrumbPanel.removeAll()

        val row = JPanel(javax.swing.BoxLayout(breadcrumbPanel, javax.swing.BoxLayout.LINE_AXIS))
        row.isOpaque = false

        entries.forEachIndexed { i, entry ->
            if (i > 0) {
                row.add(JBLabel("  >  ", SwingConstants.CENTER))
            }

            if (entry.current) {
                row.add(JBLabel(entry.label, SwingConstants.CENTER).apply { font = font.deriveFont(java.awt.Font.BOLD) })
            } else {
                row.add(ActionLink(entry.label) { onBreadcrumb(entry) })
            }
        }

        breadcrumbPanel.add(row, BorderLayout.WEST)
        breadcrumbPanel.revalidate()
        breadcrumbPanel.repaint()
    }

    private fun statusCard(title: String, lines: List<String>): JComponent {
        val pane = JEditorPane("text/html", PreviewHtml.statusHtml(title, lines))

        pane.isEditable = false

        return JBScrollPane(pane)
    }

    override fun dispose() {
        jsQuery?.dispose()
        browser?.dispose()
    }

    companion object {
        private const val CARD_DIAGRAM = "diagram"
        private const val CARD_STATUS = "status"
    }
}
