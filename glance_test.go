package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"glance/filesystem"
	"glance/llm"
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
	loadedPrompt, err := llm.LoadTemplate(testPromptPath)
	assert.NoError(t, err, "Failed to load test prompt template")
	assert.Equal(t, testPromptContent, loadedPrompt, "Loaded prompt content doesn't match expected")

	// Test with empty path - should check for prompt.txt
	// Since prompt.txt exists in this project, let's verify it loads correctly
	emptyPathResult, err := llm.LoadTemplate("")
	assert.NoError(t, err, "Loading with empty path should succeed")
	assert.NotEmpty(t, emptyPathResult, "Should return a non-empty template when path is empty")
}

// TestFileSystemPackageUsage demonstrates using the filesystem package directly
// This test is a placeholder to verify that we can use the filesystem package functions
// that replaced the removed functions in glance.go
func TestFileSystemPackageUsage(t *testing.T) {
	// Create a temporary test directory
	tempDir, err := os.MkdirTemp("", "glance-filesystem-test-*")
	assert.NoError(t, err, "Failed to create temp directory")
	defer os.RemoveAll(tempDir)

	// Demonstrate loading a gitignore using the filesystem package
	_, err = filesystem.LoadGitignore(tempDir)
	assert.NoError(t, err, "Failed to use filesystem.LoadGitignore")

	// Create an empty IgnoreChain
	ignoreChain := filesystem.IgnoreChain{}

	// Demonstrate checking if regeneration is needed
	_, err = filesystem.ShouldRegenerate(tempDir, false, ignoreChain, false)
	assert.NoError(t, err, "Failed to use filesystem.ShouldRegenerate")

	// Demonstrate bubbling up regeneration flags
	needs := make(map[string]bool)
	filesystem.BubbleUpParents(tempDir, filepath.Dir(tempDir), needs)
	
	// That's enough to verify we can use the filesystem package functions
	// directly without depending on the removed functions in glance.go
}

// Note: setupTestDir function was merged into setupIntegrationTest in integration_test.go
