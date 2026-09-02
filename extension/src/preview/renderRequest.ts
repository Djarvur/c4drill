// c4drill/renderDiagram request plumbing (issue #27): the custom method the
// Go server implements (internal/lsp render.go, wire contract on
// RenderDiagramParams / RenderDiagramResult). The result rides the pipeline
// diagnostics along with the SVG; validation failures return success with an
// empty svg and the same messages the CLI prints.

import { Diagnostic, RequestType } from 'vscode-languageclient';

import type { RenderDiagramParams } from '../core/renderParams';

export interface RenderDiagramResult {
    svg: string;
    diagnostics: Diagnostic[];
}

export const RenderDiagramRequest = new RequestType<RenderDiagramParams, RenderDiagramResult | null, void>(
    'c4drill/renderDiagram',
);
