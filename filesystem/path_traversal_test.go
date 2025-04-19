package filesystem

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPathTraversalAttempts(t *testing.T) {
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

	testCases := []struct {
		name        string
		path        string
		shouldPass  bool
		description string
	}{
		{
			name:        "SingleLevelTraversal",
			path:        filepath.Join(baseDir, "../outside"),
			shouldPass:  false,
			description: "Simple single-level directory traversal",
		},
		{
			name:        "DoubleLevelTraversal",
			path:        filepath.Join(baseDir, "../../outside"),
			shouldPass:  false,
			description: "Double-level directory traversal",
		},
		{
			name:        "MultipleLevelTraversal",
			path:        filepath.Join(baseDir, "../../../../../etc/passwd"),
			shouldPass:  false,
			description: "Multiple level directory traversal attempting to access system files",
		},
		{
			name:        "TraversalWithValidSubpath",
			path:        filepath.Join(baseDir, "../", filepath.Base(baseDir), "/subdir/testfile.txt"),
			shouldPass:  true,
			description: "Traversal that eventually returns to a valid path",
		},
		{
			name:        "TraversalWithinSubdir",
			path:        filepath.Join(subDir, "../", filepath.Base(subDir), "/testfile.txt"),
			shouldPass:  true,
			description: "Traversal within a subdirectory that resolves to a valid path",
		},
		{
			name:        "DoubleDotsInFilename",
			path:        filepath.Join(baseDir, "file..name.txt"),
			shouldPass:  true,
			description: "Double dots in a filename (should be valid)",
		},
	}

	// Windows-specific path tests
	if runtime.GOOS == "windows" {
		testCases = append(testCases, []struct {
			name        string
			path        string
			shouldPass  bool
			description string
		}{
			{
				name:        "WindowsBackslashTraversal",
				path:        baseDir + "\\..\\outside",
				shouldPass:  false,
				description: "Windows-style backslash traversal",
			},
			{
				name:        "WindowsMixedSlashTraversal",
				path:        baseDir + "\\../outside",
				shouldPass:  false,
				description: "Windows mixed forward/backslash traversal",
			},
		}...)
	}

	// Test each path traversal attempt
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ValidatePathWithinBase(tc.path, baseDir, true)

			if tc.shouldPass {
				assert.NoError(t, err, "Valid path should pass validation: "+tc.description)
			} else {
				assert.Error(t, err, "Path traversal attempt should fail: "+tc.description)
				assert.ErrorIs(t, err, ErrPathOutsideBase, "Error should be ErrPathOutsideBase")
			}
		})
	}
}

