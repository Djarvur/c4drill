package render

import (
	"strings"
	"testing"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/Djarvur/c4drill/internal/model"
)

// Note: Tests in this file do NOT use t.Parallel() because the go-graphviz
// library uses a WASM-based rendering engine that has concurrency issues.

// HTML Label Builder Tests
// These tests verify the HTML label builder functions for each unit type.

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestHTMLPersonLabel(t *testing.T) {
	label := &graph.Label{
		Name:        "Test User",
		Description: "Test description",
	}
	result := buildPersonHTMLLabel(label)
	// Should contain HTML table with emoji and name
	if !strings.Contains(result, "<table") {
		t.Error("Person label should contain HTML table")
	}

	if !strings.Contains(result, "Test User") {
		t.Error("Person label should contain name")
	}

	// Check for emoji instead of img tag
	if !strings.Contains(result, "&#x1F464;") {
		t.Error("Person label should contain person emoji")
	}

	if strings.Contains(result, "<img") {
		t.Error("Person label should NOT contain img tag")
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

	// Should NOT contain img tag
	if strings.Contains(result, "<img") {
		t.Error("DB label should NOT contain img tag")
	}

	// Should be single-column (no rowspan for icon)
	if strings.Contains(result, "rowspan") {
		t.Error("DB label should NOT contain rowspan (single-column layout)")
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

	// Should NOT contain img tag
	if strings.Contains(result, "<img") {
		t.Error("Queue label should NOT contain img tag")
	}

	// Should be single-column (no rowspan for icon)
	if strings.Contains(result, "rowspan") {
		t.Error("Queue label should NOT contain rowspan (single-column layout)")
	}

	// The pipe shape (pipe.go) carries the queue identity — no text-bar
	// graphic inside the label (double graphics).
	if strings.Contains(result, "═╦╩═╦═══") {
		t.Error("Queue label should NOT contain ASCII art graphic")
	}

	// Verify name, technology, description are present
	if !strings.Contains(result, "Test Queue") {
		t.Error("Queue label should contain name")
	}

	if !strings.Contains(result, "RabbitMQ") {
		t.Error("Queue label should contain technology")
	}

	if !strings.Contains(result, "Test description") {
		t.Error("Queue label should contain description")
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

	if !strings.Contains(result, "Test System") {
		t.Error("System label should contain name")
	}

	// Should NOT contain img tag
	if strings.Contains(result, "<img") {
		t.Error("System label should NOT contain img tag")
	}

	// Should NOT contain old monospace SYS label
	if strings.Contains(result, "SYS") {
		t.Error("System label should NOT contain old SYS monospace label")
	}

	// Should be single-column (no rowspan for icon)
	if strings.Contains(result, "rowspan") {
		t.Error("System label should NOT contain rowspan (single-column layout)")
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

	if !strings.Contains(result, "Test Container") {
		t.Error("Container label should contain name")
	}

	// Should NOT contain img tag
	if strings.Contains(result, "<img") {
		t.Error("Container label should NOT contain img tag")
	}

	// Should NOT contain old monospace CONT label
	if strings.Contains(result, "CONT") {
		t.Error("Container label should NOT contain old CONT monospace label")
	}

	// Should be single-column (no rowspan for icon)
	if strings.Contains(result, "rowspan") {
		t.Error("Container label should NOT contain rowspan (single-column layout)")
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

	if !strings.Contains(result, "Test Component") {
		t.Error("Component label should contain name")
	}

	// Should NOT contain img tag
	if strings.Contains(result, "<img") {
		t.Error("Component label should NOT contain img tag")
	}

	// Should NOT contain old monospace COMP label
	if strings.Contains(result, "COMP") {
		t.Error("Component label should NOT contain old COMP monospace label")
	}

	// Should be single-column (no rowspan for icon)
	if strings.Contains(result, "rowspan") {
		t.Error("Component label should NOT contain rowspan (single-column layout)")
	}
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestQueueShape(t *testing.T) {
	// Create a minimal graph with a Queue node and a DB node
	g := &graph.Graph{
		Nodes: []*graph.Node{
			{
				ID:   "queue1",
				Type: model.TypeQueue,
				Label: &graph.Label{
					Name: "Test Queue",
				},
			},
			{
				ID:   "db1",
				Type: model.TypeDb,
				Label: &graph.Label{
					Name: "Test DB",
				},
			},
		},
	}

	// Render to DOT and verify shapes
	dotOutput, err := RenderDOT(g)
	if err != nil {
		t.Fatalf("RenderDOT failed: %v", err)
	}

	dotStr := string(dotOutput)

	// Queue should NOT have cylinder shape
	if strings.Contains(dotStr, `queue1`) && strings.Contains(dotStr, `shape=cylinder`) &&
		strings.Index(dotStr, `queue1`) < strings.Index(dotStr, `shape=cylinder`) {
		// Check if cylinder shape is on queue1 line (within same node definition)
		queueIdx := strings.Index(dotStr, `queue1`)
		cylinderIdx := strings.Index(dotStr, `shape=cylinder`)
		nextNodeIdx := strings.Index(dotStr[queueIdx+1:], "\n\t")

		if nextNodeIdx == -1 || cylinderIdx < queueIdx+nextNodeIdx {
			t.Error("Queue node should NOT have cylinder shape")
		}
	}

	// DB should still have cylinder shape
	if !strings.Contains(dotStr, `db1`) || !strings.Contains(dotStr, `shape=cylinder`) {
		t.Error("DB node should have cylinder shape")
	}

	// Verify db1 node definition contains shape=cylinder
	dbIdx := strings.Index(dotStr, `db1`)
	if dbIdx == -1 {
		t.Fatal("DB node not found in output")
	}

	// Find end of db1 node definition (next node or closing brace)
	endIdx := strings.Index(dotStr[dbIdx:], "];")
	if endIdx == -1 {
		t.Fatal("Could not find end of DB node definition")
	}

	dbNodeDef := dotStr[dbIdx : dbIdx+endIdx]
	if !strings.Contains(dbNodeDef, `shape=cylinder`) {
		t.Error("DB node definition should contain shape=cylinder")
	}

	// Verify queue1 node definition does NOT contain shape=cylinder
	queueIdx := strings.Index(dotStr, `queue1`)
	if queueIdx == -1 {
		t.Fatal("Queue node not found in output")
	}

	endIdx = strings.Index(dotStr[queueIdx:], "];")
	if endIdx == -1 {
		t.Fatal("Could not find end of Queue node definition")
	}

	queueNodeDef := dotStr[queueIdx : queueIdx+endIdx]
	if strings.Contains(queueNodeDef, `shape=cylinder`) {
		t.Error("Queue node definition should NOT contain shape=cylinder")
	}
}
