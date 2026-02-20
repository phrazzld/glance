// Package llm provides abstractions and implementations for interacting with
// Large Language Model APIs in the glance application.
package llm

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
)

// PromptData holds the content used to generate prompts for LLM requests.
// It contains information about the directory structure, content, and summaries
// that will be filled into the prompt template.
type PromptData struct {
	// Directory is the path to the directory being processed
	Directory string

	// ProjectRoot is the path to the root directory being scanned
	ProjectRoot string

	// RelativeDirectory is the directory path relative to ProjectRoot
	RelativeDirectory string

	// ProjectMap contains a bounded directory map for global context
	ProjectMap string

	// ProjectOverview contains existing top-level glance context, if available
	ProjectOverview string

	// SubGlances contains the compiled contents of subdirectory glance.md files
	SubGlances string

	// FileContents contains the formatted contents of files in the directory
	FileContents string
}

// DefaultTemplate returns the default prompt template used for generating directory summaries.
// This template is used when no custom template is provided.
func DefaultTemplate() string {
	return `you are an expert code reviewer and technical writer.
generate a descriptive technical overview of this directory in the context of the full project:
- explain this directory's role in the overall architecture
- highlight purpose, architecture, and key file roles
- mention important dependencies or gotchas
- do NOT provide recommendations or next steps
- respond with ONLY the descriptive technical overview: no preamble or concluding remarks

project root: {{.ProjectRoot}}
directory relative path: {{.RelativeDirectory}}

project directory map:
{{.ProjectMap}}

existing top-level project overview (if available):
{{.ProjectOverview}}

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
//   - projectRoot: The root path for the entire scan
//   - dir: The directory path
//   - projectMap: A bounded directory map of the full project
//   - projectOverview: Existing top-level project context, if available
//   - subGlances: Compiled content from subdirectory glance.md files
//   - fileMap: A map of filenames to their content
//
// Returns:
//   - A populated PromptData structure
func BuildPromptData(
	projectRoot string,
	dir string,
	projectMap string,
	projectOverview string,
	subGlances string,
	fileMap map[string]string,
) *PromptData {
	relativeDir := "."
	if projectRoot != "" {
		if rel, err := filepath.Rel(projectRoot, dir); err == nil {
			relativeDir = filepath.ToSlash(rel)
			if relativeDir == "" {
				relativeDir = "."
			}
		}
	}

	return &PromptData{
		Directory:         dir,
		ProjectRoot:       projectRoot,
		RelativeDirectory: relativeDir,
		ProjectMap:        projectMap,
		ProjectOverview:   projectOverview,
		SubGlances:        subGlances,
		FileContents:      FormatFileContents(fileMap),
	}
}
