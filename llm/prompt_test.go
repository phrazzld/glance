package llm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultTemplate(t *testing.T) {
	// Get the default template
	template := DefaultTemplate()
	
	// Verify it contains the expected placeholders
	assert.Contains(t, template, "{{.Directory}}")
	assert.Contains(t, template, "{{.SubGlances}}")
	assert.Contains(t, template, "{{.FileContents}}")
}

func TestLoadTemplate(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()
	
	// Create a custom template file
	customTemplate := "Custom template with {{.Directory}} and {{.SubGlances}} and {{.FileContents}}"
	customTemplatePath := filepath.Join(tempDir, "custom.txt")
	err := os.WriteFile(customTemplatePath, []byte(customTemplate), 0644)
	require.NoError(t, err)

	// Test cases
	tests := []struct {
		name     string
		path     string
		want     string
		wantErr  bool
	}{
		{
			name:    "Default template when path is empty",
			path:    "",
			want:    DefaultTemplate(),
			wantErr: false,
		},
		{
			name:    "Custom template from valid path",
			path:    customTemplatePath,
			want:    customTemplate,
			wantErr: false,
		},
		{
			name:    "Error with non-existent path",
			path:    filepath.Join(tempDir, "nonexistent.txt"),
			want:    "",
			wantErr: true,
		},
	}
	
	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadTemplate(tt.path)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGeneratePrompt(t *testing.T) {
	// Test data
	data := &PromptData{
		Directory:    "/test/dir",
		SubGlances:   "Sub glance 1\nSub glance 2",
		FileContents: "File1: content\nFile2: content",
	}
	
	// Test cases
	tests := []struct {
		name       string
		template   string
		data       *PromptData
		wantErr    bool
		assertions func(t *testing.T, result string)
	}{
		{
			name:     "Valid template",
			template: "Dir: {{.Directory}}\nSub: {{.SubGlances}}\nFiles: {{.FileContents}}",
			data:     data,
			wantErr:  false,
			assertions: func(t *testing.T, result string) {
				assert.Contains(t, result, "Dir: /test/dir")
				assert.Contains(t, result, "Sub: Sub glance 1\nSub glance 2")
				assert.Contains(t, result, "Files: File1: content\nFile2: content")
			},
		},
		{
			name:     "Invalid template",
			template: "Dir: {{.Directory}\nSub: {{.SubGlances}}", // Missing closing brace
			data:     data,
			wantErr:  true,
			assertions: func(t *testing.T, result string) {
				// No assertions on result for error case
			},
		},
		{
			name:     "Default template",
			template: DefaultTemplate(),
			data:     data,
			wantErr:  false,
			assertions: func(t *testing.T, result string) {
				assert.Contains(t, result, data.Directory)
				assert.Contains(t, result, data.SubGlances)
				assert.Contains(t, result, data.FileContents)
			},
		},
	}
	
	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GeneratePrompt(tt.data, tt.template)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				tt.assertions(t, result)
			}
		})
	}
}

func TestFormatFileContents(t *testing.T) {
	// Test input
	fileMap := map[string]string{
		"file1.txt": "Content 1",
		"file2.go":  "package main\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}",
	}
	
	// Generate formatted content
	formatted := FormatFileContents(fileMap)
	
	// Verify results
	assert.Contains(t, formatted, "=== file: file1.txt ===")
	assert.Contains(t, formatted, "Content 1")
	assert.Contains(t, formatted, "=== file: file2.go ===")
	assert.Contains(t, formatted, "package main")
	
	// Check that files are separated
	assert.True(t, strings.Contains(formatted, "\n\n"))
}

func TestBuildPromptData(t *testing.T) {
	// Test inputs
	dir := "/test/dir"
	subGlances := "Test sub glances"
	fileMap := map[string]string{
		"file1.txt": "Content 1",
		"file2.go":  "Content 2",
	}
	
	// Build prompt data
	data := BuildPromptData(dir, subGlances, fileMap)
	
	// Verify results
	assert.Equal(t, dir, data.Directory)
	assert.Equal(t, subGlances, data.SubGlances)
	assert.Contains(t, data.FileContents, "=== file: file1.txt ===")
	assert.Contains(t, data.FileContents, "Content 1")
	assert.Contains(t, data.FileContents, "=== file: file2.go ===")
	assert.Contains(t, data.FileContents, "Content 2")
}