package dev.djarvur.c4drill.core

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class GlobMatcherTest {
    @Test
    fun `single star stays within one segment`() {
        assertTrue(globMatches("/work/models/main.toml", "models/*.toml"))
        assertFalse(globMatches("/work/models/sub/main.toml", "models/*.toml"))
    }

    @Test
    fun `double star crosses segments`() {
        assertTrue(globMatches("/a/b/c/d.toml", "**/*.toml"))
        assertTrue(globMatches("/a.toml", "**/*.toml"))
        assertTrue(globMatches("x/y/z.c4d", "x/**/*.c4d"))
        assertFalse(globMatches("y/z.c4d", "x/**/*.c4d"))
    }

    @Test
    fun `question mark matches one non-separator char`() {
        assertTrue(globMatches("/w/main1.toml", "main?.toml"))
        assertFalse(globMatches("/w/main12.toml", "main?.toml"))
        assertFalse(globMatches("/w/ma/n1.toml", "main?.toml"))
    }

    @Test
    fun `character classes and braces`() {
        assertTrue(globMatches("/w/main-a.toml", "main-[ab].toml"))
        assertFalse(globMatches("/w/main-c.toml", "main-[ab].toml"))
        assertTrue(globMatches("/w/main-b.toml", "main-[a-c].toml"))
        assertFalse(globMatches("/w/main-d.toml", "main-[a-c].toml"))
        assertTrue(globMatches("/w/main-c.toml", "main-[!ab].toml"))
        assertFalse(globMatches("/w/main-a.toml", "main-[!ab].toml"))
        assertTrue(globMatches("/w/prod.toml", "{prod,staging}.toml"))
        assertFalse(globMatches("/w/dev.toml", "{prod,staging}.toml"))
    }

    @Test
    fun `suffix matching works at a segment boundary`() {
        assertTrue(matchGlob("models/*.toml", listOf("/work/models/main.toml")))
        assertTrue(matchGlob("models/*.toml", listOf("work/models/main.toml")))
        // ".../smodels/main.toml" must NOT match "models/*.toml" at a non-boundary.
        assertFalse(matchGlob("models/*.toml", listOf("/work/smodels/main.toml")))
    }

    @Test
    fun `empty pattern matches nothing`() {
        assertFalse(matchGlob("", listOf("/anything.toml")))
    }
}

class DebouncerTest {
    private class Harness {
        var runs = 0
        val pendings = mutableListOf<Pair<Runnable, Long>>()
        var cancelCount = 0
        var lastPending: Runnable? = null
        var lastCancelable: Debouncer.Cancelable? = null

        val debouncer = Debouncer(delayMillis = 200) { task, delay ->
            pendings.add(task to delay)
            lastPending = task
            Debouncer.Cancelable { cancelCount++ }
        }
    }

    @Test
    fun `only the last trigger in the quiet window runs`() {
        val h = Harness()

        h.debouncer.trigger { h.runs++ }
        h.debouncer.trigger { h.runs++ }
        h.debouncer.trigger { h.runs++ }

        // Simulate every scheduled callback firing anyway: the generation
        // guard keeps superseded tasks from running.
        h.pendings.forEach { (task, _) -> task.run() }

        assertEquals(1, h.runs)
        assertEquals(200L, h.pendings.last().second)
    }

    @Test
    fun `canceling the last trigger suppresses its run`() {
        val h = Harness()

        h.debouncer.trigger { h.runs++ }
        val second = h.debouncer.trigger { h.runs++ }

        second.cancel()

        assertEquals(1, h.cancelCount)

        // The scheduler still fired the task (cancel came too late there) —
        // but the generation guard suppresses it.
        h.lastPending?.run()

        assertEquals(0, h.runs)
    }
}
