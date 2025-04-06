package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSetupLogging verifies that the setupLogging function properly configures the logger
func TestSetupLogging(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	originalOutput := logrus.StandardLogger().Out
	logrus.SetOutput(&buf)
	defer logrus.SetOutput(originalOutput)

	// Test verbose=true
	setupLogging(true)
	assert.Equal(t, logrus.DebugLevel, logrus.GetLevel(), "Logger should be set to debug level when verbose is true")

	// Test verbose=false
	setupLogging(false)
	assert.Equal(t, logrus.InfoLevel, logrus.GetLevel(), "Logger should be set to info level when verbose is false")

	// Test formatter settings
	formatter, ok := logrus.StandardLogger().Formatter.(*logrus.TextFormatter)
	assert.True(t, ok, "Formatter should be TextFormatter")
	assert.True(t, formatter.FullTimestamp, "FullTimestamp should be true")
	assert.True(t, formatter.ForceColors, "ForceColors should be true")
}

// TestMainWithConfig verifies that the main function works with the new config package
func TestMainWithConfig(t *testing.T) {
	// This test confirms that our refactored main function can properly use the config package
	// It's more of an integration test and depends on the compiled binary

	if os.Getenv("TEST_WITH_COMPILED_BINARY") != "true" {
		t.Skip("Skipping test that requires compiled binary. Set TEST_WITH_COMPILED_BINARY=true to run.")
	}

	if os.Getenv("GEMINI_API_KEY") == "" {
		t.Skip("Skipping test that requires GEMINI_API_KEY. Set a valid GEMINI_API_KEY environment variable to run.")
	}

	// Set up test directory structure
	testProjectDir, cleanup := setupTestProjectStructure(t)
	defer cleanup()

	// Run the glance command on the test project
	cmd := exec.Command("./glance", testProjectDir)
	output, err := cmd.CombinedOutput()

	// Command should succeed
	require.NoError(t, err, "Glance command failed with output: %s", output)

	// Verify GLANCE.md files were created
	mainGlanceFile := filepath.Join(testProjectDir, "GLANCE.md")
	assert.FileExists(t, mainGlanceFile, "GLANCE.md should exist in test directory")
}
