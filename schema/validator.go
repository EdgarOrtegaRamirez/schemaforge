package schema

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strings"
)

// ValidationError represents a single validation error
type ValidationError struct {
	Path    string `json:"path"`
	Message string `json:"message"`
	SchemaPath string `json:"schema_path,omitempty"`
}

func (e ValidationError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("%s: %s", e.Path, e.Message)
	}
	return e.Message
}

// ValidationResult contains the result of schema validation
type ValidationResult struct {
	Valid  bool               `json:"valid"`
	Errors []ValidationError  `json:"errors,omitempty"`
}

func (r *ValidationResult) AddError(path, message, schemaPath string) {
	r.Valid = false
	r.Errors = append(r.Errors, ValidationError{
		Path:       path,
		Message:    message,
		SchemaPath: schemaPath,
	})
}

// Validator validates JSON data against a JSON Schema
type Validator struct {
	schema *Schema
}

// NewValidator creates a new validator for the given schema
func NewValidator(schema *Schema) *Validator {
	return &Validator{schema: schema}
}

// Validate validates a JSON value against the schema
func (v *Validator) Validate(data interface{}) *ValidationResult {
	result := &ValidationResult{Valid: true}
	v.validateNode(v.schema, data, "", result)
	return result
}

// ValidateJSON validates a JSON byte slice against the schema
func (v *Validator) ValidateJSON(data []byte) (*ValidationResult, error) {
	var parsed interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	return v.Validate(parsed), nil
}

func (v *Validator) validateNode(schema *Schema, data interface{}, path string, result *ValidationResult) {
	if schema == nil {
		return
	}

	// Check $ref
	if schema.Ref != "" {
		return // Simplified - don't resolve refs for now
	}

	// Check const
	if schema.Const != nil {
		if !deepEqual(data, schema.Const) {
			result.AddError(path, fmt.Sprintf("expected const value %v", schema.Const), "")
			return
		}
	}

	// Check enum
	if len(schema.Enum) > 0 {
		found := false
		for _, e := range schema.Enum {
			if deepEqual(data, e) {
				found = true
				break
			}
		}
		if !found {
			result.AddError(path, fmt.Sprintf("value not in enum %v", schema.Enum), "")
			return
		}
	}

	// If no type constraint, skip type checking
	if len(schema.Type.Types) == 0 && schema.Ref == "" && len(schema.AnyOf) == 0 && len(schema.OneOf) == 0 && len(schema.AllOf) == 0 {
		return
	}

	// Check anyOf
	if len(schema.AnyOf) > 0 {
		anyValid := false
		for _, sub := range schema.AnyOf {
			vr := &ValidationResult{Valid: true}
			v.validateNode(sub, data, path, vr)
			if vr.Valid {
				anyValid = true
				break
			}
		}
		if !anyValid {
			result.AddError(path, "value does not match any of the schemas in anyOf", "")
		}
		return
	}

	// Check allOf
	if len(schema.AllOf) > 0 {
		for _, sub := range schema.AllOf {
			v.validateNode(sub, data, path, result)
		}
		return
	}

	// Check oneOf
	if len(schema.OneOf) > 0 {
		matchCount := 0
		for _, sub := range schema.OneOf {
			vr := &ValidationResult{Valid: true}
			v.validateNode(sub, data, path, vr)
			if vr.Valid {
				matchCount++
			}
		}
		if matchCount != 1 {
			result.AddError(path, fmt.Sprintf("expected exactly one match in oneOf, got %d", matchCount), "")
		}
		return
	}

	// Type checking
	actualType := jsonType(data)
	typeMatch := false
	for _, t := range schema.Type.Types {
		if t == actualType || (t == "number" && actualType == "integer") {
			typeMatch = true
			break
		}
	}

	if !typeMatch {
		result.AddError(path, fmt.Sprintf("expected type %s, got %s", strings.Join(schema.Type.Types, " or "), actualType), "")
		return
	}

	// Type-specific validation
	switch actualType {
	case "object":
		v.validateObject(schema, data.(map[string]interface{}), path, result)
	case "array":
		v.validateArray(schema, data.([]interface{}), path, result)
	case "string":
		v.validateString(schema, data.(string), path, result)
	case "number":
		v.validateNumber(schema, data, path, result)
	}
}

func (v *Validator) validateObject(schema *Schema, obj map[string]interface{}, path string, result *ValidationResult) {
	// Required properties
	for _, req := range schema.Required {
		if _, ok := obj[req]; !ok {
			result.AddError(path, fmt.Sprintf("missing required property %q", req), "")
		}
	}

	// Validate each property
	for key, val := range obj {
		propPath := joinPath(path, key)
		if propSchema, ok := schema.Properties[key]; ok {
			v.validateNode(propSchema, val, propPath, result)
		} else if schema.AdditionalProps != nil && !schema.AdditionalProps.Allowed {
			if schema.AdditionalProps.Schema != nil {
				v.validateNode(schema.AdditionalProps.Schema, val, propPath, result)
			} else {
				result.AddError(propPath, "additional property not allowed", "")
			}
		}
	}

	// Pattern properties
	for pattern, propSchema := range schema.PatternProps {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}
		for key, val := range obj {
			if re.MatchString(key) {
				propPath := joinPath(path, key)
				v.validateNode(propSchema, val, propPath, result)
			}
		}
	}

	// Min/max properties
	if schema.MinProperties != nil && len(obj) < *schema.MinProperties {
		result.AddError(path, fmt.Sprintf("object has %d properties, minimum is %d", len(obj), *schema.MinProperties), "")
	}
	if schema.MaxProperties != nil && len(obj) > *schema.MaxProperties {
		result.AddError(path, fmt.Sprintf("object has %d properties, maximum is %d", len(obj), *schema.MaxProperties), "")
	}
}

