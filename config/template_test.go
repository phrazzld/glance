package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadPromptTemplate(t *testing.T) {
	t.Run("Custom prompt file path", func(t *testing.T) {
		// Use t.TempDir() which provides a temporary directory for the test
		// that will be cleaned up when the test completes
		tempDir := t.TempDir()

		// Create a test file in the temporary directory
		promptPath := filepath.Join(tempDir, "custom.txt")
		promptContent := "custom template content"
		err := os.WriteFile(promptPath, []byte(promptContent), 0644)
		require.NoError(t, err, "Failed to create test prompt file")

		// Save original function
		originalValidateFilePath := validateFilePath
		defer func() {
			// Restore the original function after test
			validateFilePath = originalValidateFilePath
		}()

		// Mock ValidateFilePath to allow our test path
		validateFilePath = func(path, baseDir string, allowBaseDir, mustExist bool) (string, error) {
			// For this test, we'll just check if the file exists
			if mustExist {
				fileInfo, err := os.Stat(path)
				if err != nil {
					return "", err
				}
				if fileInfo.IsDir() {
					return "", fmt.Errorf("path %q is a directory, expected a file", path)
				}
			}
			return path, nil
		}

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
		assert.Contains(t, err.Error(), "path", "Error should indicate validation failure")
		assert.Empty(t, result, "Result should be empty")
	})

	t.Run("Default prompt.txt in current directory", func(t *testing.T) {
		// Save current directory to restore later
		originalDir, err := os.Getwd()
		require.NoError(t, err, "Failed to get current working directory")

		// Create a temporary directory and change to it
		tempDir := t.TempDir()
		err = os.Chdir(tempDir)
		require.NoError(t, err, "Failed to change to temp directory")

		// Restore original directory when done
		defer func() {
			err := os.Chdir(originalDir)
			if err != nil {
				t.Logf("Failed to restore original directory: %v", err)
			}
		}()

		// Create prompt.txt in temporary directory
		promptContent := "prompt from current directory"
		err = os.WriteFile("prompt.txt", []byte(promptContent), 0644)
		require.NoError(t, err, "Failed to create prompt.txt")

		// Save original function
		originalValidateFilePath := validateFilePath
		defer func() {
			// Restore the original function after test
			validateFilePath = originalValidateFilePath
		}()

		// Mock ValidateFilePath to allow our test path
		validateFilePath = func(path, baseDir string, allowBaseDir, mustExist bool) (string, error) {
			// For this test, we'll just check if the file exists
			if mustExist {
				fileInfo, err := os.Stat(path)
				if err != nil {
					return "", err
				}
				if fileInfo.IsDir() {
					return "", fmt.Errorf("path %q is a directory, expected a file", path)
				}
			}
			return path, nil
		}

		// Load with empty path
		result, err := LoadPromptTemplate("")

		// Verify
		assert.NoError(t, err, "Should not return error when prompt.txt exists")
		assert.Equal(t, promptContent, result, "Should load content from prompt.txt")
	})

	t.Run("Fallback with empty string on no prompt.txt", func(t *testing.T) {
		// Save current directory to restore later
		originalDir, err := os.Getwd()
		require.NoError(t, err, "Failed to get current working directory")

		// Create a temporary directory and change to it
		tempDir := t.TempDir()
		err = os.Chdir(tempDir)
		require.NoError(t, err, "Failed to change to temp directory")

		// Restore original directory when done
		defer func() {
			err := os.Chdir(originalDir)
			if err != nil {
				t.Logf("Failed to restore original directory: %v", err)
			}
		}()

		// Load with empty path
		result, err := LoadPromptTemplate("")

		// Verify
		assert.NoError(t, err, "Should not return error when falling back")
		assert.Empty(t, result, "Should return empty string when no template found")
	})

	t.Run("Path traversal attempt fails", func(t *testing.T) {
		// Use t.TempDir() for temporary directory
		tempDir := t.TempDir()

		// Create a safe directory with a legitimate file
		safeDir := filepath.Join(tempDir, "safe")
		err := os.MkdirAll(safeDir, 0755)
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

		// Save original function
		originalValidateFilePath := validateFilePath
		defer func() {
			// Restore the original function after test
			validateFilePath = originalValidateFilePath
		}()

		// Mock ValidateFilePath to actually enforce path traversal checks
		validateFilePath = func(path, baseDir string, allowBaseDir, mustExist bool) (string, error) {
			// For this test, we do want to enforce containment within safeDir
			absPath, err := filepath.Abs(path)
			if err != nil {
				return "", err
			}

			absSafeDir, err := filepath.Abs(safeDir)
			if err != nil {
				return "", err
			}

			// Basic containment check
			if !strings.HasPrefix(absPath, absSafeDir) && absPath != absSafeDir {
				return "", fmt.Errorf("path %q is outside of allowed directory %q", path, safeDir)
			}

			// Check existence if required
			if mustExist {
				fileInfo, err := os.Stat(absPath)
				if err != nil {
					return "", err
				}
				if fileInfo.IsDir() {
					return "", fmt.Errorf("path %q is a directory, expected a file", path)
				}
			}

			return absPath, nil
		}

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