func TestURLEncodedTraversal(t *testing.T) {
	// Create a temporary base directory for testing
	baseDir := t.TempDir()

	// Test various URL-encoded traversal attempts
	// Note: URL encoding is handled before these validation functions in real applications
	// We're documenting here that our validators don't interpret URL encoding

	// Create a test file in the base directory for reference
	testFile := filepath.Join(baseDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	t.Run("URL Encoding Is Not Interpreted", func(t *testing.T) {
		// The literal string with %2e in it (not decoded)
		literalPath := filepath.Join(baseDir, "%2e%2e", "test.txt")

		// Create the directory with the literal "%2e%2e" name
		literalDir := filepath.Join(baseDir, "%2e%2e")
		err := os.Mkdir(literalDir, 0755)
		require.NoError(t, err)

		// Create a file in this directory
		err = os.WriteFile(literalPath, []byte("literal content"), 0644)
		require.NoError(t, err)

		// This path should be valid because it's a real path with "%2e" as literal characters
		validPath, err := ValidatePathWithinBase(literalPath, baseDir, true)
		assert.NoError(t, err, "Path with literal %2e characters should be valid")
		assert.Equal(t, filepath.Clean(literalPath), filepath.Clean(validPath))

		// What would happen if we actually decoded this URL in an application?
		decodedPath := strings.ReplaceAll(literalPath, "%2e", ".")

		// The decoded path would be "../test.txt" which is a traversal attempt
		// This demonstrates that applications must URL-decode before validation
		_, err = ValidatePathWithinBase(decodedPath, baseDir, true)
		assert.Error(t, err, "Decoded path should fail validation")
		assert.ErrorIs(t, err, ErrPathOutsideBase)
	})

	t.Run("Application Must Decode URLs Before Validation", func(t *testing.T) {
		// Document that applications must handle URL decoding before validation
		// This test doesn't assert anything but serves as documentation

		encodedPaths := []string{
			"%2e%2e%2fetc%2fpasswd",            // ../etc/passwd
			"..%2f..%2f..%2f..%2fetc%2fpasswd", // ../../../../etc/passwd
			"subdir%2f..%2f..%2ffile.txt",      // subdir/../../file.txt
		}

		for _, encodedPath := range encodedPaths {
			t.Logf("URL-encoded path: %s", encodedPath)

			// If this were properly decoded in an application:
			decodedPath := strings.ReplaceAll(encodedPath, "%2e", ".")
			decodedPath = strings.ReplaceAll(decodedPath, "%2f", "/")
			t.Logf("Decoded path: %s", decodedPath)

			// The application would then pass the decoded path to validation
			// which would properly detect the traversal attempt
			t.Logf("Application should decode URLs before validation")
		}
	})
}

func TestAbsolutePathTraversal(t *testing.T) {
	// Create a temporary base directory for testing
	baseDir := t.TempDir()
	absBaseDir, err := filepath.Abs(baseDir)
	require.NoError(t, err)

	// Create a file outside the base directory
	parentDir := filepath.Dir(baseDir)
	outsideFile := filepath.Join(parentDir, "outside.txt")
	err = os.WriteFile(outsideFile, []byte("outside content"), 0644)
	require.NoError(t, err)
	absOutsideFile, err := filepath.Abs(outsideFile)
	require.NoError(t, err)

	testCases := []struct {
		name        string
		path        string
		shouldPass  bool
		description string
	}{
		{
			name:        "AbsolutePathOutsideBase",
			path:        absOutsideFile,
			shouldPass:  false,
			description: "Absolute path to a file outside the base directory",
		},
		{
			name:        "AbsolutePathToBaseDir",
			path:        absBaseDir,
			shouldPass:  true,
			description: "Absolute path to the base directory itself",
		},
		{
			name:        "AbsolutePathToParentDir",
			path:        filepath.Dir(absBaseDir),
			shouldPass:  false,
			description: "Absolute path to parent of base directory",
		},
		{
			name:        "AbsolutePathToSubDir",
			path:        filepath.Join(absBaseDir, "subdir"),
			shouldPass:  true,
			description: "Absolute path to a subdirectory",
		},
		{
			name:        "RootPath",
			path:        "/",
			shouldPass:  false,
			description: "Absolute path to the root directory",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ValidatePathWithinBase(tc.path, baseDir, true)

			if tc.shouldPass {
				assert.NoError(t, err, "Valid absolute path should pass: "+tc.description)
				assert.NotEmpty(t, result, "Valid path should return a non-empty result")
			} else {
				assert.Error(t, err, "Invalid absolute path should fail: "+tc.description)
				assert.ErrorIs(t, err, ErrPathOutsideBase, "Error should be ErrPathOutsideBase")
			}
		})
	}
}

