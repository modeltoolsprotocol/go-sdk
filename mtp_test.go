package mtp

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
)

// ── Flag extraction tests ────────────────────────────────────────────

func TestFlagBoolean(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("verbose", false, "Enable verbose output")

	schema := Describe(cmd, nil)
	assertArgType(t, schema.Commands[0], "--verbose", "boolean")
}

func TestFlagString(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("name", "", "User name")

	schema := Describe(cmd, nil)
	assertArgType(t, schema.Commands[0], "--name", "string")
}

func TestFlagIntegerOverride(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Int("port", 0, "Port number")

	schema := Describe(cmd, nil)
	assertArgType(t, schema.Commands[0], "--port", "integer")
}

func TestFlagTypeOverrideFromAnnotation(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("count", "0", "Item count")

	opts := &DescribeOptions{
		Commands: map[string]*CommandAnnotation{
			"_root": {
				ArgTypes: map[string]string{"count": "integer"},
			},
		},
	}

	schema := Describe(cmd, opts)
	assertArgType(t, schema.Commands[0], "--count", "integer")
}

func TestFlagEnum(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("format", "json", "Output format")
	EnumValues(cmd, "format", []string{"json", "csv", "yaml"})

	schema := Describe(cmd, nil)
	arg := findArg(t, schema.Commands[0], "--format")
	if arg.Type != "enum" {
		t.Errorf("expected type enum, got %s", arg.Type)
	}
	if len(arg.Values) != 3 {
		t.Errorf("expected 3 values, got %d", len(arg.Values))
	}
}

func TestFlagDefault(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("format", "json", "Output format")

	schema := Describe(cmd, nil)
	arg := findArg(t, schema.Commands[0], "--format")
	if arg.Default != "json" {
		t.Errorf("expected default 'json', got %v", arg.Default)
	}
}

func TestFlagDefaultBoolFalse(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("verbose", false, "Verbose")

	schema := Describe(cmd, nil)
	arg := findArg(t, schema.Commands[0], "--verbose")
	if arg.Default != nil {
		t.Errorf("expected nil default for false bool, got %v", arg.Default)
	}
}

func TestFlagDefaultBoolTrue(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("color", true, "Colorize output")

	schema := Describe(cmd, nil)
	arg := findArg(t, schema.Commands[0], "--color")
	if arg.Default != true {
		t.Errorf("expected default true, got %v", arg.Default)
	}
}

func TestFlagRequired(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("token", "", "Auth token")
	cmd.MarkFlagRequired("token")

	schema := Describe(cmd, nil)
	arg := findArg(t, schema.Commands[0], "--token")
	if !arg.Required {
		t.Error("expected required=true")
	}
}

func TestFlagHiddenExcluded(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("visible", "", "Visible flag")
	cmd.Flags().String("secret", "", "Secret flag")
	cmd.Flags().MarkHidden("secret")

	schema := Describe(cmd, nil)
	for _, arg := range schema.Commands[0].Args {
		if arg.Name == "--secret" {
			t.Error("hidden flag should be excluded")
		}
	}
}

func TestHelpDescribeVersionExcluded(t *testing.T) {
	cmd := &cobra.Command{Use: "test", Version: "1.0.0"}
	WithDescribe(cmd, nil)
	cmd.Flags().String("name", "", "Name")

	schema := Describe(cmd, nil)
	for _, arg := range schema.Commands[0].Args {
		switch arg.Name {
		case "--help", "--mtp-describe", "--version":
			t.Errorf("flag %s should be excluded", arg.Name)
		}
	}
}

// ── Command walking tests ────────────────────────────────────────────

func TestSingleCommand(t *testing.T) {
	cmd := &cobra.Command{Use: "tool", Short: "A tool"}

	schema := Describe(cmd, nil)
	if len(schema.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(schema.Commands))
	}
	if schema.Commands[0].Name != "_root" {
		t.Errorf("expected _root, got %s", schema.Commands[0].Name)
	}
}

func TestNestedCommands(t *testing.T) {
	root := &cobra.Command{Use: "tool"}
	sub := &cobra.Command{Use: "convert", Short: "Convert files"}
	root.AddCommand(sub)

	schema := Describe(root, nil)
	if len(schema.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(schema.Commands))
	}
	if schema.Commands[0].Name != "convert" {
		t.Errorf("expected 'convert', got %s", schema.Commands[0].Name)
	}
}

func TestDeeplyNestedCommands(t *testing.T) {
	root := &cobra.Command{Use: "tool"}
	group := &cobra.Command{Use: "db"}
	leaf := &cobra.Command{Use: "migrate", Short: "Run migrations"}
	group.AddCommand(leaf)
	root.AddCommand(group)

	schema := Describe(root, nil)
	if len(schema.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(schema.Commands))
	}
	if schema.Commands[0].Name != "db migrate" {
		t.Errorf("expected 'db migrate', got %s", schema.Commands[0].Name)
	}
}

