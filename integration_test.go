package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"glance/config"
	"glance/llm"
)

// MockClient is a mock LLM client for testing
// This duplicates the mock in llm/client_test.go since it's package-private
type MockClient struct {
	mock.Mock
}

func (m *MockClient) Generate(ctx context.Context, prompt string) (string, error) {
	args := m.Called(ctx, prompt)
	return args.String(0), args.Error(1)
}

func (m *MockClient) CountTokens(ctx context.Context, prompt string) (int, error) {
	args := m.Called(ctx, prompt)
	return args.Int(0), args.Error(1)
}

func (m *MockClient) Close() {
	m.Called()
}

// setupIntegrationTest creates a test environment for integration tests.
// It sets up a temporary directory structure and returns paths and a cleanup function.
func setupIntegrationTest(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "glance-integration-test-*")
	require.NoError(t, err, "Failed to create temp directory")

	// Create a mock project structure
	createTestProjectStructure(t, tempDir)

	// Return cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

// createTestProjectStructure creates a simple test project structure for tests
func createTestProjectStructure(t *testing.T, root string) {
	// Create directories
	dirs := []string{
		"src",
		"src/utils",
		"src/models",
		"docs",
		"tests",
	}

	for _, dir := range dirs {
		err := os.MkdirAll(filepath.Join(root, dir), 0755)
		require.NoError(t, err, "Failed to create directory: "+dir)
	}

	// Create test files
	files := map[string]string{
		"README.md":            "# Test Project\n\nThis is a test project for integration tests.\n",
		"main.go":              "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}\n",
		"src/utils/utils.go":   "package utils\n\nfunc GetMessage() string {\n\treturn \"Hello from utils\"\n}\n",
		"src/models/user.go":   "package models\n\ntype User struct {\n\tName string\n}\n",
		"docs/api.md":          "# API Documentation\n\nAPI endpoints for the test project.\n",
		"tests/main_test.go":   "package tests\n\nimport \"testing\"\n\nfunc TestMain(t *testing.T) {\n\t// Test code\n}\n",
		".gitignore":           "build/\n*.log\n",
		"prompt.txt":           "Test prompt with {{.Directory}} and {{.SubGlances}} and {{.FileContents}}",
	}

	for path, content := range files {
		err := os.WriteFile(filepath.Join(root, path), []byte(content), 0644)
		require.NoError(t, err, "Failed to create file: "+path)
	}
}

// setupMockLLMClient creates a mock LLM client that can be used for testing
func setupMockLLMClient() *MockClient {
	mockClient := new(MockClient)
	mockClient.On("Generate", mock.Anything, mock.AnythingOfType("string")).Return("Generated GLANCE content", nil)
	mockClient.On("CountTokens", mock.Anything, mock.AnythingOfType("string")).Return(100, nil)
	mockClient.On("Close").Return()
	return mockClient
}

// -----------------------------------------------------------------------------
// Integration Tests: Config + Filesystem
// -----------------------------------------------------------------------------

