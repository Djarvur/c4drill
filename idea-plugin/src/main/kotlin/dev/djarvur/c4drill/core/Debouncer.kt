// Debounce for the live preview refresh (issue #29 scope update: "debounced
// re-render on edit, ~150-250 ms"). The scheduling policy is pure and
// testable: call [Debouncer.trigger] for every document change; only the
// last trigger within the window fires [task], after a quiet period.
//
// The platform-facing wrapper wires [schedule] to a scheduled executor.

package dev.djarvur.c4drill.core

import java.util.concurrent.atomic.AtomicLong

/**
 * Debouncer coalesces bursts of triggers into one delayed call.
 *
 * [schedule] is the injection seam that actually arranges for [task] to run
 * after `delayMillis` of quiet time; the [Cancelable] it returns is invoked
 * when the trigger is superseded or canceled. The generation guard keeps a
 * late-firing run suppressed even if the underlying scheduler cannot cancel.
 */
class Debouncer(
    private val delayMillis: Long,
    private val schedule: (task: Runnable, delayMillis: Long) -> Cancelable = { _, _ ->
        throw UnsupportedOperationException("Debouncer.schedule not wired")
    },
) {
    /** Cancelable represents a scheduled (not yet run) task; cancel() must be idempotent. */
    fun interface Cancelable {
        fun cancel()
    }

    private val generation = AtomicLong(0)

    /** trigger registers a demand to run the task; only the last demand in a quiet window executes. */
    fun trigger(task: Runnable): Cancelable {
        val gen = generation.incrementAndGet()

        val scheduled = schedule({
            if (generation.get() == gen) {
                task.run()
            }
        }, delayMillis)

        return Cancelable {
            // Invalidate this trigger: anything at or before gen is suppressed.
            generation.accumulateAndGet(gen) { current, g -> maxOf(current, g + 1) }
            scheduled.cancel()
        }
    }
}
