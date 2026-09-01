// Project-level C4Drill settings (issue #29): the c4drill TOML scoping globs
// (c4drill.toml.patterns) and the set of files explicitly activated via the
// "C4Drill: Activate for This File" action. Plain TOML files never receive
// c4drill language-server features unless they opt in through one of these.

package dev.djarvur.c4drill

import com.intellij.openapi.components.PersistentStateComponent
import com.intellij.openapi.components.State
import com.intellij.openapi.components.Storage
import com.intellij.openapi.components.StoragePathMacros
import com.intellij.openapi.project.Project

@State(name = "C4drillProjectSettings", storages = [Storage(StoragePathMacros.WORKSPACE_FILE)])
class C4drillProjectSettings : PersistentStateComponent<C4drillProjectSettings.State> {
    class State {
        var tomlPatterns: MutableList<String> = mutableListOf()
        var activatedPaths: MutableSet<String> = mutableSetOf()
    }

    private val state = State()

    /** Glob patterns (absolute or project-relative, '/'-separated) selecting which .toml files are c4drill models. */
    var tomlPatterns: List<String>
        get() = state.tomlPatterns.toList()
        set(value) {
            state.tomlPatterns = value.toMutableList()
        }

    /** Absolute, '/'-separated paths explicitly activated by the user; independent of tomlPatterns. */
    val activatedPaths: Set<String>
        get() = state.activatedPaths.toSet()

    fun activate(path: String): Boolean = state.activatedPaths.add(path)

    fun deactivate(path: String): Boolean = state.activatedPaths.remove(path)

    override fun getState(): State = state

    override fun loadState(state: State) {
        this.state.tomlPatterns = state.tomlPatterns
        this.state.activatedPaths = state.activatedPaths
    }

    companion object {
        @JvmStatic
        fun getInstance(project: Project): C4drillProjectSettings = project.getService(C4drillProjectSettings::class.java)
    }
}
