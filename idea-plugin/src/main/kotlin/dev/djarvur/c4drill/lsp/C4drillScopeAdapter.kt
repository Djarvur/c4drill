// Adapter that bridges the pure scoping decision (core/DocumentScope) to
// IntelliJ VirtualFiles. Shared by the LSP provider and the preview/actions.

package dev.djarvur.c4drill

import com.intellij.openapi.project.Project
import com.intellij.openapi.roots.ProjectRootManager
import com.intellij.openapi.vfs.VirtualFile
import dev.djarvur.c4drill.core.ScopeDecisionInput
import dev.djarvur.c4drill.core.isManagedDocument
import java.nio.file.Paths

/** isManagedFile reports whether the c4drill language server should attach to this document. */
fun isManagedFile(project: Project, file: VirtualFile): Boolean {
    if (!file.isInLocalFileSystem) {
        return false
    }

    val fsPath = try {
        Paths.get(file.path).normalize().toString()
    } catch (e: Exception) {
        file.path
    }

    val projectDir = ProjectRootManager.getInstance(project).contentRoots.firstOrNull()?.path

    val relativePath = projectDir?.let { dir ->
        val normalizedDir = dir.trimEnd('/')
        if (fsPath.startsWith("$normalizedDir/")) fsPath.substring(normalizedDir.length + 1) else null
    }

    val settings = C4drillProjectSettings.getInstance(project)

    return isManagedDocument(
        ScopeDecisionInput(
            fsPath = fsPath,
            relativePath = relativePath,
            patterns = settings.tomlPatterns,
            activatedPaths = settings.activatedPaths,
        ),
    )
}
