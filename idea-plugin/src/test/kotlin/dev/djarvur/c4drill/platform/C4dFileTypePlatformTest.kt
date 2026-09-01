// Issue #35: headless platform validation of the C4D file type — the file
// type must be registered by extension in the real FileTypeManager (loaded
// from plugin.xml), .c4d documents must get the C4D language, and unrelated
// .toml files must keep their own file type.
//
// Runs against the real IntelliJ IDEA 2025.3 test distribution via the
// IntelliJ Platform test framework (see build.gradle.kts
// testFramework(TestFrameworkType.Platform)).

package dev.djarvur.c4drill.platform

import com.intellij.openapi.fileTypes.FileTypeManager
import com.intellij.testFramework.fixtures.BasePlatformTestCase
import dev.djarvur.c4drill.C4dFileType
import dev.djarvur.c4drill.C4dLanguage

class C4dFileTypePlatformTest : BasePlatformTestCase() {
    fun `test c4d extension is claimed by the plugin file type`() {
        val byExtension = FileTypeManager.getInstance().getFileTypeByExtension("c4d")

        assertSame("the c4d extension must resolve to the C4D file type registered in plugin.xml", C4dFileType.INSTANCE, byExtension)
    }

    fun `test c4d filenames resolve to the C4D file type which carries the C4D language`() {
        val fileTypeManager = FileTypeManager.getInstance()

        assertSame("a .c4d filename must resolve to the C4D file type registered in plugin.xml", C4dFileType.INSTANCE, fileTypeManager.getFileTypeByFileName("model.c4d"))
        assertSame("the file type must carry the C4D language so the TextMate highlighter and LSP wiring attach", C4dLanguage.INSTANCE, C4dFileType.INSTANCE.language)
    }

    fun `test plain toml keeps its own file type`() {
        val file = myFixture.configureByText("notes.toml", "a = 1\n")

        assertFalse("c4drill must not hijack plain TOML files", file.fileType is C4dFileType)
    }

    companion object {
        val SAMPLE_C4D: String = """
            properties {
              name: Demo
            }

            frontend: system "Frontend" {
              technology: React
              -> backend: "HTTPS | Browses"
            }

            backend: system "Backend" {
              technology: Go
            }
        """.trimIndent()
    }
}
