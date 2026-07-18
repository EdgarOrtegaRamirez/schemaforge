package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/EdgarOrtegaRamirez/schemaforge/schema"
	"github.com/spf13/cobra"
)

var version = "0.1.0"

func main() {
	rootCmd := &cobra.Command{
		Use:   "schemaforge",
		Short: "JSON Schema Toolkit - generate, validate, diff, merge, and document JSON Schemas",
		Long: `SchemaForge is a comprehensive CLI tool for working with JSON Schemas.
Generate schemas from JSON samples, validate data against schemas,
compare schema versions, merge schemas, and generate documentation.`,
		Version: version,
	}

	// Generate command
	generateCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate JSON Schema from JSON data samples",
		Long: `Analyze one or more JSON data samples and infer a JSON Schema.
Supports reading from files or stdin.`,
		RunE: runGenerate,
	}
	generateCmd.Flags().StringP("title", "t", "", "Schema title")
	generateCmd.Flags().StringP("description", "d", "", "Schema description")
	generateCmd.Flags().Bool("no-format", false, "Disable format detection")
	generateCmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")
	rootCmd.AddCommand(generateCmd)

	// Validate command
	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate JSON data against a JSON Schema",
		Long: `Validate JSON data files or stdin against a JSON Schema.
Reports all validation errors with paths.`,
		RunE: runValidate,
	}
	rootCmd.AddCommand(validateCmd)

	// Diff command
	diffCmd := &cobra.Command{
		Use:   "diff",
		Short: "Compare two JSON Schemas",
		Long: `Compare two JSON Schema files and show their differences.
Useful for tracking schema evolution.`,
		RunE: runDiff,
	}
	diffCmd.Flags().StringP("output", "o", "", "Output format: text, json (default: text)")
	rootCmd.AddCommand(diffCmd)

	// Merge command
	mergeCmd := &cobra.Command{
		Use:   "merge",
		Short: "Merge multiple JSON Schemas",
		Long: `Combine multiple JSON Schemas into a single schema.
Merges compatible schemas and uses anyOf for incompatible types.`,
		RunE: runMerge,
	}
	mergeCmd.Flags().StringP("title", "t", "", "Merged schema title")
	mergeCmd.Flags().StringP("description", "d", "", "Merged schema description")
	mergeCmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")
	rootCmd.AddCommand(mergeCmd)

	// Doc command
	docCmd := &cobra.Command{
		Use:   "doc",
		Short: "Generate documentation from a JSON Schema",
		Long: `Generate human-readable documentation from a JSON Schema.
Supports Markdown, plain text, and HTML output formats.`,
		RunE: runDoc,
	}
	docCmd.Flags().StringP("format", "f", "markdown", "Output format: markdown, text, html")
	docCmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")
	rootCmd.AddCommand(docCmd)

	// Info command
	infoCmd := &cobra.Command{
		Use:   "info",
		Short: "Display schema information and statistics",
		Long: `Show detailed information about a JSON Schema including
property count, type distribution, and constraint summary.`,
		RunE: runInfo,
	}
	rootCmd.AddCommand(infoCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runGenerate(cmd *cobra.Command, args []string) error {
	gen := schema.NewGenerator(schema.GeneratorOptions{
		DetectFormat: !cmd.Flag("no-format").Changed,
		Title:        cmd.Flag("title").Value.String(),
		Description:  cmd.Flag("description").Value.String(),
	})

	// Read from files or stdin
	if len(args) > 0 {
		for _, file := range args {
			data, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("reading %s: %w", file, err)
			}
			if err := gen.AddFromJSON(data); err != nil {
				return fmt.Errorf("parsing %s: %w", file, err)
			}
		}
	} else {
		// Read from stdin
		info, _ := os.Stdin.Stat()
		if (info.Mode() & os.ModeCharDevice) != 0 {
			return fmt.Errorf("no input provided. Use files as arguments or pipe JSON to stdin")
		}
		var data []byte
		buf := make([]byte, 4096)
		for {
			n, err := os.Stdin.Read(buf)
			data = append(data, buf[:n]...)
			if err != nil {
				break
			}
		}
		if err := gen.AddFromJSON(data); err != nil {
			return fmt.Errorf("parsing stdin: %w", err)
		}
	}

	schema, err := gen.Generate()
	if err != nil {
		return err
	}

	data, err := schema.ToJSON()
	if err != nil {
		return err
	}

	output := cmd.Flag("output").Value.String()
	if output != "" {
		return os.WriteFile(output, data, 0644)
	}
	fmt.Println(string(data))
	return nil
}

func runValidate(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: schemaforge validate <schema.json> <data.json> [data2.json ...]")
	}

	schemaData, err := os.ReadFile(args[0])
	if err != nil {
		return fmt.Errorf("reading schema: %w", err)
	}

	s, err := schema.SchemaFromJSON(schemaData)
	if err != nil {
		return fmt.Errorf("parsing schema: %w", err)
	}

	validator := schema.NewValidator(s)

	files := args[1:]
	if len(files) == 0 {
		return fmt.Errorf("no data files provided")
	}

	allValid := true
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("reading %s: %w", file, err)
		}

		result, err := validator.ValidateJSON(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error validating %s: %v\n", file, err)
			allValid = false
			continue
		}

		if result.Valid {
			fmt.Printf("✓ %s: valid\n", file)
		} else {
			allValid = false
			fmt.Printf("✗ %s: %d error(s)\n", file, len(result.Errors))
			for _, e := range result.Errors {
				fmt.Printf("  - %s\n", e.Error())
			}
		}
	}

	if !allValid {
		os.Exit(1)
	}
	return nil
}

