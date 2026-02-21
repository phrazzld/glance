package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"glance/filesystem"
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

// TestGlanceWithTestStructure tests the actual glance.md generation with a realistic directory structure
func TestGlanceWithTestStructure(t *testing.T) {
	if os.Getenv("TEST_WITH_COMPILED_BINARY") != "true" {
		t.Skip("Skipping test that requires compiled binary. Set TEST_WITH_COMPILED_BINARY=true to run.")
	}

	if os.Getenv("GEMINI_API_KEY") == "" {
		t.Skip("Skipping test that requires GEMINI_API_KEY. Set a valid GEMINI_API_KEY environment variable to run.")
	}

	// Set up test directory structure
	testProjectDir, cleanup := setupTestProjectStructure(t)
	defer cleanup()

	// Execute the glance binary on the test directory
	cmd := exec.Command("./glance", testProjectDir)
	output, err := cmd.CombinedOutput()

	// The command should succeed
	require.NoError(t, err, "Glance command failed with output: %s", output)

	// Verify glance.md files were created in each directory (except ignored ones)
	mainGlanceFile := filepath.Join(testProjectDir, filesystem.GlanceFilename)
	subdir1GlanceFile := filepath.Join(testProjectDir, "subdir1", filesystem.GlanceFilename)
	subdir2GlanceFile := filepath.Join(testProjectDir, "subdir2", filesystem.GlanceFilename)

	assert.FileExists(t, mainGlanceFile, "glance output should exist in root directory")
	assert.FileExists(t, subdir1GlanceFile, "glance output should exist in subdir1")
	assert.FileExists(t, subdir2GlanceFile, "glance output should exist in subdir2")

	// Verify that glance output was not created in ignored directory
	ignoredGlanceFile := filepath.Join(testProjectDir, "ignored_dir", filesystem.GlanceFilename)
	assert.NoFileExists(t, ignoredGlanceFile, "glance output should not exist in ignored_dir")
}

// TestGlanceForceFlag tests the behavior of the --force flag
func TestGlanceForceFlag(t *testing.T) {
	if os.Getenv("TEST_WITH_COMPILED_BINARY") != "true" {
		t.Skip("Skipping test that requires compiled binary. Set TEST_WITH_COMPILED_BINARY=true to run.")
	}

	if os.Getenv("GEMINI_API_KEY") == "" {
		t.Skip("Skipping test that requires GEMINI_API_KEY. Set a valid GEMINI_API_KEY environment variable to run.")
	}

	// Set up test directory structure
	testProjectDir, cleanup := setupTestProjectStructure(t)
	defer cleanup()

	// Create an initial glance.md file with known content
	initialContent := "# Initial content - should be replaced when --force is used"
	mainGlanceFile := filepath.Join(testProjectDir, filesystem.GlanceFilename)
	err := os.WriteFile(mainGlanceFile, []byte(initialContent), 0644)
	require.NoError(t, err, "Failed to create initial glance output file")

	// Get the creation time
	initialStat, err := os.Stat(mainGlanceFile)
	require.NoError(t, err, "Failed to stat initial glance output file")
	initialTime := initialStat.ModTime()

	// Wait a moment to ensure file timestamps will be different
	time.Sleep(1 * time.Second)

	// Run glance normally (should not replace existing glance output)
	cmd := exec.Command("./glance", testProjectDir)
	_, err = cmd.CombinedOutput()
	require.NoError(t, err, "First glance run failed")

	// Check that the file was not modified
	currentStat, err := os.Stat(mainGlanceFile)
	require.NoError(t, err, "Failed to stat glance output after first run")
	assert.Equal(t, initialTime, currentStat.ModTime(), "glance output should not have been modified")

	// Run glance with --force flag
	cmd = exec.Command("./glance", "--force", testProjectDir)
	_, err = cmd.CombinedOutput()
	require.NoError(t, err, "Glance with --force flag failed")

	// Check that the file was modified
	currentStat, err = os.Stat(mainGlanceFile)
	require.NoError(t, err, "Failed to stat glance output after force run")
	assert.NotEqual(t, initialTime, currentStat.ModTime(), "glance output should have been modified with --force")

	// Check content was changed
	content, err := os.ReadFile(mainGlanceFile)
	require.NoError(t, err, "Failed to read glance output content")
	assert.NotEqual(t, initialContent, string(content), "glance output content should have changed")
}

