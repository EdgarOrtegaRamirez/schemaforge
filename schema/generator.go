package schema

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"
)

// Generator creates JSON Schema from JSON data samples
type Generator struct {
	samples []interface{}
	options GeneratorOptions
}

// GeneratorOptions configures schema generation behavior
type GeneratorOptions struct {
	// DetectFormat tries to detect string formats (email, date, uri, etc.)
	DetectFormat bool
	// InferEnums detects enum values from samples
	InferEnums bool
	// EnumThreshold minimum number of distinct values to consider enum
	EnumThreshold int
	// MaxEnumValues maximum number of enum values
	MaxEnumValues int
	// Title schema title
	Title string
	// Description schema description
	Description string
	// Strict mode requires all samples to match the same schema
	Strict bool
}

// DefaultGeneratorOptions returns sensible defaults
func DefaultGeneratorOptions() GeneratorOptions {
	return GeneratorOptions{
		DetectFormat:  true,
		InferEnums:    true,
		EnumThreshold: 2,
		MaxEnumValues: 50,
	}
}

// NewGenerator creates a new schema generator
func NewGenerator(opts ...GeneratorOptions) *Generator {
	opts2 := DefaultGeneratorOptions()
	if len(opts) > 0 {
		// Merge: use defaults, override with provided values
		if opts[0].DetectFormat {
			opts2.DetectFormat = true
		}
		if opts[0].InferEnums {
			opts2.InferEnums = true
		}
		if opts[0].EnumThreshold > 0 {
			opts2.EnumThreshold = opts[0].EnumThreshold
		}
		if opts[0].MaxEnumValues > 0 {
			opts2.MaxEnumValues = opts[0].MaxEnumValues
		}
		if opts[0].Title != "" {
			opts2.Title = opts[0].Title
		}
		if opts[0].Description != "" {
			opts2.Description = opts[0].Description
		}
		opts2.Strict = opts[0].Strict
	}
	return &Generator{
		options: opts2,
	}
}

// AddSample adds a JSON value sample for schema inference
func (g *Generator) AddSample(data interface{}) {
	g.samples = append(g.samples, data)
}

// AddSamples adds multiple samples
func (g *Generator) AddSamples(data ...interface{}) {
	g.samples = append(g.samples, data...)
}

// AddFromJSON parses and adds a JSON byte slice as a sample
func (g *Generator) AddFromJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	g.AddSample(v)
	return nil
}

// AddFromJSONString parses and adds a JSON string as a sample
func (g *Generator) AddFromJSONString(data string) error {
	return g.AddFromJSON([]byte(data))
}

// Generate creates a JSON Schema from the added samples
func (g *Generator) Generate() (*Schema, error) {
	if len(g.samples) == 0 {
		return nil, fmt.Errorf("no samples provided")
	}

	schema := g.inferType(g.samples)

	if g.options.Title != "" {
		schema.Title = g.options.Title
	}
	if g.options.Description != "" {
		schema.Description = g.options.Description
	}

	return schema, nil
}

// inferType infers the JSON Schema type from a set of values
func (g *Generator) inferType(values []interface{}) *Schema {
	if len(values) == 0 {
		return &Schema{}
	}

	// Collect all types
	typeSet := make(map[string]bool)
	for _, v := range values {
		t := jsonType(v)
		typeSet[t] = true
	}

	// If all same type, infer specific schema
	if len(typeSet) == 1 {
		for t := range typeSet {
			switch t {
			case "object":
				return g.inferObject(values)
			case "array":
				return g.inferArray(values)
			case "string":
				return g.inferString(values)
			case "number", "integer":
				return g.inferNumber(values)
			case "boolean":
				return &Schema{Type: TypeOrArray{Types: []string{"boolean"}}}
			case "null":
				return &Schema{Type: TypeOrArray{Types: []string{"null"}}}
			}
		}
	}

	// Multiple types - use anyOf or simple union
	types := make([]string, 0, len(typeSet))
	for t := range typeSet {
		types = append(types, t)
	}
	sort.Strings(types)

	if g.options.Strict {
		return &Schema{Type: TypeOrArray{Types: types}}
	}

	// Check if all values are the same
	anyOf := make([]*Schema, 0)
	for t := range typeSet {
		sameTypeValues := make([]interface{}, 0)
		for _, v := range values {
			if jsonType(v) == t {
				sameTypeValues = append(sameTypeValues, v)
			}
		}
		if len(sameTypeValues) == 1 {
			anyOf = append(anyOf, g.inferType(sameTypeValues))
		} else {
			anyOf = append(anyOf, g.inferType(sameTypeValues))
		}
	}

	return &Schema{
		AnyOf: anyOf,
	}
}

// inferObject infers schema for object values
func (g *Generator) inferObject(values []interface{}) *Schema {
	// Merge all keys
	allKeys := make(map[string][]interface{})
	keyCounts := make(map[string]int)

	for _, v := range values {
		obj, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		for k, val := range obj {
			allKeys[k] = append(allKeys[k], val)
			keyCounts[k]++
		}
	}

	schema := &Schema{
		Type:       TypeOrArray{Types: []string{"object"}},
		Properties: make(map[string]*Schema),
	}

	// Determine required fields (present in ALL samples)
	sampleCount := len(values)
	for k, count := range keyCounts {
		if count == sampleCount {
			schema.Required = append(schema.Required, k)
		}
	}
	sort.Strings(schema.Required)

	// Infer type for each property
	for k, vals := range allKeys {
		propSchema := g.inferType(vals)
		schema.Properties[k] = propSchema
	}

	return schema
}

