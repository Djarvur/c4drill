// Bridge to IntelliJ's TextMate bundles plugin (optional dependency
// textmate.xml): reuses the exact c4d.tmLanguage.json grammar artifact from
// the VS Code extension (#27) — one grammar serves both editors.
//
// C4dTextMateBundleProvider supplies the grammar as a TextMate bundle; the
// language id / extensions in the bundle's package.json feed the TextMate
// plugin's extension->scope mapping, which C4dSyntaxHighlighterFactory
// consults for every .c4d file. The C4D file type itself is claimed
// unconditionally by the core plugin.xml — the TextMate plugin never owns
// the file type.

package dev.djarvur.c4drill.highlighting

import com.intellij.openapi.application.PathManager
import com.intellij.openapi.diagnostic.Logger
import com.intellij.openapi.fileTypes.SyntaxHighlighter
import com.intellij.openapi.fileTypes.SyntaxHighlighterFactory
import com.intellij.openapi.project.Project
import com.intellij.openapi.vfs.VirtualFile
import org.jetbrains.plugins.textmate.api.TextMateBundleProvider
import java.net.URLDecoder
import java.nio.file.Files
import java.nio.file.Path

class C4dTextMateBundleProvider : TextMateBundleProvider {
    override fun getBundles(): List<TextMateBundleProvider.PluginBundle> {
        val dir = ensureBundleOnDisk() ?: return emptyList()

        return listOf(TextMateBundleProvider.PluginBundle(BUNDLE_NAME, dir))
    }

    companion object {
        private val LOG = Logger.getInstance(C4dTextMateBundleProvider::class.java)

        const val BUNDLE_NAME: String = "C4Drill C4D"

        private const val RESOURCE_ROOT = "textmate/c4d"
        private val BUNDLE_FILES = listOf("package.json", "syntaxes/c4d.tmLanguage.json")

        /**
         * Copies the bundled grammar from the plugin jar (or dev content root)
         * into the system directory, where the TextMate plugin can load it.
         */
        private fun ensureBundleOnDisk(): Path? {
            val targetRoot = PathManager.getSystemDir().resolve("c4drill").resolve("c4d-bundle")

            return try {
                var changed = false

                for (rel in BUNDLE_FILES) {
                    val resource = C4dTextMateBundleProvider::class.java.classLoader.getResource("$RESOURCE_ROOT/$rel")
                        ?: run {
                            LOG.warn("Missing bundled C4D grammar resource: $RESOURCE_ROOT/$rel")

                            return null
                        }

                    val target = targetRoot.resolve(rel)

                    if (Files.exists(target) && Files.size(target) > 0 && sameContent(resource, target)) {
                        continue
                    }

                    Files.createDirectories(target.parent)
                    resource.openStream().use { input -> Files.copy(input, target, java.nio.file.StandardCopyOption.REPLACE_EXISTING) }
                    changed = true
                }

                if (changed) {
                    LOG.info("Extracted the C4D TextMate bundle to $targetRoot")
                }

                targetRoot
            } catch (e: Exception) {
                LOG.warn("Failed to extract the C4D TextMate bundle", e)

                null
            }
        }

        /** sameContent compares the classloader resource (file or jar entry) against the on-disk copy. */
        private fun sameContent(resource: java.net.URL, target: Path): Boolean {
            if (resource.protocol == "file") {
                val path = Path.of(URLDecoder.decode(resource.path, Charsets.UTF_8))

                return try {
                    Files.size(path) == Files.size(target) && Files.mismatch(path, target) == -1L
                } catch (e: Exception) {
                    false
                }
            }

            return try {
                resource.openStream().use { input ->
                    val bytes = input.readBytes()
                    Files.exists(target) && Files.size(target) == bytes.size.toLong() && Files.readAllBytes(target).contentEquals(bytes)
                }
            } catch (e: Exception) {
                false
            }
        }
    }
}

/**
 * Serves the C4D language with the TextMate highlighter for the scope
 * registered for the file's extension (source.c4d). Delegates to the
 * TextMate plugin's own factory so all grammar internals stay encapsulated
 * there; when the bundle has not loaded yet, files highlight as plain text
 * until the bundle registration fires a re-highlight.
 */
class C4dSyntaxHighlighterFactory : SyntaxHighlighterFactory() {
    private val delegate = org.jetbrains.plugins.textmate.language.syntax.highlighting.TextMateSyntaxHighlighterFactory()

    override fun getSyntaxHighlighter(project: Project?, virtualFile: VirtualFile?): SyntaxHighlighter =
        delegate.getSyntaxHighlighter(project, virtualFile)
}
