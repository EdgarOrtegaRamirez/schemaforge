package schema

import (
	"fmt"
	"sort"
)

// Merger combines multiple JSON Schemas into one
type Merger struct {
	schemas []*Schema
	options MergeOptions
}

// MergeOptions configures merge behavior
type MergeOptions struct {
	// Strict requires all schemas to be compatible
	Strict bool
	// Title for the merged schema
	Title string
	// Description for the merged schema
	Description string
}

// NewMerger creates a new schema merger
func NewMerger(opts ...MergeOptions) *Merger {
	opts2 := MergeOptions{}
	if len(opts) > 0 {
		opts2 = opts[0]
	}
	return &Merger{options: opts2}
}

// AddSchema adds a schema to merge
func (m *Merger) AddSchema(s *Schema) {
	m.schemas = append(m.schemas, s)
}

// Merge combines all added schemas into a single schema
func (m *Merger) Merge() (*Schema, error) {
	if len(m.schemas) == 0 {
		return nil, fmt.Errorf("no schemas to merge")
	}
	if len(m.schemas) == 1 {
		return m.schemas[0], nil
	}

	// Check if all schemas are the same type
	typeCounts := make(map[string]int)
	for _, s := range m.schemas {
		for _, t := range s.Type.Types {
			typeCounts[t]++
		}
	}

	// If all schemas have the same single type, merge them
	if len(typeCounts) == 1 {
		for t, count := range typeCounts {
			if count == len(m.schemas) {
				switch t {
				case "object":
					return m.mergeObjects()
				case "array":
					return m.mergeArrays()
				default:
					return m.mergeSimple(t)
				}
			}
		}
	}

	// Different types - use anyOf
	return m.mergeAnyOf()
}

func (m *Merger) mergeObjects() (*Schema, error) {
	merged := &Schema{
		Type:       TypeOrArray{Types: []string{"object"}},
		Properties: make(map[string]*Schema),
	}

	// Merge all properties
	propSchemas := make(map[string][]*Schema)
	for _, s := range m.schemas {
		for k, v := range s.Properties {
			propSchemas[k] = append(propSchemas[k], v)
		}
	}

	// Merge required fields
	reqCounts := make(map[string]int)
	for _, s := range m.schemas {
		for _, r := range s.Required {
			reqCounts[r]++
		}
	}
	// A field is required if it's required in ALL schemas
	for field, count := range reqCounts {
		if count == len(m.schemas) {
			merged.Required = append(merged.Required, field)
		}
	}
	sort.Strings(merged.Required)

	// Merge each property's schema
	for k, schemas := range propSchemas {
		if len(schemas) == 1 {
			merged.Properties[k] = schemas[0]
		} else {
			sub := NewMerger(m.options)
			for _, s := range schemas {
				sub.AddSchema(s)
			}
			mergedProp, err := sub.Merge()
			if err != nil {
				return nil, fmt.Errorf("failed to merge property %q: %w", k, err)
			}
			merged.Properties[k] = mergedProp
		}
	}

	// Merge title/description
	merged.Title = m.options.Title
	merged.Description = m.options.Description

	return merged, nil
}

func (m *Merger) mergeArrays() (*Schema, error) {
	merged := &Schema{
		Type: TypeOrArray{Types: []string{"array"}},
	}

	// Merge items schemas
	itemSchemas := make([]*Schema, 0)
	for _, s := range m.schemas {
		if s.Items != nil {
			itemSchemas = append(itemSchemas, s.Items)
		}
	}

	if len(itemSchemas) > 0 {
		sub := NewMerger(m.options)
		for _, s := range itemSchemas {
			sub.AddSchema(s)
		}
		mergedItems, err := sub.Merge()
		if err != nil {
			return nil, err
		}
		merged.Items = mergedItems
	}

	// Merge min/max items
	merged.MinItems = maxIntPtr(m.schemas, func(s *Schema) *int { return s.MinItems })
	merged.MaxItems = minIntPtr(m.schemas, func(s *Schema) *int { return s.MaxItems })

	merged.Title = m.options.Title
	merged.Description = m.options.Description

	return merged, nil
}

func (m *Merger) mergeSimple(typ string) (*Schema, error) {
	merged := &Schema{
		Type: TypeOrArray{Types: []string{typ}},
	}

	if typ == "string" {
		merged.MinLength = maxIntPtr(m.schemas, func(s *Schema) *int { return s.MinLength })
		merged.MaxLength = minIntPtr(m.schemas, func(s *Schema) *int { return s.MaxLength })

		// Merge formats
		formats := make(map[string]bool)
		for _, s := range m.schemas {
			if s.Format != "" {
				formats[s.Format] = true
			}
		}
		if len(formats) == 1 {
			for f := range formats {
				merged.Format = f
			}
		}

		// Merge enums
		enumSet := make(map[interface{}]bool)
		for _, s := range m.schemas {
			for _, e := range s.Enum {
				enumSet[e] = true
			}
		}
		if len(enumSet) > 0 {
			merged.Enum = make([]interface{}, 0, len(enumSet))
			for e := range enumSet {
				merged.Enum = append(merged.Enum, e)
			}
		}
	}

	if typ == "number" || typ == "integer" {
		merged.Minimum = maxFloatPtr(m.schemas, func(s *Schema) *float64 { return s.Minimum })
		merged.Maximum = minFloatPtr(m.schemas, func(s *Schema) *float64 { return s.Maximum })
	}

	merged.Title = m.options.Title
	merged.Description = m.options.Description

	return merged, nil
}

func (m *Merger) mergeAnyOf() (*Schema, error) {
	merged := &Schema{
		AnyOf: m.schemas,
		Title: m.options.Title,
		Description: m.options.Description,
	}
	return merged, nil
}

func maxIntPtr(schemas []*Schema, getter func(*Schema) *int) *int {
	var max *int
	for _, s := range schemas {
		v := getter(s)
		if v != nil {
			if max == nil || *v > *max {
				max = v
			}
		}
	}
	return max
}

func minIntPtr(schemas []*Schema, getter func(*Schema) *int) *int {
	var min *int
	for _, s := range schemas {
		v := getter(s)
		if v != nil {
			if min == nil || *v < *min {
				min = v
			}
		}
	}
	return min
}

func maxFloatPtr(schemas []*Schema, getter func(*Schema) *float64) *float64 {
	var max *float64
	for _, s := range schemas {
		v := getter(s)
		if v != nil {
			if max == nil || *v > *max {
				max = v
			}
		}
	}
	return max
}

func minFloatPtr(schemas []*Schema, getter func(*Schema) *float64) *float64 {
	var min *float64
	for _, s := range schemas {
		v := getter(s)
		if v != nil {
			if min == nil || *v < *min {
				min = v
			}
		}
	}
	return min
}
