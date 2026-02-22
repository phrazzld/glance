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
generate a concise, factual technical summary for this directory.
Use only what is present in the provided source snippets (directory summaries + file contents + explicit structure).

Hard constraints:
- do NOT describe CLI flags, command-line options, defaults, runtime modes, side effects, or performance characteristics unless they are explicitly defined in the provided source snippets.
- do NOT speculate about behavior, configuration, environment variables, dependencies, or architecture details not evidenced by the provided source snippets.
- do NOT provide recommendations, next steps, or hypothetical refactors.
- if a claim cannot be verified from the provided source snippets, omit it rather than infer.
- do NOT mention files or directories that are not listed in the provided input.

Output format:
## Purpose
One short paragraph (max 5 sentences) describing the directory-level intent.

## Key Roles
- list major files and their responsibilities
- if no obvious key roles are found, state "No dominant file roles detected."

## Dependencies and Caveats
- list important dependencies and notable caveats grounded in the provided source snippets
- max 8 bullets

Keep this output under 400 words.

respond with ONLY the sections above, in the exact order shown.

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
