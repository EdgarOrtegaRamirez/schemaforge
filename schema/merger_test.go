package schema

import (
	"testing"
)

func TestMergeSameType(t *testing.T) {
	s1 := &Schema{
		Type:      TypeOrArray{Types: []string{"string"}},
		MinLength: intPtr(1),
	}
	s2 := &Schema{
		Type:      TypeOrArray{Types: []string{"string"}},
		MaxLength: intPtr(100),
	}

	merger := NewMerger()
	merger.AddSchema(s1)
	merger.AddSchema(s2)

	merged, err := merger.Merge()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !merged.HasType("string") {
		t.Error("expected merged type string")
	}
	if merged.MinLength == nil || *merged.MinLength != 1 {
		t.Error("expected minLength 1")
	}
	if merged.MaxLength == nil || *merged.MaxLength != 100 {
		t.Error("expected maxLength 100")
	}
}

func TestMergeObjects(t *testing.T) {
	s1 := &Schema{
		Type: TypeOrArray{Types: []string{"object"}},
		Properties: map[string]*Schema{
			"name": {Type: TypeOrArray{Types: []string{"string"}}},
		},
		Required: []string{"name"},
	}
	s2 := &Schema{
		Type: TypeOrArray{Types: []string{"object"}},
		Properties: map[string]*Schema{
			"email": {Type: TypeOrArray{Types: []string{"string"}}},
		},
		Required: []string{"email"},
	}

	merger := NewMerger()
	merger.AddSchema(s1)
	merger.AddSchema(s2)

	merged, err := merger.Merge()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(merged.Properties) != 2 {
		t.Errorf("expected 2 properties, got %d", len(merged.Properties))
	}
	// name and email are each required in only 1 of 2 schemas, so not required in merged
	if len(merged.Required) != 0 {
		t.Errorf("expected 0 required (each only in 1 schema), got %d: %v", len(merged.Required), merged.Required)
	}
}

func TestMergeObjectsBothRequired(t *testing.T) {
	s1 := &Schema{
		Type: TypeOrArray{Types: []string{"object"}},
		Properties: map[string]*Schema{
			"name": {Type: TypeOrArray{Types: []string{"string"}}},
			"id":   {Type: TypeOrArray{Types: []string{"number"}}},
		},
		Required: []string{"name", "id"},
	}
	s2 := &Schema{
		Type: TypeOrArray{Types: []string{"object"}},
		Properties: map[string]*Schema{
			"name":  {Type: TypeOrArray{Types: []string{"string"}}},
			"email": {Type: TypeOrArray{Types: []string{"string"}}},
		},
		Required: []string{"name"},
	}

	merger := NewMerger()
	merger.AddSchema(s1)
	merger.AddSchema(s2)

	merged, err := merger.Merge()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// "name" is required in both schemas, so it should be required in merged
	if len(merged.Required) != 1 || merged.Required[0] != "name" {
		t.Errorf("expected required [name], got %v", merged.Required)
	}
}

func TestMergeArrays(t *testing.T) {
	s1 := &Schema{
		Type:  TypeOrArray{Types: []string{"array"}},
		Items: &Schema{Type: TypeOrArray{Types: []string{"string"}}},
	}
	s2 := &Schema{
		Type:  TypeOrArray{Types: []string{"array"}},
		Items: &Schema{Type: TypeOrArray{Types: []string{"string"}}},
	}

	merger := NewMerger()
	merger.AddSchema(s1)
	merger.AddSchema(s2)

	merged, err := merger.Merge()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !merged.HasType("array") {
		t.Error("expected merged type array")
	}
	if merged.Items == nil {
		t.Error("expected items schema")
	}
}

func TestMergeDifferentTypes(t *testing.T) {
	s1 := &Schema{Type: TypeOrArray{Types: []string{"string"}}}
	s2 := &Schema{Type: TypeOrArray{Types: []string{"number"}}}

	merger := NewMerger()
	merger.AddSchema(s1)
	merger.AddSchema(s2)

	merged, err := merger.Merge()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(merged.AnyOf) != 2 {
		t.Errorf("expected 2 anyOf schemas, got %d", len(merged.AnyOf))
	}
}

func TestMergeEmpty(t *testing.T) {
	merger := NewMerger()
	_, err := merger.Merge()
	if err == nil {
		t.Error("expected error for empty merge")
	}
}

func TestMergeSingle(t *testing.T) {
	s := &Schema{Type: TypeOrArray{Types: []string{"string"}}}
	merger := NewMerger()
	merger.AddSchema(s)

	merged, err := merger.Merge()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !merged.HasType("string") {
		t.Error("expected type string")
	}
}

func TestMergeWithTitle(t *testing.T) {
	s1 := &Schema{Type: TypeOrArray{Types: []string{"string"}}}
	s2 := &Schema{Type: TypeOrArray{Types: []string{"string"}}}

	merger := NewMerger(MergeOptions{Title: "Merged"})
	merger.AddSchema(s1)
	merger.AddSchema(s2)

	merged, err := merger.Merge()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if merged.Title != "Merged" {
		t.Errorf("expected title Merged, got %s", merged.Title)
	}
}

func intPtr(i int) *int {
	return &i
}
