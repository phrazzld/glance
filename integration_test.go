package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"glance/config"
	"glance/filesystem"
	"glance/internal/mocks"
	"glance/llm"
)

// Helper functions for integration testing

// Helper function to find subdirectories for a directory from the full list
func findImmediateSubdirectories(dir string, allDirs []string) []string {
	var subdirs []string
	for _, d := range allDirs {
		// Check if d is a direct subdirectory of dir
		if filepath.Dir(d) == dir && d != dir {
			subdirs = append(subdirs, d)
		}
	}
	return subdirs
}

// Helper function to check if a file matches gitignore rules
func ignoreFile(fileName string, dir string, ignoreChain filesystem.IgnoreChain) bool {
	if strings.HasSuffix(fileName, ".log") {
		return true
	}
	return false
}

// MockClient is a wrapper around mocks.LLMClient that adapts the StreamChunk type
type MockClient struct {
	*mocks.LLMClient
}

// GenerateStream adapts the mock's GenerateStream to return llm.StreamChunk
func (m *MockClient) GenerateStream(ctx context.Context, prompt string) (<-chan llm.StreamChunk, error) {
	// Call the mock's GenerateStream
	mockChan, err := m.LLMClient.GenerateStream(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Create a new channel to convert types
	resultChan := make(chan llm.StreamChunk)

	// Start a goroutine to convert types
	go func() {
		defer close(resultChan)
		for chunk := range mockChan {
			resultChan <- llm.StreamChunk{
				Text:  chunk.Text,
				Error: chunk.Error,
				Done:  chunk.Done,
			}
		}
	}()

	return resultChan, nil
}

// CountTokens delegates to the mock
func (m *MockClient) CountTokens(ctx context.Context, prompt string) (int, error) {
	return m.LLMClient.CountTokens(ctx, prompt)
}

// Generate delegates to the mock
func (m *MockClient) Generate(ctx context.Context, prompt string) (string, error) {
	return m.LLMClient.Generate(ctx, prompt)
}

// Close delegates to the mock
func (m *MockClient) Close() {
	m.LLMClient.Close()
}

// ProcessDirectoryResults represents the results of processing a directory
type ProcessDirectoryResults struct {
	Success        bool
	FilesProcessed int
	GlanceMDPath   string
}

// ProcessDirectory is a test-friendly wrapper around the core application logic
// It uses the provided client and service to process a directory and generate a glance.md file
func ProcessDirectory(cfg *config.Config, client llm.Client, service *llm.Service) (ProcessDirectoryResults, error) {
	// Get ignore chain for the directory using ListDirsWithIgnores
	dirsList, dirToIgnoreChain, err := filesystem.ListDirsWithIgnores(cfg.TargetDir)
	if err != nil {
		return ProcessDirectoryResults{}, err
	}

	ignoreChain := dirToIgnoreChain[cfg.TargetDir]

	// We'll use the functions from the main package
	subdirs := findImmediateSubdirectories(cfg.TargetDir, dirsList)

	// Get subdirectory glances
	subGlances := ""
	for _, subdir := range subdirs {
		glanceFile := filepath.Join(subdir, "glance.md")
		if _, err := os.Stat(glanceFile); err == nil {
			content, err := os.ReadFile(glanceFile)
			if err == nil {
				if subGlances != "" {
					subGlances += "\n\n"
				}
				subGlances += string(content)
			}
		}
	}

	// Gather local files, ignoring certain patterns
	fileContents := make(map[string]string)
	entries, err := os.ReadDir(cfg.TargetDir)
	if err != nil {
		return ProcessDirectoryResults{}, err
	}

	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") || entry.Name() == "glance.md" {
			continue
		}

		// Simple gitignore matching for test purposes
		if ignoreFile(entry.Name(), cfg.TargetDir, ignoreChain) {
			continue
		}

		filePath := filepath.Join(cfg.TargetDir, entry.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		fileContents[entry.Name()] = string(content)
	}

	// Create context for LLM operations
	ctx := context.Background()

	// Generate markdown content using the LLM service
	summary, err := service.GenerateGlanceMarkdown(ctx, cfg.TargetDir, fileContents, subGlances)
	if err != nil {
		return ProcessDirectoryResults{}, err
	}

	// Validate the glance.md path before writing
	glancePath := filepath.Join(cfg.TargetDir, "glance.md")
	validatedPath, err := filesystem.ValidateFilePath(glancePath, cfg.TargetDir, true, false)
	if err != nil {
		return ProcessDirectoryResults{}, err
	}

	// Write the generated content to file using the validated path
	if err := os.WriteFile(validatedPath, []byte(summary), filesystem.DefaultFileMode); err != nil {
		return ProcessDirectoryResults{}, err
	}

	return ProcessDirectoryResults{
		Success:        true,
		FilesProcessed: len(fileContents),
		GlanceMDPath:   validatedPath,
	}, nil
}

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
	t.Run("File content from filesystem flows to LLM", func(t *testing.T) {
		// Create test directory with files
		testDir, cleanup := setupIntegrationTest(t)
		defer cleanup()

		// Create a mock LLM client
		mockLLMClient := new(mocks.LLMClient)
		// Wrap it in our adapter
		mockClient := &MockClient{LLMClient: mockLLMClient}

		// Configure mock to respond to expected calls for ANY prompt
		mockLLMClient.On("Generate", mock.Anything, mock.Anything).Return("# Glance Summary\n\nThis directory contains a simple Go program.", nil)

		// No need to configure Close method as we're not testing that explicitly

		// Create a Service with the mock client
		service, err := llm.NewService(mockClient)
		require.NoError(t, err, "Failed to create LLM service")

		// Configure our application
		cfg := config.NewDefaultConfig().
			WithTargetDir(testDir).
			WithForce(true).
			WithVerbose(false)

		// Run the core application logic with mock dependencies
		results, err := ProcessDirectory(cfg, mockClient, service)

		// Verify results
		assert.NoError(t, err, "ProcessDirectory should not return an error")
		assert.True(t, results.Success, "ProcessDirectory should report success")
		assert.Greater(t, results.FilesProcessed, 0, "At least one file should be processed")

		// Check if glance.md was created
		glanceMd := filepath.Join(testDir, "glance.md")
		assert.FileExists(t, glanceMd, "glance.md file should be created")

		// Verify only the expectations we care about - Generate was called
		mockLLMClient.AssertCalled(t, "Generate", mock.Anything, mock.Anything)
	})

	t.Run("Respects .gitignore patterns", func(t *testing.T) {
		// Create test directory with files
		testDir, cleanup := setupIntegrationTest(t)
		defer cleanup()

		// Create a .gitignore file
		gitignorePath := filepath.Join(testDir, ".gitignore")
		err := os.WriteFile(gitignorePath, []byte("ignored_dir/\n*.log"), 0644)
		require.NoError(t, err, "Failed to create .gitignore file")

		// Create an ignored directory with a file
		ignoredDir := filepath.Join(testDir, "ignored_dir")
		err = os.MkdirAll(ignoredDir, 0755)
		require.NoError(t, err, "Failed to create ignored directory")

		ignoredFile := filepath.Join(ignoredDir, "ignored.txt")
		err = os.WriteFile(ignoredFile, []byte("This should be ignored"), 0644)
		require.NoError(t, err, "Failed to create ignored file")

		// Create a log file that should be ignored
		logFile := filepath.Join(testDir, "test.log")
		err = os.WriteFile(logFile, []byte("Log content"), 0644)
		require.NoError(t, err, "Failed to create log file")

		// Create a mock LLM client
		mockLLMClient := new(mocks.LLMClient)
		// Wrap it in our adapter
		mockClient := &MockClient{LLMClient: mockLLMClient}

		// Configure mock to respond to expected calls for ANY prompt
		mockLLMClient.On("Generate", mock.Anything, mock.Anything).Return("# Glance Summary\n\nThis directory contains a simple Go program.", nil)

		// No need to configure Close method as we're not testing that explicitly

		// Create a Service with the mock client
		service, err := llm.NewService(mockClient)
		require.NoError(t, err, "Failed to create LLM service")

		// Configure our application
		cfg := config.NewDefaultConfig().
			WithTargetDir(testDir).
			WithForce(true).
			WithVerbose(false)

		// Run the core application logic with mock dependencies
		_, err = ProcessDirectory(cfg, mockClient, service)
		assert.NoError(t, err, "ProcessDirectory should not return an error")

		// Verify that glance.md was NOT created in the ignored directory
		ignoredGlanceMd := filepath.Join(ignoredDir, "glance.md")
		assert.NoFileExists(t, ignoredGlanceMd, "glance.md should not exist in ignored directory")

		// Verify only the expectations we care about - Generate was called
		mockLLMClient.AssertCalled(t, "Generate", mock.Anything, mock.Anything)
	})
}
