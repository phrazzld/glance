package filesystem

import (
	"os"
	"path/filepath"
	"testing"

	gitignore "github.com/sabhiram/go-gitignore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldIgnoreFile(t *testing.T) {
	// Setup test directory and files
	testDir := t.TempDir()

	// Create a .gitignore file
	gitignoreContent := "*.log\n*.tmp\nignored_dir/\n"
	gitignorePath := filepath.Join(testDir, ".gitignore")
	err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644)
	require.NoError(t, err)

	// Parse the gitignore file
	gitignoreObj, err := gitignore.CompileIgnoreFile(gitignorePath)
	require.NoError(t, err)

	// Create an ignore chain
	ignoreChain := IgnoreChain{
		{
			OriginDir: testDir,
			Matcher:   gitignoreObj,
		},
	}

	tests := []struct {
		name     string
		path     string
		baseDir  string
		chain    IgnoreChain
		verbose  bool
		expected bool
	}{
		{
			name:     "Hidden file",
			path:     filepath.Join(testDir, ".hidden"),
			baseDir:  testDir,
			chain:    ignoreChain,
			verbose:  false,
			expected: true,
		},
		{
			name:     "glance.md file",
			path:     filepath.Join(testDir, "glance.md"),
			baseDir:  testDir,
			chain:    ignoreChain,
			verbose:  false,
			expected: true,
		},
		{
			name:     "Regular file",
			path:     filepath.Join(testDir, "regular.txt"),
			baseDir:  testDir,
			chain:    ignoreChain,
			verbose:  false,
			expected: false,
		},
		{
			name:     "Gitignore matched file (.log)",
			path:     filepath.Join(testDir, "test.log"),
			baseDir:  testDir,
			chain:    ignoreChain,
			verbose:  false,
			expected: true,
		},
		{
			name:     "Gitignore matched file (.tmp)",
			path:     filepath.Join(testDir, "test.tmp"),
			baseDir:  testDir,
			chain:    ignoreChain,
			verbose:  false,
			expected: true,
		},
		{
			name:     "File inside ignored directory",
			path:     filepath.Join(testDir, "ignored_dir", "file.txt"),
			baseDir:  testDir,
			chain:    ignoreChain,
			verbose:  false,
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ShouldIgnoreFile(tc.path, tc.baseDir, tc.chain, tc.verbose)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestShouldIgnoreDir(t *testing.T) {
	// Setup test directory
	testDir := t.TempDir()

	// Create a .gitignore file
	gitignoreContent := "ignored_dir/\nbuild/\n*.log\n"
	gitignorePath := filepath.Join(testDir, ".gitignore")
	err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644)
	require.NoError(t, err)

	// Parse the gitignore file
	gitignoreObj, err := gitignore.CompileIgnoreFile(gitignorePath)
	require.NoError(t, err)

	// Create an ignore chain
	ignoreChain := IgnoreChain{
		{
			OriginDir: testDir,
			Matcher:   gitignoreObj,
		},
	}

	tests := []struct {
		name     string
		path     string
		baseDir  string
		chain    IgnoreChain
		verbose  bool
		expected bool
	}{
		{
			name:     "Hidden directory",
			path:     filepath.Join(testDir, ".git"),
			baseDir:  testDir,
			chain:    ignoreChain,
			verbose:  false,
			expected: true,
		},
		{
			name:     "Node modules directory",
			path:     filepath.Join(testDir, "node_modules"),
			baseDir:  testDir,
			chain:    ignoreChain,
			verbose:  false,
			expected: true,
		},
		{
			name:     "Regular directory",
			path:     filepath.Join(testDir, "src"),
			baseDir:  testDir,
			chain:    ignoreChain,
			verbose:  false,
			expected: false,
		},
		{
			name:     "Gitignore matched directory",
			path:     filepath.Join(testDir, "ignored_dir"),
			baseDir:  testDir,
			chain:    ignoreChain,
			verbose:  false,
			expected: true,
		},
		{
			name:     "Gitignore matched build directory",
			path:     filepath.Join(testDir, "build"),
			baseDir:  testDir,
			chain:    ignoreChain,
			verbose:  false,
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ShouldIgnoreDir(tc.path, tc.baseDir, tc.chain, tc.verbose)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestMatchesGitignore(t *testing.T) {
	// Setup test directory
	testDir := t.TempDir()

	// Create a .gitignore file with various patterns
	gitignoreContent := "*.log\n*.tmp\nignored_dir/\nbuild/\n"
	gitignorePath := filepath.Join(testDir, ".gitignore")
	err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644)
	require.NoError(t, err)

	// Parse the gitignore file
	gitignoreObj, err := gitignore.CompileIgnoreFile(gitignorePath)
	require.NoError(t, err)

	// Create a nested directory with its own .gitignore
	nestedDir := filepath.Join(testDir, "nested")
	err = os.MkdirAll(nestedDir, 0755)
	require.NoError(t, err)

	nestedGitignoreContent := "*.json\n*.md\n"
	nestedGitignorePath := filepath.Join(nestedDir, ".gitignore")
	err = os.WriteFile(nestedGitignorePath, []byte(nestedGitignoreContent), 0644)
	require.NoError(t, err)

	// Parse the nested .gitignore file
	nestedGitignoreObj, err := gitignore.CompileIgnoreFile(nestedGitignorePath)
	require.NoError(t, err)

	// Create an ignore chain with both .gitignore files
	ignoreChain := IgnoreChain{
		{
			OriginDir: testDir,
			Matcher:   gitignoreObj,
		},
		{
			OriginDir: nestedDir,
			Matcher:   nestedGitignoreObj,
		},
	}

	tests := []struct {
		name     string
		path     string
		baseDir  string
		chain    IgnoreChain
		isDir    bool
		expected bool
	}{
		{
			name:     "File matching root .gitignore",
			path:     filepath.Join(testDir, "test.log"),
			baseDir:  testDir,
			chain:    ignoreChain,
			isDir:    false,
			expected: true,
		},
		{
			name:     "Dir matching root .gitignore",
			path:     filepath.Join(testDir, "ignored_dir"),
			baseDir:  testDir,
			chain:    ignoreChain,
			isDir:    true,
			expected: true,
		},
		{
			name:     "File not matching any gitignore",
			path:     filepath.Join(testDir, "test.txt"),
			baseDir:  testDir,
			chain:    ignoreChain,
			isDir:    false,
			expected: false,
		},
		{
			name:     "Nested file matching nested .gitignore",
			path:     filepath.Join(nestedDir, "config.json"),
			baseDir:  nestedDir,
			chain:    ignoreChain,
			isDir:    false,
			expected: true,
		},
		{
			name:     "Nested file matching root .gitignore",
			path:     filepath.Join(nestedDir, "test.log"),
			baseDir:  nestedDir,
			chain:    ignoreChain,
			isDir:    false,
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := MatchesGitignore(tc.path, tc.baseDir, tc.chain, tc.isDir)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestShouldIgnorePath(t *testing.T) {
	// Setup test directory and files
	testDir := t.TempDir()

	// Create a .gitignore file
	gitignoreContent := "*.log\n*.tmp\nignored_dir/\nbuild/\n"
	gitignorePath := filepath.Join(testDir, ".gitignore")
	err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644)
	require.NoError(t, err)

	// Parse the gitignore file
	gitignoreObj, err := gitignore.CompileIgnoreFile(gitignorePath)
	require.NoError(t, err)

	// Create an ignore chain
	ignoreChain := IgnoreChain{
		{
			OriginDir: testDir,
			Matcher:   gitignoreObj,
		},
	}

	// Create test directories
	err = os.MkdirAll(filepath.Join(testDir, "ignored_dir"), 0755)
	require.NoError(t, err)
	err = os.MkdirAll(filepath.Join(testDir, ".git"), 0755)
	require.NoError(t, err)
	err = os.MkdirAll(filepath.Join(testDir, "node_modules"), 0755)
	require.NoError(t, err)

	tests := []struct {
		name     string
		path     string
		baseDir  string
		chain    IgnoreChain
		isDir    bool
		verbose  bool
		expected bool
	}{
		{
			name:     "Regular file",
			path:     filepath.Join(testDir, "regular.txt"),
			baseDir:  testDir,
			chain:    ignoreChain,
			isDir:    false,
			verbose:  false,
			expected: false,
		},
		{
			name:     "Hidden file",
			path:     filepath.Join(testDir, ".hidden"),
			baseDir:  testDir,
			chain:    ignoreChain,
			isDir:    false,
			verbose:  false,
			expected: true,
		},
		{
			name:     "Hidden directory",
			path:     filepath.Join(testDir, ".git"),
			baseDir:  testDir,
			chain:    ignoreChain,
			isDir:    true,
			verbose:  false,
			expected: true,
		},
		{
			name:     "glance.md file",
			path:     filepath.Join(testDir, "glance.md"),
			baseDir:  testDir,
			chain:    ignoreChain,
			isDir:    false,
			verbose:  false,
			expected: true,
		},
		{
			name:     "node_modules directory",
			path:     filepath.Join(testDir, "node_modules"),
			baseDir:  testDir,
			chain:    ignoreChain,
			isDir:    true,
			verbose:  false,
			expected: true,
		},
		{
			name:     "Gitignore matched file",
			path:     filepath.Join(testDir, "test.log"),
			baseDir:  testDir,
			chain:    ignoreChain,
			isDir:    false,
			verbose:  false,
			expected: true,
		},
		{
			name:     "Gitignore matched directory",
			path:     filepath.Join(testDir, "ignored_dir"),
			baseDir:  testDir,
			chain:    ignoreChain,
			isDir:    true,
			verbose:  false,
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ShouldIgnorePath(tc.path, tc.baseDir, tc.chain, tc.isDir, tc.verbose)
			assert.Equal(t, tc.expected, result)
		})
	}
}
