// Issue #35: headless platform validation of settings/service wiring — the
// application- and project-level PersistentStateComponents must load as real
// platform services, the TOML scoping decision must work against real
// VirtualFiles in a fixture project, and (when the platform LSP client is
// present, as in every commercial 253+ IDE) the LSP-backed gateway must be
// registered through lsp.xml.
//
// Runs against the real IntelliJ IDEA 2025.3 test distribution.

package dev.djarvur.c4drill.platform

import com.intellij.openapi.extensions.Extensions
import com.intellij.testFramework.fixtures.BasePlatformTestCase
import dev.djarvur.c4drill.C4drillGlobalSettings
import dev.djarvur.c4drill.C4drillProjectSettings
import dev.djarvur.c4drill.isManagedFile

class C4drillSettingsWiringPlatformTest : BasePlatformTestCase() {
    fun `test global settings service loads with defaults and clamps the debounce window`() {
        val settings = C4drillGlobalSettings.getInstance()

        assertEquals("", settings.serverPath)
        assertEquals(C4drillGlobalSettings.DEFAULT_PREVIEW_DEBOUNCE_MS, settings.previewDebounceMs)

        settings.previewDebounceMs = C4drillGlobalSettings.MAX_PREVIEW_DEBOUNCE_MS + 10_000
        assertEquals(C4drillGlobalSettings.MAX_PREVIEW_DEBOUNCE_MS, settings.previewDebounceMs)

        settings.previewDebounceMs = C4drillGlobalSettings.MIN_PREVIEW_DEBOUNCE_MS - 100
        assertEquals(C4drillGlobalSettings.MIN_PREVIEW_DEBOUNCE_MS, settings.previewDebounceMs)
    }

    fun `test project settings service loads and round-trips activation`() {
        val settings = C4drillProjectSettings.getInstance(project)

        assertTrue(settings.tomlPatterns.isEmpty())
        assertTrue(settings.activatedPaths.isEmpty())

        assertTrue(settings.activate("/models/a.toml"))
        assertFalse("activating the same path twice must not report a change", settings.activate("/models/a.toml"))
        assertEquals(setOf("/models/a.toml"), settings.activatedPaths)

        assertTrue(settings.deactivate("/models/a.toml"))
        assertFalse(settings.deactivate("/models/a.toml"))
        assertTrue(settings.activatedPaths.isEmpty())
    }

    fun `test c4d documents are managed and plain toml are not`() {
        val c4d = myFixture.configureByText("model.c4d", "backend: system \"Backend\" {\n  technology: Go\n}\n")
        assertTrue(".c4d documents are always handled", isManagedFile(project, c4d.virtualFile))

        val toml = myFixture.configureByText("notes.toml", "a = 1\n")
        assertFalse("plain .toml files without opt-in must not be handled", isManagedFile(project, toml.virtualFile))
    }

    fun `test toml files opt in via the explicit activation`() {
        // Glob matching over project-relative paths is covered exhaustively by
        // the pure-JVM DocumentScopeTest; this test exercises the real
        // VirtualFile plumbing: the settings service read and the activated
        // path round trip through isManagedFile.
        // (Light fixture names cannot contain '/'; real VFS files are avoided
        // because bundled plugins with frontend-only classes break VFS
        // listeners in headless tests.)
        val model = myFixture.configureByText("model.toml", "[A]\ntype = \"system\"\nname = \"A\"\n")
        val settings = C4drillProjectSettings.getInstance(project)

        assertFalse("a .toml without any opt-in must not be handled", isManagedFile(project, model.virtualFile))

        val absolutePath = model.virtualFile.path
        settings.activate(absolutePath)
        assertTrue("an explicitly activated .toml must be handled", isManagedFile(project, model.virtualFile))

        settings.deactivate(absolutePath)
        assertFalse("after deactivation the .toml must not be handled", isManagedFile(project, model.virtualFile))
    }

    fun `test the lsp wiring from lsp xml is registered when the platform lsp client is present`() {
        // The production lookup is C4drillGatewayLookup.gateway(project). The
        // light test fixture does not bridge plugin-defined extension points
        // into per-project extension areas, so the same lsp.xml registration
        // is asserted through the root area: the gateway extension point (the
        // seam lsp.xml plugs the gateway into) must exist and resolve to the
        // LSP-backed implementation. Only OUR extension is instantiated —
        // enumerating the platform's serverSupportProvider EP would also
        // instantiate third-party providers whose classes need a frontend.
        val rootArea = Extensions.getRootArea()

        assertTrue(
            "the dev.djarvur.c4drill.gateway extension point (defined in plugin.xml) must be registered",
            rootArea.hasExtensionPoint("dev.djarvur.c4drill.gateway"),
        )

        val ep: com.intellij.openapi.extensions.ExtensionPoint<dev.djarvur.c4drill.lsp.C4drillGateway> =
            rootArea.getExtensionPoint("dev.djarvur.c4drill.gateway")
        assertTrue(
            "lsp.xml registers LspC4drillGateway behind the optional com.intellij.modules.lsp dependency; " +
                "it must resolve on the 2025.3 test distribution",
            ep.extensionList.any { it is dev.djarvur.c4drill.lsp.LspC4drillGateway },
        )
    }
}
