// types.ts mirrors the Go backend's JSON contracts (gui/internal/app).

export interface FileInfo {
  path: string;
}

export interface ProjectInfo {
  dir: string;
  files: FileInfo[];
}

export interface FileContent {
  path: string;
  text: string;
}

export interface DiagRange {
  start: { line: number; character: number };
  end: { line: number; character: number };
}

export interface Diag {
  range?: DiagRange;
  severity?: number; // 1 error, 2 warning, ...
  source?: string;
  message: string;
}

export interface DiagnosticsEvent {
  path: string;
  version?: number;
  diagnostics: Diag[];
}

export interface Breadcrumb {
  name: string;
  target: string;
}

export interface RenderOptions {
  target: string;
  allExpanded: boolean;
  expanded?: string[]; // undefined = model default; [] = collapse all
  legend?: boolean | null;
}

export interface RenderResult {
  svg: string;
  diagnostics: Diag[];
  target: string;
  allExpanded: boolean;
  breadcrumbs: Breadcrumb[];
}

export interface CompletionItem {
  label: string;
  kind?: number;
  detail?: string;
  documentation?: string;
  sortText?: string;
  filterText?: string;
  insertText?: string;
}

export interface CompletionList {
  isIncomplete: boolean;
  items: CompletionItem[];
}

export interface HoverResult {
  hover: {
    contents: { kind: string; value: string };
    range?: DiagRange;
  } | null;
}

export interface DefinitionResult {
  path: string;
  range: DiagRange;
}

export interface FormatResult {
  text: string;
}

export interface ExportResult {
  format: string;
  files: string[];
}

export interface AppInfo {
  initialDir: string;
  version: string;
}

// --- P1: AI chat -----------------------------------------------------------

/** Chat providers (issue #36): the backend dispatches on this field. */
export type ChatProvider = "openai-compatible" | "anthropic";

export interface ChatSettings {
  provider: ChatProvider;
  baseURL: string;
  model: string;
  apiKey: string;
  systemPrompt: string;
}

/** PublicChatSettings never carries the API key back to the UI. */
export interface PublicChatSettings {
  provider: ChatProvider;
  baseURL: string;
  model: string;
  hasAPIKey: boolean;
  systemPrompt: string;
}

/** ChatConfigResult is the chatConfig response: every provider slot (keys
 * masked) plus which provider the next chat request will use. */
export interface ChatConfigResult {
  provider: ChatProvider;
  config: PublicChatSettings;
  providers: Partial<Record<ChatProvider, PublicChatSettings>>;
  defaultFull: string;
}

export interface ChatMessage {
  role: "user" | "assistant" | "system";
  content: string;
}

export interface ProposedEdit {
  path: string;
  newContent: string;
  oldContent: string;
  valid: boolean;
  error: string;
}

export interface ChatEvent {
  requestID: string;
  delta?: string;
  done?: boolean;
  error?: string;
  proposals?: ProposedEdit[];
  answer?: string;
}

export interface DiffLine {
  kind: "ctx" | "add" | "del" | "hunk";
  text: string;
}

export interface ApplyEditsResult {
  applied: string[];
  errors: string[];
}
