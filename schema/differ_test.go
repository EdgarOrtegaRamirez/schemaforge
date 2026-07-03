package schema

import (
	"testing"
)

func TestDiffSchemasIdentical(t *testing.T) {
	s1 := &Schema{Type: TypeOrArray{Types: []string{"string"}}, Title: "test"}
	s2 := &Schema{Type: TypeOrArray{Types: []string{"string"}}, Title: "test"}

	result := DiffSchemas(s1, s2)
	if !result.Identical {
		t.Error("expected identical schemas")
	}
}

func TestDiffSchemasTypeChange(t *testing.T) {
	s1 := &Schema{Type: TypeOrArray{Types: []string{"string"}}}
	s2 := &Schema{Type: TypeOrArray{Types: []string{"number"}}}

	result := DiffSchemas(s1, s2)
	if result.Identical {
		t.Error("expected different schemas")
	}
	if len(result.Diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(result.Diffs))
	}
	if result.Diffs[0].Type != DiffModified {
		t.Error("expected modified diff type")
	}
}

func TestDiffSchemasPropertyAdded(t *testing.T) {
	s1 := &Schema{
		Type:       TypeOrArray{Types: []string{"object"}},
		Properties: map[string]*Schema{},
	}
	s2 := &Schema{
		Type: TypeOrArray{Types: []string{"object"}},
		Properties: map[string]*Schema{
			"name": {Type: TypeOrArray{Types: []string{"string"}}},
		},
	}

	result := DiffSchemas(s1, s2)
	if result.Identical {
		t.Error("expected different schemas")
	}
	if result.Summary.Added != 1 {
		t.Errorf("expected 1 added, got %d", result.Summary.Added)
	}
}

func TestDiffSchemasPropertyRemoved(t *testing.T) {
	s1 := &Schema{
		Type: TypeOrArray{Types: []string{"object"}},
		Properties: map[string]*Schema{
			"name": {Type: TypeOrArray{Types: []string{"string"}}},
		},
	}
	s2 := &Schema{
		Type:       TypeOrArray{Types: []string{"object"}},
		Properties: map[string]*Schema{},
	}

	result := DiffSchemas(s1, s2)
	if result.Identical {
		t.Error("expected different schemas")
	}
	if result.Summary.Removed != 1 {
		t.Errorf("expected 1 removed, got %d", result.Summary.Removed)
	}
}

func TestDiffSchemasPropertyTypeChange(t *testing.T) {
	s1 := &Schema{
		Type: TypeOrArray{Types: []string{"object"}},
		Properties: map[string]*Schema{
			"count": {Type: TypeOrArray{Types: []string{"number"}}},
		},
	}
	s2 := &Schema{
		Type: TypeOrArray{Types: []string{"object"}},
		Properties: map[string]*Schema{
			"count": {Type: TypeOrArray{Types: []string{"string"}}},
		},
	}

	result := DiffSchemas(s1, s2)
	if result.Identical {
		t.Error("expected different schemas")
	}
	if result.Summary.Modified == 0 {
		t.Error("expected at least 1 modification")
	}
}

func TestDiffSchemasTitleChange(t *testing.T) {
	s1 := &Schema{Title: "old"}
	s2 := &Schema{Title: "new"}

	result := DiffSchemas(s1, s2)
	if result.Identical {
		t.Error("expected different schemas")
	}
}

func TestDiffSchemasRequiredChange(t *testing.T) {
	s1 := &Schema{Required: []string{"name"}}
	s2 := &Schema{Required: []string{"name", "email"}}

	result := DiffSchemas(s1, s2)
	if result.Identical {
		t.Error("expected different schemas")
	}
}

func TestDiffSchemasNilComparison(t *testing.T) {
	s := &Schema{Type: TypeOrArray{Types: []string{"string"}}}

	result := DiffSchemas(nil, s)
	if result.Identical {
		t.Error("expected different schemas when one is nil")
	}

	result = DiffSchemas(s, nil)
	if result.Identical {
		t.Error("expected different schemas when one is nil")
	}

	result = DiffSchemas(nil, nil)
	if !result.Identical {
		t.Error("expected identical when both nil")
	}
}
