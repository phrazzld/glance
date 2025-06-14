// Package llm provides abstractions and implementations for interacting with
// Large Language Model APIs in the glance application.
package llm

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// PromptData holds the content used to generate prompts for LLM requests.
// It contains information about the directory structure, content, and summaries
// that will be filled into the prompt template.
type PromptData struct {
	// Directory is the path to the directory being processed
	Directory string

	// SubGlances contains the compiled contents of subdirectory glance.md files
	SubGlances string

	// FileContents contains the formatted contents of files in the directory
	FileContents string
}

// DefaultTemplate returns the default prompt template used for generating directory summaries.
// This template is used when no custom template is provided.
func DefaultTemplate() string {
	return `you are an expert code reviewer and technical writer.
generate a descriptive technical overview of this directory:
- highlight purpose, architecture, and key file roles
- mention important dependencies or gotchas
- do NOT provide recommendations or next steps
- respond with ONLY the descriptive technical overview: no preamble or concluding remarks

directory: {{.Directory}}

subdirectory summaries:
{{.SubGlances}}

local file contents:
{{.FileContents}}
`
}

// GeneratePrompt generates a prompt by filling the template with the provided data.
//
// Parameters:
//   - data: The PromptData to use for template variables
//   - templateStr: The template string to use
//
// Returns:
//   - The generated prompt as a string
//   - An error if template parsing or execution fails
func GeneratePrompt(data *PromptData, templateStr string) (string, error) {
	// Parse the template
	tmpl, err := template.New("prompt").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse prompt template: %w", err)
	}

	// Execute the template with the provided data
	var rendered bytes.Buffer
	if err = tmpl.Execute(&rendered, data); err != nil {
		return "", fmt.Errorf("failed to execute prompt template: %w", err)
	}

	return rendered.String(), nil
}

// FormatFileContents formats a map of filenames to content for inclusion in a prompt.
// The format used is "=== file: {filename} ===\n{content}\n\n".
//
// Parameters:
//   - fileMap: A map of filenames to their content
//
// Returns:
//   - A formatted string containing all file contents
func FormatFileContents(fileMap map[string]string) string {
	var builder strings.Builder

	for filename, content := range fileMap {
		builder.WriteString(fmt.Sprintf("=== file: %s ===\n%s\n\n", filename, content))
	}

	return builder.String()
}

// BuildPromptData creates a PromptData structure with the provided information.
// It formats the file contents using FormatFileContents.
//
// Parameters:
//   - dir: The directory path
//   - subGlances: Compiled content from subdirectory glance.md files
//   - fileMap: A map of filenames to their content
//
// Returns:
//   - A populated PromptData structure
func BuildPromptData(dir string, subGlances string, fileMap map[string]string) *PromptData {
	return &PromptData{
		Directory:    dir,
		SubGlances:   subGlances,
		FileContents: FormatFileContents(fileMap),
	}
}
