// Document scoping for the c4drill TOML dialect (issue #29; mirrors the VS
// Code extension's src/core/documentScope.ts from #27): the c4drill language
// server must NOT attach to unrelated TOML files, and IntelliJ ships its own
// TOML plugin which keeps handling plain TOML.
//
// Scoping rule:
//   - .c4d documents are always handled (the language is claimed unconditionally);
//   - .toml documents are handled only when they opt in, either via the
//     configured glob patterns (c4drill.toml.patterns) or via the explicit
//     "C4Drill: Activate for This File" action (persisted per project);
//   - everything else is untouched.
//
// Pure decision logic, no IntelliJ imports, so it is unit-testable with
// plain JUnit.

package dev.djarvur.c4drill.core

const val c4dExtension: String = "c4d"

/** The user-facing explanation shown when an action hits an unmanaged document. */
const val NOT_HANDLED_MESSAGE: String =
    "C4Drill: this file is not handled by c4drill. For .toml models use \"C4Drill: Activate for This File\" or c4drill.toml.patterns."

data class ScopeDecisionInput(
    /** Absolute, '/'-separated filesystem path of the document. */
    val fsPath: String,
    /** Path relative to the project content root, '/'-separated; null outside a project. */
    val relativePath: String? = null,
    /** Configured c4drill.toml.patterns globs. */
    val patterns: List<String> = emptyList(),
    /** Absolute paths the user activated explicitly via "C4Drill: Activate for This File". */
    val activatedPaths: Set<String> = emptySet(),
)

/** isManagedDocument reports whether the c4drill language server should attach to the document. */
fun isManagedDocument(args: ScopeDecisionInput): Boolean {
    if (args.fsPath in args.activatedPaths) {
        return true
    }

    val ext = extensionOf(args.fsPath)

    if (ext == c4dExtension) {
        return true
    }

    if (ext != "toml" || args.patterns.isEmpty()) {
        return false
    }

    val candidates = mutableListOf(args.fsPath)

    if (args.relativePath != null) {
        candidates.add(0, args.relativePath)
    }

    return args.patterns.any { pattern -> matchGlob(pattern, candidates) }
}

private fun extensionOf(path: String): String {
    val file = path.substringAfterLast('/', missingDelimiterValue = path).substringAfterLast('\\', missingDelimiterValue = path)
    val dot = file.lastIndexOf('.')

    return if (dot > 0) file.substring(dot + 1).lowercase() else ""
}
