package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildProjectDirectoryMap(t *testing.T) {
	root := t.TempDir()
	levelA := filepath.Join(root, "a")
	levelB := filepath.Join(levelA, "b")
	levelC := filepath.Join(root, "c")

	require.NoError(t, os.MkdirAll(levelB, 0755))
	require.NoError(t, os.MkdirAll(levelC, 0755))

	dirs := []string{levelB, root, levelC, levelA}
	projectMap := buildProjectDirectoryMap(root, dirs, 500)

	assert.Contains(t, projectMap, "project directory tree:")
	assert.Contains(t, projectMap, "- .")
	assert.Contains(t, projectMap, "  - a")
	assert.Contains(t, projectMap, "    - a/b")
	assert.Contains(t, projectMap, "  - c")
}

func TestBuildProjectDirectoryMapTruncation(t *testing.T) {
	root := t.TempDir()
	var dirs []string
	dirs = append(dirs, root)
	for i := 0; i < 20; i++ {
		dirs = append(dirs, filepath.Join(root, "dir", "nested", "deep", "path", string(rune('a'+i))))
	}

	projectMap := buildProjectDirectoryMap(root, dirs, 90)
	assert.Contains(t, projectMap, "truncated for prompt budget")
}

func TestLoadExistingProjectOverview(t *testing.T) {
	root := t.TempDir()
	glancePath := filepath.Join(root, "glance.md")
	content := "overview content that is intentionally longer than the truncation threshold"
	require.NoError(t, os.WriteFile(glancePath, []byte(content), 0644))

	overview := loadExistingProjectOverview(root, 30)
	assert.Contains(t, overview, "overview content")
	assert.Contains(t, overview, "truncated for prompt budget")
}

func TestLoadExistingProjectOverviewMissingFile(t *testing.T) {
	root := t.TempDir()
	overview := loadExistingProjectOverview(root, 100)
	assert.Empty(t, overview)
}
