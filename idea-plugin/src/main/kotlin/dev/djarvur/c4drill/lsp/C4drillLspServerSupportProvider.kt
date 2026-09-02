// The LSP integration entry (issue #29): launches `c4drill serve --lsp` for
// managed documents through the platform's built-in LSP client. Registered
// via com.intellij.platform.lsp.serverSupportProvider in lsp.xml, which is
// only parsed when the optional `com.intellij.modules.lsp` dependency is
// present (every commercial IDE on the 253+ platform).

package dev.djarvur.c4drill.lsp

import com.intellij.execution.configurations.GeneralCommandLine
import com.intellij.execution.configurations.PathEnvironmentVariableUtil
import com.intellij.openapi.project.Project
import com.intellij.openapi.vfs.VirtualFile
import com.intellij.platform.lsp.api.LspServerSupportProvider
import com.intellij.platform.lsp.api.LspServerSupportProvider.LspServerStarter
import com.intellij.platform.lsp.api.ProjectWideLspServerDescriptor
import dev.djarvur.c4drill.C4drillGlobalSettings
import dev.djarvur.c4drill.C4dFileType
import dev.djarvur.c4drill.isManagedFile
import org.eclipse.lsp4j.services.LanguageServer

class C4drillLspServerSupportProvider : LspServerSupportProvider {
    override fun fileOpened(project: Project, file: VirtualFile, serverStarter: LspServerStarter) {
        if (!isManagedFile(project, file)) {
            return
        }

        serverStarter.ensureServerStarted(C4drillLspServerDescriptor(project))
    }
}

class C4drillLspServerDescriptor(project: Project) : ProjectWideLspServerDescriptor(project, "c4drill") {
    override fun isSupportedFile(file: VirtualFile): Boolean =
        file.fileType is C4dFileType || isManagedFile(project, file)

    override fun createCommandLine(): GeneralCommandLine {
        val command = resolveServerCommand()

        return GeneralCommandLine(command).withParameters("serve", "--lsp")
    }

    override val lsp4jServerClass: Class<out LanguageServer>
        get() = C4drillLsp4jServer::class.java
}

/**
 * resolveServerCommand finds the c4drill binary: the configured override
 * first, then PATH discovery. Throws with an actionable message when neither
 * yields a usable binary — the platform surfaces the startup failure to the
 * user.
 */
fun resolveServerCommand(): String {
    val configured = C4drillGlobalSettings.getInstance().serverPath.trim()

    if (configured.isNotEmpty()) {
        return configured
    }

    val onPath = PathEnvironmentVariableUtil.findInPath("c4drill")
        ?: PathEnvironmentVariableUtil.findInPath("c4drill.exe")

    return onPath?.absolutePath
        ?: throw IllegalStateException(
            "c4drill binary not found on PATH. Install c4drill or set the binary path in Settings | Tools | C4Drill.",
        )
}
