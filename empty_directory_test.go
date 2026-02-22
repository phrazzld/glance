package main

import (
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

// TestEmptyDirectorySkipsLLM verifies that processDirectory writes a minimal stub for empty
// directories and never calls the LLM — preventing hallucination from directory path names.
func TestEmptyDirectorySkipsLLM(t *testing.T) {
	t.Run("empty directory writes stub without calling LLM", func(t *testing.T) {
		// Arrange: empty temp directory
		dir, err := os.MkdirTemp("", "glance-empty-dir-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(dir)

		// Mock client — Generate must NOT be called
		mockLLMClient := new(mocks.LLMClient)
		mockClient := &MockClient{LLMClient: mockLLMClient}

		service, err := llm.NewService(mockClient)
		require.NoError(t, err)

		cfg := config.NewDefaultConfig().WithMaxFileBytes(1 << 20)
		ignoreChain := filesystem.IgnoreChain{}

		// Act
		r := processDirectory(dir, true, ignoreChain, cfg, service)

		// Assert: success, no LLM call
		assert.True(t, r.success, "processDirectory should succeed on empty directory")
		assert.NoError(t, r.err)
		mockLLMClient.AssertNotCalled(t, "Generate", mock.Anything, mock.Anything)

		// Assert: stub file written with minimal honest content
		glancePath := filepath.Join(dir, filesystem.GlanceFilename)
		require.FileExists(t, glancePath)

		content, err := os.ReadFile(glancePath)
		require.NoError(t, err)
		body := string(content)

		assert.True(t, strings.Contains(body, "Empty directory"),
			"stub should contain 'Empty directory', got: %q", body)
	})

	t.Run("directory with only binary files writes non-empty stub without calling LLM", func(t *testing.T) {
		// Arrange: directory with a binary file that GatherLocalFiles filters out
		dir, err := os.MkdirTemp("", "glance-binary-dir-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(dir)

		// Write a binary-looking file (null bytes make it non-text)
		binaryContent := []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE}
		require.NoError(t, os.WriteFile(filepath.Join(dir, "data.bin"), binaryContent, 0600))

		mockLLMClient := new(mocks.LLMClient)
		mockClient := &MockClient{LLMClient: mockLLMClient}

		service, err := llm.NewService(mockClient)
		require.NoError(t, err)

		cfg := config.NewDefaultConfig().WithMaxFileBytes(1 << 20)
		ignoreChain := filesystem.IgnoreChain{}

		// Act
		r := processDirectory(dir, true, ignoreChain, cfg, service)

		// Assert: success, no LLM call
		assert.True(t, r.success)
		assert.NoError(t, r.err)
		mockLLMClient.AssertNotCalled(t, "Generate", mock.Anything, mock.Anything)

		// Assert: stub exists but does NOT say "Empty directory" (dir is not empty)
		glancePath := filepath.Join(dir, filesystem.GlanceFilename)
		require.FileExists(t, glancePath)
		content, err := os.ReadFile(glancePath)
		require.NoError(t, err)
		body := string(content)

		assert.False(t, strings.Contains(body, "Empty directory"),
			"stub should NOT say 'Empty directory' for a dir with binary files, got: %q", body)
		assert.True(t, strings.Contains(body, "No analyzable text content"),
			"stub should say 'No analyzable text content', got: %q", body)
	})

	t.Run("directory with only hidden files writes non-empty stub without calling LLM", func(t *testing.T) {
		// Arrange: directory with a hidden file that GatherLocalFiles skips
		dir, err := os.MkdirTemp("", "glance-hidden-dir-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(dir)

		require.NoError(t, os.WriteFile(filepath.Join(dir, ".hidden"), []byte("secret"), 0600))

		mockLLMClient := new(mocks.LLMClient)
		mockClient := &MockClient{LLMClient: mockLLMClient}

		service, err := llm.NewService(mockClient)
		require.NoError(t, err)

		cfg := config.NewDefaultConfig().WithMaxFileBytes(1 << 20)
		ignoreChain := filesystem.IgnoreChain{}

		// Act
		r := processDirectory(dir, true, ignoreChain, cfg, service)

		// Assert
		assert.True(t, r.success)
		assert.NoError(t, r.err)
		mockLLMClient.AssertNotCalled(t, "Generate", mock.Anything, mock.Anything)

		glancePath := filepath.Join(dir, filesystem.GlanceFilename)
		require.FileExists(t, glancePath)
		content, err := os.ReadFile(glancePath)
		require.NoError(t, err)
		body := string(content)

		assert.False(t, strings.Contains(body, "Empty directory"),
			"stub should NOT say 'Empty directory' for a dir with hidden files, got: %q", body)
		assert.True(t, strings.Contains(body, "No analyzable text content"),
			"stub should say 'No analyzable text content', got: %q", body)
	})

	t.Run("directory with only subglances still calls LLM", func(t *testing.T) {
		// Arrange: directory with no local files but a child has a glance summary
		dir, err := os.MkdirTemp("", "glance-subglance-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(dir)

		// Create a subdirectory with a pre-existing .glance.md
		subdir := filepath.Join(dir, "pkg")
		require.NoError(t, os.Mkdir(subdir, 0755))
		subGlancePath := filepath.Join(subdir, filesystem.GlanceFilename)
		require.NoError(t, os.WriteFile(subGlancePath, []byte("# pkg\n\nPackage code.\n"), filesystem.DefaultFileMode))

		mockLLMClient := new(mocks.LLMClient)
		mockClient := &MockClient{LLMClient: mockLLMClient}
		mockLLMClient.On("Generate", mock.Anything, mock.Anything).Return("# summary\n", nil)
		mockLLMClient.On("CountTokens", mock.Anything, mock.Anything).Return(50, nil)

		service, err := llm.NewService(mockClient)
		require.NoError(t, err)

		cfg := config.NewDefaultConfig().WithMaxFileBytes(1 << 20)
		ignoreChain := filesystem.IgnoreChain{}

		// Act
		r := processDirectory(dir, true, ignoreChain, cfg, service)

		// Assert: LLM WAS called because there is child context
		assert.True(t, r.success)
		mockLLMClient.AssertCalled(t, "Generate", mock.Anything, mock.Anything)
	})
}
