# AGENTS.md

## Project Overview

SchemaForge is a JSON Schema toolkit written in Go. It provides CLI commands and a Go library for working with JSON Schemas.

## Architecture

- `schema/types.go` — Core `Schema` struct with JSON serialization, `TypeOrArray` for flexible type fields, and helper methods
- `schema/generator.go` — Infers JSON Schema from JSON data samples using statistical analysis (type detection, format detection, enum detection)
- `schema/validator.go` — Validates JSON data against schemas with path-tracked error reporting
- `schema/differ.go` — Compares two schemas and produces a structured diff (added/removed/modified)
- `schema/merger.go` — Merges multiple schemas (combines object properties, intersects required fields, uses anyOf for incompatible types)
- `schema/documenter.go` — Generates Markdown, HTML, or plain text documentation from schemas
- `main.go` — CLI entry point using cobra with 6 commands: generate, validate, diff, merge, doc, info

## Building

```bash
go build -o schemaforge .
```

## Testing

```bash
# Unit tests (50 tests)
go test ./schema/ -v

# CLI integration tests (7 tests)
go test -v -run TestCLI

# All tests
go test ./... -v
```

## Key Design Decisions

1. **TypeOrArray** — JSON Schema's `type` field can be a string or array. Custom MarshalJSON/UnmarshalJSON handles both.
2. **AdditionalProperties** — Can be a boolean or a schema. Custom marshal/unmarshal handles both.
3. **Options merging** — GeneratorOptions uses merge semantics: provided values override defaults, zero values don't reset defaults.
4. **Diff output** — Sorted by path for consistent, diff-friendly output.
5. **Merge strategy** — Objects: combine properties, intersect required fields. Arrays: merge items schemas. Different types: use anyOf.

## Common Tasks

- **Add a new string format detector**: Add a function `is<Format>(s string) bool` in `generator.go` and register it in `detectStringFormat()`
- **Add a new CLI command**: Add a `cobra.Command` in `main.go` following the pattern of existing commands
- **Add a new diff check**: Add comparison logic in `diffSchema()` in `differ.go`
- **Add a new merge strategy**: Add handling in the appropriate `merge*()` function in `merger.go`
