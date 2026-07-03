# SchemaForge

A comprehensive JSON Schema toolkit for generating, validating, diffing, merging, and documenting JSON Schemas.

## Features

- **Generate** — Infer JSON Schema from JSON data samples with type detection, format detection (email, UUID, URI, date), and enum detection
- **Validate** — Validate JSON data against schemas with detailed error reporting and path tracking
- **Diff** — Compare two JSON Schemas and see exactly what changed (added/removed/modified properties, type changes, constraint changes)
- **Merge** — Combine multiple schemas into one with automatic property merging and required field intersection
- **Doc** — Generate human-readable documentation from schemas in Markdown, HTML, or plain text
- **Info** — Display schema statistics and property summaries

## Quick Start

```bash
# Install
go install github.com/EdgarOrtegaRamirez/schemaforge@latest

# Or build from source
git clone https://github.com/EdgarOrtegaRamirez/schemaforge.git
cd schemaforge
go build -o schemaforge .
```

## Usage

### Generate a schema from JSON data

```bash
# From a file
echo '{"name": "Alice", "age": 30, "email": "alice@example.com"}' > sample.json
schemaforge generate sample.json

# From stdin
echo '{"name": "Bob", "score": 95}' | schemaforge generate

# With title and description
schemaforge generate sample.json --title "User" --description "A user object"
```

Output:
```json
{
  "type": "object",
  "properties": {
    "email": {
      "type": "string",
      "format": "email",
      "minLength": 15,
      "maxLength": 19
    },
    "name": {
      "type": "string",
      "minLength": 5,
      "maxLength": 5
    },
    "age": {
      "type": "number",
      "minimum": 30,
      "maximum": 30
    }
  },
  "required": ["age", "email", "name"]
}
```

### Validate JSON against a schema

```bash
# Validate a file
schemaforge validate schema.json data.json

# Validate multiple files
schemaforge validate schema.json user1.json user2.json user3.json

# Output:
# ✓ user1.json: valid
# ✗ user2.json: 2 error(s)
#   - .email: missing required property "email"
#   - .age: expected type number, got string
```

### Diff two schemas

```bash
# Text output (default)
schemaforge diff schema-v1.json schema-v2.json
# Output:
# Schemas differ: 1 added, 0 removed, 1 modified
#
# [+] properties.email: property "email" added
# [~] required: required fields changed

# JSON output
schemaforge diff --output json schema-v1.json schema-v2.json
```

### Merge schemas

```bash
# Merge two schemas (object properties are combined)
schemaforge merge user-schema.json address-schema.json --title "UserWithAddress"

# Merge arrays (items schemas are combined)
schemaforge merge strings.json more-strings.json
```

### Generate documentation

```bash
# Markdown (default)
schemaforge doc schema.json

# HTML
schemaforge doc --format html schema.json > docs.html

# Plain text
schemaforge doc --format text schema.json
```

### Schema info

```bash
schemaforge info schema.json
# Output:
# Schema Information
# ==================
# Title:       User
# Description: A user object
# Type:        object
# Properties:  3
# Required:    2
# Enum values: 0
#
# Property Summary
# ----------------
#   age                  number
#   email                string (email)
#   name                 string
```

## Schema Support

SchemaForge supports JSON Schema draft-07 and 2020-12 features:

- Types: string, number, integer, boolean, array, object, null
- String constraints: minLength, maxLength, pattern, format
- Number constraints: minimum, maximum, exclusiveMinimum, exclusiveMaximum, multipleOf
- Array constraints: minItems, maxItems, uniqueItems, items
- Object constraints: properties, required, additionalProperties, minProperties, maxProperties
- Combinators: anyOf, oneOf, allOf, not
- Metadata: title, description, default, examples, enum, const
- Format detection: email, URI, UUID, date, date-time, IPv4, IPv6

## Architecture

```
schemaforge/
├── schema/
│   ├── types.go        # Core schema types and helpers
│   ├── generator.go    # Schema inference from JSON data
│   ├── validator.go    # JSON validation against schemas
│   ├── differ.go       # Schema comparison and diffing
│   ├── merger.go       # Schema merging
│   └── documenter.go   # Documentation generation
├── main.go             # CLI entry point
├── main_test.go        # CLI integration tests
└── go.mod
```

## License

MIT
