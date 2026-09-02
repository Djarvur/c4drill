// Message protocol between the preview webview and the extension host
// (issue #27 scope update). Pure types — no VS Code imports — shared by the
// panel controller and the injected webview script.

import type { PreviewViewState } from '../core/renderParams';

// ---- webview -> host ----

export interface LinkMessage {
    type: 'link';
    /** Raw href attribute value from the SVG (relative .svg or external URL). */
    href: string;
}

export interface DrillToMessage {
    type: 'drillTo';
    /** Dotted render target ('' = C1 root breadcrumb). */
    target: string;
}

export interface ToggleExpandedMessage {
    type: 'toggleExpanded';
}

export interface CollapseAllMessage {
    type: 'collapseAll';
}

export interface ReloadMessage {
    type: 'reload';
}

export interface LegendMessage {
    type: 'legend';
    value: 'default' | 'on' | 'off';
}

export interface ExpandedTextMessage {
    type: 'expandedText';
    value: string;
}

export interface ExportMessage {
    type: 'export';
    format: string;
}

export interface ReadyMessage {
    type: 'ready';
}

export type WebviewToHostMessage =
    | LinkMessage
    | DrillToMessage
    | ToggleExpandedMessage
    | CollapseAllMessage
    | ReloadMessage
    | LegendMessage
    | ExpandedTextMessage
    | ExportMessage
    | ReadyMessage;

// ---- host -> webview ----

export interface DiagnosticItem {
    /** CLI-identical diagnostic message (from the renderDiagram response). */
    message: string;
    /** Zero-based line from the diagnostic range; -1 when absent. */
    line: number;
    source: string;
}

export interface RenderedMessage {
    type: 'rendered';
    /** Render sequence number; webviews drop messages older than the last seen. */
    seq: number;
    svg: string;
    target: string;
    allExpanded: boolean;
    title: string;
}

export interface ErrorMessage {
    type: 'error';
    seq: number;
    /** Headline explaining which stage failed. */
    reason: string;
    diagnostics: DiagnosticItem[];
    target: string;
    allExpanded: boolean;
}

export interface StateMessage {
    type: 'state';
    state: PreviewViewState;
}

export interface InfoMessage {
    type: 'info';
    seq: number;
    text: string;
}

export type HostToWebviewMessage = RenderedMessage | ErrorMessage | StateMessage | InfoMessage;
