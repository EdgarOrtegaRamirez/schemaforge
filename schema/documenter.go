package schema

import (
	"fmt"
	"sort"
	"strings"
)

// Documenter generates human-readable documentation from a JSON Schema
type Documenter struct {
	schema *Schema
}

// NewDocumenter creates a new documentation generator
func NewDocumenter(schema *Schema) *Documenter {
	return &Documenter{schema: schema}
}

// ToMarkdown generates Markdown documentation from the schema
func (d *Documenter) ToMarkdown() string {
	var sb strings.Builder

	sb.WriteString("---\n")
	if d.schema.Title != "" {
		sb.WriteString(fmt.Sprintf("title: %s\n", d.schema.Title))
	}
	sb.WriteString("---\n\n")

	if d.schema.Title != "" {
		sb.WriteString(fmt.Sprintf("# %s\n\n", d.schema.Title))
	}
	if d.schema.Description != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", d.schema.Description))
	}

	d.writeSchemaDocs(&sb, d.schema, "", 0)

	return sb.String()
}

// ToText generates plain text documentation
func (d *Documenter) ToText() string {
	var sb strings.Builder

	if d.schema.Title != "" {
		sb.WriteString(fmt.Sprintf("Schema: %s\n", d.schema.Title))
		sb.WriteString(strings.Repeat("=", len(d.schema.Title)+8) + "\n\n")
	}
	if d.schema.Description != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", d.schema.Description))
	}

	d.writeSchemaDocsText(&sb, d.schema, "", 0)

	return sb.String()
}

// ToHTML generates HTML documentation
func (d *Documenter) ToHTML() string {
	var sb strings.Builder

	sb.WriteString("<!DOCTYPE html>\n<html>\n<head>\n<meta charset=\"UTF-8\">\n")
	if d.schema.Title != "" {
		sb.WriteString(fmt.Sprintf("<title>%s</title>\n", d.schema.Title))
	}
	sb.WriteString("<style>\n")
	sb.WriteString("body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }\n")
	sb.WriteString("h1 { color: #2c3e50; }\n")
	sb.WriteString("h2 { color: #34495e; border-bottom: 1px solid #eee; padding-bottom: 5px; }\n")
	sb.WriteString(".property { margin: 10px 0; padding: 10px; background: #f8f9fa; border-radius: 5px; }\n")
	sb.WriteString(".type { color: #e74c3c; font-weight: bold; }\n")
	sb.WriteString(".required { color: #e74c3c; }\n")
	sb.WriteString(".optional { color: #95a5a6; }\n")
	sb.WriteString("code { background: #f1f1f1; padding: 2px 5px; border-radius: 3px; }\n")
	sb.WriteString("</style>\n</head>\n<body>\n")

	if d.schema.Title != "" {
		sb.WriteString(fmt.Sprintf("<h1>%s</h1>\n", d.schema.Title))
	}
	if d.schema.Description != "" {
		sb.WriteString(fmt.Sprintf("<p>%s</p>\n", d.schema.Description))
	}

	d.writeSchemaDocsHTML(&sb, d.schema, "", 0)

	sb.WriteString("</body>\n</html>")
	return sb.String()
}

func (d *Documenter) writeSchemaDocs(sb *strings.Builder, s *Schema, prefix string, depth int) {
	if s == nil {
		return
	}

	// Write root type info
	if len(s.Type.Types) > 0 {
		sb.WriteString(fmt.Sprintf("**Type:** `%s`\n\n", strings.Join(s.Type.Types, " | ")))
	}

	// Write properties
	if len(s.Properties) > 0 {
		sb.WriteString("## Properties\n\n")
		sb.WriteString("| Name | Type | Required | Description |\n")
		sb.WriteString("|------|------|----------|-------------|\n")

		reqSet := s.RequiredSet()
		for _, name := range s.SortedPropertyNames() {
			prop := s.Properties[name]
			req := "No"
			if reqSet[name] {
				req = "Yes"
			}
			desc := prop.Description
			if desc == "" {
				desc = "-"
			}
			typ := prop.FormatString()
			sb.WriteString(fmt.Sprintf("| `%s` | `%s` | %s | %s |\n", name, typ, req, desc))
		}
		sb.WriteString("\n")

		// Write detailed property docs
		for _, name := range s.SortedPropertyNames() {
			prop := s.Properties[name]
			sb.WriteString(fmt.Sprintf("### `%s`\n\n", name))
			d.writeSchemaDocs(sb, prop, prefix+name+".", depth+1)
			sb.WriteString("\n")
		}
	}

	// Write array items
	if s.Items != nil {
		sb.WriteString("## Items\n\n")
		d.writeSchemaDocs(sb, s.Items, prefix+"[].", depth+1)
		sb.WriteString("\n")
	}

	// Write constraints
	constraints := d.collectConstraints(s)
	if len(constraints) > 0 {
		sb.WriteString("## Constraints\n\n")
		for _, c := range constraints {
			sb.WriteString(fmt.Sprintf("- %s\n", c))
		}
		sb.WriteString("\n")
	}

	// Write enum values
	if len(s.Enum) > 0 {
		sb.WriteString("## Allowed Values\n\n")
		sb.WriteString("```\n")
		for _, e := range s.Enum {
			sb.WriteString(fmt.Sprintf("%v\n", e))
		}
		sb.WriteString("```\n\n")
	}

	// Write anyOf/oneOf
	if len(s.AnyOf) > 0 {
		sb.WriteString("## Any Of\n\n")
		for i, sub := range s.AnyOf {
			sb.WriteString(fmt.Sprintf("### Option %d\n\n", i+1))
			d.writeSchemaDocs(sb, sub, prefix, depth+1)
		}
	}

	if len(s.OneOf) > 0 {
		sb.WriteString("## One Of\n\n")
		for i, sub := range s.OneOf {
			sb.WriteString(fmt.Sprintf("### Option %d\n\n", i+1))
			d.writeSchemaDocs(sb, sub, prefix, depth+1)
		}
	}
}

