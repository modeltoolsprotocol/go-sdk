# go-sdk (Go)

Go SDK for the [Model Tools Protocol](https://modeltoolsprotocol.org). Makes any [Cobra](https://github.com/spf13/cobra) CLI tool LLM-discoverable via `--mtp-describe`.

## Install

```bash
go get github.com/modeltoolsprotocol/go-sdk
```

## Quick Start

```go
package main

import (
    "fmt"
    "os"

    mtp "github.com/modeltoolsprotocol/go-sdk"
    "github.com/spf13/cobra"
)

func main() {
    root := &cobra.Command{
        Use:     "mytool",
        Short:   "My awesome tool",
        Version: "1.0.0",
    }

    convertCmd := &cobra.Command{
        Use:   "convert <input_file>",
        Short: "Convert a file between formats",
        Args:  cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Printf("Converting %s\n", args[0])
        },
    }

    convertCmd.Flags().StringP("format", "f", "json", "Output format")
    mtp.EnumValues(convertCmd, "format", []string{"json", "csv", "yaml"})

    root.AddCommand(convertCmd)

    opts := &mtp.DescribeOptions{
        Commands: map[string]*mtp.CommandAnnotation{
            "convert": {
                Stdin:  &mtp.IODescriptor{ContentType: "text/csv", Description: "CSV input"},
                Stdout: &mtp.IODescriptor{ContentType: "application/json", Description: "Converted output"},
                Examples: []mtp.Example{
                    {Description: "Convert CSV to JSON", Command: "mytool convert data.csv --format json"},
                },
            },
        },
    }

    mtp.WithDescribe(root, opts)

    if err := root.Execute(); err != nil {
        os.Exit(1)
    }
}
```

```bash
$ mytool --mtp-describe   # JSON schema output
$ mytool convert data.csv --format json   # normal operation
```

## API

### `mtp.WithDescribe(root, opts)`

Adds a `--mtp-describe` flag to a Cobra root command. When passed, prints the MTP JSON schema to stdout and exits.

### `mtp.Describe(root, opts)`

Returns a `*ToolSchema` without side effects. Useful for testing or programmatic access.

### `mtp.EnumValues(cmd, flagName, values)`

Annotates a flag with allowed enum values, since Cobra has no native enum support.

### `mtp.DescribeOptions`

Provides metadata that Cobra can't express natively:

- `Commands` - map of command name to `CommandAnnotation` (stdin/stdout descriptors, examples, positional arg types, auth)
- `Auth` - tool-level authentication configuration

## How It Works

Cobra already stores flag types, defaults, help strings, and usage info. The SDK reads all of this and serializes it into the MTP `--mtp-describe` JSON format. Positional args are inferred from the `Use` string convention (`<required>` and `[optional]`), with optional overrides via `CommandAnnotation.Args`.

## What Gets Auto-Extracted from Cobra

- Tool name, version, description
- Command tree (with space-separated names for nested commands)
- Flag names, types, defaults, descriptions, required status
- Positional args from `Use` string patterns

## What Needs Annotations

- stdin/stdout descriptors (content types, JSON schemas)
- Usage examples
- Authentication configuration
- Typed positional args (Cobra only has `[]string`)
- Flag type overrides (e.g. marking a string flag as `"integer"`)

## Structured IO

Arg types are flat (`string`, `boolean`, `enum`, etc.) because CLI flags are always scalar. For structured data flowing through stdin/stdout, IO descriptors support full JSON Schema (draft 2020-12): nested objects, arrays, unions, pattern validation, conditional fields.

```go
opts := &mtp.DescribeOptions{
    Commands: map[string]*mtp.CommandAnnotation{
        "process": {
            Stdin: &mtp.IODescriptor{
                ContentType: "application/json",
                Description: "Configuration to process",
                Schema: map[string]any{
                    "type": "object",
                    "properties": map[string]any{
                        "name": map[string]any{"type": "string"},
                        "settings": map[string]any{
                            "type": "object",
                            "properties": map[string]any{
                                "retries":  map[string]any{"type": "integer"},
                                "endpoints": map[string]any{
                                    "type":  "array",
                                    "items": map[string]any{"type": "string", "format": "uri"},
                                },
                            },
                        },
                    },
                    "required": []string{"name"},
                },
            },
            Stdout: &mtp.IODescriptor{
                ContentType: "application/json",
                Schema: map[string]any{
                    "type": "object",
                    "properties": map[string]any{
                        "status":  map[string]any{"type": "string", "enum": []string{"ok", "error"}},
                        "results": map[string]any{"type": "array", "items": map[string]any{"type": "object"}},
                    },
                },
            },
        },
    },
}
```

## License

Apache-2.0
