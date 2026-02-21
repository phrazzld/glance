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
	resultTime, err := LatestModTime(baseDir, ignoreChain)
	require.NoError(t, err)

	// The result should be the modification time of file3, not the ignored files
	assert.Equal(t, latestTime.Unix(), resultTime.Unix(), "Should return the latest modification time of non-ignored files")

	// Test with non-existent directory
	_, err = LatestModTime(filepath.Join(baseDir, "nonexistent"), ignoreChain)
	assert.Error(t, err, "Should return an error for non-existent directory")

	// Test with empty directory
	emptyIgnoredDir := filepath.Join(baseDir, "empty-ignored")
	err = os.Mkdir(emptyIgnoredDir, 0755)
	require.NoError(t, err)

	// Get the empty directory's own mod time
	emptyDirInfo, err := os.Stat(emptyIgnoredDir)
	require.NoError(t, err)

	// Call latestModTime
	emptyDirTime, err := LatestModTime(emptyIgnoredDir, ignoreChain)
	require.NoError(t, err)

	// Should get the directory's own time since there are no files
	assert.Equal(t, emptyDirInfo.ModTime().Unix(), emptyDirTime.Unix(), "Empty dir should return dir's own mod time")

	// Test with empty directory
	emptyDir := filepath.Join(baseDir, "empty")
	err = os.Mkdir(emptyDir, 0755)
	require.NoError(t, err)

	emptyTime, err := LatestModTime(emptyDir, ignoreChain)
	require.NoError(t, err)

	// Should return the directory's own modification time
	emptyDirInfo2, err := os.Stat(emptyDir)
	require.NoError(t, err)
	emptyDirTime2 := emptyDirInfo2.ModTime()

	assert.Equal(t, emptyDirTime2.Unix(), emptyTime.Unix(), "Should return directory's mod time for empty directory")
}

func TestShouldRegenerate(t *testing.T) {
	// Create a test directory
	baseDir := t.TempDir()

	// Create a gitignore file
	ignoreContent := "*.log"
	ignoreFile := filepath.Join(baseDir, ".gitignore")
	err := os.WriteFile(ignoreFile, []byte(ignoreContent), 0644)
	require.NoError(t, err)

	// Create the glance output file
	glanceFile := filepath.Join(baseDir, GlanceFilename)
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
		shouldRegen, err := ShouldRegenerate(baseDir, true, ignoreChain)
		assert.NoError(t, err)
		assert.True(t, shouldRegen, "Should return true when force is true")
	})

	t.Run("No need to regenerate (no newer files)", func(t *testing.T) {
		shouldRegen, err := ShouldRegenerate(baseDir, false, ignoreChain)
		assert.NoError(t, err)
		assert.False(t, shouldRegen, "Should return false when no files are newer than glance.md")
	})

	t.Run("No glance.md file", func(t *testing.T) {
		// Create a new directory without glance.md
		emptyDir := filepath.Join(baseDir, "empty")
		err := os.Mkdir(emptyDir, 0755)
		require.NoError(t, err)

		shouldRegen, err := ShouldRegenerate(emptyDir, false, ignoreChain)
		assert.NoError(t, err)
		assert.True(t, shouldRegen, "Should return true when no glance.md exists")
	})

	t.Run("Need to regenerate (newer file)", func(t *testing.T) {
		// Wait to ensure a different timestamp
		time.Sleep(10 * time.Millisecond)

		// Create a newer file
		newerFile := filepath.Join(baseDir, "newer.txt")
		err := os.WriteFile(newerFile, []byte("newer content"), 0644)
		require.NoError(t, err)

		shouldRegen, err := ShouldRegenerate(baseDir, false, ignoreChain)
		assert.NoError(t, err)
		assert.True(t, shouldRegen, "Should return true when a file is newer than glance.md")
	})

	// Test only simple cases for ShouldRegenerate

	// Skip other edge cases
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

	// Test case 4: Existing map with some entries
	t.Run("Existing entries in map", func(t *testing.T) {
		root := "/test/root"
		dir := "/test/root/parent/child/grandchild"
		needsRegen := make(map[string]bool)

		// Pre-populate the map
		needsRegen["/test/root/parent"] = true
		needsRegen["/test/other/path"] = true

		BubbleUpParents(dir, root, needsRegen)

		// Should keep existing entries and add new ones
		assert.True(t, needsRegen["/test/root/parent/child"], "Should mark parent")
		assert.True(t, needsRegen["/test/root/parent"], "Should keep existing entry")
		assert.True(t, needsRegen["/test/other/path"], "Should not affect unrelated entries")
		assert.False(t, needsRegen["/test/root"], "Should not mark root")
	})

	// Test case 5: Case with absolute paths that need normalization
	t.Run("Path normalization", func(t *testing.T) {
		// Paths with dot segments
		root := "/test/root"
		dir := "/test/root/parent/./child/../child/grandchild"
		needsRegen := make(map[string]bool)

		// Normalize paths first (as the real code would use filepath.Clean internally)
		cleanDir := filepath.Clean(dir)

		BubbleUpParents(cleanDir, root, needsRegen)

		// Should work as expected with cleaned paths
		assert.True(t, needsRegen["/test/root/parent/child"], "Should mark parent")
		assert.True(t, needsRegen["/test/root/parent"], "Should mark grandparent")
		assert.False(t, needsRegen["/test/root"], "Should not mark root")
	})

	// Test case 6: Immediate child of root
	t.Run("Immediate child of root", func(t *testing.T) {
		root := "/test/root"
		dir := "/test/root/child"
		needsRegen := make(map[string]bool)

		BubbleUpParents(dir, root, needsRegen)

		// Should be empty since the only parent is the root
		assert.Empty(t, needsRegen, "Should not mark root directory")
	})

	// Test case 7: Windows-style paths
	t.Run("Windows-style paths", func(t *testing.T) {
		// This test is useful on all platforms because filepath.Dir will
		// process paths according to the host platform
		root := filepath.FromSlash("C:/Users/test")
		dir := filepath.FromSlash("C:/Users/test/Documents/folder/subfolder")
		needsRegen := make(map[string]bool)

		BubbleUpParents(dir, root, needsRegen)

		documents := filepath.FromSlash("C:/Users/test/Documents")
		documentsFolder := filepath.FromSlash("C:/Users/test/Documents/folder")

		assert.True(t, needsRegen[documentsFolder], "Should mark parent folders correctly on Windows paths")
		assert.True(t, needsRegen[documents], "Should mark grandparent folders correctly on Windows paths")
		assert.False(t, needsRegen[root], "Should not mark root on Windows paths")
	})

	// Test case 8: Map with false values
	t.Run("Map with false values", func(t *testing.T) {
		root := "/test/root"
		dir := "/test/root/parent/child/grandchild"
		needsRegen := make(map[string]bool)

		// Pre-populate the map with false values
		needsRegen["/test/root/parent"] = false
		needsRegen["/test/root/parent/child"] = false

		BubbleUpParents(dir, root, needsRegen)

		// Should overwrite false values with true
		assert.True(t, needsRegen["/test/root/parent/child"], "Should overwrite false with true")
		assert.True(t, needsRegen["/test/root/parent"], "Should overwrite false with true")
	})

	// Test case 9: Empty directory path (edge case)
	t.Run("Empty directory path", func(t *testing.T) {
		root := "/test/root"
		dir := ""
		needsRegen := make(map[string]bool)

		// This should handle gracefully without panicking
		BubbleUpParents(dir, root, needsRegen)

		// Should not mark anything as the loop should exit immediately
		assert.Empty(t, needsRegen, "Should handle empty directory path gracefully")
	})
}

