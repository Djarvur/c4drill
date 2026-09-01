package dev.djarvur.c4drill.core

import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

/** Scoping rule tests: .c4d always managed, .toml opt-in only, everything else untouched. */
class DocumentScopeTest {
    @Test
    fun `c4d files are always managed`() {
        assertTrue(
            isManagedDocument(
                ScopeDecisionInput(fsPath = "/work/models/main.c4d"),
            ),
        )
        assertTrue(
            isManagedDocument(
                ScopeDecisionInput(fsPath = "/work/other.C4D"),
            ),
        )
    }

    @Test
    fun `plain toml is untouched without patterns`() {
        assertFalse(isManagedDocument(ScopeDecisionInput(fsPath = "/work/go.mod.toml")))
        assertFalse(
            isManagedDocument(
                ScopeDecisionInput(fsPath = "/work/Cargo.toml", patterns = emptyList()),
            ),
        )
    }

    @Test
    fun `toml with a matching relative pattern is managed`() {
        assertTrue(
            isManagedDocument(
                ScopeDecisionInput(
                    fsPath = "/work/models/main.toml",
                    relativePath = "models/main.toml",
                    patterns = listOf("models/*.toml"),
                ),
            ),
        )
    }

    @Test
    fun `toml with an absolute pattern is managed`() {
        assertTrue(
            isManagedDocument(
                ScopeDecisionInput(
                    fsPath = "/work/models/main.toml",
                    patterns = listOf("/work/models/*.toml"),
                ),
            ),
        )
    }

    @Test
    fun `non-matching toml stays untouched`() {
        assertFalse(
            isManagedDocument(
                ScopeDecisionInput(
                    fsPath = "/work/src/config.toml",
                    relativePath = "src/config.toml",
                    patterns = listOf("models/*.toml"),
                ),
            ),
        )
    }

    @Test
    fun `explicit activation wins over everything`() {
        val activated = setOf("/work/Cargo.toml")

        assertTrue(isManagedDocument(ScopeDecisionInput(fsPath = "/work/Cargo.toml", activatedPaths = activated)))
        // ...but only for the activated path.
        assertFalse(isManagedDocument(ScopeDecisionInput(fsPath = "/work/other.toml", activatedPaths = activated)))
    }

    @Test
    fun `double star patterns match nested models`() {
        assertTrue(
            isManagedDocument(
                ScopeDecisionInput(
                    fsPath = "/work/arch/systems/main.toml",
                    relativePath = "arch/systems/main.toml",
                    patterns = listOf("**/*.toml"),
                ),
            ),
        )
    }

    @Test
    fun `only toml and c4d extensions are considered`() {
        assertFalse(isManagedDocument(ScopeDecisionInput(fsPath = "/work/models/main.yaml", patterns = listOf("**/*"))))
        assertFalse(isManagedDocument(ScopeDecisionInput(fsPath = "/work/models/c4dtoml")))
    }
}
