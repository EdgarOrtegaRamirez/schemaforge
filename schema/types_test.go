package schema

import (
	"encoding/json"
	"testing"
)

func TestSchemaFromJSON(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"age": {"type": "integer"}
		},
		"required": ["name"]
	}`

	s, err := SchemaFromJSON([]byte(schemaJSON))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !s.HasType("object") {
		t.Error("expected type object")
	}
	if len(s.Properties) != 2 {
		t.Errorf("expected 2 properties, got %d", len(s.Properties))
	}
	if len(s.Required) != 1 || s.Required[0] != "name" {
		t.Errorf("expected required [name], got %v", s.Required)
	}
}

func TestSchemaToJSON(t *testing.T) {
	s := &Schema{
		Type: TypeOrArray{Types: []string{"string"}},
		Title: "test",
	}

	data, err := s.ToJSON()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if parsed["type"] != "string" {
		t.Errorf("expected type string, got %v", parsed["type"])
	}
}

func TestTypeOrArray(t *testing.T) {
	// Single type
	ta := TypeOrArray{Types: []string{"string"}}
	if !ta.HasType("string") {
		t.Error("expected HasType(string) to be true")
	}
	if ta.HasType("number") {
		t.Error("expected HasType(number) to be false")
	}
	if ta.PrimaryType() != "string" {
		t.Errorf("expected primary type string, got %s", ta.PrimaryType())
	}

	// Multiple types
	ta2 := TypeOrArray{Types: []string{"string", "null"}}
	if !ta2.HasType("string") || !ta2.HasType("null") {
		t.Error("expected HasType to match both types")
	}
}

func TestSchemaFormatString(t *testing.T) {
	tests := []struct {
		schema   *Schema
		expected string
	}{
		{&Schema{Type: TypeOrArray{Types: []string{"string"}}}, "string"},
		{&Schema{Type: TypeOrArray{Types: []string{"array"}}, Items: &Schema{Type: TypeOrArray{Types: []string{"string"}}}}, "array<string>"},
		{&Schema{Type: TypeOrArray{Types: []string{"object"}}, Properties: map[string]*Schema{"a": {}}}, "object{1 props}"},
		{&Schema{}, "any"},
		{&Schema{Ref: "#/defs/Foo"}, "$ref: #/defs/Foo"},
	}

	for _, tt := range tests {
		got := tt.schema.FormatString()
		if got != tt.expected {
			t.Errorf("FormatString() = %q, want %q", got, tt.expected)
		}
	}
}

func TestRequiredSet(t *testing.T) {
	s := &Schema{
		Required: []string{"name", "email"},
	}

	set := s.RequiredSet()
	if !set["name"] || !set["email"] {
		t.Error("expected name and email in set")
	}
	if set["age"] {
		t.Error("age should not be in set")
	}
}

func TestSortedPropertyNames(t *testing.T) {
	s := &Schema{
		Properties: map[string]*Schema{
			"zebra":  {},
			"apple":  {},
			"mango":  {},
		},
	}

	names := s.SortedPropertyNames()
	expected := []string{"apple", "mango", "zebra"}
	for i, name := range names {
		if name != expected[i] {
			t.Errorf("expected %s at index %d, got %s", expected[i], i, name)
		}
	}
}