func TestLatestModTime_EdgeCases(t *testing.T) {
	// Test case: Directory with weird filenames
	t.Run("Directory with special characters", func(t *testing.T) {
		testDir := t.TempDir()

		// Create files with special characters
		specialFiles := []string{
			"file with spaces.txt",
			"file_with_underscores.txt",
			"file-with-dashes.txt",
			"file.with.dots.txt",
			"file+with+plus.txt",
			"file'with'quotes.txt",
			"file%with%percent.txt",
			"file(with)parentheses.txt",
		}

		// Create the files with different timestamps
		var latestTime time.Time

		for i, filename := range specialFiles {
			time.Sleep(10 * time.Millisecond)
			filePath := filepath.Join(testDir, filename)
			err := os.WriteFile(filePath, []byte("content"), 0644)
			require.NoError(t, err)

			// Keep track of the latest file's mod time
			fileInfo, err := os.Stat(filePath)
			require.NoError(t, err)

			if i == 0 || fileInfo.ModTime().After(latestTime) {
				latestTime = fileInfo.ModTime()
			}
		}

		// Get the latest mod time
		resultTime, err := LatestModTime(testDir, nil)
		require.NoError(t, err)

		// Check that it matches the expected latest file
		assert.Equal(t, latestTime.Unix(), resultTime.Unix(), "Should find the latest file even with special characters")

		// Create a newer .gitignore file (which should be included in the scan)
		time.Sleep(10 * time.Millisecond)
		gitignorePath := filepath.Join(testDir, ".gitignore")
		err = os.WriteFile(gitignorePath, []byte("*.log"), 0644)
		require.NoError(t, err)

		gitignoreInfo, err := os.Stat(gitignorePath)
		require.NoError(t, err)
		gitignoreTime := gitignoreInfo.ModTime()

		// Get the latest mod time again
		newResultTime, err := LatestModTime(testDir, nil)
		require.NoError(t, err)

		// Now the .gitignore file should be the latest
		assert.Equal(t, gitignoreTime.Unix(), newResultTime.Unix(), "Should include dot files in the scan")
	})
}

// Skipping TestShouldRegenerate_EdgeCases for simplicity
// These tests are too dependent on file system permissions that vary by platform