// TestConfigFileSystemIntegration verifies that configuration settings correctly
// influence filesystem operations, particularly around the Force flag and gitignore
// functionality.
func TestConfigFileSystemIntegration(t *testing.T) {
	// Create test environment
	testDir, cleanup := setupIntegrationTest(t)
	defer cleanup()

	// 1. Test with force=false - shouldn't regenerate existing GLANCE.md
	t.Run("Respects Force flag for regeneration", func(t *testing.T) {
		// Create a GLANCE.md file with known content and timestamp
		glancePath := filepath.Join(testDir, "GLANCE.md")
		initialContent := "Initial content - should not be replaced"
		err := os.WriteFile(glancePath, []byte(initialContent), 0644)
		require.NoError(t, err, "Failed to create initial GLANCE.md file")

		// Get the initial modification time
		initialStat, err := os.Stat(glancePath)
		require.NoError(t, err)
		initialModTime := initialStat.ModTime()

		// Wait to ensure any timestamp would be different
		time.Sleep(10 * time.Millisecond)

		// Setup a configuration with force=false
		cfg := config.NewDefaultConfig().
			WithAPIKey("test-api-key").
			WithTargetDir(testDir).
			WithForce(false) // Should not overwrite existing GLANCE.md

		// Execute scanDirectories which uses the config to determine which files to scan
		dirs, ignoreChains, err := scanDirectories(cfg)
		require.NoError(t, err)
		
		// Create a mock LLM service
		mockClient := setupMockLLMClient()
		llmService, err := llm.NewService(mockClient)
		require.NoError(t, err)

		// Process directories
		results := processDirectories(dirs, ignoreChains, cfg, llmService)

		// Verify results
		assert.True(t, len(results) > 0, "Should have processed at least one directory")
		
		// Check that the root directory was processed successfully but didn't change the file
		for _, r := range results {
			if r.dir == testDir {
				assert.True(t, r.success, "Root directory should be processed successfully")
				assert.Equal(t, 0, r.attempts, "Should not have attempted to regenerate existing GLANCE.md")
			}
		}

		// Verify GLANCE.md was not modified
		currentStat, err := os.Stat(glancePath)
		require.NoError(t, err)
		assert.Equal(t, initialModTime, currentStat.ModTime(), "GLANCE.md should not have been modified")

		// Verify content was not changed
		content, err := os.ReadFile(glancePath)
		require.NoError(t, err)
		assert.Equal(t, initialContent, string(content), "GLANCE.md content should not have changed")
	})

	// 2. Test with force=true - should regenerate GLANCE.md even if it exists
	t.Run("Force flag regenerates existing files", func(t *testing.T) {
		// Create a GLANCE.md file with known content and timestamp
		glancePath := filepath.Join(testDir, "GLANCE.md")
		initialContent := "Initial content - should be replaced"
		err := os.WriteFile(glancePath, []byte(initialContent), 0644)
		require.NoError(t, err, "Failed to create initial GLANCE.md file")

		// Get the initial modification time
		initialStat, err := os.Stat(glancePath)
		require.NoError(t, err)
		initialModTime := initialStat.ModTime()

		// Wait to ensure any timestamp would be different
		time.Sleep(10 * time.Millisecond)

		// Setup a configuration with force=true
		cfg := config.NewDefaultConfig().
			WithAPIKey("test-api-key").
			WithTargetDir(testDir).
			WithForce(true) // Should overwrite existing GLANCE.md

		// Execute scanDirectories which uses the config to determine which files to scan
		dirs, ignoreChains, err := scanDirectories(cfg)
		require.NoError(t, err)
		
		// Create a mock LLM service
		mockClient := setupMockLLMClient()
		llmService, err := llm.NewService(mockClient)
		require.NoError(t, err)

		// Process directories
		results := processDirectories(dirs, ignoreChains, cfg, llmService)

		// Verify results
		assert.True(t, len(results) > 0, "Should have processed at least one directory")
		
		// Check that the root directory was processed successfully and regenerated the file
		rootProcessed := false
		for _, r := range results {
			if r.dir == testDir {
				rootProcessed = true
				assert.True(t, r.success, "Root directory should be processed successfully")
				assert.Equal(t, 1, r.attempts, "Should have attempted to regenerate GLANCE.md")
			}
		}
		assert.True(t, rootProcessed, "Root directory should have been processed")

		// Verify GLANCE.md was modified
		currentStat, err := os.Stat(glancePath)
		require.NoError(t, err)
		assert.NotEqual(t, initialModTime, currentStat.ModTime(), "GLANCE.md should have been modified")

		// Verify content was changed
		content, err := os.ReadFile(glancePath)
		require.NoError(t, err)
		assert.NotEqual(t, initialContent, string(content), "GLANCE.md content should have changed")
	})
}

// -----------------------------------------------------------------------------
// Integration Tests: Filesystem + LLM
// -----------------------------------------------------------------------------

