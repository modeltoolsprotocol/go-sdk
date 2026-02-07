package mtp

// ToolSchema is the top-level --describe output for a CLI tool.
type ToolSchema struct {
	SpecVersion string              `json:"specVersion"`
	Name        string              `json:"name"`
	Version     string              `json:"version"`
	Description string              `json:"description"`
	Commands    []CommandDescriptor `json:"commands"`
	Auth        *AuthConfig         `json:"auth,omitempty"`
}

// CommandDescriptor describes a single command within a tool.
type CommandDescriptor struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Args        []ArgDescriptor `json:"args,omitempty"`
	Stdin       *IODescriptor   `json:"stdin,omitempty"`
	Stdout      *IODescriptor   `json:"stdout,omitempty"`
	Examples    []Example       `json:"examples,omitempty"`
	Auth        *CommandAuth    `json:"auth,omitempty"`
}

// ArgDescriptor describes a single argument (flag or positional) for a command.
type ArgDescriptor struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Required    bool     `json:"required,omitempty"`
	Default     any      `json:"default,omitempty"`
	Values      []string `json:"values,omitempty"`
}

// IODescriptor describes stdin or stdout for a command.
type IODescriptor struct {
	ContentType string         `json:"contentType,omitempty"`
	Description string         `json:"description,omitempty"`
	Schema      map[string]any `json:"schema,omitempty"`
}

// Example is a usage example for a command.
type Example struct {
	Description string `json:"description,omitempty"`
	Command     string `json:"command"`
	Output      string `json:"output,omitempty"`
}

// AuthConfig describes the authentication requirements for a tool.
type AuthConfig struct {
	Required  bool           `json:"required,omitempty"`
	EnvVar    string         `json:"envVar"`
	Providers []AuthProvider `json:"providers"`
}

// AuthProvider describes a single authentication provider.
type AuthProvider struct {
	ID               string   `json:"id"`
	Type             string   `json:"type"`
	DisplayName      string   `json:"displayName,omitempty"`
	AuthorizationURL string   `json:"authorizationUrl,omitempty"`
	TokenURL         string   `json:"tokenUrl,omitempty"`
	Scopes           []string `json:"scopes,omitempty"`
	ClientID         string   `json:"clientId,omitempty"`
	RegistrationURL  string   `json:"registrationUrl,omitempty"`
	Instructions     string   `json:"instructions,omitempty"`
}

// CommandAuth describes per-command authentication requirements.
type CommandAuth struct {
	Required bool     `json:"required,omitempty"`
	Scopes   []string `json:"scopes,omitempty"`
}

// DescribeOptions provides metadata that Cobra doesn't natively expose.
type DescribeOptions struct {
	Commands map[string]*CommandAnnotation
	Auth     *AuthConfig
}

// CommandAnnotation supplements a command with MTP metadata.
type CommandAnnotation struct {
	Args     []ArgDescriptor   // Positional args (Cobra has no typed positional args)
	ArgTypes map[string]string // Flag name -> MTP type override (e.g. "port" -> "integer")
	Stdin    *IODescriptor
	Stdout   *IODescriptor
	Examples []Example
	Auth     *CommandAuth
}
