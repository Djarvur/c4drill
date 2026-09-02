// Grammar smoke test (issue #27): loads syntaxes/c4d.tmLanguage.json through
// vscode-textmate (the same tokenizer engine VS Code uses) and asserts that
// the constructs from the real example models tokenize into the expected
// scopes. The grammar is the shared artifact other clients reference, so its
// behavior is pinned here.

import { test, before } from 'node:test';
import * as assert from 'node:assert/strict';
import * as fs from 'node:fs';
import * as path from 'node:path';

import type { IGrammar, IToken, StateStack } from 'vscode-textmate';

let grammar: IGrammar;

// Line tokenizer with a per-line token cache so assertions can ask for the
// scopes covering a substring (the way humans read the model text).
interface Line {
    text: string;
    scopesAt(substr: string): string[];
    tokens(): IToken[];
}

function buildLine(text: string, tokens: IToken[]): Line {
    return {
        text,
        tokens: () => tokens,
        scopesAt: (substr: string): string[] => {
            const idx = text.indexOf(substr);
            assert.notEqual(idx, -1, `substring not found on line: ${JSON.stringify(text)} — ${substr}`);

            for (const t of tokens) {
                if (idx >= t.startIndex && idx < t.endIndex) {
                    return t.scopes;
                }
            }

            return [];
        },
    };
}

async function tokenize(text: string): Promise<Line[]> {
    const out: Line[] = [];
    let stack: StateStack | null = null;

    for (const lineText of text.split('\n')) {
        const res = grammar.tokenizeLine(lineText, stack);
        stack = res.ruleStack;
        out.push(buildLine(lineText, res.tokens));
    }

    return out;
}

before(async () => {
    const vscodeOniguruma = await import('vscode-oniguruma');
    const wasm = fs.readFileSync(require.resolve('vscode-oniguruma/release/onig.wasm'));

    await vscodeOniguruma.loadWASM(wasm);

    const vscodeTextmate = await import('vscode-textmate');
    const grammarPath = path.resolve(__dirname, '..', '..', '..', 'syntaxes', 'c4d.tmLanguage.json');
    const registry = new vscodeTextmate.Registry({
        onigLib: Promise.resolve({
            createOnigScanner: (sources: string[]) => new vscodeOniguruma.OnigScanner(sources),
            createOnigString: (s: string) => new vscodeOniguruma.OnigString(s),
        }),
        loadGrammar: async (scopeName: string) => {
            if (scopeName !== 'source.c4d') {
                return null;
            }

            return vscodeTextmate.parseRawGrammar(fs.readFileSync(grammarPath, 'utf8'), grammarPath);
        },
    });

    const loaded: IGrammar | null = await registry.loadGrammar('source.c4d');
    if (loaded === null) {
        throw new Error('source.c4d grammar must load');
    }

    grammar = loaded;
});

test('unit header: id, type, external modifier and quoted name', async () => {
    const [line] = await tokenize('monitoring: system external "Monitoring Service" {');

    assert.ok(line.scopesAt('monitoring').includes('entity.name.unit.c4d'), 'id');
    assert.ok(line.scopesAt('system').includes('storage.type.unit.c4d'), 'type');
    assert.ok(line.scopesAt('external').includes('storage.modifier.external.c4d'), 'external');
    assert.ok(line.scopesAt('Monitoring Service').includes('string.quoted.double.c4d'), 'name');
});

test('properties block and its fields', async () => {
    const lines = await tokenize('properties {\n  name: Styling Demo\n  edges: spline\n}');

    assert.ok(lines[0].scopesAt('properties').includes('keyword.other.properties.c4d'), 'keyword');
    assert.ok(lines[1].scopesAt('name').includes('support.type.property-name.c4d'), 'field key');
    assert.ok(lines[2].scopesAt('spline').includes('constant.language.c4d'), 'enum value');
});