// TestFileSystemLLMIntegration verifies the integration between the filesystem
// package and the LLM package, particularly the flow of scanning files and
// generating GLANCE.md content.
func TestFileSystemLLMIntegration(t *testing.T) {
	// Create test environment
	testDir, cleanup := setupIntegrationTest(t)
	defer cleanup()

	t.Run("File content from filesystem flows to LLM", func(t *testing.T) {
		// Setup a custom mock client that validates input data 
		validatingMockClient := new(MockClient)
		
		// Capture the prompt to analyze its contents
		var capturedPrompt string
		validatingMockClient.On("Generate", mock.Anything, mock.AnythingOfType("string")).
			Run(func(args mock.Arguments) {
				capturedPrompt = args.String(1)
			}).
			Return("Generated GLANCE content", nil)
		validatingMockClient.On("CountTokens", mock.Anything, mock.AnythingOfType("string")).Return(100, nil)
		validatingMockClient.On("Close").Return()
		
		// Setup configuration
		cfg := config.NewDefaultConfig().
			WithAPIKey("test-api-key").
			WithTargetDir(testDir).
			WithForce(true)

		// Execute scanDirectories which uses the config to determine which files to scan
		dirs, ignoreChains, err := scanDirectories(cfg)
		require.NoError(t, err)
		
		// Create LLM service with the validating mock client
		llmService, err := llm.NewService(validatingMockClient)
		require.NoError(t, err)

		// Process directories
		results := processDirectories(dirs, ignoreChains, cfg, llmService)

		// Verify results
		assert.True(t, len(results) > 0, "Should have processed at least one directory")
		
		// Verify at least one GLANCE.md was created
		glancePath := filepath.Join(testDir, "GLANCE.md")
		assert.FileExists(t, glancePath, "GLANCE.md should exist in root directory")
		
		// Verify that file content was properly passed to the LLM
		assert.Contains(t, capturedPrompt, "main.go", "Prompt should include main.go file content")
		assert.Contains(t, capturedPrompt, "README.md", "Prompt should include README.md file content")
		assert.Contains(t, capturedPrompt, "func main()", "Prompt should include source code content")
	})

	t.Run("Respects .gitignore patterns", func(t *testing.T) {
		// Add some ignored files that shouldn't be processed
		ignoreDir := filepath.Join(testDir, "build")
		err := os.MkdirAll(ignoreDir, 0755)
		require.NoError(t, err)
		
		ignoredFile := filepath.Join(ignoreDir, "ignored.txt")
		err = os.WriteFile(ignoredFile, []byte("This file should be ignored"), 0644)
		require.NoError(t, err)
		
		logFile := filepath.Join(testDir, "test.log")
		err = os.WriteFile(logFile, []byte("This log file should be ignored"), 0644)
		require.NoError(t, err)
		
		// Setup a custom mock client that validates input data 
		validatingMockClient := new(MockClient)
		
		// Capture the prompt to analyze its contents
		var capturedPrompt string
		validatingMockClient.On("Generate", mock.Anything, mock.AnythingOfType("string")).
			Run(func(args mock.Arguments) {
				capturedPrompt = args.String(1)
			}).
			Return("Generated GLANCE content", nil)
		validatingMockClient.On("CountTokens", mock.Anything, mock.AnythingOfType("string")).Return(100, nil)
		validatingMockClient.On("Close").Return()
		
		// Setup configuration
		cfg := config.NewDefaultConfig().
			WithAPIKey("test-api-key").
			WithTargetDir(testDir).
			WithForce(true)

		// Execute scanDirectories which uses the config to determine which files to scan
		dirs, ignoreChains, err := scanDirectories(cfg)
		require.NoError(t, err)
		
		// Create LLM service with the validating mock client
		llmService, err := llm.NewService(validatingMockClient)
		require.NoError(t, err)

		// Process directories
		processDirectories(dirs, ignoreChains, cfg, llmService)
		
		// Verify that ignored files were not passed to the LLM
		assert.NotContains(t, capturedPrompt, "ignored.txt", "Prompt should not include ignored.txt file content")
		assert.NotContains(t, capturedPrompt, "test.log", "Prompt should not include test.log file content")
		assert.NotContains(t, capturedPrompt, "This file should be ignored", "Prompt should not include content from ignored files")
		
		// Verify GLANCE.md wasn't created in ignored directories
		ignoredGlanceFile := filepath.Join(ignoreDir, "GLANCE.md")
		assert.NoFileExists(t, ignoredGlanceFile, "GLANCE.md should not exist in ignored directory")
	})
}

// -----------------------------------------------------------------------------
// Integration Tests: LLM + UI
// -----------------------------------------------------------------------------

