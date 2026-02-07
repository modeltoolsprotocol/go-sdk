package mtp

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// pflagTypeToMTP maps pflag type strings to MTP type strings.
func pflagTypeToMTP(f *pflag.Flag) string {
	switch f.Value.Type() {
	case "bool":
		return "boolean"
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64":
		return "integer"
	case "float32", "float64":
		return "number"
	case "stringSlice", "intSlice", "stringArray", "uintSlice":
		return "array"
	default:
		return "string"
	}
}

// flagDefault returns a typed default value for a flag, or nil if the
// default is the zero value for its type.
func flagDefault(f *pflag.Flag) any {
	switch f.Value.Type() {
	case "bool":
		if f.DefValue == "true" {
			return true
		}
		return nil
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64":
		if f.DefValue == "" || f.DefValue == "0" {
			return nil
		}
		return f.DefValue
	case "float32", "float64":
		if f.DefValue == "" || f.DefValue == "0" {
			return nil
		}
		return f.DefValue
	default:
		if f.DefValue == "" || f.DefValue == "[]" {
			return nil
		}
		return f.DefValue
	}
}

// skippedFlags are flags that should never appear in --describe output.
var skippedFlags = map[string]bool{
	"help":         true,
	"mtp-describe": true,
	"version":      true,
}

// extractFlags builds ArgDescriptors from a command's flags.
func extractFlags(cmd *cobra.Command, ann *CommandAnnotation) []ArgDescriptor {
	var args []ArgDescriptor

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if skippedFlags[f.Name] || f.Hidden {
			return
		}

		typ := pflagTypeToMTP(f)
		if ann != nil {
			if override, ok := ann.ArgTypes[f.Name]; ok {
				typ = override
			}
		}

		arg := ArgDescriptor{
			Name:        "--" + f.Name,
			Type:        typ,
			Description: f.Usage,
		}

		// Cobra stores required-flag info as an annotation.
		if ann, ok := f.Annotations["cobra_annotation_bash_completion_one_required_flag"]; ok && len(ann) > 0 {
			arg.Required = true
		}

		if def := flagDefault(f); def != nil {
			arg.Default = def
		}

		// Enum values stored via EnumValues helper.
		if vals, ok := f.Annotations["values"]; ok && len(vals) > 0 {
			arg.Type = "enum"
			arg.Values = vals
		}

		args = append(args, arg)
	})

	return args
}

// parseUseArgs extracts positional arg descriptors from a Cobra Use string.
// Convention: "command <required> [optional]"
func parseUseArgs(use string) []ArgDescriptor {
	parts := strings.Fields(use)
	if len(parts) <= 1 {
		return nil
	}

	var args []ArgDescriptor
	for _, part := range parts[1:] {
		if strings.HasPrefix(part, "<") && strings.HasSuffix(part, ">") {
			name := strings.Trim(part, "<>")
			args = append(args, ArgDescriptor{
				Name:     name,
				Type:     "string",
				Required: true,
			})
		} else if strings.HasPrefix(part, "[") && strings.HasSuffix(part, "]") {
			name := strings.Trim(part, "[]")
			args = append(args, ArgDescriptor{
				Name: name,
				Type: "string",
			})
		}
	}
	return args
}

// extractCommand builds a CommandDescriptor from a single Cobra command.
func extractCommand(cmd *cobra.Command, name string, ann *CommandAnnotation) CommandDescriptor {
	desc := strings.TrimSpace(cmd.Short)
	if desc == "" {
		desc = strings.TrimSpace(cmd.Long)
	}

	cd := CommandDescriptor{
		Name:        name,
		Description: desc,
	}

	// Positional args: annotation overrides Use string parsing.
	if ann != nil && len(ann.Args) > 0 {
		cd.Args = append(cd.Args, ann.Args...)
	} else {
		cd.Args = append(cd.Args, parseUseArgs(cmd.Use)...)
	}

	// Flags
	cd.Args = append(cd.Args, extractFlags(cmd, ann)...)

	// Annotation-only fields
	if ann != nil {
		cd.Stdin = ann.Stdin
		cd.Stdout = ann.Stdout
		cd.Examples = ann.Examples
		cd.Auth = ann.Auth
	}

	return cd
}

// skippedCommands are auto-generated commands that should be excluded.
var skippedCommands = map[string]bool{
	"help":       true,
	"completion": true,
}

// walkCommands recursively extracts CommandDescriptors from a Cobra command tree.
func walkCommands(cmd *cobra.Command, prefix string, opts *DescribeOptions) []CommandDescriptor {
	var commands []CommandDescriptor

	visible := visibleSubcommands(cmd)
	if len(visible) == 0 {
		// Leaf command (or single-command tool)
		name := prefix
		if name == "" {
			name = "_root"
		}
		var ann *CommandAnnotation
		if opts != nil && opts.Commands != nil {
			ann = opts.Commands[name]
		}
		commands = append(commands, extractCommand(cmd, name, ann))
		return commands
	}

	for _, sub := range visible {
		subName := sub.Name()
		if prefix != "" {
			subName = prefix + " " + sub.Name()
		}
		commands = append(commands, walkCommands(sub, subName, opts)...)
	}

	return commands
}

// visibleSubcommands returns non-hidden, non-skipped subcommands.
func visibleSubcommands(cmd *cobra.Command) []*cobra.Command {
	var visible []*cobra.Command
	for _, sub := range cmd.Commands() {
		if sub.Hidden || skippedCommands[sub.Name()] {
			continue
		}
		visible = append(visible, sub)
	}
	return visible
}