test('edge statement: arrow, dotted peer, label, inline option block', async () => {
    const lines = await tokenize([
        'web: system "Web" {',
        '  -> api.auth: "HTTPS | Validates" { color: "#1565C0" style: dashed }',
        '}',
    ].join('\n'));

    assert.ok(lines[1].scopesAt('->').includes('keyword.operator.arrow.c4d'), 'arrow');
    assert.ok(lines[1].scopesAt('api.auth').includes('entity.other.attribute-name.peer.c4d'), 'peer');
    assert.ok(lines[1].scopesAt('Validates').includes('string.quoted.double.c4d'), 'label');
    assert.ok(lines[1].scopesAt('color').includes('support.type.property-name.c4d'), 'option key');
    assert.ok(lines[1].scopesAt('dashed').includes('constant.language.c4d'), 'enum');
    // The option block's closing brace must not terminate the unit block:
    // the unit's closing brace on the next line is the block end.
    assert.ok(lines[1].tokens().some((t) => t.scopes.includes('punctuation.section.block.end.c4d')), 'inline block end');
});

test('comments, include, template, use and ${param} tokens', async () => {
    const lines = await tokenize([
        '# a comment',
        'include templates.c4d once',
        'template microservice(name, tech) {',
        '  type: container',
        '  name: "${name} Service"',
        '  use dataService(name: users, tech: Go)',
        '}',
        'properties {',
        '  legendLine: ["Nightly batch|#E65100|dashed"]',
        '}',
    ].join('\n'));

    assert.ok(lines[0].scopesAt('# a comment').includes('comment.line.number-sign.c4d'), 'comment');
    assert.ok(lines[1].scopesAt('include').includes('keyword.control.include.c4d'), 'include');
    assert.ok(lines[1].scopesAt('templates.c4d').includes('string.unquoted.include-path.c4d'), 'path');
    assert.ok(lines[1].scopesAt('once').includes('keyword.modifier.once.c4d'), 'once');
    assert.ok(lines[2].scopesAt('template').includes('keyword.declaration.template.c4d'), 'template');
    assert.ok(lines[2].scopesAt('name, tech').includes('variable.parameter.c4d'), 'params');
    assert.ok(lines[3].scopesAt('container').includes('storage.type.unit.c4d'), 'template type stmt');
    assert.ok(lines[4].scopesAt('${name}').includes('variable.other.template-param.c4d'), 'template token in string');
    assert.ok(lines[5].scopesAt('use').includes('keyword.control.use.c4d'), 'use');
});

test('triple-quoted multi-line string spans lines', async () => {
    const lines = await tokenize('desc: """first\nsecond"""\nedges: spline');

    assert.ok(lines[0].scopesAt('"""first').some((s) => s.startsWith('string.quoted.double.multiline')), 'first line');
    assert.ok(lines[1].scopesAt('second').some((s) => s.startsWith('string.quoted.double.multiline')), 'second line');
    assert.ok(lines[2].scopesAt('spline').includes('constant.language.c4d'), 'code after the string');
});

test('all 17 unit type keywords tokenize as storage.type.unit', async () => {
    const types = [
        'person', 'personExternal', 'system', 'systemExternal', 'db', 'dbExternal',
        'queue', 'queueExternal', 'box',
        'container', 'containerDb', 'containerQueue', 'containerBox',
        'component', 'componentDb', 'componentQueue', 'componentBox',
    ];

    assert.equal(types.length, 17, 'the 17 types');

    const text = types.map((t, i) => `u${i}: ${t} "N${i}" {\n}`).join('\n');
    const lines = await tokenize(text);

    for (let i = 0; i < types.length; i++) {
        assert.ok(
            lines[i * 2].scopesAt(types[i]).includes('storage.type.unit.c4d'),
            `type ${types[i]}`,
        );
        assert.ok(
            lines[i * 2].tokens().some((t) => t.scopes.includes('punctuation.section.block.begin.c4d')),
            `opening brace for ${types[i]}`,
        );
    }
});
