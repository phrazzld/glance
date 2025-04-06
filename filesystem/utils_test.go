package filesystem

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	gitignore "github.com/sabhiram/go-gitignore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLatestModTime(t *testing.T) {
	// Create a test directory structure
	baseDir := t.TempDir()
	
	// Create a gitignore file
	ignoreContent := "*.log\nignore-dir/"
	ignoreFile := filepath.Join(baseDir, ".gitignore")
	err := os.WriteFile(ignoreFile, []byte(ignoreContent), 0644)
	require.NoError(t, err)
	
	// Create subdirectories and files with different modification times
	subDir1 := filepath.Join(baseDir, "subdir1")
	err = os.Mkdir(subDir1, 0755)
	require.NoError(t, err)
	
	subDir2 := filepath.Join(baseDir, "subdir2")
	err = os.Mkdir(subDir2, 0755)
	require.NoError(t, err)
	
	ignoreDir := filepath.Join(baseDir, "ignore-dir")
	err = os.Mkdir(ignoreDir, 0755)
	require.NoError(t, err)
	
	hiddenDir := filepath.Join(baseDir, ".hidden")
	err = os.Mkdir(hiddenDir, 0755)
	require.NoError(t, err)
	
	// Create files with different timestamps
	file1 := filepath.Join(subDir1, "file1.txt")
	err = os.WriteFile(file1, []byte("file1 content"), 0644)
	require.NoError(t, err)
	
	// Wait a bit to ensure different timestamps
	time.Sleep(10 * time.Millisecond)
	
	file2 := filepath.Join(subDir2, "file2.txt")
	err = os.WriteFile(file2, []byte("file2 content"), 0644)
	require.NoError(t, err)
	
	// Wait again
	time.Sleep(10 * time.Millisecond)
	
	// This will be the latest file
	file3 := filepath.Join(baseDir, "file3.txt")
	err = os.WriteFile(file3, []byte("file3 content"), 0644)
	require.NoError(t, err)
	
	// Get file info for the latest file to check against
	latestInfo, err := os.Stat(file3)
	require.NoError(t, err)
	latestTime := latestInfo.ModTime()
	
	// Create an ignored file with a newer timestamp
	time.Sleep(10 * time.Millisecond)
	ignoredFile := filepath.Join(ignoreDir, "ignored.txt")
	err = os.WriteFile(ignoredFile, []byte("ignored content"), 0644)
	require.NoError(t, err)
	
	hiddenFile := filepath.Join(hiddenDir, "hidden.txt")
	err = os.WriteFile(hiddenFile, []byte("hidden content"), 0644)
	require.NoError(t, err)
	
	logFile := filepath.Join(baseDir, "test.log")
	err = os.WriteFile(logFile, []byte("log content"), 0644)
	require.NoError(t, err)
	
	// Create ignore chain for testing
	gitignoreMatcher, err := gitignore.CompileIgnoreFile(ignoreFile)
	require.NoError(t, err)
	
	ignoreChain := IgnoreChain{
		{
			OriginDir: baseDir,
			Matcher:   gitignoreMatcher,
		},
	}
	
	// Test the function
	resultTime, err := LatestModTime(baseDir, ignoreChain, true)
	require.NoError(t, err)
	
	// The result should be the modification time of file3, not the ignored files
	assert.Equal(t, latestTime.Unix(), resultTime.Unix(), "Should return the latest modification time of non-ignored files")
	
	// Test with non-existent directory
	_, err = LatestModTime(filepath.Join(baseDir, "nonexistent"), ignoreChain, true)
	assert.Error(t, err, "Should return an error for non-existent directory")
}

