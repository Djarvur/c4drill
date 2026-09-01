// prompt.go is the chat's context assembly: the SKILL.md-seeded system
// prompt plus the authoring context (open files, diagnostics, selection)
// that makes the assistant answer about the architecture the user is
// actually editing. Pure functions — fully unit-tested.

package ai

import (
	_ "embed"
	"strings"
)

//go:embed skill_seed.md
var skillSeed string

// editProtocol is the structured-edit contract the assistant must follow for
// every proposed change (see edits.go for the parser).
const editProtocol = `
## Proposing edits

When the user asks you to change the architecture code, propose edits as
complete-file replacement blocks embedded in your reply, using exactly this
fenced format (one block per file; the path is relative to the project root):

` + "```" + `c4drill-edit path=systems/auth.toml
<the COMPLETE new content of the file>
` + "```" + `

Rules:
- always emit the full file content, not a fragment;
- never touch files outside the opened project;
- explain what you changed and why in plain text around the blocks.
`

// AuthoringContext is the snapshot of what the user is working on, injected
// into the user message so answers are grounded in the real model.
type AuthoringContext struct {
	// ActiveFile is the project-relative path of the file being edited.
	ActiveFile string `json:"activeFile"`
	// ActiveContent is the current buffer content of ActiveFile.
	ActiveContent string `json:"activeContent"`
	// OpenFiles are the other open buffers (path → content). Their content
	// is included truncated: the include graph usually makes them relevant.
	OpenFiles map[string]string `json:"openFiles,omitempty"`
	// Diagnostics are the active LSP diagnostics, rendered "path: message".
	Diagnostics []string `json:"diagnostics,omitempty"`
	// Selection is the text selected in the editor, if any.
	Selection string `json:"selection,omitempty"`
}

// maxOpenFileChars bounds each non-active open file in the context.
const maxOpenFileChars = 8000

// SystemPrompt is the system message: the c4drill format knowledge (SKILL.md
// snapshot) plus the edit-proposal protocol, optionally extended by the
// user's custom suffix from settings.
func SystemPrompt(customSuffix string) string {
	var b strings.Builder

	b.WriteString("You are the c4drill architecture assistant, embedded in the c4drill ")
	b.WriteString("desktop app. You help the author understand, generate and modify ")
	b.WriteString("c4drill architecture models (.toml and .c4d).\n\n")
	b.WriteString(strings.TrimSpace(skillSeed))
	b.WriteString("\n\n")
	b.WriteString(strings.TrimSpace(editProtocol))

	if suffix := strings.TrimSpace(customSuffix); suffix != "" {
		b.WriteString("\n\n## Additional instructions from the author\n\n")
		b.WriteString(suffix)
	}

	return b.String()
}

// BuildMessages assembles the full request transcript: the system prompt,
// the conversation so far, then the new user message with the authoring
// context appended.
func BuildMessages(systemPrompt string, history []Message, userText string, ctx *AuthoringContext) []Message {
	messages := make([]Message, 0, len(history)+2)

	messages = append(messages, Message{Role: "system", Content: systemPrompt})
	messages = append(messages, history...)
	messages = append(messages, Message{Role: "user", Content: UserContent(userText, ctx)})

	return messages
}

// UserContent renders the user's text together with the authoring context.
func UserContent(userText string, ctx *AuthoringContext) string {
	if ctx == nil {
		return userText
	}

	var b strings.Builder

	b.WriteString("<authoring-context>\n")

	if ctx.ActiveFile != "" {
		b.WriteString("Active file: " + ctx.ActiveFile + "\n")
		b.WriteString("```\n")
		b.WriteString(ctx.ActiveContent)

		if !strings.HasSuffix(ctx.ActiveContent, "\n") {
			b.WriteString("\n")
		}

		b.WriteString("```\n")
	}

	for path, content := range ctx.OpenFiles {
		b.WriteString("Open file: " + path + "\n```\n")
		b.WriteString(truncate(content, maxOpenFileChars))
		b.WriteString("\n```\n")
	}

	if len(ctx.Diagnostics) > 0 {
		b.WriteString("Active diagnostics:\n")

		for _, d := range ctx.Diagnostics {
			b.WriteString("- " + d + "\n")
		}
	}

	if sel := strings.TrimSpace(ctx.Selection); sel != "" {
		b.WriteString("Selected text:\n```\n" + sel + "\n```\n")
	}

	b.WriteString("</authoring-context>\n\n")
	b.WriteString(userText)

	return b.String()
}

// truncate cuts s to at most n runes, noting the cut.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}

	cut := s[:n]
	if i := strings.LastIndexByte(cut, '\n'); i > 0 {
		cut = cut[:i] // keep whole lines
	}

	return cut + "\n… (truncated)"
}
