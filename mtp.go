package mtp

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// MTPSpecVersion is the version of the MTP specification implemented by this SDK.
const MTPSpecVersion = "2026-02-07"

// Describe extracts a ToolSchema from a Cobra command tree.
// This is a pure function with no side effects, useful for testing
// or programmatic access to the schema.
func Describe(root *cobra.Command, opts *DescribeOptions) *ToolSchema {
	desc := strings.TrimSpace(root.Short)
	if desc == "" {
		desc = strings.TrimSpace(root.Long)
	}

	schema := &ToolSchema{
		SpecVersion: MTPSpecVersion,
		Name:        root.Name(),
		Version:     root.Version,
		Description: desc,
		Commands:    walkCommands(root, "", opts),
	}

	if opts != nil && opts.Auth != nil {
		schema.Auth = opts.Auth
	}

	return schema
}

// WithDescribe adds a --describe flag to the root command.
// When --describe is passed, it prints the JSON schema to stdout and exits 0.
func WithDescribe(root *cobra.Command, opts *DescribeOptions) {
	var describeFlag bool

	root.PersistentFlags().BoolVar(
		&describeFlag,
		"mtp-describe",
		false,
		"Output machine-readable JSON schema for this tool",
	)

	printAndExit := func() {
		schema := Describe(root, opts)
		enc := json.NewEncoder(os.Stdout)
		if err := enc.Encode(schema); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding schema: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Chain with any existing PersistentPreRunE or PersistentPreRun.
	existingE := root.PersistentPreRunE
	existingPlain := root.PersistentPreRun

	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if describeFlag {
			printAndExit()
		}

		if existingE != nil {
			return existingE(cmd, args)
		}
		if existingPlain != nil {
			existingPlain(cmd, args)
		}
		return nil
	}
	// Clear PersistentPreRun so Cobra doesn't complain about both being set.
	root.PersistentPreRun = nil

	// If root has no Run/RunE (common for tools with subcommands), Cobra
	// shows help instead of executing hooks. Set RunE so --describe works
	// when invoked on the root command directly (e.g. "tool --describe").
	if root.RunE == nil && root.Run == nil {
		root.RunE = func(cmd *cobra.Command, args []string) error {
			if describeFlag {
				printAndExit()
			}
			return cmd.Help()
		}
	}
}

// EnumValues annotates a flag with allowed enum values.
// Call after adding the flag to the command:
//
//	cmd.Flags().String("format", "json", "Output format")
//	mtp.EnumValues(cmd, "format", []string{"json", "csv", "yaml"})
func EnumValues(cmd *cobra.Command, flagName string, values []string) {
	f := cmd.Flags().Lookup(flagName)
	if f == nil {
		return
	}
	if f.Annotations == nil {
		f.Annotations = map[string][]string{}
	}
	f.Annotations["values"] = values
}
