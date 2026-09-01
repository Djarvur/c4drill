// Export: saves the currently rendered SVG from the preview surface. The
// LSP render contract (v1) exposes only svg — html/dot exports stay a CLI
// operation until the server grows them.

package dev.djarvur.c4drill.preview

import com.intellij.openapi.diagnostic.Logger
import com.intellij.openapi.fileChooser.FileChooserFactory
import com.intellij.openapi.fileChooser.FileSaverDescriptor
import com.intellij.openapi.project.Project
import com.intellij.openapi.vfs.VirtualFile

object PreviewExport {
    /** exportSvg shows a save dialog and writes [svg]; returns true when written. */
    fun exportSvg(project: Project, svg: String): Boolean {
        val descriptor = FileSaverDescriptor("Export SVG", "Save the rendered diagram as an .svg file", "svg")
        val dialog = FileChooserFactory.getInstance().createSaveFileDialog(descriptor, project)

        val wrapper = dialog.save(null as VirtualFile?, "diagram.svg") ?: return false

        val file = wrapper.file ?: return false

        file.parentFile?.mkdirs()

        return try {
            file.writeText(svg)
            true
        } catch (e: Exception) {
            Logger.getInstance(PreviewExport::class.java).warn("Failed to write SVG export", e)

            false
        }
    }
}