func TestShouldRegenerate(t *testing.T) {
	// Create a test directory
	baseDir := t.TempDir()
	
	// Create a gitignore file
	ignoreContent := "*.log"
	ignoreFile := filepath.Join(baseDir, ".gitignore")
	err := os.WriteFile(ignoreFile, []byte(ignoreContent), 0644)
	require.NoError(t, err)
	
	// Create a GLANCE.md file
	glanceFile := filepath.Join(baseDir, "GLANCE.md")
	err = os.WriteFile(glanceFile, []byte("# Glance\n\nTest summary"), 0644)
	require.NoError(t, err)
	
	// Get the current time for reference
	_, err = os.Stat(glanceFile)
	require.NoError(t, err)
	
	// Create an ignore chain
	gitignoreMatcher, err := gitignore.CompileIgnoreFile(ignoreFile)
	require.NoError(t, err)
	
	ignoreChain := IgnoreChain{
		{
			OriginDir: baseDir,
			Matcher:   gitignoreMatcher,
		},
	}
	
	// Test cases
	t.Run("Force regeneration", func(t *testing.T) {
		shouldRegen, err := ShouldRegenerate(baseDir, true, ignoreChain, false)
		assert.NoError(t, err)
		assert.True(t, shouldRegen, "Should return true when force is true")
	})
	
	t.Run("No need to regenerate (no newer files)", func(t *testing.T) {
		shouldRegen, err := ShouldRegenerate(baseDir, false, ignoreChain, false)
		assert.NoError(t, err)
		assert.False(t, shouldRegen, "Should return false when no files are newer than GLANCE.md")
	})
	
	t.Run("No GLANCE.md file", func(t *testing.T) {
		// Create a new directory without GLANCE.md
		emptyDir := filepath.Join(baseDir, "empty")
		err := os.Mkdir(emptyDir, 0755)
		require.NoError(t, err)
		
		shouldRegen, err := ShouldRegenerate(emptyDir, false, ignoreChain, false)
		assert.NoError(t, err)
		assert.True(t, shouldRegen, "Should return true when no GLANCE.md exists")
	})
	
	t.Run("Need to regenerate (newer file)", func(t *testing.T) {
		// Wait to ensure a different timestamp
		time.Sleep(10 * time.Millisecond)
		
		// Create a newer file
		newerFile := filepath.Join(baseDir, "newer.txt")
		err := os.WriteFile(newerFile, []byte("newer content"), 0644)
		require.NoError(t, err)
		
		shouldRegen, err := ShouldRegenerate(baseDir, false, ignoreChain, false)
		assert.NoError(t, err)
		assert.True(t, shouldRegen, "Should return true when a file is newer than GLANCE.md")
	})
}

func TestBubbleUpParents(t *testing.T) {
	// Test case 1: Standard case with multiple parent directories
	t.Run("Standard case", func(t *testing.T) {
		root := "/test/root"
		dir := "/test/root/parent/child/grandchild"
		needsRegen := make(map[string]bool)
		
		BubbleUpParents(dir, root, needsRegen)
		
		// Should mark all parents up to (but not including) root
		assert.True(t, needsRegen["/test/root/parent/child"], "Should mark parent")
		assert.True(t, needsRegen["/test/root/parent"], "Should mark grandparent")
		assert.False(t, needsRegen["/test/root"], "Should not mark root")
		assert.False(t, needsRegen["/test"], "Should not mark directories above root")
	})
	
	// Test case 2: Directory is the root
	t.Run("Dir is root", func(t *testing.T) {
		root := "/test/root"
		dir := "/test/root"
		needsRegen := make(map[string]bool)
		
		BubbleUpParents(dir, root, needsRegen)
		
		// Should not mark anything
		assert.Empty(t, needsRegen, "Should not mark any directories")
	})
	
	// Test case 3: Path shorter than root (edge case, shouldn't happen)
	t.Run("Path shorter than root", func(t *testing.T) {
		root := "/test/root/deep"
		dir := "/test/root"
		needsRegen := make(map[string]bool)
		
		BubbleUpParents(dir, root, needsRegen)
		
		// Should not mark anything
		assert.Empty(t, needsRegen, "Should not mark any directories")
	})
}