func TestHiddenCommandsExcluded(t *testing.T) {
	root := &cobra.Command{Use: "tool"}
	visible := &cobra.Command{Use: "visible", Short: "Visible"}
	hidden := &cobra.Command{Use: "hidden", Short: "Hidden", Hidden: true}
	root.AddCommand(visible, hidden)

	schema := Describe(root, nil)
	if len(schema.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(schema.Commands))
	}
	if schema.Commands[0].Name != "visible" {
		t.Errorf("expected 'visible', got %s", schema.Commands[0].Name)
	}
}

func TestHelpCompletionExcluded(t *testing.T) {
	root := &cobra.Command{Use: "tool"}
	sub := &cobra.Command{Use: "convert", Short: "Convert"}
	root.AddCommand(sub)

	// Cobra auto-adds help and completion; calling Commands() triggers init.
	_ = root.Commands()

	schema := Describe(root, nil)
	for _, cmd := range schema.Commands {
		if cmd.Name == "help" || cmd.Name == "completion" {
			t.Errorf("command %s should be excluded", cmd.Name)
		}
	}
}

func TestMultipleCommands(t *testing.T) {
	root := &cobra.Command{Use: "tool"}
	root.AddCommand(
		&cobra.Command{Use: "convert", Short: "Convert"},
		&cobra.Command{Use: "validate", Short: "Validate"},
		&cobra.Command{Use: "process", Short: "Process"},
	)

	schema := Describe(root, nil)
	if len(schema.Commands) != 3 {
		t.Fatalf("expected 3 commands, got %d", len(schema.Commands))
	}
}

// ── Annotation merging tests ─────────────────────────────────────────

func TestAnnotationsMerged(t *testing.T) {
	root := &cobra.Command{Use: "tool"}
	sub := &cobra.Command{Use: "fetch", Short: "Fetch data"}
	sub.Flags().Bool("verbose", false, "Verbose")
	root.AddCommand(sub)

	opts := &DescribeOptions{
		Commands: map[string]*CommandAnnotation{
			"fetch": {
				Stdin:  &IODescriptor{ContentType: "application/json", Description: "Input data"},
				Stdout: &IODescriptor{ContentType: "application/json", Description: "Output data"},
				Examples: []Example{
					{Description: "Fetch something", Command: "tool fetch --verbose"},
				},
				Auth: &CommandAuth{Required: true, Scopes: []string{"read"}},
			},
		},
	}

	schema := Describe(root, opts)
	cmd := schema.Commands[0]

	if cmd.Stdin == nil || cmd.Stdin.ContentType != "application/json" {
		t.Error("stdin not merged")
	}
	if cmd.Stdout == nil || cmd.Stdout.ContentType != "application/json" {
		t.Error("stdout not merged")
	}
	if len(cmd.Examples) != 1 {
		t.Error("examples not merged")
	}
	if cmd.Auth == nil || !cmd.Auth.Required {
		t.Error("auth not merged")
	}
}

// ── Schema generation tests ──────────────────────────────────────────

func TestSchemaMetadata(t *testing.T) {
	root := &cobra.Command{
		Use:     "mytool",
		Short:   "My awesome tool",
		Version: "2.1.0",
	}

	schema := Describe(root, nil)
	if schema.SpecVersion != MTPSpecVersion {
		t.Errorf("expected specVersion %q, got %q", MTPSpecVersion, schema.SpecVersion)
	}
	if schema.Name != "mytool" {
		t.Errorf("expected name 'mytool', got %s", schema.Name)
	}
	if schema.Version != "2.1.0" {
		t.Errorf("expected version '2.1.0', got %s", schema.Version)
	}
	if schema.Description != "My awesome tool" {
		t.Errorf("expected description 'My awesome tool', got %s", schema.Description)
	}
}

func TestSchemaAuth(t *testing.T) {
	root := &cobra.Command{Use: "tool", Short: "A tool"}

	opts := &DescribeOptions{
		Auth: &AuthConfig{
			Required: true,
			EnvVar:   "TOOL_TOKEN",
			Providers: []AuthProvider{
				{
					ID:           "github",
					Type:         "oauth2",
					DisplayName:  "GitHub",
					TokenURL:     "https://github.com/login/oauth/access_token",
					Scopes:       []string{"repo", "read:org"},
					ClientID:     "abc123",
					Instructions: "Create a GitHub OAuth app",
				},
			},
		},
	}

	schema := Describe(root, opts)
	if schema.Auth == nil {
		t.Fatal("expected auth config")
	}
	if !schema.Auth.Required {
		t.Error("expected auth required=true")
	}
	if schema.Auth.EnvVar != "TOOL_TOKEN" {
		t.Errorf("expected envVar TOOL_TOKEN, got %s", schema.Auth.EnvVar)
	}
	if len(schema.Auth.Providers) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(schema.Auth.Providers))
	}
	if schema.Auth.Providers[0].Type != "oauth2" {
		t.Errorf("expected provider type oauth2, got %s", schema.Auth.Providers[0].Type)
	}
}

