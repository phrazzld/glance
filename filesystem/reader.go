// Package filesystem provides functionality for scanning, reading, and managing
// filesystem operations in the glance application.
package filesystem

import (
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

// MaxDefaultFileSize is the default maximum file size in bytes for file reading (5MB)
const MaxDefaultFileSize = 5 * 1024 * 1024

// ReadTextFile reads a file at the given path and returns its contents as a string.
// It validates UTF-8 encoding and handles errors.
//
// Parameters:
//   - path: The absolute path to the file to read
//   - maxBytes: The maximum number of bytes to read (0 for unlimited)
//
// Returns:
//   - The contents of the file as a string
//   - An error, if any occurred during reading
func ReadTextFile(path string, maxBytes int64) (string, error) {
	// #nosec G304 -- This is a core function to read files by path, security validation happens at caller level
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	// Validate UTF-8 by replacing invalid sequences with the replacement character
	contentStr := strings.ToValidUTF8(string(content), "�")

	// Truncate if needed
	if maxBytes > 0 && int64(len(contentStr)) > maxBytes {
		contentStr = TruncateContent(contentStr, maxBytes)
	}

	return contentStr, nil
}

// TruncateContent truncates a string to a maximum size in bytes and
// adds an indicator that the content was truncated.
//
// Parameters:
//   - content: The string to truncate
//   - maxBytes: The maximum number of bytes to keep
//
// Returns:
//   - The truncated string with an indicator
func TruncateContent(content string, maxBytes int64) string {
	// If maxBytes is 0 or negative, return the full content (no truncation)
	if maxBytes <= 0 {
		return content
	}

	// If content is shorter than the max, return the full content
	if int64(len(content)) <= maxBytes {
		return content
	}

	// Otherwise, truncate and add indicator
	return content[:maxBytes] + "...(truncated)"
}

// IsTextFile checks if a file's content type indicates it is a text-based file
// by reading its first 512 bytes.
//
// Parameters:
//   - path: The path to the file to check
//
// Returns:
//   - true if the file appears to be text-based, false otherwise
//   - an error, if any occurred during the check
func IsTextFile(path string) (bool, error) {
	// #nosec G304 -- This is a core function to open files by path, security validation happens at caller level
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	// Handle Close error properly
	defer func() {
		_ = f.Close() // explicitly ignore the error as we're in a read-only context
	}()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return false, err
	}

	ctype := http.DetectContentType(buf[:n])
	if strings.HasPrefix(ctype, "text/") ||
		strings.HasPrefix(ctype, "application/json") ||
		strings.HasPrefix(ctype, "application/xml") ||
		strings.Contains(ctype, "yaml") {
		return true, nil
	}

	return false, nil
}

// GatherLocalFiles reads immediate files in a directory and returns a map of
// relative path to file content for text-based files.
//
// Parameters:
//   - dir: The directory to scan for files
//   - ignoreChain: A chain of gitignore matchers to check for ignored files
//   - maxFileBytes: The maximum number of bytes to read from each file
//   - verbose: Whether to log verbose debug information
//
// Returns:
//   - A map of relative file paths to their contents as strings
//   - An error, if any occurred during scanning or reading
func GatherLocalFiles(dir string, ignoreChain []IgnoreRule, maxFileBytes int64, verbose bool) (map[string]string, error) {
	files := make(map[string]string)

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, werr error) error {
		if werr != nil {
			return werr
		}

		// Skip subdirectories (beyond the current dir)
		if d.IsDir() && path != dir {
			return fs.SkipDir
		}

		// Skip directories, glance.md, and hidden files
		if d.IsDir() || d.Name() == "glance.md" || strings.HasPrefix(d.Name(), ".") {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			if verbose && logrus.IsLevelEnabled(logrus.DebugLevel) {
				logrus.Debugf("Error calculating relative path for %s from %s: %v",
					path, dir, err)
			}
			return nil
		}

		// Check if the file should be ignored
		shouldInclude := true
		for _, rule := range ignoreChain {
			// Skip rules from directories that are not ancestors of the current path
			if !strings.HasPrefix(dir, rule.OriginDir) {
				continue
			}

			// Get the path relative to the rule's origin
			ruleRelPath, err := filepath.Rel(rule.OriginDir, path)
			if err != nil {
				if verbose && logrus.IsLevelEnabled(logrus.DebugLevel) {
					logrus.Debugf("Error calculating relative path for %s from %s: %v",
						path, rule.OriginDir, err)
				}
				continue
			}

			// Convert to slash path for consistent matching
			ruleRelPath = filepath.ToSlash(ruleRelPath)

			if rule.Matcher.MatchesPath(ruleRelPath) {
				shouldInclude = false
				if verbose && logrus.IsLevelEnabled(logrus.DebugLevel) {
					logrus.Debugf("Ignoring file via .gitignore chain: %s", relPath)
				}
				break
			}
		}

		if !shouldInclude {
			return nil
		}

		// Check if file is text-based
		isText, errCheck := IsTextFile(path)
		if errCheck != nil && verbose && logrus.IsLevelEnabled(logrus.DebugLevel) {
			logrus.Debugf("Error checking if file is text: %s => %v", path, errCheck)
		}

		if !isText {
			if verbose && logrus.IsLevelEnabled(logrus.DebugLevel) {
				logrus.Debugf("📊 Skipping binary/non-text file: %s", path)
			}
			return nil
		}

		// Read file content
		content, err := ReadTextFile(path, maxFileBytes)
		if err != nil {
			if verbose && logrus.IsLevelEnabled(logrus.DebugLevel) {
				logrus.Debugf("Error reading file %s: %v", path, err)
			}
			return nil
		}

		files[relPath] = content
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}
