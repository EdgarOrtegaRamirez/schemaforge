package schema

import (
	"testing"
)

func TestGenerateString(t *testing.T) {
	gen := NewGenerator()
	gen.AddFromJSONString(`"hello"`)
	gen.AddFromJSONString(`"world"`)
	gen.AddFromJSONString(`"test"`)

	s, err := gen.Generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !s.HasType("string") {
		t.Error("expected type string")
	}
}

func TestGenerateNumber(t *testing.T) {
	gen := NewGenerator()
	gen.AddFromJSONString(`42`)
	gen.AddFromJSONString(`17`)
	gen.AddFromJSONString(`99`)

	s, err := gen.Generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !s.HasType("number") && !s.HasType("integer") {
		t.Errorf("expected number or integer type, got %v", s.Type.Types)
	}
}

func TestGenerateBoolean(t *testing.T) {
	gen := NewGenerator()
	gen.AddFromJSONString(`true`)
	gen.AddFromJSONString(`false`)

	s, err := gen.Generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !s.HasType("boolean") {
		t.Error("expected type boolean")
	}
}

func TestGenerateObject(t *testing.T) {
	gen := NewGenerator()
	gen.AddFromJSONString(`{"name": "Alice", "age": 30}`)
	gen.AddFromJSONString(`{"name": "Bob", "age": 25}`)

	s, err := gen.Generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !s.HasType("object") {
		t.Error("expected type object")
	}
	if len(s.Properties) != 2 {
		t.Errorf("expected 2 properties, got %d", len(s.Properties))
	}
	if len(s.Required) != 2 {
		t.Errorf("expected 2 required fields, got %d", len(s.Required))
	}
}

func TestGenerateArray(t *testing.T) {
	gen := NewGenerator()
	gen.AddFromJSONString(`["hello", "world"]`)
	gen.AddFromJSONString(`["foo", "bar", "baz"]`)

	s, err := gen.Generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !s.HasType("array") {
		t.Error("expected type array")
	}
	if s.Items == nil {
		t.Fatal("expected items schema")
	}
	if !s.Items.HasType("string") {
		t.Error("expected items to be string type")
	}
}

func TestGenerateEmailFormat(t *testing.T) {
	gen := NewGenerator(GeneratorOptions{DetectFormat: true})
	gen.AddFromJSONString(`"alice@example.com"`)
	gen.AddFromJSONString(`"bob@test.org"`)

	s, err := gen.Generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.Format != "email" {
		t.Errorf("expected format email, got %s", s.Format)
	}
}

func TestGenerateUUIDFormat(t *testing.T) {
	gen := NewGenerator(GeneratorOptions{DetectFormat: true})
	gen.AddFromJSONString(`"550e8400-e29b-41d4-a716-446655440000"`)
	gen.AddFromJSONString(`"6ba7b810-9dad-11d1-80b4-00c04fd430c8"`)

	s, err := gen.Generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.Format != "uuid" {
		t.Errorf("expected format uuid, got %s", s.Format)
	}
}

func TestGenerateWithEnum(t *testing.T) {
	gen := NewGenerator(GeneratorOptions{InferEnums: true, EnumThreshold: 2})
	gen.AddFromJSONString(`"red"`)
	gen.AddFromJSONString(`"blue"`)
	gen.AddFromJSONString(`"red"`)

	s, err := gen.Generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(s.Enum) != 2 {
		t.Errorf("expected 2 enum values, got %d", len(s.Enum))
	}
}

func TestGenerateWithNull(t *testing.T) {
	gen := NewGenerator()
	gen.AddSample(nil)

	s, err := gen.Generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !s.HasType("null") {
		t.Error("expected type null")
	}
}

func TestGenerateEmpty(t *testing.T) {
	gen := NewGenerator()
	_, err := gen.Generate()
	if err == nil {
		t.Error("expected error for empty generator")
	}
}

func TestGenerateNestedObject(t *testing.T) {
	gen := NewGenerator()
	gen.AddFromJSONString(`{"user": {"name": "Alice", "email": "a@b.com"}, "count": 5}`)
	gen.AddFromJSONString(`{"user": {"name": "Bob", "email": "b@c.com"}, "count": 10}`)

	s, err := gen.Generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !s.HasType("object") {
		t.Error("expected type object")
	}
	userProp, ok := s.Properties["user"]
	if !ok {
		t.Fatal("expected user property")
	}
	if !userProp.HasType("object") {
		t.Error("expected user to be object type")
	}
}

func TestGenerateTitleDescription(t *testing.T) {
	gen := NewGenerator(GeneratorOptions{
		Title:       "User",
		Description: "A user object",
	})
	gen.AddFromJSONString(`{"name": "Alice"}`)

	s, err := gen.Generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.Title != "User" {
		t.Errorf("expected title User, got %s", s.Title)
	}
	if s.Description != "A user object" {
		t.Errorf("expected description, got %s", s.Description)
	}
}
