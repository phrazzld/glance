package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"glance/filesystem"
)

func TestReadSubdirectories(t *testing.T) {
	// Create a temporary directory for testing
	testDir := t.TempDir()

	// Create subdirectories
	subDir1 := filepath.Join(testDir, "subdir1")
	subDir2 := filepath.Join(testDir, "subdir2")
	subDir3 := filepath.Join(testDir, "subdir3")
	nestedDir := filepath.Join(subDir1, "nested")
	hiddenDir := filepath.Join(testDir, ".hidden")

	for _, dir := range []string{subDir1, subDir2, subDir3, nestedDir, hiddenDir} {
		err := os.MkdirAll(dir, 0755)
		require.NoError(t, err)
	}

	// Create a .gitignore file to exclude subdir3
	gitignorePath := filepath.Join(testDir, ".gitignore")
	err := os.WriteFile(gitignorePath, []byte("subdir3\n"), 0644)
	require.NoError(t, err)

	// Create an empty IgnoreChain for tests that don't need gitignore rules
	emptyIgnoreChain := filesystem.IgnoreChain{}

	// Create an IgnoreChain with the gitignore that excludes subdir3
	gitignore, err := filesystem.LoadGitignore(testDir)
	require.NoError(t, err)
	require.NotNil(t, gitignore, "Gitignore should be loaded successfully")

	ignoreChain := filesystem.IgnoreChain{
		{
			OriginDir: testDir,
			Matcher:   gitignore,
		},
	}
	require.NoError(t, err)

	t.Run("ValidDirectory", func(t *testing.T) {
		// Test with a valid directory
		subdirs, err := readSubdirectories(testDir, emptyIgnoreChain)

		assert.NoError(t, err)
		assert.Contains(t, subdirs, subDir1)
		assert.Contains(t, subdirs, subDir2)
		assert.Contains(t, subdirs, subDir3)
		assert.NotContains(t, subdirs, nestedDir) // Not a direct subdirectory of testDir
		assert.NotContains(t, subdirs, hiddenDir) // Hidden directories should be skipped
	})

	t.Run("NestedDirectory", func(t *testing.T) {
		// Test with a nested directory
		subdirs, err := readSubdirectories(subDir1, emptyIgnoreChain)

		assert.NoError(t, err)
		assert.Contains(t, subdirs, nestedDir)
		assert.NotContains(t, subdirs, subDir2)
	})

	t.Run("GitignoreRespected", func(t *testing.T) {
		// Test that gitignore rules are respected
		subdirs, err := readSubdirectories(testDir, ignoreChain)

		assert.NoError(t, err)
		assert.Contains(t, subdirs, subDir1)
		assert.Contains(t, subdirs, subDir2)
		assert.NotContains(t, subdirs, subDir3) // Should be excluded by .gitignore
	})

	t.Run("AttemptedTraversalWithParentRef", func(t *testing.T) {
		// Create a path with parent directory reference
		// This should be caught by the path validation
		invalidDirPath := filepath.Join(testDir, "subdir1", "..", "outside")

		// We can't directly test readSubdirectories with this path
		// because it's validated before the function gets to list entries
		// So instead, we'll verify that ValidateDirPath rejects it

		_, err := filesystem.ValidateDirPath(invalidDirPath, testDir, true, true)
		assert.Error(t, err)
	})

	t.Run("AttemptedTraversalDuringEnumeration", func(t *testing.T) {
		// Create a special test directory with a deceptive subdirectory
		traversalTestDir := filepath.Join(testDir, "traversal-test")
		err := os.MkdirAll(traversalTestDir, 0755)
		require.NoError(t, err)

		// Create a symbolic link that points outside the directory
		outsideTargetDir := filepath.Join(testDir, "outside-target")
		err = os.MkdirAll(outsideTargetDir, 0755)
		require.NoError(t, err)

		// Create a file in the outside directory to verify it's not accessible
		outsideFile := filepath.Join(outsideTargetDir, "secret.txt")
		err = os.WriteFile(outsideFile, []byte("secret content"), 0644)
		require.NoError(t, err)

		// Now create a symbolic link in the traversalTestDir pointing to the outside directory
		// Skip this test on platforms where symlinks aren't properly supported
		symlinkPath := filepath.Join(traversalTestDir, "symlink-outside")
		err = os.Symlink(outsideTargetDir, symlinkPath)
		if err != nil {
			t.Skip("Skipping symlink test - symlinks not supported on this platform")
		}

		// Now call readSubdirectories on the traversalTestDir
		subdirs, err := readSubdirectories(traversalTestDir, emptyIgnoreChain)

		// It should successfully return but the symlink should not be in the results
		assert.NoError(t, err)
		for _, subdir := range subdirs {
			assert.NotEqual(t, symlinkPath, subdir, "Symbolic link pointing outside should be excluded")
			// Additionally verify it's not the target either
			assert.NotEqual(t, outsideTargetDir, subdir, "Target of symbolic link should not be included")
		}
	})

	t.Run("NonExistentDirectory", func(t *testing.T) {
		// Test with a directory that doesn't exist
		nonExistentDir := filepath.Join(testDir, "nonexistent")
		_, err := readSubdirectories(nonExistentDir, emptyIgnoreChain)

		assert.Error(t, err)
	})
}
