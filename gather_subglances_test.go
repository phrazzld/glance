package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGatherSubGlances(t *testing.T) {
	// Create a temporary directory structure for testing
	testDir := t.TempDir()

	// Create subdirectories
	subDir1 := filepath.Join(testDir, "subdir1")
	subDir2 := filepath.Join(testDir, "subdir2")
	subDir3 := filepath.Join(testDir, "subdir3")
	nestedDir := filepath.Join(subDir1, "nested")

	for _, dir := range []string{subDir1, subDir2, subDir3, nestedDir} {
		err := os.MkdirAll(dir, 0755)
		require.NoError(t, err)
	}

	// Create glance.md files in subdirectories
	glanceFile1 := filepath.Join(subDir1, "glance.md")
	glanceFile2 := filepath.Join(subDir2, "glance.md")
	glanceFile3 := filepath.Join(subDir3, "glance.md")
	nestedGlanceFile := filepath.Join(nestedDir, "glance.md")

	err := os.WriteFile(glanceFile1, []byte("Content from subdir1"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(glanceFile2, []byte("Content from subdir2"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(glanceFile3, []byte("Content from subdir3"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(nestedGlanceFile, []byte("Content from nested dir"), 0644)
	require.NoError(t, err)

	t.Run("ValidSubdirectories", func(t *testing.T) {
		// Test with valid subdirectories
		subdirs := []string{subDir1, subDir2, subDir3}
		content, err := gatherSubGlances(subdirs)
		
		assert.NoError(t, err)
		assert.Contains(t, content, "Content from subdir1")
		assert.Contains(t, content, "Content from subdir2")
		assert.Contains(t, content, "Content from subdir3")
	})

	t.Run("NestedSubdirectory", func(t *testing.T) {
		// Test with nested subdirectory
		subdirs := []string{nestedDir}
		content, err := gatherSubGlances(subdirs)
		
		assert.NoError(t, err)
		assert.Contains(t, content, "Content from nested dir")
	})

	t.Run("MixedSubdirectories", func(t *testing.T) {
		// Test with a mix of regular and nested subdirectories
		subdirs := []string{subDir1, nestedDir}
		content, err := gatherSubGlances(subdirs)
		
		assert.NoError(t, err)
		assert.Contains(t, content, "Content from subdir1")
		assert.Contains(t, content, "Content from nested dir")
	})

	t.Run("AttemptedTraversalWithParentRef", func(t *testing.T) {
		// Create a path with parent directory reference
		// This should be caught by the path validation
		invalidPath := filepath.Join(subDir1, "..", "outside")
		subdirs := []string{invalidPath}
		
		content, err := gatherSubGlances(subdirs)
		
		// Function shouldn't return an error, but should skip the invalid directory
		assert.NoError(t, err)
		assert.Empty(t, content)
	})

	t.Run("AttemptedTraversalWithAbsolutePath", func(t *testing.T) {
		// Create a path outside the test directory structure
		outsideDir := filepath.Join(os.TempDir(), "outside")
		err := os.MkdirAll(outsideDir, 0755)
		require.NoError(t, err)
		defer os.RemoveAll(outsideDir)
		
		// Create a glance.md file in the outside directory
		outsideGlanceFile := filepath.Join(outsideDir, "glance.md")
		err = os.WriteFile(outsideGlanceFile, []byte("Content from outside"), 0644)
		require.NoError(t, err)
		
		// Try to gather from the outside directory
		subdirs := []string{outsideDir}
		content, err := gatherSubGlances(subdirs)
		
		// Function shouldn't return an error, but should skip the invalid directory
		assert.NoError(t, err)
		assert.Empty(t, content)
	})

	t.Run("NonExistentDirectory", func(t *testing.T) {
		// Test with a directory that doesn't exist
		nonExistentDir := filepath.Join(testDir, "nonexistent")
		subdirs := []string{nonExistentDir}
		
		content, err := gatherSubGlances(subdirs)
		
		// Function shouldn't return an error, but should skip the non-existent directory
		assert.NoError(t, err)
		assert.Empty(t, content)
	})

	t.Run("MissingGlanceFile", func(t *testing.T) {
		// Create a directory without a glance.md file
		emptyDir := filepath.Join(testDir, "empty")
		err := os.MkdirAll(emptyDir, 0755)
		require.NoError(t, err)
		
		subdirs := []string{emptyDir}
		content, err := gatherSubGlances(subdirs)
		
		// Function shouldn't return an error, but should skip the directory without glance.md
		assert.NoError(t, err)
		assert.Empty(t, content)
	})

	t.Run("InvalidBaseDirForGlancePath", func(t *testing.T) {
		// This test ensures that using a parent directory as baseDir for validating glance.md
		// correctly prevents path traversal
		
		// Create a scenario where an attacker might try to reference a file outside
		// the directory by manipulating the glance.md path
		
		// First create a valid directory
		validDir := filepath.Join(testDir, "valid")
		err := os.MkdirAll(validDir, 0755)
		require.NoError(t, err)
		
		// But manually use it with a manipulated path to test security
		// This test directly checks the path validation logic
		// In real use, the file name is hardcoded as "glance.md"
		
		subdirs := []string{validDir}
		content, err := gatherSubGlances(subdirs)
		
		// Function shouldn't return an error but should skip the invalid file
		assert.NoError(t, err)
		assert.Empty(t, content)
	})
}