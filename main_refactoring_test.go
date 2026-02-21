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

	"glance/filesystem"
)

// TestSetupLogging verifies that the setupLogging function properly configures the logger
// based on the GLANCE_LOG_LEVEL environment variable
func TestSetupLogging(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	originalOutput := logrus.StandardLogger().Out
	logrus.SetOutput(&buf)
	defer logrus.SetOutput(originalOutput)

	// Save current log level to restore after test
	originalLevel := logrus.GetLevel()
	defer logrus.SetLevel(originalLevel)

	// Test cases for various log level settings
	testCases := []struct {
		name          string
		envValue      string
		expectedLevel logrus.Level
	}{
		{
			name:          "debug level",
			envValue:      "debug",
			expectedLevel: logrus.DebugLevel,
		},
		{
			name:          "info level",
			envValue:      "info",
			expectedLevel: logrus.InfoLevel,
		},
		{
			name:          "warn level",
			envValue:      "warn",
			expectedLevel: logrus.WarnLevel,
		},
		{
			name:          "warning level (alternative)",
			envValue:      "warning",
			expectedLevel: logrus.WarnLevel,
		},
		{
			name:          "error level",
			envValue:      "error",
			expectedLevel: logrus.ErrorLevel,
		},
		{
			name:          "empty string defaults to info",
			envValue:      "",
			expectedLevel: logrus.InfoLevel,
		},
		{
			name:          "invalid value defaults to info",
			envValue:      "invalid_level",
			expectedLevel: logrus.InfoLevel,
		},
		{
			name:          "case insensitivity",
			envValue:      "DEBUG",
			expectedLevel: logrus.DebugLevel,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set environment variable for this test case
			if tc.envValue != "" {
				os.Setenv("GLANCE_LOG_LEVEL", tc.envValue)
				defer os.Unsetenv("GLANCE_LOG_LEVEL")
			} else {
				os.Unsetenv("GLANCE_LOG_LEVEL")
			}

			// Run the function being tested
			setupLogging()

			// Verify the log level was set correctly
			assert.Equal(t, tc.expectedLevel, logrus.GetLevel())
		})
	}

	// Test formatter settings (independent of log level)
	t.Run("formatter settings", func(t *testing.T) {
		os.Unsetenv("GLANCE_LOG_LEVEL")
		setupLogging()
		formatter, ok := logrus.StandardLogger().Formatter.(*logrus.TextFormatter)
		assert.True(t, ok, "Formatter should be TextFormatter")
		assert.True(t, formatter.FullTimestamp, "FullTimestamp should be true")
		assert.True(t, formatter.ForceColors, "ForceColors should be true")
	})
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

	// Verify glance.md files were created
	mainGlanceFile := filepath.Join(testProjectDir, filesystem.GlanceFilename)
	assert.FileExists(t, mainGlanceFile, "glance output should exist in test directory")
}
