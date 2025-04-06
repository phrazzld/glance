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
	
	// Verify it contains essential prompt instructions
	assert.Contains(t, template, "expert code reviewer")
	assert.Contains(t, template, "technical overview")
	assert.Contains(t, template, "highlight purpose")
}

func TestLoadTemplate(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()
	
	// Create a custom template file
	customTemplate := "Custom template with {{.Directory}} and {{.SubGlances}} and {{.FileContents}}"
	customTemplatePath := filepath.Join(tempDir, "custom.txt")
	err := os.WriteFile(customTemplatePath, []byte(customTemplate), 0644)
	require.NoError(t, err)
	
	// Create a default location template file in current directory (will be cleaned up)
	defaultLocTemplate := "Default location template with {{.Directory}}"
	defaultPath := "prompt.txt"
	currentDir, err := os.Getwd()
	require.NoError(t, err)
	fullDefaultPath := filepath.Join(currentDir, defaultPath)
	
	// Use cleanup to ensure we remove the prompt.txt file after tests
	t.Cleanup(func() {
		os.Remove(fullDefaultPath)
	})

	// Test cases
	tests := []struct {
		name     string
		setupFn  func() error        // Setup function to run before the test
		cleanupFn func() error       // Cleanup function to run after the test
		path     string
		want     string
		wantErr  bool
	}{
		{
			name:    "Default template when path is empty and no prompt.txt exists",
			setupFn: func() error { return nil },
			cleanupFn: func() error { return nil },
			path:    "",
			want:    DefaultTemplate(),
			wantErr: false,
		},
		{
			name:    "Custom template from valid path",
			setupFn: func() error { return nil },
			cleanupFn: func() error { return nil },
			path:    customTemplatePath,
			want:    customTemplate,
			wantErr: false,
		},
		{
			name:    "Error with non-existent path",
			setupFn: func() error { return nil },
			cleanupFn: func() error { return nil },
			path:    filepath.Join(tempDir, "nonexistent.txt"),
			want:    "",
			wantErr: true,
		},
		{
			name: "Use prompt.txt from current directory when path is empty",
			setupFn: func() error {
				return os.WriteFile(fullDefaultPath, []byte(defaultLocTemplate), 0644)
			},
			cleanupFn: func() error {
				return os.Remove(fullDefaultPath)
			},
			path:    "",
			want:    defaultLocTemplate,
			wantErr: false,
		},
	}
	
	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run setup
			if tt.setupFn != nil {
				err := tt.setupFn()
				require.NoError(t, err)
			}
			
			// Test cleanup
			if tt.cleanupFn != nil {
				defer func() {
					err := tt.cleanupFn()
					if err != nil {
						t.Logf("cleanup error: %v", err)
					}
				}()
			}
			
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
		{
			name:     "Template with unknown variable",
			template: "Dir: {{.Directory}}\nUnknown: {{.UnknownVar}}",
			data:     data,
			wantErr:  true,
			assertions: func(t *testing.T, result string) {
				// No assertions on result for error case
			},
		},
		{
			name:     "Empty template",
			template: "",
			data:     data,
			wantErr:  false,
			assertions: func(t *testing.T, result string) {
				assert.Empty(t, result)
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
	
	// Test with nil data - will handle nil gracefully in the template execution
	t.Run("Nil data", func(t *testing.T) {
		// Go's text/template will handle nil data by providing zero values
		// rather than failing with an error, so we're testing that it works
		result, err := GeneratePrompt(nil, "template")
		assert.NoError(t, err)
		assert.Equal(t, "template", result)
	})
}

func TestFormatFileContents(t *testing.T) {
	// Test with normal input
	t.Run("Normal file map", func(t *testing.T) {
		fileMap := map[string]string{
			"file1.txt": "Content 1",
			"file2.go":  "package main\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}",
		}
		
		formatted := FormatFileContents(fileMap)
		
		assert.Contains(t, formatted, "=== file: file1.txt ===")
		assert.Contains(t, formatted, "Content 1")
		assert.Contains(t, formatted, "=== file: file2.go ===")
		assert.Contains(t, formatted, "package main")
		assert.True(t, strings.Contains(formatted, "\n\n"))
	})
	
	// Test with empty map
	t.Run("Empty file map", func(t *testing.T) {
		fileMap := map[string]string{}
		formatted := FormatFileContents(fileMap)
		assert.Empty(t, formatted)
	})
	
	// Test with empty content
	t.Run("Files with empty content", func(t *testing.T) {
		fileMap := map[string]string{
			"empty.txt": "",
		}
		formatted := FormatFileContents(fileMap)
		assert.Contains(t, formatted, "=== file: empty.txt ===")
		assert.Contains(t, formatted, "===\n\n\n") // Empty content followed by newlines
	})
	
	// Test with special characters
	t.Run("Files with special characters", func(t *testing.T) {
		fileMap := map[string]string{
			"special.txt": "Content with special chars: ©®™",
		}
		formatted := FormatFileContents(fileMap)
		assert.Contains(t, formatted, "=== file: special.txt ===")
		assert.Contains(t, formatted, "Content with special chars: ©®™")
	})
	
	// Test with multi-line content
	t.Run("Multi-line content", func(t *testing.T) {
		fileMap := map[string]string{
			"multiline.txt": "Line 1\nLine 2\nLine 3",
		}
		formatted := FormatFileContents(fileMap)
		assert.Contains(t, formatted, "=== file: multiline.txt ===")
		assert.Contains(t, formatted, "Line 1\nLine 2\nLine 3")
	})
}

func TestBuildPromptData(t *testing.T) {
	// Test normal inputs
	t.Run("Normal inputs", func(t *testing.T) {
		dir := "/test/dir"
		subGlances := "Test sub glances"
		fileMap := map[string]string{
			"file1.txt": "Content 1",
			"file2.go":  "Content 2",
		}
		
		data := BuildPromptData(dir, subGlances, fileMap)
		
		assert.Equal(t, dir, data.Directory)
		assert.Equal(t, subGlances, data.SubGlances)
		assert.Contains(t, data.FileContents, "=== file: file1.txt ===")
		assert.Contains(t, data.FileContents, "Content 1")
		assert.Contains(t, data.FileContents, "=== file: file2.go ===")
		assert.Contains(t, data.FileContents, "Content 2")
	})
	
	// Test with empty inputs
	t.Run("Empty inputs", func(t *testing.T) {
		data := BuildPromptData("", "", map[string]string{})
		
		assert.Empty(t, data.Directory)
		assert.Empty(t, data.SubGlances)
		assert.Empty(t, data.FileContents)
	})
	
	// Test with nil file map
	t.Run("Nil file map", func(t *testing.T) {
		data := BuildPromptData("/test/dir", "Sub glances", nil)
		
		assert.Equal(t, "/test/dir", data.Directory)
		assert.Equal(t, "Sub glances", data.SubGlances)
		assert.Empty(t, data.FileContents)
	})
	
	// Test with large input
	t.Run("Large input", func(t *testing.T) {
		largeContent := strings.Repeat("Large content line\n", 1000)
		fileMap := map[string]string{
			"large.txt": largeContent,
		}
		
		data := BuildPromptData("/test/dir", "Sub glances", fileMap)
		
		assert.Equal(t, "/test/dir", data.Directory)
		assert.Equal(t, "Sub glances", data.SubGlances)
		assert.Contains(t, data.FileContents, "=== file: large.txt ===")
		assert.Contains(t, data.FileContents, "Large content line")
		// The content should be preserved
		assert.True(t, strings.Count(data.FileContents, "Large content line") > 100)
	})
}