// TestLLMUIIntegration verifies that the LLM operations properly report
// progress via UI components.
func TestLLMUIIntegration(t *testing.T) {
	// Save and restore log level
	originalLevel := logrus.GetLevel()
	logrus.SetLevel(logrus.DebugLevel)
	defer logrus.SetLevel(originalLevel)
	
	// Create a buffer to capture log output
	var logBuffer logCapture
	originalOutput := logrus.StandardLogger().Out
	logrus.SetOutput(&logBuffer)
	defer logrus.SetOutput(originalOutput)

	// Create test environment
	testDir, cleanup := setupIntegrationTest(t)
	defer cleanup()

	t.Run("Progress reporting during LLM generation", func(t *testing.T) {
		// Setup configuration
		cfg := config.NewDefaultConfig().
			WithAPIKey("test-api-key").
			WithTargetDir(testDir).
			WithForce(true).
			WithVerbose(true) // Enable verbose mode to see more output

		// Execute scanDirectories which uses the config to determine which files to scan
		dirs, ignoreChains, err := scanDirectories(cfg)
		require.NoError(t, err)
		
		// Create a mock LLM service 
		mockClient := setupMockLLMClient()
		llmService, err := llm.NewService(mockClient, llm.WithVerbose(true))
		require.NoError(t, err)

		// Process directories
		processDirectories(dirs, ignoreChains, cfg, llmService)
		
		// Verify that progress indicators were reported
		logs := logBuffer.String()
		assert.Contains(t, logs, "Scanning", "Log should contain scanning indicator")
		assert.Contains(t, logs, "Preparing to generate", "Log should contain generation preparation message")
		
		// Reset the log buffer for the next test
		logBuffer.Reset()
	})
}

// -----------------------------------------------------------------------------
// Integration Tests: Config + LLM
// -----------------------------------------------------------------------------

// TestConfigLLMIntegration verifies that the configuration is properly used
// by the LLM services, particularly API key handling and prompt template loading.
func TestConfigLLMIntegration(t *testing.T) {
	// Create test environment
	testDir, cleanup := setupIntegrationTest(t)
	defer cleanup()

	t.Run("Custom prompt template flows through to LLM", func(t *testing.T) {
		// Create a custom prompt template
		customPromptPath := filepath.Join(testDir, "custom_prompt.txt")
		customPromptContent := "Custom prompt for {{.Directory}} with special marker CUSTOM_PROMPT_TEST"
		err := os.WriteFile(customPromptPath, []byte(customPromptContent), 0644)
		require.NoError(t, err)
		
		// Setup a custom mock client that validates input data 
		validatingMockClient := new(MockClient)
		
		// Capture the prompt to analyze its contents
		var capturedPrompt string
		validatingMockClient.On("Generate", mock.Anything, mock.AnythingOfType("string")).
			Run(func(args mock.Arguments) {
				capturedPrompt = args.String(1)
			}).
			Return("Generated GLANCE content", nil)
		validatingMockClient.On("CountTokens", mock.Anything, mock.AnythingOfType("string")).Return(100, nil)
		validatingMockClient.On("Close").Return()
		
		// Setup configuration with custom prompt template
		cfg := config.NewDefaultConfig().
			WithAPIKey("test-api-key").
			WithTargetDir(testDir).
			WithForce(true).
			WithPromptTemplate(customPromptContent)

		// Execute scanDirectories which uses the config to determine which files to scan
		dirs, ignoreChains, err := scanDirectories(cfg)
		require.NoError(t, err)
		
		// Create LLM service
		llmService, err := llm.NewService(validatingMockClient)
		require.NoError(t, err)

		// Process directories
		processDirectories(dirs, ignoreChains, cfg, llmService)
		
		// Verify that the custom prompt template was used
		assert.Contains(t, capturedPrompt, "CUSTOM_PROMPT_TEST", 
			"Prompt should include the custom template marker")
		assert.Contains(t, capturedPrompt, testDir, 
			"Prompt should include the directory name from template substitution")
	})
}

// -----------------------------------------------------------------------------
// Log Capture Helper
// -----------------------------------------------------------------------------

// logCapture is a simple io.Writer that captures log output
type logCapture struct {
	content string
}

func (c *logCapture) Write(p []byte) (n int, err error) {
	c.content += string(p)
	return len(p), nil
}

func (c *logCapture) String() string {
	return c.content
}

func (c *logCapture) Reset() {
	c.content = ""
}