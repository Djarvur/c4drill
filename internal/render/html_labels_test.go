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
	result := buildPersonHTMLLabel(label)
	// Should contain HTML table with icon and name
	if !strings.Contains(result, "<table") {
		t.Error("Person label should contain HTML table")
	}
	if !strings.Contains(result, "Test User") {
		t.Error("Person label should contain name")
	}
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestHTMLDbLabel(t *testing.T) {
	label := &graph.Label{
		Name:        "Test DB",
		Technology:  "PostgreSQL",
		Description: "Test description",
	}
	result := buildDbHTMLLabel(label)
	if !strings.Contains(result, "<table") {
		t.Error("DB label should contain HTML table")
	}
	if !strings.Contains(result, "Test DB") {
		t.Error("DB label should contain name")
	}
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestHTMLQueueLabel(t *testing.T) {
	label := &graph.Label{
		Name:        "Test Queue",
		Technology:  "RabbitMQ",
		Description: "Test description",
	}
	result := buildQueueHTMLLabel(label)
	if !strings.Contains(result, "<table") {
		t.Error("Queue label should contain HTML table")
	}
	// Queue has NO rowspan - 4 separate rows
	if strings.Contains(result, "rowspan") {
		t.Error("Queue label should NOT contain rowspan")
	}
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestHTMLSystemLabel(t *testing.T) {
	label := &graph.Label{
		Name:        "Test System",
		Technology:  "Go",
		Description: "Test description",
	}
	result := buildSystemHTMLLabel(label)
	if !strings.Contains(result, "<table") {
		t.Error("System label should contain HTML table")
	}
	if !strings.Contains(result, "SYS") {
		t.Error("System label should contain SYS label")
	}
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestHTMLContainerLabel(t *testing.T) {
	label := &graph.Label{
		Name:        "Test Container",
		Technology:  "Docker",
		Description: "Test description",
	}
	result := buildContainerHTMLLabel(label)
	if !strings.Contains(result, "<table") {
		t.Error("Container label should contain HTML table")
	}
	if !strings.Contains(result, "CONT") {
		t.Error("Container label should contain CONT label")
	}
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestHTMLComponentLabel(t *testing.T) {
	label := &graph.Label{
		Name:        "Test Component",
		Technology:  "Go",
		Description: "Test description",
	}
	result := buildComponentHTMLLabel(label)
	if !strings.Contains(result, "<table") {
		t.Error("Component label should contain HTML table")
	}
	if !strings.Contains(result, "COMP") {
		t.Error("Component label should contain COMP label")
	}
}
