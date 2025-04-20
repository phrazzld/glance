package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// Tests are skipped, but would use internal/mocks.LLMClient when implemented

// Additional setup specific to this integration test
func setupIntegrationTest(t *testing.T) (string, func()) {
	testDir, err := os.MkdirTemp("", "glance-integration-test-*")
	require.NoError(t, err, "Failed to create temp test directory")

	// Create test files
	mainGo := filepath.Join(testDir, "main.go")
	err = os.WriteFile(mainGo, []byte("package main\n\nfunc main() {\n\t// Test\n}\n"), 0644)
	require.NoError(t, err, "Failed to create main.go")

	readmeMd := filepath.Join(testDir, "README.md")
	err = os.WriteFile(readmeMd, []byte("# Test Project\n\nDescription."), 0644)
	require.NoError(t, err, "Failed to create README.md")

	// Return cleanup function
	return testDir, func() {
		err := os.RemoveAll(testDir)
		if err != nil {
			t.Logf("Warning: failed to clean up test directory: %v", err)
		}
	}
}

// TestFileSystemLLMIntegration verifies the integration between the filesystem
// package and the LLM package, particularly the flow of scanning files and
// generating glance.md content.
func TestFileSystemLLMIntegration(t *testing.T) {
	// Skip all integration tests for now - they require more setup for MockClient
	t.Skip("Skipping integration tests - Need to fix MockClient implementation")

	t.Run("File content from filesystem flows to LLM", func(t *testing.T) {
		t.Skip("Test needs to be fixed")
	})

	t.Run("Respects .gitignore patterns", func(t *testing.T) {
		t.Skip("Test needs to be fixed")
	})
}
