// Issue #35: headless platform validation of the C4D TextMate bundle — the
// provider must extract the #27 grammar from the plugin into the system dir,
// the TextMate plugin must map the c4d extension onto the source.c4d scope,
// and the C4D syntax highlighter must tokenize an example file with real
// TextMate tokens (not the plain-text fallback).
//
// Runs against the real IntelliJ IDEA 2025.3 test distribution with the
// bundled TextMate plugin loaded (optional dependency satisfied).

package dev.djarvur.c4drill.platform

import com.intellij.openapi.fileTypes.SyntaxHighlighterFactory
import com.intellij.testFramework.fixtures.BasePlatformTestCase
import dev.djarvur.c4drill.C4dLanguage
import dev.djarvur.c4drill.highlighting.C4dTextMateBundleProvider
import org.jetbrains.plugins.textmate.TextMateService
import org.jetbrains.plugins.textmate.api.TextMateBundleProvider
import org.jetbrains.plugins.textmate.language.syntax.lexer.TextMateElementType
import java.nio.file.Files
import kotlin.io.path.readText

class C4dTextMatePlatformTest : BasePlatformTestCase() {
    fun `test bundle provider extracts the C4D grammar from the plugin`() {
        val bundles: List<TextMateBundleProvider.PluginBundle> = C4dTextMateBundleProvider().getBundles()

        assertEquals("exactly one bundle: the C4D grammar", 1, bundles.size)
        assertEquals(C4dTextMateBundleProvider.BUNDLE_NAME, bundles[0].name)

        val root = bundles[0].path
        assertTrue("package.json must be extracted into the system dir", Files.exists(root.resolve("package.json")))
        assertTrue("the grammar file must be extracted next to it", Files.exists(root.resolve("syntaxes/c4d.tmLanguage.json")))
        assertTrue(
            "the grammar must be the #27 artifact rooted at source.c4d",
            root.resolve("syntaxes/c4d.tmLanguage.json").readText().contains("\"source.c4d\""),
        )
    }

    fun `test registered bundle produces real TextMate highlighting for an example c4d file`() {
        // Load the bundles the same way the TextMate plugin does at runtime:
        // through the bundleProvider extension point our textmate.xml declares.
        // reloadEnabledBundles applies the extension->scope mapping
        // asynchronously, so keep nudging it (and pumping the IDE event queue)
        // until the c4d mapping lands.
        val tmService = TextMateService.getInstance()
        tmService.reloadEnabledBundles()

        var descriptor: org.jetbrains.plugins.textmate.language.TextMateLanguageDescriptor? = null
        val deadline = System.currentTimeMillis() + 30_000

        while (System.currentTimeMillis() < deadline) {
            descriptor = tmService.getLanguageDescriptorByExtension("c4d")

            if (descriptor != null) {
                break
            }

            com.intellij.util.ui.UIUtil.dispatchAllInvocationEvents()
            Thread.sleep(200)
            tmService.reloadEnabledBundles()
        }

        assertNotNull("the c4d extension must map onto a registered TextMate language after bundle registration", descriptor)
        assertEquals(
            "the registered grammar must be the source.c4d scope from the #27 artifact",
            "source.c4d",
            descriptor!!.rootScopeName.toString(),
        )

        val file = myFixture.configureByText("model.c4d", C4dFileTypePlatformTest.SAMPLE_C4D)
        val highlighter = SyntaxHighlighterFactory.getSyntaxHighlighter(C4dLanguage.INSTANCE, project, file.virtualFile)
        val lexer = highlighter.highlightingLexer

        val text = file.text
        lexer.start(text)

        var tokenCount = 0
        var textMateTokens = 0

        while (lexer.tokenType != null) {
            tokenCount++
            if (lexer.tokenType is TextMateElementType) {
                textMateTokens++
            }
            lexer.advance()
        }

        assertTrue("the example must tokenize into more than one token (got $tokenCount)", tokenCount > 1)
        assertTrue(
            "expected real TextMate tokens for the example file (got $tokenCount tokens, $textMateTokens TextMate) — " +
                "the plain-text fallback means the bundle did not load",
            textMateTokens > 0,
        )
    }
}
