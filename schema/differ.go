package schema

import (
	"fmt"
	"sort"
)

// DiffType represents the type of difference between schemas
type DiffType string

const (
	DiffAdded    DiffType = "added"
	DiffRemoved  DiffType = "removed"
	DiffModified DiffType = "modified"
)

// SchemaDiff represents a single difference between two schemas
type SchemaDiff struct {
	Type     DiffType    `json:"type"`
	Path     string      `json:"path"`
	Message  string      `json:"message"`
	OldValue interface{} `json:"old_value,omitempty"`
	NewValue interface{} `json:"new_value,omitempty"`
}

// DiffResult contains the result of comparing two schemas
type DiffResult struct {
	Identical bool         `json:"identical"`
	Diffs     []SchemaDiff `json:"diffs,omitempty"`
	Summary   DiffSummary  `json:"summary"`
}

// DiffSummary provides counts of different diff types
type DiffSummary struct {
	Added    int `json:"added"`
	Removed  int `json:"removed"`
	Modified int `json:"modified"`
}

// DiffSchemas compares two JSON Schemas and returns their differences
func DiffSchemas(a, b *Schema) *DiffResult {
	result := &DiffResult{Identical: true}
	diffSchema(a, b, "", result)

	// Sort diffs by path for consistent output
	sort.Slice(result.Diffs, func(i, j int) bool {
		return result.Diffs[i].Path < result.Diffs[j].Path
	})

	// Build summary
	for _, d := range result.Diffs {
		switch d.Type {
		case DiffAdded:
			result.Summary.Added++
		case DiffRemoved:
			result.Summary.Removed++
		case DiffModified:
			result.Summary.Modified++
		}
	}

	if len(result.Diffs) > 0 {
		result.Identical = false
	}

	return result
}

func diffSchema(a, b *Schema, path string, result *DiffResult) {
	if a == nil && b == nil {
		return
	}
	if a == nil {
		addDiff(result, DiffAdded, path, fmt.Sprintf("schema added: %s", b.FormatString()), nil, b)
		return
	}
	if b == nil {
		addDiff(result, DiffRemoved, path, fmt.Sprintf("schema removed: %s", a.FormatString()), a, nil)
		return
	}

	// Compare type
	if !typesEqual(a.Type.Types, b.Type.Types) {
		addDiff(result, DiffModified, joinPath(path, "type"),
			fmt.Sprintf("type changed from %v to %v", a.Type.Types, b.Type.Types),
			a.Type.Types, b.Type.Types)
	}

	// Compare title
	if a.Title != b.Title {
		addDiff(result, DiffModified, joinPath(path, "title"),
			fmt.Sprintf("title changed from %q to %q", a.Title, b.Title),
			a.Title, b.Title)
	}

	// Compare description
	if a.Description != b.Description {
		addDiff(result, DiffModified, joinPath(path, "description"),
			fmt.Sprintf("description changed from %q to %q", a.Description, b.Description),
			a.Description, b.Description)
	}

	// Compare required
	if !stringSliceEqual(a.Required, b.Required) {
		addDiff(result, DiffModified, joinPath(path, "required"),
			"required fields changed",
			a.Required, b.Required)
	}

	// Compare enum
	if !interfaceSliceEqual(a.Enum, b.Enum) {
		addDiff(result, DiffModified, joinPath(path, "enum"),
			"enum values changed",
			a.Enum, b.Enum)
	}

	// Compare properties
	diffProperties(a.Properties, b.Properties, path, result)

	// Compare items
	if a.Items != nil || b.Items != nil {
		diffSchema(a.Items, b.Items, joinPath(path, "items"), result)
	}

	// Compare string constraints
	diffIntPtr("minLength", a.MinLength, b.MinLength, path, result)
	diffIntPtr("maxLength", a.MaxLength, b.MaxLength, path, result)
	if a.Pattern != b.Pattern {
		addDiff(result, DiffModified, joinPath(path, "pattern"),
			fmt.Sprintf("pattern changed from %q to %q", a.Pattern, b.Pattern),
			a.Pattern, b.Pattern)
	}
	if a.Format != b.Format {
		addDiff(result, DiffModified, joinPath(path, "format"),
			fmt.Sprintf("format changed from %q to %q", a.Format, b.Format),
			a.Format, b.Format)
	}

	// Compare number constraints
	diffFloatPtr("minimum", a.Minimum, b.Minimum, path, result)
	diffFloatPtr("maximum", a.Maximum, b.Maximum, path, result)
	diffFloatPtr("exclusiveMinimum", a.ExclusiveMinimum, b.ExclusiveMinimum, path, result)
	diffFloatPtr("exclusiveMaximum", a.ExclusiveMaximum, b.ExclusiveMaximum, path, result)
	diffFloatPtr("multipleOf", a.MultipleOf, b.MultipleOf, path, result)

	// Compare array constraints
	diffIntPtr("minItems", a.MinItems, b.MinItems, path, result)
	diffIntPtr("maxItems", a.MaxItems, b.MaxItems, path, result)

	// Compare object constraints
	diffIntPtr("minProperties", a.MinProperties, b.MinProperties, path, result)
	diffIntPtr("maxProperties", a.MaxProperties, b.MaxProperties, path, result)
}

