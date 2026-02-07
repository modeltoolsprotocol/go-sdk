package main

import (
	"fmt"
	"os"

	mtp "github.com/modeltoolsprotocol/go-sdk"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:     "filetool",
		Short:   "Convert and validate files between formats",
		Version: "1.2.0",
	}

	// ── convert command ──────────────────────────────────────────

	var convertFormat string
	var convertPretty bool

	convertCmd := &cobra.Command{
		Use:   "convert <input_file>",
		Short: "Convert a file from one format to another",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Converting %s to %s (pretty=%v)\n", args[0], convertFormat, convertPretty)
		},
	}

	convertCmd.Flags().StringVarP(&convertFormat, "format", "f", "json", "Output format")
	mtp.EnumValues(convertCmd, "format", []string{"json", "csv", "yaml"})
	convertCmd.Flags().BoolVar(&convertPretty, "pretty", false, "Pretty-print output")

	// ── validate command ─────────────────────────────────────────

	var validateStrict bool

	validateCmd := &cobra.Command{
		Use:   "validate <input_file>",
		Short: "Check if a file is well-formed and valid",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Validating %s (strict=%v)\n", args[0], validateStrict)
		},
	}

	validateCmd.Flags().BoolVar(&validateStrict, "strict", false, "Enable strict validation mode")

	// ── process command ──────────────────────────────────────────

	var processVerbose bool

	processCmd := &cobra.Command{
		Use:   "process",
		Short: "Process structured JSON input from stdin",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Processing (verbose=%v)\n", processVerbose)
		},
	}

	processCmd.Flags().BoolVar(&processVerbose, "verbose", false, "Enable verbose output")

	// ── assemble ─────────────────────────────────────────────────

	root.AddCommand(convertCmd, validateCmd, processCmd)

	opts := &mtp.DescribeOptions{
		Commands: map[string]*mtp.CommandAnnotation{
			"convert": {
				Stdin:  &mtp.IODescriptor{ContentType: "text/plain", Description: "Raw input data (alternative to file path)"},
				Stdout: &mtp.IODescriptor{ContentType: "application/json", Description: "Converted output"},
				Examples: []mtp.Example{
					{
						Description: "Convert a CSV file to JSON",
						Command:     "filetool convert data.csv --format json --pretty",
						Output:      "[{\n  \"name\": \"Alice\",\n  \"age\": 30\n}]",
					},
					{
						Description: "Pipe from stdin",
						Command:     "cat data.csv | filetool convert - --format yaml",
					},
				},
			},
			"validate": {
				Examples: []mtp.Example{
					{
						Description: "Validate a JSON file",
						Command:     "filetool validate config.json",
						Output:      "{\"valid\": true, \"errors\": []}",
					},
				},
			},
			"process": {
				Stdin: &mtp.IODescriptor{
					ContentType: "application/json",
					Description: "JSON object to process",
					Schema: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"name":  map[string]any{"type": "string", "description": "Item name"},
							"count": map[string]any{"type": "integer", "description": "Number of items"},
						},
						"required": []string{"name"},
					},
				},
				Stdout: &mtp.IODescriptor{ContentType: "application/json", Description: "Processing result"},
				Examples: []mtp.Example{
					{
						Description: "Process a JSON object from stdin",
						Command:     "echo '{\"name\":\"foo\",\"count\":3}' | filetool process",
						Output:      "{\"status\": \"ok\", \"processed\": \"foo\"}",
					},
				},
			},
		},
	}

	mtp.WithDescribe(root, opts)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
