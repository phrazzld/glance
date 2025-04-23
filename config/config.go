// Package config provides configuration management for the glance application.
package config

// Config holds the application configuration parameters.
// This structure centralizes all application settings, making them easier to
// manage, test, and extend in the future.
type Config struct {
	// APIKey is the Gemini API key used for generating content
	APIKey string

	// TargetDir is the directory to scan and generate glance.md files for
	TargetDir string

	// Force indicates whether to regenerate existing glance.md files
	Force bool

	// PromptTemplate contains the template text used for generating prompts
	PromptTemplate string

	// MaxRetries defines how many times to retry API calls before giving up
	MaxRetries int

	// MaxFileBytes is the maximum file size in bytes to process (larger files are truncated)
	MaxFileBytes int64
}

// Default constants used in configuration
const (
	// DefaultMaxRetries is the default number of API retry attempts
	DefaultMaxRetries = 3

	// DefaultMaxFileBytes is the default maximum file size (5MB)
	DefaultMaxFileBytes = 5 * 1024 * 1024
)

// Default prompt template used when no custom template is provided
var defaultPromptTemplate = `you are an expert code reviewer and technical writer.
generate a descriptive technical overview of this directory:
- highlight purpose, architecture, and key file roles
- mention important dependencies or gotchas
- do NOT provide recommendations or next steps

directory: {{.Directory}}

subdirectory summaries:
{{.SubGlances}}

local file contents:
{{.FileContents}}
`

// NewDefaultConfig creates a new Config with default values.
// This provides a starting point for configuration that can be
// customized using the With* methods.
func NewDefaultConfig() *Config {
	return &Config{
		APIKey:         "",
		TargetDir:      "",
		Force:          false,
		PromptTemplate: defaultPromptTemplate,
		MaxRetries:     DefaultMaxRetries,
		MaxFileBytes:   DefaultMaxFileBytes,
	}
}

// WithAPIKey returns a new Config with the specified API key.
func (c *Config) WithAPIKey(apiKey string) *Config {
	// Create a copy of the config to ensure immutability
	newConfig := *c
	newConfig.APIKey = apiKey // pragma: allowlist secret
	return &newConfig
}

// WithTargetDir returns a new Config with the specified target directory.
func (c *Config) WithTargetDir(targetDir string) *Config {
	newConfig := *c
	newConfig.TargetDir = targetDir
	return &newConfig
}

// WithForce returns a new Config with the specified force flag value.
func (c *Config) WithForce(force bool) *Config {
	newConfig := *c
	newConfig.Force = force
	return &newConfig
}

// WithPromptTemplate returns a new Config with the specified prompt template.
func (c *Config) WithPromptTemplate(template string) *Config {
	newConfig := *c
	newConfig.PromptTemplate = template
	return &newConfig
}

// WithMaxRetries returns a new Config with the specified max retries value.
func (c *Config) WithMaxRetries(maxRetries int) *Config {
	newConfig := *c
	newConfig.MaxRetries = maxRetries
	return &newConfig
}

// WithMaxFileBytes returns a new Config with the specified max file bytes value.
func (c *Config) WithMaxFileBytes(maxFileBytes int64) *Config {
	newConfig := *c
	newConfig.MaxFileBytes = maxFileBytes
	return &newConfig
}
