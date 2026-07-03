package schema

import (
	"strings"
	"testing"
)

func TestDocumenterMarkdown(t *testing.T) {
	s := &Schema{
		Title:       "User",
		Description: "A user object",
		Type:        TypeOrArray{Types: []string{"object"}},
		Properties: map[string]*Schema{
			"name":  {Type: TypeOrArray{Types: []string{"string"}}, Description: "User name"},
			"email": {Type: TypeOrArray{Types: []string{"string"}}, Format: "email"},
		},
		Required: []string{"name"},
	}

	doc := NewDocumenter(s)
	md := doc.ToMarkdown()

	if !strings.Contains(md, "# User") {
		t.Error("expected title in markdown")
	}
	if !strings.Contains(md, "A user object") {
		t.Error("expected description in markdown")
	}
	if !strings.Contains(md, "`name`") {
		t.Error("expected name property")
	}
	if !strings.Contains(md, "`email`") {
		t.Error("expected email property")
	}
}

func TestDocumenterText(t *testing.T) {
	s := &Schema{
		Title: "Test",
		Type:  TypeOrArray{Types: []string{"object"}},
		Properties: map[string]*Schema{
			"id": {Type: TypeOrArray{Types: []string{"number"}}},
		},
	}

	doc := NewDocumenter(s)
	text := doc.ToText()

	if !strings.Contains(text, "Schema: Test") {
		t.Error("expected title in text")
	}
	if !strings.Contains(text, "id") {
		t.Error("expected id property")
	}
}

func TestDocumenterHTML(t *testing.T) {
	s := &Schema{
		Title: "Test",
		Type:  TypeOrArray{Types: []string{"object"}},
		Properties: map[string]*Schema{
			"name": {Type: TypeOrArray{Types: []string{"string"}}},
		},
	}

	doc := NewDocumenter(s)
	html := doc.ToHTML()

	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("expected HTML doctype")
	}
	if !strings.Contains(html, "Test") {
		t.Error("expected title in HTML")
	}
	if !strings.Contains(html, "name") {
		t.Error("expected name property in HTML")
	}
}

func TestDocumenterConstraints(t *testing.T) {
	minLen := 5
	maxLen := 100
	s := &Schema{
		Type:      TypeOrArray{Types: []string{"string"}},
		MinLength: &minLen,
		MaxLength: &maxLen,
		Pattern:   "^[a-z]+$",
		Format:    "email",
	}

	doc := NewDocumenter(s)
	md := doc.ToMarkdown()

	if !strings.Contains(md, "Min length") {
		t.Error("expected min length constraint")
	}
	if !strings.Contains(md, "Max length") {
		t.Error("expected max length constraint")
	}
}

func TestDocumenterEnum(t *testing.T) {
	s := &Schema{
		Type: TypeOrArray{Types: []string{"string"}},
		Enum: []interface{}{"red", "green", "blue"},
	}

	doc := NewDocumenter(s)
	md := doc.ToMarkdown()

	if !strings.Contains(md, "Allowed Values") {
		t.Error("expected enum section")
	}
	if !strings.Contains(md, "red") {
		t.Error("expected red in enum values")
	}
}

func TestDocumenterEmpty(t *testing.T) {
	s := &Schema{}
	doc := NewDocumenter(s)
	md := doc.ToMarkdown()

	// Should not panic
	if md == "" {
		t.Error("expected non-empty output")
	}
}