func TestSchemaJSON(t *testing.T) {
	root := &cobra.Command{
		Use:     "tool",
		Short:   "A tool",
		Version: "1.0.0",
	}
	sub := &cobra.Command{Use: "run", Short: "Run something"}
	sub.Flags().String("target", "", "Target to run")
	root.AddCommand(sub)

	schema := Describe(root, nil)

	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Round-trip through JSON to verify structure.
	var decoded ToolSchema
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.SpecVersion != MTPSpecVersion {
		t.Errorf("expected specVersion %q, got %q", MTPSpecVersion, decoded.SpecVersion)
	}
	if decoded.Name != "tool" {
		t.Errorf("expected name 'tool', got %s", decoded.Name)
	}
	if len(decoded.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(decoded.Commands))
	}
	if decoded.Commands[0].Name != "run" {
		t.Errorf("expected command 'run', got %s", decoded.Commands[0].Name)
	}
}

// ── WithDescribe tests ───────────────────────────────────────────────

func TestWithDescribeAddsFlag(t *testing.T) {
	root := &cobra.Command{Use: "tool"}
	WithDescribe(root, nil)

	f := root.PersistentFlags().Lookup("mtp-describe")
	if f == nil {
		t.Fatal("--mtp-describe flag not added")
	}
	if f.Usage != "Output machine-readable JSON schema for this tool" {
		t.Errorf("unexpected usage: %s", f.Usage)
	}
}

func TestWithDescribeChainsPreRun(t *testing.T) {
	var chainCalled bool
	root := &cobra.Command{
		Use: "tool",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			chainCalled = true
		},
		Run: func(cmd *cobra.Command, args []string) {},
	}
	WithDescribe(root, nil)

	root.SetArgs([]string{})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if !chainCalled {
		t.Error("existing PersistentPreRun was not chained")
	}
}

func TestWithDescribeChainsPreRunE(t *testing.T) {
	var chainCalled bool
	root := &cobra.Command{
		Use: "tool",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			chainCalled = true
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {},
	}
	WithDescribe(root, nil)

	root.SetArgs([]string{})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if !chainCalled {
		t.Error("existing PersistentPreRunE was not chained")
	}
}

// ── Positional arg tests ─────────────────────────────────────────────

func TestPositionalArgsFromUse(t *testing.T) {
	cmd := &cobra.Command{Use: "convert <input> [output]", Short: "Convert"}

	schema := Describe(cmd, nil)
	args := schema.Commands[0].Args
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(args))
	}
	if args[0].Name != "input" || !args[0].Required {
		t.Errorf("expected required 'input', got %+v", args[0])
	}
	if args[1].Name != "output" || args[1].Required {
		t.Errorf("expected optional 'output', got %+v", args[1])
	}
}

func TestPositionalArgsAnnotationOverride(t *testing.T) {
	root := &cobra.Command{Use: "tool"}
	sub := &cobra.Command{Use: "convert <input>", Short: "Convert"}
	root.AddCommand(sub)

	opts := &DescribeOptions{
		Commands: map[string]*CommandAnnotation{
			"convert": {
				Args: []ArgDescriptor{
					{Name: "input_file", Type: "string", Required: true, Description: "Input file path"},
					{Name: "output_file", Type: "string", Description: "Output file path"},
				},
			},
		},
	}

	schema := Describe(root, opts)
	args := schema.Commands[0].Args
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(args))
	}
	if args[0].Name != "input_file" {
		t.Errorf("expected annotation arg 'input_file', got %s", args[0].Name)
	}
	if args[0].Description != "Input file path" {
		t.Errorf("expected description from annotation, got %s", args[0].Description)
	}
}

func TestPositionalArgsWithFlags(t *testing.T) {
	cmd := &cobra.Command{Use: "convert <input>", Short: "Convert"}
	cmd.Flags().String("format", "json", "Output format")

	schema := Describe(cmd, nil)
	args := schema.Commands[0].Args
	if len(args) != 2 {
		t.Fatalf("expected 2 args (1 positional + 1 flag), got %d", len(args))
	}
	if args[0].Name != "input" {
		t.Errorf("first arg should be positional 'input', got %s", args[0].Name)
	}
	if args[1].Name != "--format" {
		t.Errorf("second arg should be flag '--format', got %s", args[1].Name)
	}
}

// ── EnumValues helper test ───────────────────────────────────────────

func TestEnumValuesNonexistentFlag(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	// Should not panic on nonexistent flag.
	EnumValues(cmd, "nonexistent", []string{"a", "b"})
}

// ── Helpers ──────────────────────────────────────────────────────────

func findArg(t *testing.T, cmd CommandDescriptor, name string) ArgDescriptor {
	t.Helper()
	for _, arg := range cmd.Args {
		if arg.Name == name {
			return arg
		}
	}
	t.Fatalf("arg %s not found in command %s", name, cmd.Name)
	return ArgDescriptor{}
}

func assertArgType(t *testing.T, cmd CommandDescriptor, name, expectedType string) {
	t.Helper()
	arg := findArg(t, cmd, name)
	if arg.Type != expectedType {
		t.Errorf("expected %s type %s, got %s", name, expectedType, arg.Type)
	}
}