// TestGlanceWithModifiedFiles tests that glance.md is regenerated when files in the directory are modified
func TestGlanceWithModifiedFiles(t *testing.T) {
	if os.Getenv("TEST_WITH_COMPILED_BINARY") != "true" {
		t.Skip("Skipping test that requires compiled binary. Set TEST_WITH_COMPILED_BINARY=true to run.")
	}

	if os.Getenv("GEMINI_API_KEY") == "" {
		t.Skip("Skipping test that requires GEMINI_API_KEY. Set a valid GEMINI_API_KEY environment variable to run.")
	}

	// Set up test directory structure
	testProjectDir, cleanup := setupTestProjectStructure(t)
	defer cleanup()

	// Run glance to generate initial glance.md files
	cmd := exec.Command("./glance", testProjectDir)
	_, err := cmd.CombinedOutput()
	require.NoError(t, err, "Initial glance run failed")

	// Get the initial modification time of the glance.md file
	mainGlanceFile := filepath.Join(testProjectDir, filesystem.GlanceFilename)
	initialStat, err := os.Stat(mainGlanceFile)
	require.NoError(t, err, "Failed to stat initial glance output file")
	initialTime := initialStat.ModTime()

	// Wait a moment to ensure file timestamps will be different
	time.Sleep(1 * time.Second)

	// Modify a file in the test directory
	testFile := filepath.Join(testProjectDir, "main.go")
	err = os.WriteFile(testFile, []byte("package main\n\nfunc main() {\n\tprintln(\"Updated content!\")\n}\n"), 0644)
	require.NoError(t, err, "Failed to modify test file")

	// Run glance again (should detect modified file and regenerate glance.md)
	cmd = exec.Command("./glance", testProjectDir)
	_, err = cmd.CombinedOutput()
	require.NoError(t, err, "Second glance run failed")

	// Check that the glance.md file was regenerated
	currentStat, err := os.Stat(mainGlanceFile)
	require.NoError(t, err, "Failed to stat glance output after file modification")
	assert.NotEqual(t, initialTime, currentStat.ModTime(), "glance output should have been regenerated after file modification")
}

// TestDebugLogging verifies that debug level logs are present by default
func TestDebugLogging(t *testing.T) {
	if os.Getenv("TEST_WITH_COMPILED_BINARY") != "true" {
		t.Skip("Skipping test that requires compiled binary. Set TEST_WITH_COMPILED_BINARY=true to run.")
	}

	if os.Getenv("GEMINI_API_KEY") == "" {
		t.Skip("Skipping test that requires GEMINI_API_KEY. Set a valid GEMINI_API_KEY environment variable to run.")
	}

	// Set up test directory structure
	testProjectDir, cleanup := setupTestProjectStructure(t)
	defer cleanup()

	// Run glance with default settings
	cmd := exec.Command("./glance", testProjectDir)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Glance execution failed")

	// Check for debug level messages in the output
	outputStr := string(output)
	assert.Contains(t, outputStr, "level=debug", "Output should contain debug level messages by default")
}

// TestGlanceWithCustomPromptFile tests using a custom prompt file
func TestGlanceWithCustomPromptFile(t *testing.T) {
	if os.Getenv("TEST_WITH_COMPILED_BINARY") != "true" {
		t.Skip("Skipping test that requires compiled binary. Set TEST_WITH_COMPILED_BINARY=true to run.")
	}

	if os.Getenv("GEMINI_API_KEY") == "" {
		t.Skip("Skipping test that requires GEMINI_API_KEY. Set a valid GEMINI_API_KEY environment variable to run.")
	}

	// Set up test directory structure
	testProjectDir, cleanup := setupTestProjectStructure(t)
	defer cleanup()

	// Create a custom prompt file
	customPromptFile := filepath.Join(testProjectDir, "custom-prompt.txt")
	customPromptContent := "Custom prompt template for testing {{.Directory}}"
	err := os.WriteFile(customPromptFile, []byte(customPromptContent), 0644)
	require.NoError(t, err, "Failed to create custom prompt file")

	// Run glance with custom prompt file
	cmd := exec.Command("./glance", "--prompt-file", customPromptFile, testProjectDir)
	_, err = cmd.CombinedOutput()
	require.NoError(t, err, "Glance with custom prompt file failed")

	// Verify glance.md was created
	mainGlanceFile := filepath.Join(testProjectDir, filesystem.GlanceFilename)
	assert.FileExists(t, mainGlanceFile, "glance output should exist when using custom prompt file")
}

// TestGlanceChangePropagation tests that changes in subdirectories trigger regeneration in parent directories
func TestGlanceChangePropagation(t *testing.T) {
	if os.Getenv("TEST_WITH_COMPILED_BINARY") != "true" {
		t.Skip("Skipping test that requires compiled binary. Set TEST_WITH_COMPILED_BINARY=true to run.")
	}

	if os.Getenv("GEMINI_API_KEY") == "" {
		t.Skip("Skipping test that requires GEMINI_API_KEY. Set a valid GEMINI_API_KEY environment variable to run.")
	}

	// Set up test directory structure
	testProjectDir, cleanup := setupTestProjectStructure(t)
	defer cleanup()

	// Run glance to generate initial glance.md files
	cmd := exec.Command("./glance", testProjectDir)
	_, err := cmd.CombinedOutput()
	require.NoError(t, err, "Initial glance run failed")

	// Get the initial modification time of the root glance.md file
	mainGlanceFile := filepath.Join(testProjectDir, filesystem.GlanceFilename)

	mainInitialStat, err := os.Stat(mainGlanceFile)
	require.NoError(t, err, "Failed to stat initial root glance.md")
	mainInitialTime := mainInitialStat.ModTime()

	// Wait a moment to ensure file timestamps will be different
	time.Sleep(1 * time.Second)

	// Force regeneration of a subdirectory's glance.md file
	cmd = exec.Command("./glance", "--force", filepath.Join(testProjectDir, "subdir1"))
	_, err = cmd.CombinedOutput()
	require.NoError(t, err, "Force regeneration of subdir1 glance.md failed")

	// Run glance on the entire directory again
	cmd = exec.Command("./glance", testProjectDir)
	_, err = cmd.CombinedOutput()
	require.NoError(t, err, "Second glance run failed")

	// Check that the parent glance.md file was also regenerated due to changes in the subdirectory
	mainCurrentStat, err := os.Stat(mainGlanceFile)
	require.NoError(t, err, "Failed to stat root glance.md after subdirectory change")
	assert.NotEqual(t, mainInitialTime, mainCurrentStat.ModTime(),
		"Root glance.md should have been regenerated after subdirectory's glance.md changed")
}

