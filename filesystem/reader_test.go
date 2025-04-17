package filesystem

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	gitignore "github.com/sabhiram/go-gitignore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadTextFile(t *testing.T) {
	// Setup test directory and files
	testDir := t.TempDir()

	// Create a text file
	textFile := filepath.Join(testDir, "test.txt")
	textContent := "This is a test file with some content.\nIt has multiple lines."
	err := os.WriteFile(textFile, []byte(textContent), 0644)
	require.NoError(t, err)

	// Create a file with invalid UTF-8
	invalidUTF8 := filepath.Join(testDir, "invalid.txt")
	invalidContent := []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f, 0xff, 0xfe, 0x57, 0x6f, 0x72, 0x6c, 0x64} // "Hello" + invalid bytes + "World"
	err = os.WriteFile(invalidUTF8, invalidContent, 0644)
	require.NoError(t, err)

	// Create a large file for truncation testing
	largeFile := filepath.Join(testDir, "large.txt")
	largeContent := strings.Repeat("Large file test content. ", 1000) // Approx 23KB
	err = os.WriteFile(largeFile, []byte(largeContent), 0644)
	require.NoError(t, err)

	tests := []struct {
		name     string
		path     string
		maxBytes int64
		expect   string
		wantErr  bool
	}{
		{
			name:     "Read entire text file",
			path:     textFile,
			maxBytes: 0,
			expect:   textContent,
			wantErr:  false,
		},
		{
			name:     "Read with truncation",
			path:     textFile,
			maxBytes: 10,
			expect:   "This is a ...(truncated)",
			wantErr:  false,
		},
		{
			name:     "Read file with invalid UTF-8",
			path:     invalidUTF8,
			maxBytes: 0,
			expect:   "Hello�World", // Expect replacement characters
			wantErr:  false,
		},
		{
			name:     "Read nonexistent file",
			path:     filepath.Join(testDir, "nonexistent.txt"),
			maxBytes: 0,
			expect:   "",
			wantErr:  true,
		},
		{
			name:     "Truncate large file",
			path:     largeFile,
			maxBytes: 100,
			expect:   largeContent[:100] + "...(truncated)",
			wantErr:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			content, err := ReadTextFile(tc.path, tc.maxBytes)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expect, content)
			}
		})
	}
}

func TestTruncateContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		maxBytes int64
		expected string
	}{
		{
			name:     "Content shorter than max",
			content:  "Short content",
			maxBytes: 20,
			expected: "Short content",
		},
		{
			name:     "Content equal to max",
			content:  "12345",
			maxBytes: 5,
			expected: "12345",
		},
		{
			name:     "Content longer than max",
			content:  "This is a very long string that will be truncated",
			maxBytes: 10,
			expected: "This is a ...(truncated)",
		},
		{
			name:     "Zero max bytes (no truncation)",
			content:  "Some content",
			maxBytes: 0,
			expected: "Some content",
		},
		{
			name:     "Negative max bytes (no truncation)",
			content:  "Some content",
			maxBytes: -1,
			expected: "Some content",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := TruncateContent(tc.content, tc.maxBytes)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsTextFile(t *testing.T) {
	// Setup test directory and files
	testDir := t.TempDir()

	// Create a text file
	textFile := filepath.Join(testDir, "test.txt")
	err := os.WriteFile(textFile, []byte("This is a text file"), 0644)
	require.NoError(t, err)

	// Create a JSON file
	jsonFile := filepath.Join(testDir, "test.json")
	err = os.WriteFile(jsonFile, []byte(`{"key": "value"}`), 0644)
	require.NoError(t, err)

	// Create a XML file
	xmlFile := filepath.Join(testDir, "test.xml")
	err = os.WriteFile(xmlFile, []byte(`<root><item>value</item></root>`), 0644)
	require.NoError(t, err)

	// Create a YAML file
	yamlFile := filepath.Join(testDir, "test.yaml")
	err = os.WriteFile(yamlFile, []byte(`key: value\nlist:\n  - item1\n  - item2`), 0644)
	require.NoError(t, err)

	// Create a binary file
	binFile := filepath.Join(testDir, "test.bin")
	binData := make([]byte, 100)
	for i := range binData {
		binData[i] = byte(i)
	}
	err = os.WriteFile(binFile, binData, 0644)
	require.NoError(t, err)

	// Create an empty file
	emptyFile := filepath.Join(testDir, "empty.txt")
	err = os.WriteFile(emptyFile, []byte{}, 0644)
	require.NoError(t, err)

	tests := []struct {
		name     string
		path     string
		expected bool
		wantErr  bool
	}{
		{
			name:     "Text file",
			path:     textFile,
			expected: true,
			wantErr:  false,
		},
		{
			name:     "JSON file",
			path:     jsonFile,
			expected: true,
			wantErr:  false,
		},
		{
			name:     "XML file",
			path:     xmlFile,
			expected: true,
			wantErr:  false,
		},
		{
			name:     "YAML file",
			path:     yamlFile,
			expected: true,
			wantErr:  false,
		},
		{
			name:     "Binary file",
			path:     binFile,
			expected: false,
			wantErr:  false,
		},
		{
			name:     "Empty file",
			path:     emptyFile,
			expected: true, // Empty files are typically treated as text
			wantErr:  false,
		},
		{
			name:     "Nonexistent file",
			path:     filepath.Join(testDir, "nonexistent.txt"),
			expected: false,
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isText, err := IsTextFile(tc.path)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, isText)
			}
		})
	}
}

