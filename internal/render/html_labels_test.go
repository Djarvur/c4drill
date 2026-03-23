package render

import (
	"strings"
	"testing"

	"github.com/Djarvur/c4drill/internal/graph"
)

// Note: Tests in this file do NOT use t.Parallel() because the go-graphviz
// library uses a WASM-based rendering engine that has concurrency issues.

// HTML Label Builder Tests
// These tests verify the HTML label builder functions for each unit type.
// The implementations will be added in Wave 1 (plan 12-01).

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestHTMLPersonLabel(t *testing.T) {
	label := &graph.Label{
		Name:        "Test User",
		Description: "Test description",
	}
	result := buildPersonHTMLLabel(label, ".icons/person-3C7FC0.svg")
	// Should contain HTML table with icon and name
	if !strings.Contains(result, "<table") {
		t.Error("Person label should contain HTML table")
	}

	if !strings.Contains(result, "Test User") {
		t.Error("Person label should contain name")
	}

	if !strings.Contains(result, `<img src=".icons/person-3C7FC0.svg"`) {
		t.Error("Person label should contain img tag with icon path")
	}
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestHTMLDbLabel(t *testing.T) {
	label := &graph.Label{
		Name:        "Test DB",
		Technology:  "PostgreSQL",
		Description: "Test description",
	}

	result := buildDbHTMLLabel(label, ".icons/db-3C7FC0.svg")
	if !strings.Contains(result, "<table") {
		t.Error("DB label should contain HTML table")
	}

	if !strings.Contains(result, "Test DB") {
		t.Error("DB label should contain name")
	}

	if !strings.Contains(result, `<img src=".icons/db-3C7FC0.svg"`) {
		t.Error("DB label should contain img tag with icon path")
	}
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestHTMLQueueLabel(t *testing.T) {
	label := &graph.Label{
		Name:        "Test Queue",
		Technology:  "RabbitMQ",
		Description: "Test description",
	}

	result := buildQueueHTMLLabel(label, ".icons/pipe-3C7FC0.svg")
	if !strings.Contains(result, "<table") {
		t.Error("Queue label should contain HTML table")
	}
	// Queue has NO rowspan - 4 separate rows
	if strings.Contains(result, "rowspan") {
		t.Error("Queue label should NOT contain rowspan")
	}

	if !strings.Contains(result, `<img src=".icons/pipe-3C7FC0.svg"`) {
		t.Error("Queue label should contain img tag with icon path")
	}
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestHTMLSystemLabel(t *testing.T) {
	label := &graph.Label{
		Name:        "Test System",
		Technology:  "Go",
		Description: "Test description",
	}

	result := buildSystemHTMLLabel(label, ".icons/system-3C7FC0.svg")
	if !strings.Contains(result, "<table") {
		t.Error("System label should contain HTML table")
	}

	if !strings.Contains(result, "Test System") {
		t.Error("System label should contain name")
	}

	if !strings.Contains(result, `<img src=".icons/system-3C7FC0.svg"`) {
		t.Error("System label should contain img tag with icon path")
	}
	// Should NOT contain old monospace SYS label anymore
	if strings.Contains(result, "SYS") {
		t.Error("System label should NOT contain old SYS monospace label")
	}
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestHTMLContainerLabel(t *testing.T) {
	label := &graph.Label{
		Name:        "Test Container",
		Technology:  "Docker",
		Description: "Test description",
	}

	result := buildContainerHTMLLabel(label, ".icons/container-3C7FC0.svg")
	if !strings.Contains(result, "<table") {
		t.Error("Container label should contain HTML table")
	}

	if !strings.Contains(result, "Test Container") {
		t.Error("Container label should contain name")
	}

	if !strings.Contains(result, `<img src=".icons/container-3C7FC0.svg"`) {
		t.Error("Container label should contain img tag with icon path")
	}
	// Should NOT contain old monospace CONT label anymore
	if strings.Contains(result, "CONT") {
		t.Error("Container label should NOT contain old CONT monospace label")
	}
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestHTMLComponentLabel(t *testing.T) {
	label := &graph.Label{
		Name:        "Test Component",
		Technology:  "Go",
		Description: "Test description",
	}

	result := buildComponentHTMLLabel(label, ".icons/component-78A8D8.svg")
	if !strings.Contains(result, "<table") {
		t.Error("Component label should contain HTML table")
	}

	if !strings.Contains(result, "Test Component") {
		t.Error("Component label should contain name")
	}

	if !strings.Contains(result, `<img src=".icons/component-78A8D8.svg"`) {
		t.Error("Component label should contain img tag with icon path")
	}
	// Should NOT contain old monospace COMP label anymore
	if strings.Contains(result, "COMP") {
		t.Error("Component label should NOT contain old COMP monospace label")
	}
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestHTMLLabelEmptyIconPath(t *testing.T) {
	label := &graph.Label{
		Name:        "Test User",
		Description: "Test description",
	}
	result := buildPersonHTMLLabel(label, "")
	// Should still contain HTML table with name even without icon
	if !strings.Contains(result, "<table") {
		t.Error("Person label should contain HTML table")
	}

	if !strings.Contains(result, "Test User") {
		t.Error("Person label should contain name")
	}
	// Should NOT contain img tag when iconRelPath is empty
	if strings.Contains(result, "<img") {
		t.Error("Person label should NOT contain img tag when iconRelPath is empty")
	}
}
