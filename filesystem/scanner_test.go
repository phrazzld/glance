package filesystem

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDirectory creates a temporary directory structure for testing
// Returns the root directory path and a cleanup function
func setupTestDirectory(t *testing.T) (string, func()) {
	t.Helper()

	// Create root test directory
	root, err := os.MkdirTemp("", "filesystem-scanner-test-*")
	require.NoError(t, err, "Failed to create temp root directory")

	// Create subdirectories
	subDirs := []string{
		"dir1",
		"dir1/subdir1",
		"dir1/subdir2",
		"dir2",
		"dir2/subdir1",
		".hidden_dir",
		"node_modules",
		"ignored_dir",
	}

	for _, dir := range subDirs {
		err := os.MkdirAll(filepath.Join(root, dir), 0755)
		require.NoError(t, err, "Failed to create subdirectory: "+dir)
	}

	// Create sample files in each directory
	files := []string{
		"file.txt",
		"dir1/file1.txt",
		"dir1/subdir1/file11.txt",
		"dir1/subdir2/file12.txt",
		"dir2/file2.txt",
		"dir2/subdir1/file21.txt",
		".hidden_dir/hidden_file.txt",
		"node_modules/module_file.txt",
		"ignored_dir/ignored_file.txt",
	}

	for _, file := range files {
		err := os.WriteFile(filepath.Join(root, file), []byte("test content"), 0644)
		require.NoError(t, err, "Failed to create file: "+file)
	}

	// Create .gitignore files
	gitignoreContents := map[string]string{
		"": "ignored_dir/\n*.log\n",
		"dir1": "subdir2/\n",
	}

	for dir, content := range gitignoreContents {
		err := os.WriteFile(filepath.Join(root, dir, ".gitignore"), []byte(content), 0644)
		require.NoError(t, err, "Failed to create .gitignore in "+dir)
	}

	// Return cleanup function
	cleanup := func() {
		os.RemoveAll(root)
	}

	return root, cleanup
}