// TestBinaryFileHandling tests that binary files are properly ignored
func TestBinaryFileHandling(t *testing.T) {
	if os.Getenv("TEST_WITH_COMPILED_BINARY") != "true" {
		t.Skip("Skipping test that requires compiled binary. Set TEST_WITH_COMPILED_BINARY=true to run.")
	}

	if os.Getenv("GEMINI_API_KEY") == "" {
		t.Skip("Skipping test that requires GEMINI_API_KEY. Set a valid GEMINI_API_KEY environment variable to run.")
	}

	// Set up test directory structure
	testProjectDir, cleanup := setupTestProjectStructure(t)
	defer cleanup()

	// Create a binary file (using PNG header)
	binaryFileContent := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, // IHDR chunk length
		0x49, 0x48, 0x44, 0x52, // "IHDR"
		// Add some random bytes to make it look like a real binary file
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10,
	}

	binaryFile := filepath.Join(testProjectDir, "binary.png")
	err := os.WriteFile(binaryFile, binaryFileContent, 0644)
	require.NoError(t, err, "Failed to create binary file")

	// Run glance (debug logging is now on by default)
	cmd := exec.Command("./glance", testProjectDir)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Glance run failed")

	// Check debug output for binary file detection
	outputStr := string(output)
	assert.Contains(t, outputStr, "binary", "Output should mention binary file detection")

	// Verify glance output was created
	mainGlanceFile := filepath.Join(testProjectDir, filesystem.GlanceFilename)
	content, err := os.ReadFile(mainGlanceFile)
	require.NoError(t, err, "Failed to read glance output content")

	// The content should not reference the binary file content
	contentStr := string(content)
	assert.NotContains(t, contentStr, string(binaryFileContent),
		"glance output should not contain binary file content")
}

// setupTestProjectStructure creates a test directory structure for testing the glance tool
// Returns the path to the test directory and a cleanup function
func setupTestProjectStructure(t *testing.T) (string, func()) {
	t.Helper()

	// Create a root directory for the test project
	testProjectDir, err := os.MkdirTemp("", "glance-project-test-*")
	require.NoError(t, err, "Failed to create test project directory")

	// Create subdirectories
	subdir1 := filepath.Join(testProjectDir, "subdir1")
	subdir2 := filepath.Join(testProjectDir, "subdir2")
	ignoredDir := filepath.Join(testProjectDir, "ignored_dir")

	for _, dir := range []string{subdir1, subdir2, ignoredDir} {
		err := os.MkdirAll(dir, 0755)
		require.NoError(t, err, "Failed to create subdirectory: "+dir)
	}

	// Create files in the root directory
	mainGo := `package main

import (
	"fmt"
)

func main() {
	fmt.Println("Hello from test project")
}
`
	err = os.WriteFile(filepath.Join(testProjectDir, "main.go"), []byte(mainGo), 0644)
	require.NoError(t, err, "Failed to create main.go file")

	readmeMd := `# Test Project

This is a test project for testing the glance tool.
`
	err = os.WriteFile(filepath.Join(testProjectDir, "README.md"), []byte(readmeMd), 0644)
	require.NoError(t, err, "Failed to create README.md file")

	// Create a .gitignore file that ignores ignored_dir
	gitignore := `ignored_dir/
*.log
`
	err = os.WriteFile(filepath.Join(testProjectDir, ".gitignore"), []byte(gitignore), 0644)
	require.NoError(t, err, "Failed to create .gitignore file")

	// Create files in subdirectories
	err = os.WriteFile(filepath.Join(subdir1, "file1.txt"), []byte("Content of file1.txt"), 0644)
	require.NoError(t, err, "Failed to create file1.txt")

	err = os.WriteFile(filepath.Join(subdir2, "file2.txt"), []byte("Content of file2.txt"), 0644)
	require.NoError(t, err, "Failed to create file2.txt")

	// Even though this is in an ignored directory, create a file for thoroughness
	err = os.WriteFile(filepath.Join(ignoredDir, "ignored.txt"), []byte("This file should be ignored"), 0644)
	require.NoError(t, err, "Failed to create ignored.txt")

	// Return the test directory and a cleanup function
	cleanup := func() {
		os.RemoveAll(testProjectDir)
	}

	return testProjectDir, cleanup
}
