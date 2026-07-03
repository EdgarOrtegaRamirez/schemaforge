package schema

import (
	"testing"
)

func TestValidateString(t *testing.T) {
	s := &Schema{Type: TypeOrArray{Types: []string{"string"}}}
	v := NewValidator(s)

	result := v.Validate("hello")
	if !result.Valid {
		t.Error("expected valid")
	}

	result = v.Validate(42)
	if result.Valid {
		t.Error("expected invalid for number against string schema")
	}
}

func TestValidateNumber(t *testing.T) {
	min := 0.0
	max := 100.0
	s := &Schema{
		Type:    TypeOrArray{Types: []string{"number"}},
		Minimum: &min,
		Maximum: &max,
	}
	v := NewValidator(s)

	result := v.Validate(50.0)
	if !result.Valid {
		t.Error("expected valid")
	}

	result = v.Validate(150.0)
	if result.Valid {
		t.Error("expected invalid for value above max")
	}

	result = v.Validate(-10.0)
	if result.Valid {
		t.Error("expected invalid for value below min")
	}
}

func TestValidateObject(t *testing.T) {
	s := &Schema{
		Type: TypeOrArray{Types: []string{"object"}},
		Properties: map[string]*Schema{
			"name": {Type: TypeOrArray{Types: []string{"string"}}},
			"age":  {Type: TypeOrArray{Types: []string{"number"}}},
		},
		Required: []string{"name"},
	}
	v := NewValidator(s)

	// Valid
	result := v.Validate(map[string]interface{}{
		"name": "Alice",
		"age":  30.0,
	})
	if !result.Valid {
		t.Errorf("expected valid, got errors: %v", result.Errors)
	}

	// Valid without optional field
	result = v.Validate(map[string]interface{}{
		"name": "Bob",
	})
	if !result.Valid {
		t.Errorf("expected valid without optional field, got errors: %v", result.Errors)
	}

	// Invalid - missing required
	result = v.Validate(map[string]interface{}{
		"age": 25.0,
	})
	if result.Valid {
		t.Error("expected invalid for missing required field")
	}

	// Invalid - wrong type
	result = v.Validate(map[string]interface{}{
		"name": 123,
	})
	if result.Valid {
		t.Error("expected invalid for wrong type")
	}
}

func TestValidateArray(t *testing.T) {
	minItems := 1
	maxItems := 3
	s := &Schema{
		Type:     TypeOrArray{Types: []string{"array"}},
		Items:    &Schema{Type: TypeOrArray{Types: []string{"string"}}},
		MinItems: &minItems,
		MaxItems: &maxItems,
	}
	v := NewValidator(s)

	// Valid
	result := v.Validate([]interface{}{"a", "b"})
	if !result.Valid {
		t.Errorf("expected valid, got errors: %v", result.Errors)
	}

	// Invalid - empty array
	result = v.Validate([]interface{}{})
	if result.Valid {
		t.Error("expected invalid for empty array")
	}

	// Invalid - too many items
	result = v.Validate([]interface{}{"a", "b", "c", "d"})
	if result.Valid {
		t.Error("expected invalid for too many items")
	}

	// Invalid - wrong item type
	result = v.Validate([]interface{}{"a", 123})
	if result.Valid {
		t.Error("expected invalid for wrong item type")
	}
}

func TestValidateEnum(t *testing.T) {
	s := &Schema{
		Type: TypeOrArray{Types: []string{"string"}},
		Enum: []interface{}{"red", "green", "blue"},
	}
	v := NewValidator(s)

	result := v.Validate("red")
	if !result.Valid {
		t.Error("expected valid")
	}

	result = v.Validate("yellow")
	if result.Valid {
		t.Error("expected invalid for non-enum value")
	}
}

func TestValidateConst(t *testing.T) {
	s := &Schema{
		Const: "fixed-value",
	}
	v := NewValidator(s)

	result := v.Validate("fixed-value")
	if !result.Valid {
		t.Error("expected valid")
	}

	result = v.Validate("other")
	if result.Valid {
		t.Error("expected invalid for non-const value")
	}
}

func TestValidateStringConstraints(t *testing.T) {
	minLen := 3
	maxLen := 10
	s := &Schema{
		Type:      TypeOrArray{Types: []string{"string"}},
		MinLength: &minLen,
		MaxLength: &maxLen,
		Pattern:   "^[a-z]+$",
	}
	v := NewValidator(s)

	// Valid
	result := v.Validate("hello")
	if !result.Valid {
		t.Errorf("expected valid, got errors: %v", result.Errors)
	}

	// Invalid - too short
	result = v.Validate("hi")
	if result.Valid {
		t.Error("expected invalid for too short")
	}

	// Invalid - too long
	result = v.Validate("this is too long")
	if result.Valid {
		t.Error("expected invalid for too long")
	}

	// Invalid - doesn't match pattern
	result = v.Validate("Hello")
	if result.Valid {
		t.Error("expected invalid for pattern mismatch")
	}
}

func TestValidateAnyOf(t *testing.T) {
	s := &Schema{
		AnyOf: []*Schema{
			{Type: TypeOrArray{Types: []string{"string"}}},
			{Type: TypeOrArray{Types: []string{"number"}}},
		},
	}
	v := NewValidator(s)

	result := v.Validate("hello")
	if !result.Valid {
		t.Error("expected valid for string")
	}

	result = v.Validate(42.0)
	if !result.Valid {
		t.Error("expected valid for number")
	}

	result = v.Validate(true)
	if result.Valid {
		t.Error("expected invalid for boolean")
	}
}

func TestValidateNestedObject(t *testing.T) {
	s := &Schema{
		Type: TypeOrArray{Types: []string{"object"}},
		Properties: map[string]*Schema{
			"user": {
				Type: TypeOrArray{Types: []string{"object"}},
				Properties: map[string]*Schema{
					"name": {Type: TypeOrArray{Types: []string{"string"}}},
					"age":  {Type: TypeOrArray{Types: []string{"number"}}},
				},
				Required: []string{"name"},
			},
		},
		Required: []string{"user"},
	}
	v := NewValidator(s)

	// Valid
	result := v.Validate(map[string]interface{}{
		"user": map[string]interface{}{
			"name": "Alice",
			"age":  30.0,
		},
	})
	if !result.Valid {
		t.Errorf("expected valid, got errors: %v", result.Errors)
	}

	// Invalid - nested missing required
	result = v.Validate(map[string]interface{}{
		"user": map[string]interface{}{
			"age": 30.0,
		},
	})
	if result.Valid {
		t.Error("expected invalid for missing nested required")
	}
}

func TestValidateJSONBytes(t *testing.T) {
	s := &Schema{Type: TypeOrArray{Types: []string{"string"}}}
	v := NewValidator(s)

	result, err := v.ValidateJSON([]byte(`"hello"`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Valid {
		t.Error("expected valid")
	}

	result, err = v.ValidateJSON([]byte(`{invalid`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
