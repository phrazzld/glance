package config

import (
	"fmt"
	"os"
	"path/filepath"

	"glance/filesystem"
)

// LoadPromptTemplate loads a prompt template from the specified file path.
// If the path is empty, it attempts to load from "prompt.txt" in the current directory.
// If neither is available, it returns an empty string (caller should use default template).
// All file paths are securely validated to prevent path traversal vulnerabilities.
//
// Parameters:
//   - path: The path to the template file (can be empty)
//
// Returns:
//   - The template content as a string
//   - An error if loading fails
func LoadPromptTemplate(path string) (string, error) {
	// Get current working directory as the base for validation
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	// For custom path, properly validate against the entire filesystem
	if path != "" {
		// Clean and absolutize the path first, then validate against filesystem root
		cleanPath := filepath.Clean(path)
		absPath, err := filepath.Abs(cleanPath)
		if err != nil {
			return "", fmt.Errorf("invalid prompt template path: %w", err)
		}

		// Use filesystem root ("/") as baseDir to enforce path validity
		// but not containment within any specific directory
		// allowBaseDir=true because the root itself is allowed
		// mustExist=true to ensure the file exists and is not a directory
		rootDir := "/"
		validPath, err := filesystem.ValidateFilePath(absPath, rootDir, true, true)
		if err != nil {
			return "", fmt.Errorf("failed to validate prompt template path: %w", err)
		}

		// Read the validated file path
		// #nosec G304 -- The path has been validated using filesystem.ValidateFilePath
		data, err := os.ReadFile(validPath)
		if err != nil {
			return "", fmt.Errorf("failed to read custom prompt template from '%s': %w", validPath, err)
		}
		return string(data), nil
	}

	// Try the default prompt.txt in the current directory
	defaultPromptPath := filepath.Join(cwd, "prompt.txt")
	// Check if the file exists, but don't return an error if it doesn't
	if _, err := os.Stat(defaultPromptPath); err == nil {
		// Validate the default path against the current working directory
		// allowBaseDir=false because prompt.txt should be in CWD, not be CWD itself
		// mustExist=true to ensure the file exists and is a file
		validDefaultPath, err := filesystem.ValidateFilePath(defaultPromptPath, cwd, false, true)
		if err != nil {
			// We just log this rather than fail, to maintain backward compatibility
			// with existing behavior of falling back to default template
			return "", fmt.Errorf("failed to validate default prompt template: %w", err)
		}

		// Read the validated file
		// #nosec G304 -- The path has been validated using filesystem.ValidateFilePath
		if data, err := os.ReadFile(validDefaultPath); err == nil {
			return string(data), nil
		}
	}

	// Return empty string - caller should use default template
	return "", nil
}
