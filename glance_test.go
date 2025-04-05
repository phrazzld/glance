package main

import (
	"os"
	"path/filepath"
	"testing"

	gitignore "github.com/sabhiram/go-gitignore"
	"github.com/stretchr/testify/assert"
)

// TestLoadPromptTemplate verifies the prompt template loading functionality
func TestLoadPromptTemplate(t *testing.T) {
	// Create a temporary test directory
	tempDir, err := os.MkdirTemp("", "glance-prompt-test-*")
	assert.NoError(t, err, "Failed to create temp directory")
	defer os.RemoveAll(tempDir)

	// Create a test prompt file
	testPromptPath := filepath.Join(tempDir, "test-prompt.txt")
	testPromptContent := "test prompt template {{.Directory}}"
	err = os.WriteFile(testPromptPath, []byte(testPromptContent), 0644)
	assert.NoError(t, err, "Failed to create test prompt file")

	// Test loading the custom prompt file
	loadedPrompt, err := loadPromptTemplate(testPromptPath)
	assert.NoError(t, err, "Failed to load test prompt template")
	assert.Equal(t, testPromptContent, loadedPrompt, "Loaded prompt content doesn't match expected")

	// Test with empty path - should check for prompt.txt
	// Since prompt.txt exists in this project, let's verify it loads correctly
	emptyPathResult, err := loadPromptTemplate("")
	assert.NoError(t, err, "Loading with empty path should succeed")
	assert.NotEmpty(t, emptyPathResult, "Should return a non-empty template when path is empty")
}

// TestIsIgnored verifies .gitignore pattern matching
func TestIsIgnored(t *testing.T) {
	// This test doesn't need any setup, as isIgnored is a pure function
	// Create some mock ignores for testing
	mockIgnoreContent := []byte("*.log\ntmp/\n")

	// Create a temporary gitignore file
	tempDir, err := os.MkdirTemp("", "glance-gitignore-test-*")
	assert.NoError(t, err, "Failed to create temp directory")
	defer os.RemoveAll(tempDir)

	gitignorePath := filepath.Join(tempDir, ".gitignore")
	err = os.WriteFile(gitignorePath, mockIgnoreContent, 0644)
	assert.NoError(t, err, "Failed to create test .gitignore file")

	// Load the gitignore
	mockIgnore, err := loadGitignore(tempDir)
	assert.NoError(t, err, "Failed to load test gitignore")

	// Test cases
	testCases := []struct {
		path     string
		expected bool
	}{
		{"file.txt", false},         // Regular file, not ignored
		{"file.log", true},          // Matches *.log pattern
		{"tmp/file.txt", true},      // Inside ignored directory
		{"logs/file.txt", false},    // Not ignored
		{"tmp/logs/file.txt", true}, // Inside ignored directory
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			result := isIgnored(tc.path, []*gitignore.GitIgnore{mockIgnore})
			assert.Equal(t, tc.expected, result, "isIgnored(%q) should return %v", tc.path, tc.expected)
		})
	}
}

// TestHelpers contains reusable test utilities
func setupTestDir(t *testing.T, prefix string) (string, func()) {
	tempDir, err := os.MkdirTemp("", prefix)
	assert.NoError(t, err, "Failed to create temp directory")
	
	cleanup := func() {
		os.RemoveAll(tempDir)
	}
	
	return tempDir, cleanup
}