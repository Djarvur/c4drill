// Application-level C4Drill settings (issue #29): the c4drill server binary
// location (c4drill.server.path) and the preview debounce window. Blank
// serverPath means "discover `c4drill` on PATH".

package dev.djarvur.c4drill

import com.intellij.openapi.application.ApplicationManager
import com.intellij.openapi.components.PersistentStateComponent
import com.intellij.openapi.components.State
import com.intellij.openapi.components.Storage

@State(name = "C4drillGlobalSettings", storages = [Storage("c4drill.xml")])
class C4drillGlobalSettings : PersistentStateComponent<C4drillGlobalSettings.State> {
    data class State(
        var serverPath: String = "",
        var previewDebounceMs: Int = DEFAULT_PREVIEW_DEBOUNCE_MS,
    )

    private val state = State()

    var serverPath: String
        get() = state.serverPath
        set(value) {
            state.serverPath = value
        }

    var previewDebounceMs: Int
        get() = state.previewDebounceMs
        set(value) {
            state.previewDebounceMs = value.coerceIn(MIN_PREVIEW_DEBOUNCE_MS, MAX_PREVIEW_DEBOUNCE_MS)
        }

    override fun getState(): State = state

    override fun loadState(state: State) {
        this.state.serverPath = state.serverPath
        this.state.previewDebounceMs = state.previewDebounceMs
    }

    companion object {
        const val DEFAULT_PREVIEW_DEBOUNCE_MS: Int = 200
        const val MIN_PREVIEW_DEBOUNCE_MS: Int = 50
        const val MAX_PREVIEW_DEBOUNCE_MS: Int = 5000

        @JvmStatic
        fun getInstance(): C4drillGlobalSettings = ApplicationManager.getApplication().getService(C4drillGlobalSettings::class.java)
    }
}
