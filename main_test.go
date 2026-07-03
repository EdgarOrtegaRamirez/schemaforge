package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCLIGenerate(t *testing.T) {
	bin := buildBinary(t)
	defer os.Remove(bin)

	// Test generate from file
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.json")
	os.WriteFile(inputFile, []byte(`{"name": "Alice", "age": 30}`), 0644)

	cmd := exec.Command(bin, "generate", inputFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("generate failed: %v\n%s", err, output)
	}

	var schema map[string]interface{}
	if err := json.Unmarshal(output, &schema); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if schema["type"] != "object" {
		t.Errorf("expected type object, got %v", schema["type"])
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok || len(props) != 2 {
		t.Errorf("expected 2 properties, got %v", props)
	}
}

func TestCLIValidate(t *testing.T) {
	bin := buildBinary(t)
	defer os.Remove(bin)

	tmpDir := t.TempDir()

	// Create schema
	schemaFile := filepath.Join(tmpDir, "schema.json")
	os.WriteFile(schemaFile, []byte(`{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"age": {"type": "number"}
		},
		"required": ["name"]
	}`), 0644)

	// Valid data
	validFile := filepath.Join(tmpDir, "valid.json")
	os.WriteFile(validFile, []byte(`{"name": "Alice", "age": 30}`), 0644)

	// Invalid data
	invalidFile := filepath.Join(tmpDir, "invalid.json")
	os.WriteFile(invalidFile, []byte(`{"age": 30}`), 0644)

	// Validate valid data
	cmd := exec.Command(bin, "validate", schemaFile, validFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("validate failed: %v\n%s", err, output)
	}
	if !bytes.Contains(output, []byte("valid")) {
		t.Errorf("expected valid, got: %s", output)
	}

	// Validate invalid data
	cmd = exec.Command(bin, "validate", schemaFile, invalidFile)
	output, err = cmd.CombinedOutput()
	if err == nil {
		t.Error("expected exit code 1 for invalid data")
	}
	if !bytes.Contains(output, []byte("error")) {
		t.Errorf("expected error message, got: %s", output)
	}
}

func TestCLIDiff(t *testing.T) {
	bin := buildBinary(t)
	defer os.Remove(bin)

	tmpDir := t.TempDir()

	schema1 := filepath.Join(tmpDir, "schema1.json")
	os.WriteFile(schema1, []byte(`{"type": "object", "properties": {"name": {"type": "string"}}}`), 0644)

	schema2 := filepath.Join(tmpDir, "schema2.json")
	os.WriteFile(schema2, []byte(`{"type": "object", "properties": {"name": {"type": "string"}, "email": {"type": "string"}}}`), 0644)

	cmd := exec.Command(bin, "diff", schema1, schema2)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("diff failed: %v\n%s", err, output)
	}

	if !bytes.Contains(output, []byte("differ")) {
		t.Errorf("expected diff output, got: %s", output)
	}
}

func TestCLIDiffJSON(t *testing.T) {
	bin := buildBinary(t)
	defer os.Remove(bin)

	tmpDir := t.TempDir()

	schema1 := filepath.Join(tmpDir, "s1.json")
	os.WriteFile(schema1, []byte(`{"type": "string"}`), 0644)

	schema2 := filepath.Join(tmpDir, "s2.json")
	os.WriteFile(schema2, []byte(`{"type": "number"}`), 0644)

	cmd := exec.Command(bin, "diff", "--output", "json", schema1, schema2)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("diff failed: %v\n%s", err, output)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if result["identical"] != false {
		t.Error("expected non-identical")
	}
}

func TestCLIMerge(t *testing.T) {
	bin := buildBinary(t)
	defer os.Remove(bin)

	tmpDir := t.TempDir()

	schema1 := filepath.Join(tmpDir, "s1.json")
	os.WriteFile(schema1, []byte(`{"type": "object", "properties": {"name": {"type": "string"}}}`), 0644)

	schema2 := filepath.Join(tmpDir, "s2.json")
	os.WriteFile(schema2, []byte(`{"type": "object", "properties": {"email": {"type": "string"}}}`), 0644)

	cmd := exec.Command(bin, "merge", schema1, schema2, "--title", "Merged")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("merge failed: %v\n%s", err, output)
	}

	var schema map[string]interface{}
	if err := json.Unmarshal(output, &schema); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if schema["title"] != "Merged" {
		t.Errorf("expected title Merged, got %v", schema["title"])
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok || len(props) != 2 {
		t.Errorf("expected 2 properties, got %v", props)
	}
}

func TestCLIDoc(t *testing.T) {
	bin := buildBinary(t)
	defer os.Remove(bin)

	tmpDir := t.TempDir()

	schemaFile := filepath.Join(tmpDir, "schema.json")
	os.WriteFile(schemaFile, []byte(`{
		"title": "User",
		"description": "A user object",
		"type": "object",
		"properties": {
			"name": {"type": "string", "description": "User name"}
		},
		"required": ["name"]
	}`), 0644)

	cmd := exec.Command(bin, "doc", schemaFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("doc failed: %v\n%s", err, output)
	}

	if !bytes.Contains(output, []byte("# User")) {
		t.Errorf("expected title in output, got: %s", output)
	}
}

func TestCLIInfo(t *testing.T) {
	bin := buildBinary(t)
	defer os.Remove(bin)

	tmpDir := t.TempDir()

	schemaFile := filepath.Join(tmpDir, "schema.json")
	os.WriteFile(schemaFile, []byte(`{
		"title": "Test",
		"type": "object",
		"properties": {
			"name": {"type": "string"}
		},
		"required": ["name"]
	}`), 0644)

	cmd := exec.Command(bin, "info", schemaFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("info failed: %v\n%s", err, output)
	}

	if !bytes.Contains(output, []byte("Schema Information")) {
		t.Errorf("expected info header, got: %s", output)
	}
	if !bytes.Contains(output, []byte("Test")) {
		t.Errorf("expected title, got: %s", output)
	}
}

func buildBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "schemaforge")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %v\n%s", err, output)
	}
	return bin
}