func TestSymlinkTraversal(t *testing.T) {
	// Skip on Windows as symlinks require elevated privileges
	if runtime.GOOS == "windows" {
		t.Skip("Skipping symlink tests on Windows")
	}

	// Create a temporary base directory for testing
	baseDir := t.TempDir()

	// Create a directory outside the base directory
	outsideDir := filepath.Join(filepath.Dir(baseDir), "outside")
	err := os.Mkdir(outsideDir, 0755)
	require.NoError(t, err)

	// Create a file in the outside directory
	outsideFile := filepath.Join(outsideDir, "outside.txt")
	err = os.WriteFile(outsideFile, []byte("outside content"), 0644)
	require.NoError(t, err)

	// Create a subdirectory in the base directory
	subDir := filepath.Join(baseDir, "subdir")
	err = os.Mkdir(subDir, 0755)
	require.NoError(t, err)

	// Create a symlink in the base directory pointing to the outside directory
	symlinkToOutsideDir := filepath.Join(baseDir, "symlink-to-outside")
	err = os.Symlink(outsideDir, symlinkToOutsideDir)
	require.NoError(t, err)

	// Create a symlink in the base directory pointing to the outside file
	symlinkToOutsideFile := filepath.Join(baseDir, "symlink-to-outside-file")
	err = os.Symlink(outsideFile, symlinkToOutsideFile)
	require.NoError(t, err)

	// Create a file in the base directory
	insideFile := filepath.Join(baseDir, "inside.txt")
	err = os.WriteFile(insideFile, []byte("inside content"), 0644)
	require.NoError(t, err)

	// Create a symlink to a file within the base directory (should be valid)
	symlinkToInsideFile := filepath.Join(subDir, "symlink-to-inside")
	err = os.Symlink(insideFile, symlinkToInsideFile)
	require.NoError(t, err)

	t.Run("SymlinkToOutsideDir", func(t *testing.T) {
		// ValidatePathWithinBase doesn't follow symlinks, so it should pass
		validPath, err := ValidatePathWithinBase(symlinkToOutsideDir, baseDir, true)
		assert.NoError(t, err, "Symlink within base dir should pass basic validation")
		assert.Equal(t, filepath.Clean(symlinkToOutsideDir), filepath.Clean(validPath))

		// But ValidateDirPath should detect that it's a symlink to outside
		// This depends on how the function is implemented:
		// Some implementations may follow symlinks, others may not
		// We document the current behavior (doesn't follow symlinks)
		validPath, err = ValidateDirPath(symlinkToOutsideDir, baseDir, true, true)

		// Current implementation doesn't follow symlinks during validation
		// Uncomment below tests if it's decided that symlink following should be added

		// Symlinks should be resolved in a real security context,
		// but we document current behavior which doesn't follow links
		assert.NoError(t, err, "Current implementation doesn't follow symlinks")
	})

	t.Run("SymlinkToOutsideFile", func(t *testing.T) {
		// ValidatePathWithinBase doesn't follow symlinks, so it should pass
		validPath, err := ValidatePathWithinBase(symlinkToOutsideFile, baseDir, true)
		assert.NoError(t, err, "Symlink within base dir should pass basic validation")
		assert.Equal(t, filepath.Clean(symlinkToOutsideFile), filepath.Clean(validPath))

		// But ValidateFilePath should detect that it's a symlink to outside
		// This depends on how the function is implemented:
		// Some implementations may follow symlinks, others may not
		// We document the current behavior (doesn't follow symlinks)
		validPath, err = ValidateFilePath(symlinkToOutsideFile, baseDir, true, true)

		// Current implementation doesn't follow symlinks during validation
		assert.NoError(t, err, "Current implementation doesn't follow symlinks")
	})

	t.Run("SymlinkToInsideFile", func(t *testing.T) {
		// Symlink to a file within the base directory should be valid
		validPath, err := ValidatePathWithinBase(symlinkToInsideFile, baseDir, true)
		assert.NoError(t, err, "Symlink to a file within base dir should pass validation")
		assert.Equal(t, filepath.Clean(symlinkToInsideFile), filepath.Clean(validPath))

		validPath, err = ValidateFilePath(symlinkToInsideFile, baseDir, true, true)
		assert.NoError(t, err, "Symlink to a file within base dir should pass validation")
		assert.Equal(t, filepath.Clean(symlinkToInsideFile), filepath.Clean(validPath))
	})

	t.Run("AccessThroughSymlink", func(t *testing.T) {
		// Try to access a file through the symlink
		fileViaSymlink := filepath.Join(symlinkToOutsideDir, "outside.txt")

		// ValidatePathWithinBase checks the path string, not the resolved path
		// Since the string starts with baseDir, it should initially pass
		validPath, err := ValidatePathWithinBase(fileViaSymlink, baseDir, true)
		assert.NoError(t, err, "Path string validation doesn't follow symlinks")
		assert.Equal(t, filepath.Clean(fileViaSymlink), filepath.Clean(validPath))

		// But ValidateFilePath should ideally detect the traversal through symlink
		// if it follows symlinks during validation (current implementation doesn't)
		validPath, err = ValidateFilePath(fileViaSymlink, baseDir, true, true)

		// Current implementation doesn't follow symlinks during validation
		// so this test documents that if symlink traversal detection is needed,
		// it must be implemented separately
		if err == nil {
			t.Log("Note: Current implementation doesn't detect traversal through symlinks")
		}
	})
}

