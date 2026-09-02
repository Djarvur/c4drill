// View state and c4drill/renderDiagram parameter construction for the live
// preview panel (issue #27 scope update). The wire contract is documented on
// internal/lsp RenderDiagramParams / RenderDiagramResult:
//
//	{ textDocument: { uri }, target, allExpanded, expanded, legend, format }
//
//	target       "" = C1 context, one segment = that unit's C2, deeper = C3
//	allExpanded  the single all-nested diagram (--expanded); overrides target
//	expanded     replacement for [properties].expanded C1 drill-down set;
//	             an EMPTY array is meaningful ("collapse all") — the field is
//	             deliberately sent even when empty, while "model default" is
//	             expressed by OMITTING the field (undefined here)
//	legend       true/false overrides properties.legend; undefined = default
//	format       only "svg" in v1
//
// Pure logic, no VS Code imports, so it is unit-testable with node:test.

export type LegendChoice = 'default' | 'on' | 'off';

export interface PreviewViewState {
    /** Render target path; '' is the C1 context diagram. */
    readonly target: string;
    /** All-expanded diagram mode (the CLI --expanded mode); overrides target. */
    readonly allExpanded: boolean;
    /** Comma/whitespace-separated unit paths replacing [properties].expanded; empty = model default. */
    readonly expandedText: string;
    /**
     * Explicit "collapse all" override: sends expanded: [] (an EMPTY array is
     * meaningful on the wire — the server treats it as "collapse everything").
     * Takes precedence over expandedText; cleared as soon as the user edits
     * the expanded-set input.
     */
    readonly collapseAll: boolean;
    readonly legend: LegendChoice;
}

export const initialViewState: PreviewViewState = {
    target: '',
    allExpanded: false,
    expandedText: '',
    collapseAll: false,
    legend: 'default',
};

export interface RenderDiagramParams {
    textDocument: { uri: string };
    target?: string;
    allExpanded?: boolean;
    /** Absent = model default; present (possibly []) = explicit override. */
    expanded?: string[];
    legend?: boolean;
    format?: string;
}

export const renderDiagramFormat = 'svg';

// buildRenderParams maps view state + document URI to the renderDiagram
// request params, honoring the empty-array-means-collapse-all contract.
export function buildRenderParams(uri: string, state: PreviewViewState): RenderDiagramParams {
    const params: RenderDiagramParams = {
        textDocument: { uri },
        format: renderDiagramFormat,
    };

    if (state.allExpanded) {
        params.allExpanded = true;
    } else {
        params.target = state.target;
    }

    if (state.collapseAll) {
        // The wire contract: an empty array IS the "collapse all" override.
        params.expanded = [];
    } else {
        const expanded = parseExpandedSet(state.expandedText);
        if (expanded !== undefined) {
            params.expanded = expanded;
        }
    }

    if (state.legend === 'on') {
        params.legend = true;
    } else if (state.legend === 'off') {
        params.legend = false;
    }

    return params;
}

// parseExpandedSet splits the expanded-set input on commas and whitespace.
// Empty/blank input yields undefined (model default); non-empty input always
// yields an array (possibly empty entries are filtered — explicit entries
// only). An input of only separators (",,") therefore resets to the model
// default, while listing zero units is not expressible via the text box
// (the UI's "Collapse all" button sends the empty-array override instead).
export function parseExpandedSet(text: string): string[] | undefined {
    const entries = text
        .split(/[\s,]+/)
        .map((s) => s.trim())
        .filter((s) => s !== '');

    return entries.length === 0 ? undefined : entries;
}

// breadcrumbTrail returns the clickable ancestor chain for the current
// target: the C1 root entry plus one entry per dotted segment. The last
// entry is the current view (not navigable).
export interface BreadcrumbEntry {
    readonly label: string;
    readonly target: string;
    readonly current: boolean;
}

export function breadcrumbTrail(state: PreviewViewState): BreadcrumbEntry[] {
    const trail: BreadcrumbEntry[] = [
        { label: 'C1', target: '', current: !state.allExpanded && state.target === '' },
    ];

    if (state.allExpanded) {
        trail.push({ label: 'expanded', target: '', current: true });

        return trail;
    }

    if (state.target === '') {
        return trail;
    }

    const segments = state.target.split('.');

    for (let i = 0; i < segments.length; i++) {
        const t = segments.slice(0, i + 1).join('.');

        trail.push({ label: segments[i], target: t, current: i === segments.length - 1 });
    }

    return trail;
}
