package schema

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Schema represents a JSON Schema document (draft-07/2020-12 compatible)
type Schema struct {
	Schema          string                `json:"$schema,omitempty"`
	ID              string                `json:"$id,omitempty"`
	Title           string                `json:"title,omitempty"`
	Description     string                `json:"description,omitempty"`
	Type            TypeOrArray           `json:"type,omitempty"`
	Properties      map[string]*Schema    `json:"properties,omitempty"`
	Required        []string              `json:"required,omitempty"`
	Items           *Schema               `json:"items,omitempty"`
	AdditionalProps *AdditionalProperties `json:"additionalProperties,omitempty"`
	Enum            []interface{}         `json:"enum,omitempty"`
	Const           interface{}           `json:"const,omitempty"`
	Default         interface{}           `json:"default,omitempty"`
	Examples        []interface{}         `json:"examples,omitempty"`
	AllOf           []*Schema             `json:"allOf,omitempty"`
	AnyOf           []*Schema             `json:"anyOf,omitempty"`
	OneOf           []*Schema             `json:"oneOf,omitempty"`
	Not             *Schema               `json:"not,omitempty"`
	If              *Schema               `json:"if,omitempty"`
	Then            *Schema               `json:"then,omitempty"`
	Else            *Schema               `json:"else,omitempty"`
	Ref             string                `json:"$ref,omitempty"`
	DynamicRef      string                `json:"$dynamicRef,omitempty"`
	Defs            map[string]*Schema    `json:"$defs,omitempty"`
	Definitions     map[string]*Schema    `json:"definitions,omitempty"`

	// String constraints
	MinLength *int   `json:"minLength,omitempty"`
	MaxLength *int   `json:"maxLength,omitempty"`
	Pattern   string `json:"pattern,omitempty"`
	Format    string `json:"format,omitempty"`

	// Number constraints
	MultipleOf       *float64 `json:"multipleOf,omitempty"`
	Minimum          *float64 `json:"minimum,omitempty"`
	Maximum          *float64 `json:"maximum,omitempty"`
	ExclusiveMinimum *float64 `json:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum *float64 `json:"exclusiveMaximum,omitempty"`

	// Array constraints
	MinItems    *int    `json:"minItems,omitempty"`
	MaxItems    *int    `json:"maxItems,omitempty"`
	UniqueItems *bool   `json:"uniqueItems,omitempty"`
	Contains    *Schema `json:"contains,omitempty"`

	// Object constraints
	MinProperties *int               `json:"minProperties,omitempty"`
	MaxProperties *int               `json:"maxProperties,omitempty"`
	PatternProps  map[string]*Schema `json:"patternProperties,omitempty"`
	PropertyNames *Schema            `json:"propertyNames,omitempty"`

	// Deprecated
	Deprecated *bool `json:"deprecated,omitempty"`

	// Read-only/write-only
	ReadOnly  *bool `json:"readOnly,omitempty"`
	WriteOnly *bool `json:"writeOnly,omitempty"`

	// ContentMediaType, ContentEncoding
	ContentMediaType string `json:"contentMediaType,omitempty"`
	ContentEncoding  string `json:"contentEncoding,omitempty"`
}

// TypeOrArray handles JSON Schema's "type" field which can be a string or array
type TypeOrArray struct {
	Types []string
}

func (t *TypeOrArray) MarshalJSON() ([]byte, error) {
	if len(t.Types) == 0 {
		return nil, nil
	}
	if len(t.Types) == 1 {
		return json.Marshal(t.Types[0])
	}
	return json.Marshal(t.Types)
}

func (t *TypeOrArray) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		t.Types = []string{s}
		return nil
	}
	var arr []string
	if err := json.Unmarshal(data, &arr); err == nil {
		t.Types = arr
		return nil
	}
	return fmt.Errorf("type must be a string or array of strings")
}

func (t *TypeOrArray) HasType(typ string) bool {
	for _, t2 := range t.Types {
		if t2 == typ {
			return true
		}
	}
	return false
}

func (t *TypeOrArray) PrimaryType() string {
	if len(t.Types) > 0 {
		return t.Types[0]
	}
	return ""
}

// AdditionalProperties can be bool or schema
type AdditionalProperties struct {
	Allowed bool
	Schema  *Schema
}

func (a *AdditionalProperties) MarshalJSON() ([]byte, error) {
	if a.Schema != nil {
		return json.Marshal(a.Schema)
	}
	return json.Marshal(a.Allowed)
}

func (a *AdditionalProperties) UnmarshalJSON(data []byte) error {
	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		a.Allowed = b
		return nil
	}
	var s Schema
	if err := json.Unmarshal(data, &s); err == nil {
		a.Schema = &s
		return nil
	}
	return fmt.Errorf("additionalProperties must be a boolean or schema")
}

// SchemaFromJSON parses a JSON byte slice into a Schema
func SchemaFromJSON(data []byte) (*Schema, error) {
	var s Schema
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// ToJSON serializes the schema to JSON bytes
func (s *Schema) ToJSON() ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}

// ToJSONCompact serializes the schema to compact JSON
func (s *Schema) ToJSONCompact() ([]byte, error) {
	return json.Marshal(s)
}

// Types returns the list of types this schema accepts
func (s *Schema) Types() []string {
	return s.Type.Types
}

// HasType checks if this schema accepts a given type
func (s *Schema) HasType(typ string) bool {
	return s.Type.HasType(typ)
}

// IsEmpty returns true if the schema has no constraints
func (s *Schema) IsEmpty() bool {
	return s.Title == "" && s.Description == "" &&
		len(s.Type.Types) == 0 && len(s.Properties) == 0 &&
		len(s.Required) == 0 && s.Items == nil &&
		len(s.Enum) == 0 && s.Const == nil &&
		len(s.AllOf) == 0 && len(s.AnyOf) == 0 && len(s.OneOf) == 0 &&
		s.Not == nil && s.Ref == ""
}

// CollectDefs collects all $defs and definitions into a single map
func (s *Schema) CollectDefs() map[string]*Schema {
	defs := make(map[string]*Schema)
	for k, v := range s.Defs {
		defs[k] = v
	}
	for k, v := range s.Definitions {
		defs[k] = v
	}
	return defs
}

// RequiredSet returns required fields as a set
func (s *Schema) RequiredSet() map[string]bool {
	set := make(map[string]bool, len(s.Required))
	for _, r := range s.Required {
		set[r] = true
	}
	return set
}

// SortedPropertyNames returns property names in sorted order
func (s *Schema) SortedPropertyNames() []string {
	names := make([]string, 0, len(s.Properties))
	for k := range s.Properties {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

// FormatString returns a human-readable description of the schema type
func (s *Schema) FormatString() string {
	if len(s.Type.Types) == 0 {
		if s.Ref != "" {
			return fmt.Sprintf("$ref: %s", s.Ref)
		}
		if len(s.AnyOf) > 0 {
			return "anyOf"
		}
		if len(s.OneOf) > 0 {
			return "oneOf"
		}
		if len(s.AllOf) > 0 {
			return "allOf"
		}
		return "any"
	}
	if len(s.Type.Types) == 1 {
		t := s.Type.Types[0]
		if t == "array" && s.Items != nil {
			inner := s.Items.FormatString()
			return fmt.Sprintf("array<%s>", inner)
		}
		if t == "object" && len(s.Properties) > 0 {
			return fmt.Sprintf("object{%d props}", len(s.Properties))
		}
		return t
	}
	return strings.Join(s.Type.Types, " | ")
}
