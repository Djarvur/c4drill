plugins {
    // Auto-provisions the JDK 21 toolchain (no local JDK needed to build).
    id("org.gradle.toolchains.foojay-resolver-convention") version "1.0.0"
}

rootProject.name = "c4drill-idea-plugin"
