// The C4D language and file type (issue #29): claimed unconditionally —
// every .c4d document gets the C4Drill editor experience. Highlighting is
// delivered by the TextMate grammar reused verbatim from the VS Code
// extension (#27), see textmate.xml and C4dTextMateBundleProvider.

package dev.djarvur.c4drill

import com.intellij.icons.AllIcons
import com.intellij.lang.Language
import com.intellij.openapi.fileTypes.LanguageFileType
import com.intellij.openapi.util.IconLoader
import javax.swing.Icon

object C4drillIcons {
    @JvmField
    val C4DRILL: Icon = IconLoader.getIcon("/icons/c4drill.svg", C4drillIcons::class.java)

    @JvmField
    val TOOL_WINDOW: Icon = AllIcons.Nodes.Module
}

class C4dLanguage private constructor() : Language("C4D") {
    companion object {
        @JvmStatic
        val INSTANCE: C4dLanguage = C4dLanguage()
    }
}

class C4dFileType private constructor() : LanguageFileType(C4dLanguage.INSTANCE) {
    override fun getName(): String = "C4D"

    override fun getDescription(): String = "C4Drill C4D architecture model"

    override fun getDefaultExtension(): String = "c4d"

    override fun getIcon(): Icon = C4drillIcons.C4DRILL

    companion object {
        @JvmStatic
        val INSTANCE: C4dFileType = C4dFileType()
    }
}

/** The single TextMate scope name the #27 grammar roots at (documentation anchor). */
const val C4D_TEXTMATE_SCOPE: String = "source.c4d"