func (v *Validator) validateArray(schema *Schema, arr []interface{}, path string, result *ValidationResult) {
	// Min/max items
	if schema.MinItems != nil && len(arr) < *schema.MinItems {
		result.AddError(path, fmt.Sprintf("array has %d items, minimum is %d", len(arr), *schema.MinItems), "")
	}
	if schema.MaxItems != nil && len(arr) > *schema.MaxItems {
		result.AddError(path, fmt.Sprintf("array has %d items, maximum is %d", len(arr), *schema.MaxItems), "")
	}

	// Unique items
	if schema.UniqueItems != nil && *schema.UniqueItems {
		seen := make(map[string]bool)
		for i, item := range arr {
			key := fmt.Sprintf("%v", item)
			if seen[key] {
				result.AddError(joinPath(path, fmt.Sprintf("[%d]", i)), "duplicate item in uniqueItems array", "")
			}
			seen[key] = true
		}
	}

	// Validate items
	if schema.Items != nil {
		for i, item := range arr {
			itemPath := fmt.Sprintf("%s[%d]", path, i)
			v.validateNode(schema.Items, item, itemPath, result)
		}
	}
}

func (v *Validator) validateString(schema *Schema, s string, path string, result *ValidationResult) {
	if schema.MinLength != nil && len(s) < *schema.MinLength {
		result.AddError(path, fmt.Sprintf("string length %d is less than minimum %d", len(s), *schema.MinLength), "")
	}
	if schema.MaxLength != nil && len(s) > *schema.MaxLength {
		result.AddError(path, fmt.Sprintf("string length %d exceeds maximum %d", len(s), *schema.MaxLength), "")
	}
	if schema.Pattern != "" {
		re, err := regexp.Compile(schema.Pattern)
		if err == nil && !re.MatchString(s) {
			result.AddError(path, fmt.Sprintf("string does not match pattern %q", schema.Pattern), "")
		}
	}
	if schema.Format != "" {
		if !validateFormat(s, schema.Format) {
			result.AddError(path, fmt.Sprintf("string does not match format %q", schema.Format), "")
		}
	}
}

func (v *Validator) validateNumber(schema *Schema, data interface{}, path string, result *ValidationResult) {
	var n float64
	switch v := data.(type) {
	case float64:
		n = v
	case json.Number:
		f, err := v.Float64()
		if err != nil {
			return
		}
		n = f
	}

	if schema.Minimum != nil && n < *schema.Minimum {
		result.AddError(path, fmt.Sprintf("value %g is less than minimum %g", n, *schema.Minimum), "")
	}
	if schema.Maximum != nil && n > *schema.Maximum {
		result.AddError(path, fmt.Sprintf("value %g exceeds maximum %g", n, *schema.Maximum), "")
	}
	if schema.ExclusiveMinimum != nil && n <= *schema.ExclusiveMinimum {
		result.AddError(path, fmt.Sprintf("value %g must be greater than %g", n, *schema.ExclusiveMinimum), "")
	}
	if schema.ExclusiveMaximum != nil && n >= *schema.ExclusiveMaximum {
		result.AddError(path, fmt.Sprintf("value %g must be less than %g", n, *schema.ExclusiveMaximum), "")
	}
	if schema.MultipleOf != nil && *schema.MultipleOf != 0 {
		if math.Remainder(n, *schema.MultipleOf) != 0 {
			result.AddError(path, fmt.Sprintf("value %g is not a multiple of %g", n, *schema.MultipleOf), "")
		}
	}
}

func validateFormat(s, format string) bool {
	switch format {
	case "email":
		return isEmail(s)
	case "uri", "url":
		return isURI(s)
	case "uuid":
		return isUUID(s)
	case "date":
		return isDate(s)
	case "date-time":
		return isDateTime(s)
	case "ipv4":
		return isIPv4(s)
	case "ipv6":
		return isIPv6(s)
	case "hostname":
		return len(s) > 0 && !strings.Contains(s, " ")
	case "ipv4-cidr", "ipv6-cidr":
		return strings.Contains(s, "/")
	default:
		return true // Unknown formats pass
	}
}

func joinPath(base, field string) string {
	if base == "" {
		return field
	}
	if strings.HasPrefix(field, "[") {
		return base + field
	}
	return base + "." + field
}

func deepEqual(a, b interface{}) bool {
	// Handle json.Number
	if aNum, ok := a.(json.Number); ok {
		if bNum, ok := b.(json.Number); ok {
			return aNum.String() == bNum.String()
		}
		if bFloat, ok := b.(float64); ok {
			aFloat, err := aNum.Float64()
			if err != nil {
				return false
			}
			return aFloat == bFloat
		}
	}
	if bNum, ok := b.(json.Number); ok {
		if aFloat, ok := a.(float64); ok {
			bFloat, err := bNum.Float64()
			if err != nil {
				return false
			}
			return aFloat == bFloat
		}
	}
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}