func TestGatherLocalFiles(t *testing.T) {
	// Setup test directory and files
	testDir := t.TempDir()

	// Create test files
	files := map[string]string{
		"file1.txt":   "Content of file1",
		"file2.json":  `{"key":"value"}`,
		".hidden.txt": "Hidden file",                    // Should be ignored (hidden)
		"glance.md":   "glance file",                    // Should be ignored (glance.md)
		"binary.bin":  string([]byte{0, 1, 2, 3, 4, 5}), // Should be ignored (binary)
	}

	for name, content := range files {
		path := filepath.Join(testDir, name)
		err := os.WriteFile(path, []byte(content), 0644)
		require.NoError(t, err)
	}

	// Create a nested directory (should be skipped)
	nestedDir := filepath.Join(testDir, "nested")
	err := os.Mkdir(nestedDir, 0755)
	require.NoError(t, err)

	nestedFile := filepath.Join(nestedDir, "nested.txt")
	err = os.WriteFile(nestedFile, []byte("Nested file content"), 0644)
	require.NoError(t, err)

	// Test with no ignore rules
	t.Run("Basic gathering with no ignore rules", func(t *testing.T) {
		results, err := GatherLocalFiles(testDir, nil, 0, true)
		assert.NoError(t, err)

		// Should find exactly 2 files (file1.txt and file2.json)
		assert.Len(t, results, 2)
		assert.Contains(t, results, "file1.txt")
		assert.Contains(t, results, "file2.json")

		// Should not contain hidden, glance.md, or nested files
		assert.NotContains(t, results, ".hidden.txt")
		assert.NotContains(t, results, "glance.md")
		assert.NotContains(t, results, "nested/nested.txt")

		// Content should match
		assert.Equal(t, "Content of file1", results["file1.txt"])
		assert.Equal(t, `{"key":"value"}`, results["file2.json"])
	})

	// Test with truncation
	t.Run("Truncation of large files", func(t *testing.T) {
		results, err := GatherLocalFiles(testDir, nil, 5, true)
		assert.NoError(t, err)

		// Content should be truncated
		assert.Equal(t, "Conte...(truncated)", results["file1.txt"])
	})

	// Test with gitignore rules
	t.Run("Respecting gitignore rules", func(t *testing.T) {
		// Create a .gitignore file
		gitignoreContent := "*.json\n"
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

		results, err := GatherLocalFiles(testDir, ignoreChain, 0, true)
		assert.NoError(t, err)

		// Should only find file1.txt as file2.json is ignored by gitignore
		assert.Len(t, results, 1)
		assert.Contains(t, results, "file1.txt")
		assert.NotContains(t, results, "file2.json")
	})

	// Test with non-existent directory
	t.Run("Error handling for non-existent directory", func(t *testing.T) {
		nonExistentDir := filepath.Join(testDir, "does-not-exist")
		_, err := GatherLocalFiles(nonExistentDir, nil, 0, true)
		assert.Error(t, err)
	})

	// Test with empty directory
	t.Run("Empty directory", func(t *testing.T) {
		emptyDir := filepath.Join(testDir, "empty")
		err := os.Mkdir(emptyDir, 0755)
		require.NoError(t, err)

		results, err := GatherLocalFiles(emptyDir, nil, 0, true)
		assert.NoError(t, err)
		assert.Empty(t, results, "Empty directory should return empty results map")
	})
}
