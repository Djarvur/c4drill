// Minimal VS Code-style glob matcher for `c4drill.toml.patterns` (issue #27:
// a .toml file is a c4drill model only when it opts in via a glob or the
// explicit activate command — the extension must not hijack plain TOML).
//
// Supported syntax (the subset VS Code's own glob implementation covers for
// settings like files.exclude):
//
//	*            any characters except '/'
//	?            any single character except '/'
//	**           any number of path segments (also collapses the adjacent '/')
//	{a,b}        brace alternation (may contain globs)
//	[abc] [a-z]  character classes; [!abc] negation
//
// Pure logic, no VS Code imports, so it is unit-testable with node:test.

export function matchGlob(pattern: string, candidates: string[]): boolean {
    const regexes = globToRegExp(pattern);

    return candidates.some((c) => regexes.some((re) => re.test(c)));
}

// globToRegExp compiles one glob into one or more anchored regexes (brace
// expansion can yield several). The input is expected to use '/' separators.
export function globToRegExp(pattern: string): RegExp[] {
    const expanded = expandBraces(pattern);

    return expanded.map((p) => new RegExp(`^${globSegmentToRegex(p)}$`));
}

// expandBraces expands {a,bc} alternations (non-nested, like VS Code).
export function expandBraces(pattern: string): string[] {
    const open = pattern.indexOf('{');
    if (open === -1) {
        return [pattern];
    }

    const close = matchingBrace(pattern, open);
    if (close === -1) {
        return [pattern]; // unbalanced brace: literal
    }

    const head = pattern.slice(0, open);
    const tail = pattern.slice(close + 1);
    const body = pattern.slice(open + 1, close);

    const out: string[] = [];

    for (const alt of body.split(',')) {
        for (const rest of expandBraces(tail)) {
            out.push(head + alt + rest);
        }
    }

    return out;
}

function matchingBrace(pattern: string, open: number): number {
    let depth = 0;

    for (let i = open; i < pattern.length; i++) {
        switch (pattern[i]) {
            case '\\':
                i++; // skip escaped character

                break;
            case '{':
                depth++;

                break;
            case '}':
                depth--;
                if (depth === 0) {
                    return i;
                }

                break;
            default:
                break;
        }
    }

    return -1;
}

// globSegmentToRegex translates the glob body into a regex source. Patterns
// without a '/' also match a file name at any depth (VS Code "matchBase"
// behavior): `*.arch.toml` matches `models/service.arch.toml`.
function globSegmentToRegex(pattern: string): string {
    const re = translateGlob(pattern);

    // matchBase on the ORIGINAL pattern: the translated regex always
    // contains '/' (inside `[^/]*` classes), so the check cannot use it.
    return pattern.includes('/') ? re : `(?:${re}|.*/${re})`;
}

// translateGlob converts the glob body into a regex source.
function translateGlob(pattern: string): string {
    let re = '';

    for (let i = 0; i < pattern.length; i++) {
        const ch = pattern[i];

        switch (ch) {
            case '*':
                if (pattern[i + 1] === '*') {
                    // Globstar. Consume the '**' plus any following '/' run:
                    // `**/` (including a leading one) crosses zero or more
                    // whole segments; a bare or trailing `**` is `.*`.
                    let j = i + 2;
                    let slashes = 0;
                    while (j < pattern.length && pattern[j] === '/') {
                        j++;
                        slashes++;
                    }

                    if (slashes > 0) {
                        re += '(?:.*/)?';
                    } else {
                        re += '.*';
                    }

                    i = j - 1;
                } else {
                    re += '[^/]*';
                }

                break;
            case '?':
                re += '[^/]';

                break;
            case '[': {
                const cls = charClass(pattern, i);
                if (cls === undefined) {
                    re += '\\[';

                    break;
                }

                re += cls.source;
                i = cls.end;

                break;
            }
            case '\\':
                i++;
                if (i < pattern.length) {
                    re += escapeRegex(pattern[i]);
                }

                break;
            default:
                re += escapeRegex(ch);

                break;
        }
    }

    return re;
}

// charClass parses `[abc]`, `[a-z]`, `[!abc]` starting at the '['.
function charClass(pattern: string, start: number): { source: string; end: number } | undefined {
    let i = start + 1;
    let negated = false;

    if (pattern[i] === '!' || pattern[i] === '^') {
        negated = true;
        i++;
    }

    let body = '';
    // A ']' as the first character is a literal member.
    if (pattern[i] === ']') {
        body += '\\]';
        i++;
    }

    while (i < pattern.length && pattern[i] !== ']') {
        if (pattern[i] === '\\') {
            i++;

            if (i >= pattern.length) {
                return undefined;
            }

            body += escapeRegex(pattern[i]);
        } else if (pattern[i + 1] === '-' && pattern[i + 2] !== undefined && pattern[i + 2] !== ']') {
            body += `${escapeRegex(pattern[i])}-${escapeRegex(pattern[i + 2])}`;
            i += 2;
        } else {
            body += escapeRegex(pattern[i]);
        }

        i++;
    }

    if (i >= pattern.length) {
        return undefined; // unterminated class: literal '['
    }

    return { source: `[${negated ? '^' : ''}${body}]`, end: i };
}

function escapeRegex(ch: string): string {
    return /[a-zA-Z0-9_/-]/.test(ch) ? ch : `\\${ch}`;
}
