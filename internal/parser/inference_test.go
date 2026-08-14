// Package parser_test — exported inference helpers (Plan 35-05 Task 1).
//
// defaultTypeForParent and inferGenericType were unexported hooks inside
// parseUnitWithOrder; the C4D front-end's ToModel must apply the IDENTICAL
// post-parse inference (D-02 parity: both front-ends produce the same
// *model.Unit for equivalent documents), so the helpers are exported
// in place (the 35-PATTERNS export/move recommendation — export chosen to
// avoid churning internal/model). These tests pin the exported surface and
// the value tables the TOML parser has always produced.
package parser_test

import (
	"testing"

	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/stretchr/testify/assert"
)

// TestDefaultTypeForParent pins the omitted-type default per parent type:
// C1 root -> system, system -> container, box -> system (same-level C1
// grouping), container -> component, containerBox -> container,
// componentBox -> component, and every non-grouping parent falls back to
// system (C1).
func TestDefaultTypeForParent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		parentType model.UnitType
		want       model.UnitType
	}{
		{name: "no parent (C1 root)", parentType: "", want: model.TypeSystem},
		{name: "system (C2)", parentType: model.TypeSystem, want: model.TypeContainer},
		{name: "box (C1 grouping)", parentType: model.TypeBox, want: model.TypeSystem},
		{name: "container (C3)", parentType: model.TypeContainer, want: model.TypeComponent},
		{name: "containerBox (C2 grouping)", parentType: model.TypeContainerBox, want: model.TypeContainer},
		{name: "componentBox (C3 grouping)", parentType: model.TypeComponentBox, want: model.TypeComponent},
		{name: "db parent falls back to C1", parentType: model.TypeDb, want: model.TypeSystem},
		{name: "person parent falls back to C1", parentType: model.TypePerson, want: model.TypeSystem},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, parser.DefaultTypeForParent(tt.parentType),
				"DefaultTypeForParent(%q)", tt.parentType)
		})
	}
}

// TestInferGenericType pins the generic (level-agnostic) type promotion:
// db/queue/box resolve to their level-specific variants by parent, explicit
// level-specific variants pass through, and non-generic types are unchanged.
func TestInferGenericType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		unitType   model.UnitType
		parentType model.UnitType
		want       model.UnitType
	}{
		// C1 level (root or C1 box): generics unchanged.
		{name: "db at C1", unitType: model.TypeDb, parentType: "", want: model.TypeDb},
		{name: "queue at C1", unitType: model.TypeQueue, parentType: "", want: model.TypeQueue},
		{name: "box at C1", unitType: model.TypeBox, parentType: model.TypeBox, want: model.TypeBox},
		// C2 level (inside system or containerBox).
		{name: "db in system", unitType: model.TypeDb, parentType: model.TypeSystem, want: model.TypeContainerDb},
		{name: "queue in system", unitType: model.TypeQueue, parentType: model.TypeSystem, want: model.TypeContainerQueue},
		{name: "box in system", unitType: model.TypeBox, parentType: model.TypeSystem, want: model.TypeContainerBox},
		{name: "db in containerBox", unitType: model.TypeDb, parentType: model.TypeContainerBox, want: model.TypeContainerDb},
		// C3 level (inside container or componentBox).
		{name: "db in container", unitType: model.TypeDb, parentType: model.TypeContainer, want: model.TypeComponentDb},
		{name: "queue in container", unitType: model.TypeQueue, parentType: model.TypeContainer, want: model.TypeComponentQueue},
		{name: "box in container", unitType: model.TypeBox, parentType: model.TypeContainer, want: model.TypeComponentBox},
		{name: "db in componentBox", unitType: model.TypeDb, parentType: model.TypeComponentBox, want: model.TypeComponentDb},
		// Explicit level-specific variants pass through (not generic types).
		{name: "explicit containerDb", unitType: model.TypeContainerDb, parentType: model.TypeSystem, want: model.TypeContainerDb},
		{name: "explicit componentQueue", unitType: model.TypeComponentQueue, parentType: model.TypeContainer, want: model.TypeComponentQueue},
		{name: "explicit containerBox stays", unitType: model.TypeContainerBox, parentType: model.TypeSystem, want: model.TypeContainerBox},
		// Non-generic types are returned as-is regardless of parent.
		{name: "system unchanged", unitType: model.TypeSystem, parentType: model.TypeSystem, want: model.TypeSystem},
		{name: "person unchanged", unitType: model.TypePerson, parentType: "", want: model.TypePerson},
		{name: "external db unchanged", unitType: model.TypeDbExternal, parentType: "", want: model.TypeDbExternal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, parser.InferGenericType(tt.unitType, tt.parentType),
				"InferGenericType(%q, %q)", tt.unitType, tt.parentType)
		})
	}
}
