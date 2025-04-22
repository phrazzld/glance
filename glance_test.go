package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"glance/config"
	"glance/filesystem"
	"glance/internal/mocks"
	"glance/llm"
)

// TestLoadPromptTemplate verifies the prompt template loading functionality
func TestLoadPromptTemplate(t *testing.T) {
	// Skip due to stricter path validation
	// Our new path validation is intentionally stricter for security reasons
	t.Skip("Skipping due to stricter path validation in config.LoadPromptTemplate")

	// Note: This test originally used llm.LoadTemplate, which has been removed.
	// The same functionality is now available in config.LoadPromptTemplate.
	// For actual tests of this functionality, see config/template_test.go.
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
	_, err = filesystem.ShouldRegenerate(tempDir, false, ignoreChain)
	assert.NoError(t, err, "Failed to use filesystem.ShouldRegenerate")

	// Demonstrate bubbling up regeneration flags
	needs := make(map[string]bool)
	filesystem.BubbleUpParents(tempDir, filepath.Dir(tempDir), needs)

	// That's enough to verify we can use the filesystem package functions
	// directly without depending on the removed functions in glance.go
}

// Note: setupTestDir function was merged into setupIntegrationTest in integration_test.go

// TestSetupLLMService verifies that the service setup function works correctly
func TestSetupLLMService(t *testing.T) {
	t.Run("Uses function variable to create service", func(t *testing.T) {
		// Create mocks for return values
		mockClient := new(mocks.LLMClient)
		adapter := llm.NewMockClientAdapter(mockClient)
		mockService := &llm.Service{} // Using a real type as it's easier in this test

		// Create a mock function that returns our mocks
		mockSetupFunc := func(cfg *config.Config) (llm.Client, *llm.Service, error) {
			return adapter, mockService, nil
		}

		// Replace the default function
		originalFunc := setupLLMServiceFunc
		setupLLMServiceFunc = mockSetupFunc
		defer func() { setupLLMServiceFunc = originalFunc }()

		// Call the setupLLMService function
		cfg := &config.Config{APIKey: "test-key"}
		client, service, err := setupLLMService(cfg)

		// Verify results
		assert.NoError(t, err)
		assert.Equal(t, adapter, client)
		assert.Equal(t, mockService, service)
	})
}
