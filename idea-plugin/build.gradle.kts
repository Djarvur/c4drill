// C4Drill JetBrains plugin build (issue #29).
//
// Targets the IntelliJ Platform 2025.3 line (since-build 253): the first
// platform release where the bundled LSP client is available to third-party
// plugins in every commercial IDE AND surfaces textDocument/documentSymbol in
// the Structure view (2025.3 addition), and where the unified IntelliJ IDEA
// distribution ships LSP support to all users.

plugins {
    id("java")
    kotlin("jvm") version "2.2.20"
    id("org.jetbrains.intellij.platform") version "2.18.1"
}

group = "dev.djarvur"
version = "0.1.0"

kotlin {
    jvmToolchain(21)
}

repositories {
    mavenCentral()
    intellijPlatform {
        defaultRepositories()
    }
}

dependencies {
    intellijPlatform {
        // Unified IntelliJ IDEA 2025.3 distribution: embeds the platform LSP
        // client (module com.intellij.modules.lsp) our client is built on.
        //
        // C4DRILL_IDEA_HOME optionally points at a local 2025.3 distribution
        // (e.g. an extracted .app) for environments where
        // download.jetbrains.com is unreachable; when unset the standard
        // remote dependency is used.
        val localIde = providers.environmentVariable("C4DRILL_IDEA_HOME").orNull

        if (localIde.isNullOrBlank()) {
            intellijIdeaUltimate("2025.3")
        } else {
            local(localIde)
        }

        // Platform LSP client (embedded module, all commercial IDEs 253+).
        bundledModule("intellij.platform.lsp.impl")
        // lsp4j wire types (platform library the LSP client speaks).
        bundledLibrary("lib.eclipse.lsp4j")
        // TextMate bundles support for the C4D grammar.
        bundledPlugin("org.jetbrains.plugins.textmate")
        pluginVerifier()
    }

    testImplementation("junit:junit:4.13.2")
}

intellijPlatform {
    // Plain-JVM Kotlin plugin: no forms/sequence instrumentation needed, and
    // skipping it avoids pulling the remote JetBrains java-compiler-ant-tasks
    // artifact in restricted-network environments.
    instrumentCode = false

    pluginConfiguration {
        id = "dev.djarvur.c4drill"
        name = "C4Drill"
        version = project.version.toString()
        description = """
            C4Drill architecture-as-code support: syntax highlighting, autocomplete, diagnostics, formatting and a live
            diagram preview for the c4drill TOML dialect and the C4D format, powered by the <code>c4drill</code> language
            server (<code>c4drill serve --lsp</code>).
        """.trimIndent()

        ideaVersion {
            sinceBuild = "253"
            // Left open: the plugin only uses platform APIs documented as
            // stable across the 253+ releases it is tested against.
            untilBuild = provider { null }
        }
    }

    pluginVerification {
        ides {
            recommended()
        }
    }
}

tasks {
    wrapper {
        gradleVersion = "9.7.1"
    }
}
