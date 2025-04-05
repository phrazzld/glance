package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockDirectoryChecker implements directoryChecker for testing
type mockDirectoryChecker struct {
	shouldPass bool
	errorMsg   string
}

func (m *mockDirectoryChecker) CheckDirectory(path string) error {
	if !m.shouldPass {
		return errors.New(m.errorMsg)
	}
	return nil
}

// Setup and teardown for directory checker
func setupMockDirectoryChecker(shouldPass bool, errorMsg string) func() {
	original := dirChecker
	dirChecker = &mockDirectoryChecker{shouldPass: shouldPass, errorMsg: errorMsg}
	
	return func() {
		dirChecker = original
	}
}

func TestLoadConfig(t *testing.T) {
	// Setup the mock directory checker to pass
	cleanup := setupMockDirectoryChecker(true, "")
	defer cleanup()

	// Save and restore environment variables
	origEnv := os.Getenv("GEMINI_API_KEY")
	defer os.Setenv("GEMINI_API_KEY", origEnv)

	// Set test API key in environment
	testAPIKey := "test-gemini-api-key"
	os.Setenv("GEMINI_API_KEY", testAPIKey)

	// Create test arguments
	args := []string{"glance", "--force", "--verbose", "/test/dir"}

	// Run the function
	cfg, err := LoadConfig(args)

	// Verify no error
	require.NoError(t, err, "LoadConfig should not return an error with valid inputs")

	// Check the configuration values
	assert.Equal(t, testAPIKey, cfg.APIKey, "API Key should be set from environment")
	assert.Equal(t, "/test/dir", cfg.TargetDir, "Target directory should be set from arguments")
	assert.True(t, cfg.Force, "Force flag should be true")
	assert.True(t, cfg.Verbose, "Verbose flag should be true")
	assert.NotEmpty(t, cfg.PromptTemplate, "Prompt template should not be empty")
	assert.Equal(t, DefaultMaxRetries, cfg.MaxRetries, "MaxRetries should have default value")
	assert.Equal(t, int64(DefaultMaxFileBytes), cfg.MaxFileBytes, "MaxFileBytes should have default value")
}

func TestLoadConfigWithCustomPromptFile(t *testing.T) {
	// Setup the mock directory checker to pass
	cleanup := setupMockDirectoryChecker(true, "")
	defer cleanup()

	// Save and restore environment variables
	origEnv := os.Getenv("GEMINI_API_KEY")
	defer os.Setenv("GEMINI_API_KEY", origEnv)
	
	// Set test API key in environment
	os.Setenv("GEMINI_API_KEY", "test-api-key")

	// Create a temporary prompt file
	tempDir, err := os.MkdirTemp("", "glance-test-*")
	require.NoError(t, err, "Failed to create temp directory")
	defer os.RemoveAll(tempDir)

	customPromptPath := filepath.Join(tempDir, "custom-prompt.txt")
	customPromptContent := "custom prompt template for testing {{.Directory}}"
	err = os.WriteFile(customPromptPath, []byte(customPromptContent), 0644)
	require.NoError(t, err, "Failed to create custom prompt file")

	// Create test arguments with custom prompt file
	args := []string{"glance", "--prompt-file", customPromptPath, "/test/dir"}

	// Run the function
	cfg, err := LoadConfig(args)

	// Verify no error
	require.NoError(t, err, "LoadConfig should not return an error with valid inputs")

	// Check the prompt template was loaded correctly
	assert.Equal(t, customPromptContent, cfg.PromptTemplate, "Prompt template should be loaded from file")
}

func TestLoadConfigMissingAPIKey(t *testing.T) {
	// Setup the mock directory checker to pass
	cleanup := setupMockDirectoryChecker(true, "")
	defer cleanup()

	// Save and restore environment variables
	origEnv := os.Getenv("GEMINI_API_KEY")
	defer os.Setenv("GEMINI_API_KEY", origEnv)
	
	// Clear the API key from environment
	os.Setenv("GEMINI_API_KEY", "")

	// Create test arguments
	args := []string{"glance", "/test/dir"}

	// Run the function
	_, err := LoadConfig(args)

	// Verify error for missing API key
	assert.Error(t, err, "LoadConfig should return an error when GEMINI_API_KEY is missing")
	assert.Contains(t, err.Error(), "GEMINI_API_KEY", "Error should mention missing API key")
}

func TestLoadConfigMissingTargetDir(t *testing.T) {
	// Save and restore environment variables
	origEnv := os.Getenv("GEMINI_API_KEY")
	defer os.Setenv("GEMINI_API_KEY", origEnv)
	
	// Set test API key in environment
	os.Setenv("GEMINI_API_KEY", "test-api-key")

	// Create test arguments without target directory
	args := []string{"glance"}

	// Run the function
	_, err := LoadConfig(args)

	// Verify error for missing target directory
	assert.Error(t, err, "LoadConfig should return an error when target directory is missing")
	assert.Contains(t, err.Error(), "directory", "Error should mention missing directory")
}

func TestLoadConfigInvalidPromptFile(t *testing.T) {
	// Setup the mock directory checker to pass
	cleanup := setupMockDirectoryChecker(true, "")
	defer cleanup()

	// Save and restore environment variables
	origEnv := os.Getenv("GEMINI_API_KEY")
	defer os.Setenv("GEMINI_API_KEY", origEnv)
	
	// Set test API key in environment
	os.Setenv("GEMINI_API_KEY", "test-api-key")

	// Create test arguments with non-existent prompt file
	args := []string{"glance", "--prompt-file", "/path/to/nonexistent/prompt.txt", "/test/dir"}

	// Run the function
	_, err := LoadConfig(args)

	// Verify error for invalid prompt file
	assert.Error(t, err, "LoadConfig should return an error when prompt file doesn't exist")
	assert.Contains(t, err.Error(), "prompt", "Error should mention prompt file issue")
}

func TestLoadConfigInvalidDirectory(t *testing.T) {
	// Setup the mock directory checker to fail
	dirErrorMsg := "cannot access directory: permission denied"
	cleanup := setupMockDirectoryChecker(false, dirErrorMsg)
	defer cleanup()

	// Save and restore environment variables
	origEnv := os.Getenv("GEMINI_API_KEY")
	defer os.Setenv("GEMINI_API_KEY", origEnv)
	
	// Set test API key in environment
	os.Setenv("GEMINI_API_KEY", "test-api-key")

	// Create test arguments
	args := []string{"glance", "/test/dir"}

	// Run the function
	_, err := LoadConfig(args)

	// Verify error for invalid directory
	assert.Error(t, err, "LoadConfig should return an error when directory check fails")
	assert.Contains(t, err.Error(), dirErrorMsg, "Error should contain the directory error message")
}