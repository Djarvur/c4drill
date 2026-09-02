// prompt_test.go pins the context assembly: the SKILL.md seed and edit
// protocol land in the system prompt, custom suffixes append, and the
// authoring context (active file, open files, diagnostics, selection)
// reaches the user message.

package ai_test

import (
	"strings"
	"testing"

	"github.com/Djarvur/c4drill/internal/gui/ai"
	"github.com/stretchr/testify/assert"
)

func TestSystemPromptSeedsSkillAndProtocol(t *testing.T) {
	t.Parallel()

	prompt := ai.SystemPrompt("")

	assert.Contains(t, prompt, "c4drill architecture assistant")
	// The SKILL.md snapshot's format knowledge is embedded.
	assert.Contains(t, prompt, "[properties]")
	assert.Contains(t, prompt, "c4drill-edit path=")
	// The custom suffix is empty: nothing extra appended.
	assert.NotContains(t, prompt, "Additional instructions from the author")
}

func TestSystemPromptAppendsCustomSuffix(t *testing.T) {
	t.Parallel()

	prompt := ai.SystemPrompt("Always answer in Spanish.")
	assert.Contains(t, prompt, "Always answer in Spanish.")
}

func TestUserContentCarriesAuthoringContext(t *testing.T) {
	t.Parallel()

	ctx := &ai.AuthoringContext{
		ActiveFile:    "demo.toml",
		ActiveContent: "[properties]\nname = \"Demo\"\n",
		OpenFiles: map[string]string{
			"other.toml": strings.Repeat("line\n", 10),
		},
		Diagnostics: []string{"demo.toml: unit \"x\" has no links"},
		Selection:   "name = \"Demo\"",
	}

	content := ai.UserContent("What is wrong?", ctx)

	assert.Contains(t, content, "<authoring-context>")
	assert.Contains(t, content, "Active file: demo.toml")
	assert.Contains(t, content, "name = \"Demo\"")
	assert.Contains(t, content, "other.toml")
	assert.Contains(t, content, "unit \"x\" has no links")
	assert.Contains(t, content, "Selected text:")
	assert.Contains(t, content, "What is wrong?")
}

func TestUserContentNilContext(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "plain question", ai.UserContent("plain question", nil))
}

func TestBuildMessagesOrder(t *testing.T) {
	t.Parallel()

	history := []ai.Message{{Role: "user", Content: "hi"}, {Role: "assistant", Content: "hello"}}

	messages := ai.BuildMessages("SYS", history, "next", nil)

	assert.Len(t, messages, 4)
	assert.Equal(t, "system", messages[0].Role)
	assert.Equal(t, "SYS", messages[0].Content)
	assert.Equal(t, "hi", messages[1].Content)
	assert.Equal(t, "hello", messages[2].Content)
	assert.Contains(t, messages[3].Content, "next")
}