func TestEdgeCases(t *testing.T) {
	// Create a temporary base directory for testing
	baseDir := t.TempDir()

	testCases := []struct {
		name        string
		path        string
		baseDir     string
		shouldPass  bool
		description string
	}{
		{
			name:        "EmptyPath",
			path:        "",
			baseDir:     baseDir,
			shouldPass:  false,
			description: "Empty path",
		},
		{
			name:        "EmptyBaseDir",
			path:        baseDir,
			baseDir:     "",
			shouldPass:  false,
			description: "Empty base directory",
		},
		{
			name:        "DotPath",
			path:        ".",
			baseDir:     baseDir,
			shouldPass:  false,
			description: "Path is just a dot",
		},
		{
			name:        "VeryLongPath",
			path:        filepath.Join(baseDir, strings.Repeat("a/", 100)+"file.txt"),
			baseDir:     baseDir,
			shouldPass:  true,
			description: "Very long path with many subdirectories",
		},
		{
			name:        "PathWithSpaces",
			path:        filepath.Join(baseDir, "folder with spaces", "file with spaces.txt"),
			baseDir:     baseDir,
			shouldPass:  true,
			description: "Path with spaces in directory and filename",
		},
		{
			name:        "PathWithSpecialChars",
			path:        filepath.Join(baseDir, "folder-!@#$%^&*()_+", "file-!@#$%^&*()_+.txt"),
			baseDir:     baseDir,
			shouldPass:  true,
			description: "Path with special characters",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ValidatePathWithinBase(tc.path, tc.baseDir, true)

			if tc.shouldPass {
				assert.NoError(t, err, "Valid path should pass: "+tc.description)
				assert.NotEmpty(t, result, "Valid path should return a non-empty result")
			} else {
				assert.Error(t, err, "Invalid path should fail: "+tc.description)
			}
		})
	}
}

func TestNormalizationSecurity(t *testing.T) {
	// Create a temporary base directory for testing
	baseDir := t.TempDir()

	// Test normalization cases
	testCases := []struct {
		name        string
		path        string
		shouldPass  bool
		description string
	}{
		{
			name:        "MultipleSlashes",
			path:        filepath.Join(baseDir, "///subdir///file.txt"),
			shouldPass:  true,
			description: "Path with multiple slashes",
		},
		{
			name:        "DotSlashPrefix",
			path:        filepath.Join(baseDir, "./subdir/./file.txt"),
			shouldPass:  true,
			description: "Path with ./ prefixes",
		},
		{
			name:        "DotDotSlashWithinAllowed",
			path:        filepath.Join(baseDir, "subdir/../subdir/file.txt"),
			shouldPass:  true,
			description: "Path with ../ that stays within the allowed directory",
		},
		{
			name:        "ComplexNormalization",
			path:        filepath.Join(baseDir, "./subdir/..////./subdir/././file.txt"),
			shouldPass:  true,
			description: "Complex path with multiple normalization components",
		},
		{
			name:        "NormalizationTraversal",
			path:        filepath.Join(baseDir, "subdir/../../outside/file.txt"),
			shouldPass:  false,
			description: "Path using normalization to traverse outside",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ValidatePathWithinBase(tc.path, baseDir, true)

			if tc.shouldPass {
				assert.NoError(t, err, "Valid normalized path should pass: "+tc.description)
				assert.NotEmpty(t, result, "Valid path should return a non-empty result")
			} else {
				assert.Error(t, err, "Invalid path should fail: "+tc.description)
				assert.ErrorIs(t, err, ErrPathOutsideBase, "Error should be ErrPathOutsideBase")
			}
		})
	}
}
