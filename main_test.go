package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCLIExecution is a basic test to verify the CLI can be executed.
// This serves as an initial end-to-end test to ensure the binary runs successfully.
func TestCLIExecution(t *testing.T) {
	// Skip this test when running with "go test" directly as it requires
	// the compiled binary. This test is meant to be run after building.
	if os.Getenv("TEST_WITH_COMPILED_BINARY") != "true" {
		t.Skip("Skipping test that requires compiled binary. Set TEST_WITH_COMPILED_BINARY=true to run.")
	}

	// Create a temporary test directory
	tempDir, err := os.MkdirTemp("", "glance-test-*")
	assert.NoError(t, err, "Failed to create temp directory")
	defer os.RemoveAll(tempDir)

	// Create a simple file structure for testing
	testFile := filepath.Join(tempDir, "test.go")
	err = os.WriteFile(testFile, []byte("package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}\n"), 0644)
	assert.NoError(t, err, "Failed to create test file")

	// Execute the compiled binary with --help flag (doesn't require API key)
	cmd := exec.Command("./glance", "--help")
	output, err := cmd.CombinedOutput()
	
	// We just check if it runs without crashing - not checking specific output
	// since it will change as the CLI evolves
	assert.NoError(t, err, "CLI execution failed with error: %v, output: %s", err, output)
	assert.NotEmpty(t, output, "CLI execution produced no output")
}

// TestUsage verifies that the CLI reports correct usage when run with no arguments
func TestUsage(t *testing.T) {
	if os.Getenv("TEST_WITH_COMPILED_BINARY") != "true" {
		t.Skip("Skipping test that requires compiled binary. Set TEST_WITH_COMPILED_BINARY=true to run.")
	}

	// Run glance with no arguments - should exit with non-zero status
	cmd := exec.Command("./glance")
	output, err := cmd.CombinedOutput()
	
	// Should fail due to missing arguments
	assert.Error(t, err, "Expected CLI to fail with no arguments")
	assert.Contains(t, string(output), "Usage:", "Expected usage information in output")
}