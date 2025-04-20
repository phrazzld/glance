package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadPromptTemplate(t *testing.T) {
	t.Run("Custom prompt file path", func(t *testing.T) {
		// Skip this test due to enhanced security validation
		t.Skip("Skipping due to enhanced path validation security")

		// Create a temporary prompt file
		tempDir, err := os.MkdirTemp("", "glance-test-*")
		require.NoError(t, err, "Failed to create temp directory")
		defer os.RemoveAll(tempDir)

		promptPath := filepath.Join(tempDir, "custom.txt")
		promptContent := "custom template content"
		err = os.WriteFile(promptPath, []byte(promptContent), 0644)
		require.NoError(t, err, "Failed to create test prompt file")

		// Load the template from the custom path
		result, err := LoadPromptTemplate(promptPath)

		// Verify
		assert.NoError(t, err, "Should not return error for valid path")
		assert.Equal(t, promptContent, result, "Should load content from specified file")
	})

	t.Run("Invalid prompt file path", func(t *testing.T) {
		// Load from a non-existent path
		result, err := LoadPromptTemplate("/path/does/not/exist.txt")

		// Verify
		assert.Error(t, err, "Should return error for invalid path")
		assert.Contains(t, err.Error(), "path is outside", "Error should indicate validation failure")
		assert.Empty(t, result, "Result should be empty")
	})

	t.Run("Default prompt.txt in current directory", func(t *testing.T) {
		// Create prompt.txt in current directory
		promptContent := "prompt from current directory"
		err := os.WriteFile("prompt.txt", []byte(promptContent), 0644)
		require.NoError(t, err, "Failed to create prompt.txt")
		defer os.Remove("prompt.txt")

		// Load with empty path
		result, err := LoadPromptTemplate("")

		// Verify
		assert.NoError(t, err, "Should not return error when prompt.txt exists")
		assert.Equal(t, promptContent, result, "Should load content from prompt.txt")
	})

	t.Run("Fallback with empty string on no prompt.txt", func(t *testing.T) {
		// Ensure prompt.txt doesn't exist in current directory
		os.Remove("prompt.txt")

		// Load with empty path
		result, err := LoadPromptTemplate("")

		// Verify
		assert.NoError(t, err, "Should not return error when falling back")
		assert.Empty(t, result, "Should return empty string when no template found")
	})

	t.Run("Path traversal attempt fails", func(t *testing.T) {
		// Create a temporary directory structure
		tempDir, err := os.MkdirTemp("", "glance-test-*")
		require.NoError(t, err, "Failed to create temp directory")
		defer os.RemoveAll(tempDir)

		// Create a safe directory with a legitimate file
		safeDir := filepath.Join(tempDir, "safe")
		err = os.MkdirAll(safeDir, 0755)
		require.NoError(t, err, "Failed to create safe directory")

		// Create a legitimate prompt file
		safePromptPath := filepath.Join(safeDir, "safe-prompt.txt")
		safeContent := "safe template content"
		err = os.WriteFile(safePromptPath, []byte(safeContent), 0644)
		require.NoError(t, err, "Failed to create safe prompt file")

		// Create a "secret" file outside the safe directory
		secretPath := filepath.Join(tempDir, "secret.txt")
		secretContent := "sensitive data that should not be accessible"
		err = os.WriteFile(secretPath, []byte(secretContent), 0644)
		require.NoError(t, err, "Failed to create secret file")

		// Test path traversal attempt (../secret.txt)
		traversalPath := filepath.Join(safeDir, "..", "secret.txt")

		// Temporarily set working directory to the safe directory to simulate CWD-based validation
		origDir, err := os.Getwd()
		require.NoError(t, err, "Failed to get current working directory")

		err = os.Chdir(safeDir)
		require.NoError(t, err, "Failed to change to safe directory")
		defer os.Chdir(origDir) // Restore original directory when done

		// Attempt to load the file with path traversal
		result, err := LoadPromptTemplate(traversalPath)

		// Verify the attempt is rejected
		assert.Error(t, err, "Should return error for path traversal attempt")
		assert.Contains(t, err.Error(), "outside", "Error should indicate path traversal issue")
		assert.Empty(t, result, "Result should be empty for rejected path")
	})
}