func runDiff(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: schemaforge diff <schema1.json> <schema2.json>")
	}

	data1, err := os.ReadFile(args[0])
	if err != nil {
		return fmt.Errorf("reading %s: %w", args[0], err)
	}
	data2, err := os.ReadFile(args[1])
	if err != nil {
		return fmt.Errorf("reading %s: %w", args[1], err)
	}

	s1, err := schema.SchemaFromJSON(data1)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", args[0], err)
	}
	s2, err := schema.SchemaFromJSON(data2)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", args[1], err)
	}

	result := schema.DiffSchemas(s1, s2)

	outputFormat := cmd.Flag("output").Value.String()
	if outputFormat == "json" {
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	if result.Identical {
		fmt.Println("Schemas are identical")
		return nil
	}

	fmt.Printf("Schemas differ: %d added, %d removed, %d modified\n\n",
		result.Summary.Added, result.Summary.Removed, result.Summary.Modified)

	for _, d := range result.Diffs {
		symbol := "~"
		switch d.Type {
		case schema.DiffAdded:
			symbol = "+"
		case schema.DiffRemoved:
			symbol = "-"
		case schema.DiffModified:
			symbol = "~"
		}
		fmt.Printf("[%s] %s: %s\n", symbol, d.Path, d.Message)
	}

	return nil
}

func runMerge(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: schemaforge merge <schema1.json> <schema2.json> [...]")
	}

	merger := schema.NewMerger(schema.MergeOptions{
		Title:       cmd.Flag("title").Value.String(),
		Description: cmd.Flag("description").Value.String(),
	})

	for _, file := range args {
		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("reading %s: %w", file, err)
		}
		s, err := schema.SchemaFromJSON(data)
		if err != nil {
			return fmt.Errorf("parsing %s: %w", file, err)
		}
		merger.AddSchema(s)
	}

	merged, err := merger.Merge()
	if err != nil {
		return err
	}

	data, err := merged.ToJSON()
	if err != nil {
		return err
	}

	output := cmd.Flag("output").Value.String()
	if output != "" {
		return os.WriteFile(output, data, 0644)
	}
	fmt.Println(string(data))
	return nil
}

func runDoc(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: schemaforge doc <schema.json>")
	}

	data, err := os.ReadFile(args[0])
	if err != nil {
		return fmt.Errorf("reading schema: %w", err)
	}

	s, err := schema.SchemaFromJSON(data)
	if err != nil {
		return fmt.Errorf("parsing schema: %w", err)
	}

	doc := schema.NewDocumenter(s)

	format := cmd.Flag("format").Value.String()
	var output string
	switch strings.ToLower(format) {
	case "text", "txt":
		output = doc.ToText()
	case "html":
		output = doc.ToHTML()
	default:
		output = doc.ToMarkdown()
	}

	outFile := cmd.Flag("output").Value.String()
	if outFile != "" {
		return os.WriteFile(outFile, []byte(output), 0644)
	}
	fmt.Print(output)
	return nil
}

func runInfo(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: schemaforge info <schema.json>")
	}

	data, err := os.ReadFile(args[0])
	if err != nil {
		return fmt.Errorf("reading schema: %w", err)
	}

	s, err := schema.SchemaFromJSON(data)
	if err != nil {
		return fmt.Errorf("parsing schema: %w", err)
	}

	fmt.Println("Schema Information")
	fmt.Println("==================")
	fmt.Printf("Title:       %s\n", coalesce(s.Title, "(none)"))
	fmt.Printf("Description: %s\n", coalesce(s.Description, "(none)"))
	fmt.Printf("Type:        %s\n", coalesce(strings.Join(s.Type.Types, " | "), "(any)"))
	fmt.Printf("Properties:  %d\n", len(s.Properties))
	fmt.Printf("Required:    %d\n", len(s.Required))
	fmt.Printf("Enum values: %d\n", len(s.Enum))

	if len(s.Properties) > 0 {
		fmt.Println("\nProperty Summary")
		fmt.Println("----------------")
		for _, name := range s.SortedPropertyNames() {
			prop := s.Properties[name]
			fmt.Printf("  %-20s %s\n", name, prop.FormatString())
		}
	}

	if len(s.AnyOf) > 0 {
		fmt.Printf("\nAnyOf variants: %d\n", len(s.AnyOf))
	}
	if len(s.OneOf) > 0 {
		fmt.Printf("OneOf variants: %d\n", len(s.OneOf))
	}
	if len(s.AllOf) > 0 {
		fmt.Printf("AllOf schemas:  %d\n", len(s.AllOf))
	}

	return nil
}

func coalesce(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