func diffProperties(aProps, bProps map[string]*Schema, path string, result *DiffResult) {
	if aProps == nil {
		aProps = make(map[string]*Schema)
	}
	if bProps == nil {
		bProps = make(map[string]*Schema)
	}

	// Find all property names
	allKeys := make(map[string]bool)
	for k := range aProps {
		allKeys[k] = true
	}
	for k := range bProps {
		allKeys[k] = true
	}

	keys := make([]string, 0, len(allKeys))
	for k := range allKeys {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		propPath := joinPath(path, "properties."+k)
		aProp, aOk := aProps[k]
		bProp, bOk := bProps[k]

		if aOk && !bOk {
			addDiff(result, DiffRemoved, propPath,
				fmt.Sprintf("property %q removed", k), aProp, nil)
		} else if !aOk && bOk {
			addDiff(result, DiffAdded, propPath,
				fmt.Sprintf("property %q added", k), nil, bProp)
		} else if aOk && bOk {
			diffSchema(aProp, bProp, propPath, result)
		}
	}
}

func diffIntPtr(name string, a, b *int, path string, result *DiffResult) {
	if (a == nil) != (b == nil) {
		addDiff(result, DiffModified, joinPath(path, name),
			fmt.Sprintf("%s changed", name), a, b)
		return
	}
	if a != nil && b != nil && *a != *b {
		addDiff(result, DiffModified, joinPath(path, name),
			fmt.Sprintf("%s changed from %d to %d", name, *a, *b), *a, *b)
	}
}

func diffFloatPtr(name string, a, b *float64, path string, result *DiffResult) {
	if (a == nil) != (b == nil) {
		addDiff(result, DiffModified, joinPath(path, name),
			fmt.Sprintf("%s changed", name), a, b)
		return
	}
	if a != nil && b != nil && *a != *b {
		addDiff(result, DiffModified, joinPath(path, name),
			fmt.Sprintf("%s changed from %g to %g", name, *a, *b), *a, *b)
	}
}

func addDiff(result *DiffResult, diffType DiffType, path, message string, oldVal, newVal interface{}) {
	result.Diffs = append(result.Diffs, SchemaDiff{
		Type:     diffType,
		Path:     path,
		Message:  message,
		OldValue: oldVal,
		NewValue: newVal,
	})
}

func typesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	aCopy := make([]string, len(a))
	bCopy := make([]string, len(b))
	copy(aCopy, a)
	copy(bCopy, b)
	sort.Strings(aCopy)
	sort.Strings(bCopy)
	for i := range aCopy {
		if aCopy[i] != bCopy[i] {
			return false
		}
	}
	return true
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func interfaceSliceEqual(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if fmt.Sprintf("%v", a[i]) != fmt.Sprintf("%v", b[i]) {
			return false
		}
	}
	return true
}
