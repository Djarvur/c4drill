// Glob matching for the c4drill TOML scoping rule (issue #29; mirrors the VS
// Code extension's src/core/globMatch.ts from #27). Supports the glob subset
// users need for path patterns: `**` (any path segments), `*` (within one
// segment), `?` (single non-separator char), `[abc]`/`[a-z]` character
// classes (with `!` negation), and `{a,b}` brace alternation.
//
// Pure logic, no IntelliJ imports, so it is unit-testable with plain JUnit.

package dev.djarvur.c4drill.core

/**
 * matchGlob reports whether any of the candidate paths matches the pattern.
 * Candidates are '/'-separated, normalized paths (absolute or relative); the
 * pattern matches when it equals a candidate or matches its path suffix at a
 * segment boundary (so a "models/<one-segment>.toml" pattern matches the
 * absolute "/work/models/main.toml").
 */
fun matchGlob(pattern: String, candidates: List<String>): Boolean =
    candidates.any { globMatches(it, pattern) }

/** globMatches checks one pattern against one normalized '/'-separated path. */
fun globMatches(path: String, pattern: String): Boolean {
    if (pattern.isEmpty()) {
        return false
    }

    val regex = globToRegex(pattern)

    if (regex.matches(path)) {
        return true
    }

    // Segment-boundary suffix match: "models/*.toml" matches
    // "/work/models/main.toml" and "work/models/main.toml".
    val segments = path.split('/').filter { it.isNotEmpty() }

    for (i in segments.indices) {
        val candidate = segments.subList(i, segments.size).joinToString("/")

        if (regex.matches(candidate)) {
            return true
        }
    }

    return false
}

private val globCache = java.util.concurrent.ConcurrentHashMap<String, Regex>()

/** globToRegex translates a glob pattern into a regex (see file comment for the supported subset). */
fun globToRegex(pattern: String): Regex = globCache.computeIfAbsent(pattern) { compileGlob(it) }

private fun compileGlob(pattern: String): Regex {
    val re = StringBuilder()
    var i = 0
    val n = pattern.length

    while (i < n) {
        val c = pattern[i]

        when {
            c == '*' && i + 1 < n && pattern[i + 1] == '*' -> {
                // `**` — any number of path segments (or none).
                val doubleStarEndsSegment = i + 2 >= n || pattern[i + 2] == '/'
                if (doubleStarEndsSegment) {
                    re.append("(?:[^/]*/)*")
                    i += if (i + 2 < n) 3 else 2
                } else {
                    // `**` inside a segment (e.g. `a**b`) degrades to `*`.
                    re.append("[^/]*")
                    i += 2
                }
            }

            c == '*' -> {
                re.append("[^/]*")
                i++
            }

            c == '?' -> {
                re.append("[^/]")
                i++
            }

            c == '[' -> {
                val close = pattern.indexOf(']', i + 1)
                if (close < 0) {
                    re.append(Regex.escape(c.toString()))
                    i++
                } else {
                    var body = pattern.substring(i + 1, close)
                    val negated = body.startsWith("!") || body.startsWith("^")
                    if (negated) body = body.substring(1)
                    // Escape regex metacharacters but keep ranges intact (`-`
                    // stays literal so `[a-z]` keeps working as a range).
                    val escaped = body.replace(Regex("([\\\\^\\]])"), "\\\\$1")
                    re.append(if (negated) "[^$escaped]" else "[$escaped]")
                    i = close + 1
                }
            }

            c == '{' -> {
                val close = pattern.indexOf('}', i + 1)
                if (close < 0) {
                    re.append(Regex.escape(c.toString()))
                    i++
                } else {
                    val body = pattern.substring(i + 1, close)
                    val alternatives = body.split(',').joinToString("|") { Regex.escape(it) }
                    re.append("(?:").append(alternatives).append(")")
                    i = close + 1
                }
            }

            else -> {
                re.append(Regex.escape(c.toString()))
                i++
            }
        }
    }

    return Regex(re.toString())
}
