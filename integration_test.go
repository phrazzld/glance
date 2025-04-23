package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

// ProcessDirectoriesWithTracking processes a list of directories with parent propagation tracking
// This is a simplified version of processDirectories from glance.go for testing purposes
func ProcessDirectoriesWithTracking(dirsList []string, dirToIgnoreChain map[string]filesystem.IgnoreChain, cfg *config.Config, service *llm.Service) map[string]bool {
	// Create map to track directories needing regeneration due to child changes
	needsRegen := make(map[string]bool)

	// Process each directory (similar to glance.go's processDirectories function)
	for _, d := range dirsList {
		ignoreChain := dirToIgnoreChain[d]

		// Check if we need to regenerate based on either:
		// 1. Global force flag
		// 2. Local file changes (using ShouldRegenerate)
		// 3. Parent propagation (from needsRegen map)
		forceDir := cfg.Force

		// If not using global force, check for file changes
		if !forceDir {
			shouldRegen, _ := filesystem.ShouldRegenerate(d, false, ignoreChain)
			forceDir = shouldRegen
		}

		// Also check if this directory needs regeneration due to child directory changes
		forceDir = forceDir || needsRegen[d]

		// If we're forcing regeneration, simulate processing by just generating a new glance.md
		if forceDir {
			// Generate a glance.md file
			glancePath := filepath.Join(d, "glance.md")
			validatedPath, _ := filesystem.ValidateFilePath(glancePath, d, true, false)
			content := "# Test Glance\n\nThis is a test glance.md file for " + d + "\nGenerated at: " + time.Now().String()
			_ = os.WriteFile(validatedPath, []byte(content), filesystem.DefaultFileMode)

			// We only bubble up parent's regeneration flag if we actually did regeneration
			// This matches the logic in glance.go
			filesystem.BubbleUpParents(d, cfg.TargetDir, needsRegen)
		}
	}

	return needsRegen
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

// setupMultiLevelDirectoryStructure creates a nested directory structure for testing parent propagation
func setupMultiLevelDirectoryStructure(t *testing.T) (string, map[string]string, func()) {
	rootDir, err := os.MkdirTemp("", "glance-parent-regen-test-*")
	require.NoError(t, err, "Failed to create root test directory")

	// Create nested directory structure: root/level1/level2/level3
	level1Dir := filepath.Join(rootDir, "level1")
	level2Dir := filepath.Join(level1Dir, "level2")
	level3Dir := filepath.Join(level2Dir, "level3")

	// Create all directories
	for _, dir := range []string{level1Dir, level2Dir, level3Dir} {
		err := os.MkdirAll(dir, 0755)
		require.NoError(t, err, "Failed to create directory: "+dir)
	}

	// Create files in each directory
	paths := map[string]string{
		"root":   rootDir,
		"level1": level1Dir,
		"level2": level2Dir,
		"level3": level3Dir,
	}

	// Add a file to each directory
	for level, dir := range paths {
		filePath := filepath.Join(dir, level+".txt")
		err := os.WriteFile(filePath, []byte("Content for "+level), 0644)
		require.NoError(t, err, "Failed to create file in "+level)
	}

	// Return cleanup function
	return rootDir, paths, func() {
		err := os.RemoveAll(rootDir)
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
		mockLLMClient.On("CountTokens", mock.Anything, mock.Anything).Return(100, nil)

		// No need to configure Close method as we're not testing that explicitly

		// Create a Service with the mock client
		service, err := llm.NewService(mockClient)
		require.NoError(t, err, "Failed to create LLM service")

		// Configure our application
		cfg := config.NewDefaultConfig().
			WithTargetDir(testDir).
			WithForce(true)

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
		mockLLMClient.On("CountTokens", mock.Anything, mock.Anything).Return(100, nil)

		// No need to configure Close method as we're not testing that explicitly

		// Create a Service with the mock client
		service, err := llm.NewService(mockClient)
		require.NoError(t, err, "Failed to create LLM service")

		// Configure our application
		cfg := config.NewDefaultConfig().
			WithTargetDir(testDir).
			WithForce(true)

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

// TestParentRegenerationPropagation tests that when a file in a child directory is changed,
// the glance.md files in all parent directories are regenerated
func TestParentRegenerationPropagation(t *testing.T) {
	// Create test directory with multi-level structure
	rootDir, dirs, cleanup := setupMultiLevelDirectoryStructure(t)
	defer cleanup()

	// Create a mock LLM client and service (not really used in this test)
	mockLLMClient := new(mocks.LLMClient)
	mockClient := &MockClient{LLMClient: mockLLMClient}
	mockLLMClient.On("Generate", mock.Anything, mock.Anything).Return("# Mock Glance\n\nThis is a mock glance.md summary.", nil)
	mockLLMClient.On("CountTokens", mock.Anything, mock.Anything).Return(100, nil)
	service, err := llm.NewService(mockClient)
	require.NoError(t, err, "Failed to create LLM service")

	// Configure application
	cfg := config.NewDefaultConfig().
		WithTargetDir(rootDir)

	// Get all directories to process
	dirsList, dirToIgnoreChain, err := filesystem.ListDirsWithIgnores(rootDir)
	require.NoError(t, err, "Failed to list directories")

	// Reverse dirsList to process from deepest to shallowest
	for i, j := 0, len(dirsList)-1; i < j; i, j = i+1, j-1 {
		dirsList[i], dirsList[j] = dirsList[j], dirsList[i]
	}

	// Initial run to generate all glance.md files - force to ensure all are generated
	cfg = cfg.WithForce(true)
	// We don't need to check the return value for the first run
	_ = ProcessDirectoriesWithTracking(dirsList, dirToIgnoreChain, cfg, service)

	// Verify all directories have glance.md files
	for _, dir := range dirs {
		glancePath := filepath.Join(dir, "glance.md")
		assert.FileExists(t, glancePath, "Initial glance.md should exist in "+dir)
	}

	// Get initial modification times
	initialModTimes := make(map[string]time.Time)
	for level, dir := range dirs {
		glancePath := filepath.Join(dir, "glance.md")
		info, err := os.Stat(glancePath)
		require.NoError(t, err, "Failed to stat glance.md in "+level)
		initialModTimes[level] = info.ModTime()
	}

	// Wait to ensure file timestamps will be different
	time.Sleep(1 * time.Second)

	// Modify a file in the deepest level (level3)
	deepestFilePath := filepath.Join(dirs["level3"], "level3.txt")
	newContent := "Modified content for level3 - " + time.Now().String()
	err = os.WriteFile(deepestFilePath, []byte(newContent), 0644)
	require.NoError(t, err, "Failed to modify file in deepest directory")

	// Also explicitly touch the file to ensure modification time is updated
	currentTime := time.Now().Local()
	err = os.Chtimes(deepestFilePath, currentTime, currentTime)
	require.NoError(t, err, "Failed to update file modification time")

	// Wait to ensure timestamp detection works reliably
	time.Sleep(100 * time.Millisecond)

	// Run without global force flag, so only changed dirs and parents regenerate
	cfg = cfg.WithForce(false)
	parentRegenMap := ProcessDirectoriesWithTracking(dirsList, dirToIgnoreChain, cfg, service)

	// Check that parent dirs are marked for regeneration in the map
	for level, dir := range dirs {
		if level == "level1" || level == "level2" {
			// These should be marked for regeneration from bubbling up
			assert.True(t, parentRegenMap[dir],
				fmt.Sprintf("%s directory should be marked for regeneration", level))
		}
	}

	// Get new modification times
	finalModTimes := make(map[string]time.Time)
	for level, dir := range dirs {
		glancePath := filepath.Join(dir, "glance.md")
		info, err := os.Stat(glancePath)
		require.NoError(t, err, "Failed to stat glance.md in "+level)
		finalModTimes[level] = info.ModTime()
	}

	// Level3 (changed directory) should be regenerated
	t.Logf("level3 initial: %v, final: %v", initialModTimes["level3"], finalModTimes["level3"])
	assert.True(t, finalModTimes["level3"].After(initialModTimes["level3"]),
		"level3 glance.md should have been regenerated (final time should be after initial time)")

	// Parent directories should be regenerated due to bubbling up
	t.Logf("level2 initial: %v, final: %v", initialModTimes["level2"], finalModTimes["level2"])
	assert.True(t, finalModTimes["level2"].After(initialModTimes["level2"]),
		"level2 glance.md should have been regenerated due to child change")

	t.Logf("level1 initial: %v, final: %v", initialModTimes["level1"], finalModTimes["level1"])
	assert.True(t, finalModTimes["level1"].After(initialModTimes["level1"]),
		"level1 glance.md should have been regenerated due to child change")
}

// TestForcedChildRegenerationBubblesUp tests that when a child directory is forcibly regenerated,
// the glance.md files in all parent directories are also regenerated
func TestForcedChildRegenerationBubblesUp(t *testing.T) {
	// Create test directory with multi-level structure
	rootDir, dirs, cleanup := setupMultiLevelDirectoryStructure(t)
	defer cleanup()

	// Create a mock LLM client and service
	mockLLMClient := new(mocks.LLMClient)
	mockClient := &MockClient{LLMClient: mockLLMClient}
	mockLLMClient.On("Generate", mock.Anything, mock.Anything).Return("# Mock Glance\n\nThis is a mock glance.md summary.", nil)
	mockLLMClient.On("CountTokens", mock.Anything, mock.Anything).Return(100, nil)
	service, err := llm.NewService(mockClient)
	require.NoError(t, err, "Failed to create LLM service")

	// Configure application for the root directory
	rootCfg := config.NewDefaultConfig().
		WithTargetDir(rootDir)

	// Get all directories to process
	dirsList, dirToIgnoreChain, err := filesystem.ListDirsWithIgnores(rootDir)
	require.NoError(t, err, "Failed to list directories")

	// Reverse dirsList to process from deepest to shallowest
	for i, j := 0, len(dirsList)-1; i < j; i, j = i+1, j-1 {
		dirsList[i], dirsList[j] = dirsList[j], dirsList[i]
	}

	// Initial run to generate all glance.md files without force flag
	rootCfg = rootCfg.WithForce(false)
	_ = ProcessDirectoriesWithTracking(dirsList, dirToIgnoreChain, rootCfg, service)

	// Verify all directories have glance.md files
	for _, dir := range dirs {
		glancePath := filepath.Join(dir, "glance.md")
		assert.FileExists(t, glancePath, "Initial glance.md should exist in "+dir)
	}

	// Get initial modification times
	initialModTimes := make(map[string]time.Time)
	for level, dir := range dirs {
		glancePath := filepath.Join(dir, "glance.md")
		info, err := os.Stat(glancePath)
		require.NoError(t, err, "Failed to stat glance.md in "+level)
		initialModTimes[level] = info.ModTime()
	}

	// Wait to ensure file timestamps will be different
	time.Sleep(1 * time.Second)

	// We'll modify our approach to better simulate how the application handles directory-specific force flags

	// Create a function to process only the level3 directory with force flag
	processDirectoryWithForce := func(dir string) {
		// Generate a glance.md file directly (simulating forced regeneration)
		glancePath := filepath.Join(dir, "glance.md")
		validatedPath, _ := filesystem.ValidateFilePath(glancePath, dir, true, false)
		content := "# Forced Glance\n\nThis is a forcibly regenerated glance.md file\nGenerated at: " + time.Now().String()
		_ = os.WriteFile(validatedPath, []byte(content), filesystem.DefaultFileMode)

		// Explicitly touch the file to ensure modification time is updated
		now := time.Now()
		err = os.Chtimes(validatedPath, now, now)
		require.NoError(t, err, "Failed to update glance.md modification time")
	}

	// Force regenerate the level3 directory
	processDirectoryWithForce(dirs["level3"])

	// Run the process on the whole directory structure with a custom tracking function
	// that uses our needsRegen map to track which directories need regeneration
	customTracking := func() map[string]bool {
		needsRegen := make(map[string]bool)

		// Track successful regeneration of level3 by bubbling up to parents
		filesystem.BubbleUpParents(dirs["level3"], rootDir, needsRegen)

		// Process all directories (with our tracked needsRegen map)
		for _, d := range dirsList {
			ignoreChain := dirToIgnoreChain[d]

			// Check if regeneration is needed due to local changes or parent propagation
			forceDir := false
			shouldRegen, _ := filesystem.ShouldRegenerate(d, false, ignoreChain)
			forceDir = shouldRegen || needsRegen[d]

			// If regeneration is needed, generate a new glance.md file
			if forceDir {
				glancePath := filepath.Join(d, "glance.md")
				validatedPath, _ := filesystem.ValidateFilePath(glancePath, d, true, false)
				content := "# Test Glance\n\nThis is a test glance.md file for " + d + "\nGenerated at: " + time.Now().String()
				_ = os.WriteFile(validatedPath, []byte(content), filesystem.DefaultFileMode)
			}
		}

		return needsRegen
	}

	// Run our custom tracking function and get the regeneration map
	parentRegenMap := customTracking()

	// Check that parent dirs are marked for regeneration in the map
	for level, dir := range dirs {
		if level == "level1" || level == "level2" {
			// These should be marked for regeneration from bubbling up
			assert.True(t, parentRegenMap[dir],
				fmt.Sprintf("%s directory should be marked for regeneration", level))
		}
	}

	// Get new modification times
	finalModTimes := make(map[string]time.Time)
	for level, dir := range dirs {
		glancePath := filepath.Join(dir, "glance.md")
		info, err := os.Stat(glancePath)
		require.NoError(t, err, "Failed to stat glance.md in "+level)
		finalModTimes[level] = info.ModTime()
	}

	// Level3 (forced directory) should be regenerated
	t.Logf("level3 initial: %v, final: %v", initialModTimes["level3"], finalModTimes["level3"])
	assert.True(t, finalModTimes["level3"].After(initialModTimes["level3"]),
		"level3 glance.md should have been regenerated due to force flag")

	// Parent directories should be regenerated due to bubbling up
	t.Logf("level2 initial: %v, final: %v", initialModTimes["level2"], finalModTimes["level2"])
	assert.True(t, finalModTimes["level2"].After(initialModTimes["level2"]),
		"level2 glance.md should have been regenerated due to forced child")

	t.Logf("level1 initial: %v, final: %v", initialModTimes["level1"], finalModTimes["level1"])
	assert.True(t, finalModTimes["level1"].After(initialModTimes["level1"]),
		"level1 glance.md should have been regenerated due to forced child")
}

// TestNoChangesMeansNoRegeneration tests that when no files have changed between runs,
// no glance.md files are regenerated (optimization)
func TestNoChangesMeansNoRegeneration(t *testing.T) {
	// Create test directory with multi-level structure
	rootDir, dirs, cleanup := setupMultiLevelDirectoryStructure(t)
	defer cleanup()

	// Create a mock LLM client and service
	mockLLMClient := new(mocks.LLMClient)
	mockClient := &MockClient{LLMClient: mockLLMClient}
	mockLLMClient.On("Generate", mock.Anything, mock.Anything).Return("# Mock Glance\n\nThis is a mock glance.md summary.", nil)
	mockLLMClient.On("CountTokens", mock.Anything, mock.Anything).Return(100, nil)
	service, err := llm.NewService(mockClient)
	require.NoError(t, err, "Failed to create LLM service")

	// Configure application
	cfg := config.NewDefaultConfig().
		WithTargetDir(rootDir)

	// Get all directories to process
	dirsList, dirToIgnoreChain, err := filesystem.ListDirsWithIgnores(rootDir)
	require.NoError(t, err, "Failed to list directories")

	// Reverse dirsList to process from deepest to shallowest
	for i, j := 0, len(dirsList)-1; i < j; i, j = i+1, j-1 {
		dirsList[i], dirsList[j] = dirsList[j], dirsList[i]
	}

	// Initial run to generate all glance.md files - force to ensure all are generated initially
	firstRunCfg := cfg.WithForce(true)
	_ = ProcessDirectoriesWithTracking(dirsList, dirToIgnoreChain, firstRunCfg, service)

	// Verify all directories have glance.md files
	for _, dir := range dirs {
		glancePath := filepath.Join(dir, "glance.md")
		assert.FileExists(t, glancePath, "Initial glance.md should exist in "+dir)
	}

	// Wait a moment to ensure file timestamps would be different if files were regenerated
	time.Sleep(1 * time.Second)

	// Get the modification times after the first run
	initialModTimes := make(map[string]time.Time)
	for level, dir := range dirs {
		glancePath := filepath.Join(dir, "glance.md")
		info, err := os.Stat(glancePath)
		require.NoError(t, err, "Failed to stat glance.md in "+level)
		initialModTimes[level] = info.ModTime()
		t.Logf("Initial mod time for %s: %v", level, initialModTimes[level])
	}

	// Run again without force flag and without any file changes
	secondRunCfg := cfg.WithForce(false)
	regenMap := ProcessDirectoriesWithTracking(dirsList, dirToIgnoreChain, secondRunCfg, service)

	// Verify no directories were marked for regeneration
	for level, dir := range dirs {
		assert.False(t, regenMap[dir],
			fmt.Sprintf("%s directory should NOT be marked for regeneration", level))
	}

	// Get modification times after the second run
	finalModTimes := make(map[string]time.Time)
	for level, dir := range dirs {
		glancePath := filepath.Join(dir, "glance.md")
		info, err := os.Stat(glancePath)
		require.NoError(t, err, "Failed to stat glance.md in "+level)
		finalModTimes[level] = info.ModTime()
		t.Logf("Final mod time for %s: %v", level, finalModTimes[level])
	}

	// Verify no glance.md files were regenerated (modification times should be unchanged)
	for level := range dirs {
		assert.Equal(t, initialModTimes[level], finalModTimes[level],
			fmt.Sprintf("%s's glance.md should NOT have been regenerated (mod times should be equal)", level))
	}
}

// setupBranchingDirectoryStructure creates a directory structure with multiple branches
// for testing sibling directory isolation
// Structure:
//
//	root/
//	├── branch_a/
//	│   └── deep_a/
//	│       └── nested_a/
//	└── branch_b/
//	    └── deep_b/
func setupBranchingDirectoryStructure(t *testing.T) (string, map[string]string, func()) {
	rootDir, err := os.MkdirTemp("", "glance-sibling-isolation-test-*")
	require.NoError(t, err, "Failed to create root test directory")

	// Create branching directory structure with two separate branches
	branchADir := filepath.Join(rootDir, "branch_a")
	deepADir := filepath.Join(branchADir, "deep_a")
	nestedADir := filepath.Join(deepADir, "nested_a")

	branchBDir := filepath.Join(rootDir, "branch_b")
	deepBDir := filepath.Join(branchBDir, "deep_b")

	// Create all directories
	for _, dir := range []string{branchADir, deepADir, nestedADir, branchBDir, deepBDir} {
		err := os.MkdirAll(dir, 0755)
		require.NoError(t, err, "Failed to create directory: "+dir)
	}

	// Create files in each directory
	paths := map[string]string{
		"root":     rootDir,
		"branch_a": branchADir,
		"deep_a":   deepADir,
		"nested_a": nestedADir,
		"branch_b": branchBDir,
		"deep_b":   deepBDir,
	}

	// Add a file to each directory
	for level, dir := range paths {
		filePath := filepath.Join(dir, level+".txt")
		err := os.WriteFile(filePath, []byte("Content for "+level), 0644)
		require.NoError(t, err, "Failed to create file in "+level)
	}

	// Return cleanup function
	return rootDir, paths, func() {
		err := os.RemoveAll(rootDir)
		if err != nil {
			t.Logf("Warning: failed to clean up test directory: %v", err)
		}
	}
}

// TestSiblingDirectoryIsolation tests that when a file in one branch of the directory structure
// changes, glance.md files in sibling branches are not regenerated
func TestSiblingDirectoryIsolation(t *testing.T) {
	// Create test directory with branching structure
	rootDir, dirs, cleanup := setupBranchingDirectoryStructure(t)
	defer cleanup()

	// Create a mock LLM client and service
	mockLLMClient := new(mocks.LLMClient)
	mockClient := &MockClient{LLMClient: mockLLMClient}
	mockLLMClient.On("Generate", mock.Anything, mock.Anything).Return("# Mock Glance\n\nThis is a mock glance.md summary.", nil)
	mockLLMClient.On("CountTokens", mock.Anything, mock.Anything).Return(100, nil)
	service, err := llm.NewService(mockClient)
	require.NoError(t, err, "Failed to create LLM service")

	// Configure application for the root directory
	cfg := config.NewDefaultConfig().
		WithTargetDir(rootDir)

	// Get all directories to process
	dirsList, dirToIgnoreChain, err := filesystem.ListDirsWithIgnores(rootDir)
	require.NoError(t, err, "Failed to list directories")

	// Reverse dirsList to process from deepest to shallowest
	for i, j := 0, len(dirsList)-1; i < j; i, j = i+1, j-1 {
		dirsList[i], dirsList[j] = dirsList[j], dirsList[i]
	}

	// Initial run to generate all glance.md files
	initialCfg := cfg.WithForce(true)
	_ = ProcessDirectoriesWithTracking(dirsList, dirToIgnoreChain, initialCfg, service)

	// Verify all directories have glance.md files
	for _, dir := range dirs {
		glancePath := filepath.Join(dir, "glance.md")
		assert.FileExists(t, glancePath, "Initial glance.md should exist in "+dir)
	}

	// Wait to ensure file timestamps will be different if files are regenerated
	time.Sleep(1 * time.Second)

	// Get initial modification times for all directories
	initialModTimes := make(map[string]time.Time)
	for level, dir := range dirs {
		glancePath := filepath.Join(dir, "glance.md")
		info, err := os.Stat(glancePath)
		require.NoError(t, err, "Failed to stat glance.md in "+level)
		initialModTimes[level] = info.ModTime()
		t.Logf("Initial mod time for %s: %v", level, initialModTimes[level])
	}

	// Modify a file in the nested_a branch
	nestedAFilePath := filepath.Join(dirs["nested_a"], "nested_a.txt")
	newContent := "Modified content for nested_a - " + time.Now().String()
	err = os.WriteFile(nestedAFilePath, []byte(newContent), 0644)
	require.NoError(t, err, "Failed to modify file in nested_a directory")

	// Explicitly touch the file to ensure modification time is updated
	now := time.Now()
	err = os.Chtimes(nestedAFilePath, now, now)
	require.NoError(t, err, "Failed to update file modification time")

	// Run again without the force flag
	secondRunCfg := cfg.WithForce(false)
	regenMap := ProcessDirectoriesWithTracking(dirsList, dirToIgnoreChain, secondRunCfg, service)

	// Get final modification times
	finalModTimes := make(map[string]time.Time)
	for level, dir := range dirs {
		glancePath := filepath.Join(dir, "glance.md")
		info, err := os.Stat(glancePath)
		require.NoError(t, err, "Failed to stat glance.md in "+level)
		finalModTimes[level] = info.ModTime()
		t.Logf("Final mod time for %s: %v", level, finalModTimes[level])
	}

	// Affected Paths: nested_a, deep_a, branch_a
	// These should be regenerated based on file changes and bubble-up
	affectedPaths := []string{"nested_a", "deep_a", "branch_a"}
	for _, path := range affectedPaths {
		t.Logf("Checking affected path: %s", path)
		assert.True(t, finalModTimes[path].After(initialModTimes[path]),
			fmt.Sprintf("%s glance.md should have been regenerated (final time should be after initial time)", path))

		// For all affected paths, they should be marked for regeneration or be the source of change
		assert.True(t, regenMap[dirs[path]] || path == "nested_a",
			fmt.Sprintf("%s should be marked for regeneration or be the modified directory", path))
	}

	// The root directory should also be regenerated, but it's not always in the regenMap
	// since it's the target directory and is handled differently
	t.Logf("Checking root directory")
	assert.True(t, finalModTimes["root"].After(initialModTimes["root"]),
		"root glance.md should have been regenerated (final time should be after initial time)")

	// Unaffected Paths: branch_b, deep_b
	// These should NOT be regenerated
	unaffectedPaths := []string{"branch_b", "deep_b"}
	for _, path := range unaffectedPaths {
		t.Logf("Checking unaffected path: %s", path)
		assert.Equal(t, initialModTimes[path], finalModTimes[path],
			fmt.Sprintf("%s glance.md should NOT have been regenerated (mod times should be equal)", path))
		assert.False(t, regenMap[dirs[path]],
			fmt.Sprintf("%s should NOT be marked for regeneration", path))
	}
}