func TestListDirsWithIgnores(t *testing.T) {
	// Set up test directory
	root, cleanup := setupTestDirectory(t)
	defer cleanup()

	// Print content of .gitignore files to debug
	rootGitignore, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	require.NoError(t, err, "Failed to read root .gitignore")
	t.Logf("Root .gitignore content: %s", string(rootGitignore))
	
	dir1Gitignore, err := os.ReadFile(filepath.Join(root, "dir1", ".gitignore"))
	require.NoError(t, err, "Failed to read dir1 .gitignore")
	t.Logf("dir1 .gitignore content: %s", string(dir1Gitignore))
	
	// Check direct pattern matching with loaded .gitignore files for debugging
	rootGI, err := LoadGitignore(root)
	require.NoError(t, err, "Failed to load root .gitignore")
	
	t.Logf("Direct .gitignore matching checks:")
	t.Logf("- root gitignore.MatchesPath('ignored_dir') = %v", rootGI.MatchesPath("ignored_dir"))
	t.Logf("- root gitignore.MatchesPath('ignored_dir/') = %v", rootGI.MatchesPath("ignored_dir/"))
	
	dir1GI, err := LoadGitignore(filepath.Join(root, "dir1"))
	require.NoError(t, err, "Failed to load dir1 .gitignore")
	
	t.Logf("- dir1 gitignore.MatchesPath('subdir2') = %v", dir1GI.MatchesPath("subdir2"))
	t.Logf("- dir1 gitignore.MatchesPath('subdir2/') = %v", dir1GI.MatchesPath("subdir2/"))
	
	// Call the function we want to test
	dirs, ignoreChains, err := ListDirsWithIgnores(root)
	
	// Verify no error occurred
	require.NoError(t, err, "ListDirsWithIgnores should not return an error with valid directory")
	
	// Print the directories we found for debugging
	t.Logf("Found directories: %v", dirs)
	
	// Check that we got the expected directories (and not the ignored ones)
	assert.Contains(t, dirs, root, "Result should include the root directory")
	assert.Contains(t, dirs, filepath.Join(root, "dir1"), "Result should include dir1")
	assert.Contains(t, dirs, filepath.Join(root, "dir1/subdir1"), "Result should include dir1/subdir1")
	assert.Contains(t, dirs, filepath.Join(root, "dir2"), "Result should include dir2")
	assert.Contains(t, dirs, filepath.Join(root, "dir2/subdir1"), "Result should include dir2/subdir1")
	
	// Check that ignored dirs are NOT included
	assert.NotContains(t, dirs, filepath.Join(root, ".hidden_dir"), "Result should not include .hidden_dir")
	assert.NotContains(t, dirs, filepath.Join(root, "node_modules"), "Result should not include node_modules")
	assert.NotContains(t, dirs, filepath.Join(root, "ignored_dir"), "Result should not include ignored_dir")
	assert.NotContains(t, dirs, filepath.Join(root, "dir1/subdir2"), "Result should not include dir1/subdir2 (ignored by local .gitignore)")

	// Verify ignore chains are correct
	// Root should have an ignore chain entry
	rootChain, ok := ignoreChains[root]
	assert.True(t, ok, "Root directory should have an ignore chain entry")
	
	// Root should have 1 rule after initialization (its own .gitignore)
	assert.Equal(t, 1, len(rootChain), "Root should have 1 ignore rule (its own .gitignore)")
	if len(rootChain) > 0 {
		rootRule := rootChain[0]
		assert.Equal(t, root, rootRule.OriginDir, "Rule's origin directory should match root")
		assert.NotNil(t, rootRule.Matcher, "Root's matcher should not be nil")
		
		// Test specific matches using the matcher
		assert.True(t, rootRule.Matcher.MatchesPath("ignored_dir/file.txt"), "Root matcher should ignore 'ignored_dir/file.txt'")
		assert.True(t, rootRule.Matcher.MatchesPath("test.log"), "Root matcher should ignore '*.log'")
	}

	// dir1 should have root's .gitignore + its own
	dir1Path := filepath.Join(root, "dir1")
	dir1Chain, ok := ignoreChains[dir1Path]
	assert.True(t, ok, "dir1 should have an ignore chain entry")
	
	// dir1 should have 2 rules: inherited from root + its own
	assert.Equal(t, 2, len(dir1Chain), "dir1 should have 2 ignore rules (root's + its own)")
	
	if len(dir1Chain) >= 2 {
		// First rule is inherited from root
		rootRule := dir1Chain[0]
		assert.Equal(t, root, rootRule.OriginDir, "First rule's origin directory should match root")
		
		// Second rule is dir1's own
		dir1Rule := dir1Chain[1]
		assert.Equal(t, dir1Path, dir1Rule.OriginDir, "Second rule's origin directory should match dir1")
		assert.NotNil(t, dir1Rule.Matcher, "dir1's matcher should not be nil")
		
		// Test specific matches using the matcher
		assert.True(t, dir1Rule.Matcher.MatchesPath("subdir2/file.txt"), "dir1 matcher should ignore 'subdir2/file.txt'")
	}
}

func TestLoadGitignore(t *testing.T) {
	// Set up test directory with a .gitignore file
	tempDir, err := os.MkdirTemp("", "gitignore-test-*")
	require.NoError(t, err, "Failed to create temp directory")
	defer os.RemoveAll(tempDir)

	// Create a .gitignore file
	gitignoreContent := "*.log\ntmp/\n"
	err = os.WriteFile(filepath.Join(tempDir, ".gitignore"), []byte(gitignoreContent), 0644)
	require.NoError(t, err, "Failed to create .gitignore file")

	// Test the LoadGitignore function
	gitignoreObj, err := LoadGitignore(tempDir)
	require.NoError(t, err, "LoadGitignore should not return an error with valid .gitignore file")
	require.NotNil(t, gitignoreObj, "LoadGitignore should return a non-nil GitIgnore object")

	// Test that the patterns were correctly loaded
	assert.True(t, gitignoreObj.MatchesPath("test.log"), "test.log should match *.log pattern")
	assert.True(t, gitignoreObj.MatchesPath("tmp/file.txt"), "tmp/file.txt should match tmp/ pattern")
	assert.False(t, gitignoreObj.MatchesPath("test.txt"), "test.txt should not match any pattern")

	// Test with non-existent .gitignore
	emptyDir, err := os.MkdirTemp("", "empty-gitignore-test-*")
	require.NoError(t, err, "Failed to create temp directory")
	defer os.RemoveAll(emptyDir)

	emptyGitignore, err := LoadGitignore(emptyDir)
	assert.Nil(t, err, "LoadGitignore should not return an error when .gitignore doesn't exist")
	assert.Nil(t, emptyGitignore, "LoadGitignore should return nil for GitIgnore when .gitignore doesn't exist")
}