func (d *Documenter) writeSchemaDocsText(sb *strings.Builder, s *Schema, prefix string, depth int) {
	if s == nil {
		return
	}

	indent := strings.Repeat("  ", depth)

	if len(s.Type.Types) > 0 {
		sb.WriteString(fmt.Sprintf("%sType: %s\n", indent, strings.Join(s.Type.Types, " | ")))
	}

	if len(s.Properties) > 0 {
		reqSet := s.RequiredSet()
		sb.WriteString(fmt.Sprintf("%sProperties:\n", indent))
		for _, name := range s.SortedPropertyNames() {
			prop := s.Properties[name]
			req := "optional"
			if reqSet[name] {
				req = "required"
			}
			sb.WriteString(fmt.Sprintf("%s  - %s (%s, %s)\n", indent, name, prop.FormatString(), req))
			if prop.Description != "" {
				sb.WriteString(fmt.Sprintf("%s    %s\n", indent, prop.Description))
			}
		}
	}

	constraints := d.collectConstraints(s)
	if len(constraints) > 0 {
		sb.WriteString(fmt.Sprintf("%sConstraints:\n", indent))
		for _, c := range constraints {
			sb.WriteString(fmt.Sprintf("%s  - %s\n", indent, c))
		}
	}

	if len(s.Enum) > 0 {
		sb.WriteString(fmt.Sprintf("%sEnum: %v\n", indent, s.Enum))
	}
}

func (d *Documenter) writeSchemaDocsHTML(sb *strings.Builder, s *Schema, prefix string, depth int) {
	if s == nil {
		return
	}

	if len(s.Type.Types) > 0 {
		sb.WriteString(fmt.Sprintf("<p><strong>Type:</strong> <code>%s</code></p>\n", strings.Join(s.Type.Types, " | ")))
	}

	if len(s.Properties) > 0 {
		sb.WriteString("<h2>Properties</h2>\n")
		reqSet := s.RequiredSet()
		for _, name := range s.SortedPropertyNames() {
			prop := s.Properties[name]
			sb.WriteString("<div class=\"property\">\n")
			sb.WriteString(fmt.Sprintf("  <strong><code>%s</code></strong> ", name))
			sb.WriteString(fmt.Sprintf("<span class=\"type\">%s</span> ", prop.FormatString()))
			if reqSet[name] {
				sb.WriteString("<span class=\"required\">required</span> ")
			} else {
				sb.WriteString("<span class=\"optional\">optional</span> ")
			}
			if prop.Description != "" {
				sb.WriteString(fmt.Sprintf("\n  <p>%s</p>\n", prop.Description))
			}
			sb.WriteString("</div>\n")
		}
	}
}

func (d *Documenter) collectConstraints(s *Schema) []string {
	var constraints []string

	if s.MinLength != nil {
		constraints = append(constraints, fmt.Sprintf("Min length: %d", *s.MinLength))
	}
	if s.MaxLength != nil {
		constraints = append(constraints, fmt.Sprintf("Max length: %d", *s.MaxLength))
	}
	if s.Pattern != "" {
		constraints = append(constraints, fmt.Sprintf("Pattern: `%s`", s.Pattern))
	}
	if s.Format != "" {
		constraints = append(constraints, fmt.Sprintf("Format: `%s`", s.Format))
	}
	if s.Minimum != nil {
		constraints = append(constraints, fmt.Sprintf("Minimum: %g", *s.Minimum))
	}
	if s.Maximum != nil {
		constraints = append(constraints, fmt.Sprintf("Maximum: %g", *s.Maximum))
	}
	if s.ExclusiveMinimum != nil {
		constraints = append(constraints, fmt.Sprintf("Exclusive minimum: %g", *s.ExclusiveMinimum))
	}
	if s.ExclusiveMaximum != nil {
		constraints = append(constraints, fmt.Sprintf("Exclusive maximum: %g", *s.ExclusiveMaximum))
	}
	if s.MultipleOf != nil {
		constraints = append(constraints, fmt.Sprintf("Multiple of: %g", *s.MultipleOf))
	}
	if s.MinItems != nil {
		constraints = append(constraints, fmt.Sprintf("Min items: %d", *s.MinItems))
	}
	if s.MaxItems != nil {
		constraints = append(constraints, fmt.Sprintf("Max items: %d", *s.MaxItems))
	}
	if s.UniqueItems != nil && *s.UniqueItems {
		constraints = append(constraints, "Unique items: true")
	}
	if s.MinProperties != nil {
		constraints = append(constraints, fmt.Sprintf("Min properties: %d", *s.MinProperties))
	}
	if s.MaxProperties != nil {
		constraints = append(constraints, fmt.Sprintf("Max properties: %d", *s.MaxProperties))
	}
	if s.Deprecated != nil && *s.Deprecated {
		constraints = append(constraints, "Deprecated: true")
	}
	if s.ReadOnly != nil && *s.ReadOnly {
		constraints = append(constraints, "Read-only: true")
	}
	if s.WriteOnly != nil && *s.WriteOnly {
		constraints = append(constraints, "Write-only: true")
	}

	sort.Strings(constraints)
	return constraints
}
