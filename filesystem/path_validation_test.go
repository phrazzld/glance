package filesystem

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatePathWithinBase(t *testing.T) {
	// Create a temporary base directory for testing
	baseDir := t.TempDir()

	// Create a subdirectory for testing
	subDir := filepath.Join(baseDir, "subdir")
	err := os.Mkdir(subDir, 0755)
	require.NoError(t, err)

	// Create a file for testing
	testFile := filepath.Join(subDir, "testfile.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	t.Run("Empty baseDir is rejected", func(t *testing.T) {
		_, err := ValidatePathWithinBase(testFile, "", true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "baseDir cannot be empty")
	})

	t.Run("Valid paths within base", func(t *testing.T) {
		// Test with base directory itself
		validPath, err := ValidatePathWithinBase(baseDir, baseDir, true)
		assert.NoError(t, err)
		assert.Equal(t, filepath.Clean(baseDir), filepath.Clean(validPath))

		// Test with subdirectory
		validPath, err = ValidatePathWithinBase(subDir, baseDir, true)
		assert.NoError(t, err)
		assert.Equal(t, filepath.Clean(subDir), filepath.Clean(validPath))

		// Test with file in subdirectory
		validPath, err = ValidatePathWithinBase(testFile, baseDir, true)
		assert.NoError(t, err)
		assert.Equal(t, filepath.Clean(testFile), filepath.Clean(validPath))
	})

	t.Run("Base directory not allowed when allowBaseDir is false", func(t *testing.T) {
		_, err := ValidatePathWithinBase(baseDir, baseDir, false)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrPathOutsideBase)
	})

	t.Run("Path outside base directory", func(t *testing.T) {
		// Test with parent of base directory
		parentDir := filepath.Dir(baseDir)
		_, err := ValidatePathWithinBase(parentDir, baseDir, true)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrPathOutsideBase)

		// Test with sibling directory
		siblingDir := filepath.Join(parentDir, "sibling")
		_, err = ValidatePathWithinBase(siblingDir, baseDir, true)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrPathOutsideBase)

		// Test with path traversal attempt
		traversalPath := filepath.Join(baseDir, "../outside")
		_, err = ValidatePathWithinBase(traversalPath, baseDir, true)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrPathOutsideBase)
	})

	t.Run("Path normalization", func(t *testing.T) {
		// Test with path containing . and ..
		normalizedPath := filepath.Join(baseDir, "subdir/../subdir/testfile.txt")
		validPath, err := ValidatePathWithinBase(normalizedPath, baseDir, true)
		assert.NoError(t, err)
		assert.Equal(t, filepath.Clean(testFile), filepath.Clean(validPath))
	})
}

func TestValidateFilePath(t *testing.T) {
	// Create a temporary base directory for testing
	baseDir := t.TempDir()

	// Create a subdirectory for testing
	subDir := filepath.Join(baseDir, "subdir")
	err := os.Mkdir(subDir, 0755)
	require.NoError(t, err)

	// Create a file for testing
	testFile := filepath.Join(subDir, "testfile.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	t.Run("Empty baseDir is rejected", func(t *testing.T) {
		_, err := ValidateFilePath(testFile, "", true, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "baseDir cannot be empty")
	})

	t.Run("Valid file path that exists", func(t *testing.T) {
		validPath, err := ValidateFilePath(testFile, baseDir, true, true)
		assert.NoError(t, err)
		assert.Equal(t, filepath.Clean(testFile), filepath.Clean(validPath))
	})

	t.Run("Non-existent file when mustExist is true", func(t *testing.T) {
		nonExistentFile := filepath.Join(subDir, "nonexistent.txt")
		_, err := ValidateFilePath(nonExistentFile, baseDir, true, true)
		assert.Error(t, err)
		// Not checking the exact error type since it's a wrapped error from os.Stat
	})

	t.Run("Non-existent file when mustExist is false", func(t *testing.T) {
		nonExistentFile := filepath.Join(subDir, "nonexistent.txt")
		validPath, err := ValidateFilePath(nonExistentFile, baseDir, true, false)
		assert.NoError(t, err)
		assert.Equal(t, filepath.Clean(nonExistentFile), filepath.Clean(validPath))
	})

	t.Run("Directory path when expecting file", func(t *testing.T) {
		_, err := ValidateFilePath(subDir, baseDir, true, true)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotFile)
	})

	t.Run("Path outside base directory", func(t *testing.T) {
		outsideFile := filepath.Join(filepath.Dir(baseDir), "outside.txt")
		_, err := ValidateFilePath(outsideFile, baseDir, true, false)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrPathOutsideBase)
	})
}

func TestValidateDirPath(t *testing.T) {
	// Create a temporary base directory for testing
	baseDir := t.TempDir()

	// Create a subdirectory for testing
	subDir := filepath.Join(baseDir, "subdir")
	err := os.Mkdir(subDir, 0755)
	require.NoError(t, err)

	// Create a file for testing
	testFile := filepath.Join(subDir, "testfile.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	t.Run("Empty baseDir is rejected", func(t *testing.T) {
		_, err := ValidateDirPath(subDir, "", true, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "baseDir cannot be empty")
	})

	t.Run("Valid directory path that exists", func(t *testing.T) {
		validPath, err := ValidateDirPath(subDir, baseDir, true, true)
		assert.NoError(t, err)
		assert.Equal(t, filepath.Clean(subDir), filepath.Clean(validPath))
	})

	t.Run("Non-existent directory when mustExist is true", func(t *testing.T) {
		nonExistentDir := filepath.Join(baseDir, "nonexistent")
		_, err := ValidateDirPath(nonExistentDir, baseDir, true, true)
		assert.Error(t, err)
		// Not checking the exact error type since it's a wrapped error from os.Stat
	})

	t.Run("Non-existent directory when mustExist is false", func(t *testing.T) {
		nonExistentDir := filepath.Join(baseDir, "nonexistent")
		validPath, err := ValidateDirPath(nonExistentDir, baseDir, true, false)
		assert.NoError(t, err)
		assert.Equal(t, filepath.Clean(nonExistentDir), filepath.Clean(validPath))
	})

	t.Run("File path when expecting directory", func(t *testing.T) {
		_, err := ValidateDirPath(testFile, baseDir, true, true)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotDirectory)
	})

	t.Run("Path outside base directory", func(t *testing.T) {
		outsideDir := filepath.Join(filepath.Dir(baseDir), "outside")
		_, err := ValidateDirPath(outsideDir, baseDir, true, false)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrPathOutsideBase)
	})
}