// inferArray infers schema for array values
func (g *Generator) inferArray(values []interface{}) *Schema {
	// Collect all items
	allItems := make([]interface{}, 0)
	for _, v := range values {
		arr, ok := v.([]interface{})
		if !ok {
			continue
		}
		allItems = append(allItems, arr...)
	}

	schema := &Schema{
		Type: TypeOrArray{Types: []string{"array"}},
	}

	if len(allItems) > 0 {
		schema.Items = g.inferType(allItems)
	}

	// Check min/max items
	minItems := len(allItems)
	maxItems := 0
	for _, v := range values {
		arr, ok := v.([]interface{})
		if !ok {
			continue
		}
		if len(arr) < minItems {
			minItems = len(arr)
		}
		if len(arr) > maxItems {
			maxItems = len(arr)
		}
	}

	if minItems == maxItems && minItems > 0 {
		schema.MinItems = &minItems
		schema.MaxItems = &maxItems
	} else if minItems > 0 {
		schema.MinItems = &minItems
	}

	return schema
}

// inferString infers schema for string values
func (g *Generator) inferString(values []interface{}) *Schema {
	strs := make([]string, 0, len(values))
	for _, v := range values {
		s, ok := v.(string)
		if !ok {
			continue
		}
		strs = append(strs, s)
	}

	schema := &Schema{
		Type: TypeOrArray{Types: []string{"string"}},
	}

	if len(strs) == 0 {
		return schema
	}

	// Min/max length
	minLen := len(strs[0])
	maxLen := len(strs[0])
	for _, s := range strs {
		if len(s) < minLen {
			minLen = len(s)
		}
		if len(s) > maxLen {
			maxLen = len(s)
		}
	}

	if minLen == maxLen {
		schema.MinLength = &minLen
		schema.MaxLength = &maxLen
	} else {
		schema.MinLength = &minLen
		schema.MaxLength = &maxLen
	}

	// Detect format
	if g.options.DetectFormat {
		format := detectStringFormat(strs)
		if format != "" {
			schema.Format = format
		}
	}

	// Detect enum
	if g.options.InferEnums && len(strs) <= g.options.MaxEnumValues {
		unique := make(map[string]bool)
		for _, s := range strs {
			unique[s] = true
		}
		if len(strs) >= g.options.EnumThreshold && len(unique) <= g.options.MaxEnumValues {
			enumVals := make([]interface{}, 0, len(unique))
			for s := range unique {
				enumVals = append(enumVals, s)
			}
			sort.Slice(enumVals, func(i, j int) bool {
				return enumVals[i].(string) < enumVals[j].(string)
			})
			schema.Enum = enumVals
		}
	}

	return schema
}

// inferNumber infers schema for numeric values
func (g *Generator) inferNumber(values []interface{}) *Schema {
	nums := make([]float64, 0, len(values))
	allInt := true
	for _, v := range values {
		switch n := v.(type) {
		case float64:
			nums = append(nums, n)
			if n != math.Trunc(n) {
				allInt = false
			}
		case json.Number:
			f, err := n.Float64()
			if err == nil {
				nums = append(nums, f)
				if f != math.Trunc(f) {
					allInt = false
				}
			}
		case int:
			nums = append(nums, float64(n))
		case int64:
			nums = append(nums, float64(n))
		}
	}

	typ := "number"
	if allInt {
		typ = "integer"
	}

	schema := &Schema{
		Type: TypeOrArray{Types: []string{typ}},
	}

	if len(nums) == 0 {
		return schema
	}

	// Min/max
	minVal := nums[0]
	maxVal := nums[0]
	for _, n := range nums {
		if n < minVal {
			minVal = n
		}
		if n > maxVal {
			maxVal = n
		}
	}

	if minVal == maxVal {
		schema.Const = minVal
	} else {
		schema.Minimum = &minVal
		schema.Maximum = &maxVal
	}

	return schema
}

// detectStringFormat tries to detect the format of string values
func detectStringFormat(strs []string) string {
	if len(strs) == 0 {
		return ""
	}

	formats := map[string]func(string) bool{
		"email":     isEmail,
		"uri":       isURI,
		"uuid":      isUUID,
		"date":      isDate,
		"date-time": isDateTime,
		"ipv4":      isIPv4,
		"ipv6":      isIPv6,
	}

	for format, check := range formats {
		allMatch := true
		for _, s := range strs {
			if !check(s) {
				allMatch = false
				break
			}
		}
		if allMatch {
			return format
		}
	}

	return ""
}

// Simple format detection functions
func isEmail(s string) bool {
	return strings.Contains(s, "@") && strings.Contains(s, ".") && len(s) > 5
}

func isURI(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") || strings.HasPrefix(s, "ftp://")
}

func isUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, c := range s {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
		} else if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func isDate(s string) bool {
	if len(s) != 10 {
		return false
	}
	return s[4] == '-' && s[7] == '-'
}

func isDateTime(s string) bool {
	return strings.Contains(s, "T") && len(s) >= 19
}

func isIPv4(s string) bool {
	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		return false
	}
	for _, p := range parts {
		if len(p) == 0 || len(p) > 3 {
			return false
		}
		for _, c := range p {
			if c < '0' || c > '9' {
				return false
			}
		}
	}
	return true
}

func isIPv6(s string) bool {
	return strings.Contains(s, ":") && len(s) > 6
}

// jsonType returns the JSON type name for a Go value
func jsonType(v interface{}) string {
	if v == nil {
		return "null"
	}
	switch v.(type) {
	case bool:
		return "boolean"
	case float64, int, int64, json.Number:
		return "number"
	case string:
		return "string"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		// Try reflection for other numeric types
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return "number"
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return "number"
		case reflect.Float32, reflect.Float64:
			return "number"
		}
		return "unknown"
	}
}
