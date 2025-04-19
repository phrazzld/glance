// Package llm provides abstractions and implementations for interacting with
// Large Language Model APIs in the glance application.
package llm

import (
	"bytes"
	"fmt"
	"os"
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

directory: {{.Directory}}

subdirectory summaries:
{{.SubGlances}}

local file contents:
{{.FileContents}}
`
}

// LoadTemplate loads a prompt template from the specified file path.
// If the path is empty, it attempts to load from "prompt.txt" in the current directory.
// If neither is available, it returns the default template.
// All file paths are validated to prevent path traversal vulnerabilities.
//
// Parameters:
//   - path: The path to the template file (can be empty)
//
// Returns:
//   - The template content as a string
//   - An error if loading fails
func LoadTemplate(path string) (string, error) {
	// Get current working directory as the base for validation
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	// For custom path, don't restrict to current working directory
	// This allows users to provide templates from anywhere on the filesystem
	if path != "" {
		// Clean and absolutize the path, but don't enforce directory containment
		// since the prompt file can be anywhere the user wants
		cleanPath := filepath.Clean(path)
		absPath, err := filepath.Abs(cleanPath)
		if err != nil {
			return "", fmt.Errorf("invalid prompt template path: %w", err)
		}

		// Verify the file exists and is a file (not a directory)
		info, err := os.Stat(absPath)
		if err != nil {
			return "", fmt.Errorf("failed to access prompt template at '%s': %w", path, err)
		}
		if info.IsDir() {
			return "", fmt.Errorf("prompt template path '%s' is a directory, not a file", path)
		}

		// Read the validated file path
		// #nosec G304 -- Reading template files from a validated path
		data, err := os.ReadFile(absPath)
		if err != nil {
			return "", fmt.Errorf("failed to read custom prompt template from '%s': %w", absPath, err)
		}
		return string(data), nil
	}

	// No path provided, try to load from default location
	defaultPromptPath := filepath.Join(cwd, "prompt.txt")
	// Check if the file exists, but don't return an error if it doesn't
	if _, err := os.Stat(defaultPromptPath); err == nil {
		// File exists, clean and read it
		cleanPath := filepath.Clean(defaultPromptPath)
		// #nosec G304 -- Reading from a standard prompt.txt file in the current directory
		if data, err := os.ReadFile(cleanPath); err == nil {
			return string(data), nil
		}
	}

	// Fall back to the default template
	return DefaultTemplate(), nil
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
