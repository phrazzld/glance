// Package filesystem provides functionality for scanning, reading, and managing
// filesystem operations in the glance application.
package filesystem

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// DefaultFileMode defines the file permission mode for files created by the application.
// Value 0o600 (rw-------) ensures files are only readable and writable by the owner
// and not accessible to group members or other users. This is important for security
// as glance.md files may contain sensitive code analysis or information derived
// from private code repositories.
const DefaultFileMode = 0o600

// LatestModTime finds the most recent modification time of any file or directory
// in the specified directory (recursively searched).
//
// Parameters:
//   - dir: The directory to search for the latest modification time
//   - ignoreChain: A chain of gitignore matchers to check for ignored files/directories
//
// Returns:
//   - The most recent modification time found
//   - An error, if any occurred during the search
func LatestModTime(dir string, ignoreChain IgnoreChain) (time.Time, error) {
	var latest time.Time

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, werr error) error {
		if werr != nil {
			return werr
		}

		// For directories (except the root dir), check if we should skip them
		if d.IsDir() && path != dir {
			// Check if the directory should be ignored
			if ShouldIgnoreDir(path, dir, ignoreChain) {
				return fs.SkipDir
			}
		}

		// Get file info for modification time
		info, errStat := d.Info()
		if errStat != nil {
			if logrus.IsLevelEnabled(logrus.DebugLevel) {
				logrus.Debugf("Error getting file info for %s: %v", path, errStat)
			}
			return nil
		}

		// Update latest time if this file/dir is newer
		if info.ModTime().After(latest) {
			latest = info.ModTime()
		}

		return nil
	})

	return latest, err
}

// ShouldRegenerate determines if a glance.md file in a directory needs to be regenerated.
// Regeneration is needed if:
// - Force is true
// - glance.md doesn't exist
// - Any file in the directory is newer than glance.md
//
// Parameters:
//   - dir: The directory to check for regeneration need
//   - globalForce: Whether regeneration is forced globally
//   - ignoreChain: A chain of gitignore matchers to check for ignored files/directories
//
// Returns:
//   - true if regeneration is needed, false otherwise
//   - an error, if any occurred during the check
func ShouldRegenerate(dir string, globalForce bool, ignoreChain IgnoreChain) (bool, error) {
	// Always regenerate if force is true
	if globalForce {
		if logrus.IsLevelEnabled(logrus.DebugLevel) {
			logrus.Debugf("Force regeneration for %s", dir)
		}
		return true, nil
	}

	// Check if glance.md exists
	glancePath := filepath.Join(dir, GlanceFilename)
	glanceInfo, err := os.Stat(glancePath)
	if err != nil {
		if logrus.IsLevelEnabled(logrus.DebugLevel) {
			logrus.Debugf("glance.md not found in %s, will generate", dir)
		}
		return true, nil
	}

	// Check if any file is newer than glance.md
	latest, err := LatestModTime(dir, ignoreChain)
	if err != nil {
		return false, err
	}

	if latest.After(glanceInfo.ModTime()) {
		if logrus.IsLevelEnabled(logrus.DebugLevel) {
			logrus.Debugf("Found newer files in %s, will regenerate glance.md", dir)
		}
		return true, nil
	}

	return false, nil
}

// BubbleUpParents marks all parent directories of a given directory for regeneration,
// up to but not including the root directory.
//
// Parameters:
//   - dir: The starting directory whose parents should be marked
//   - root: The root directory (ancestors of this won't be marked)
//   - needs: A map to track which directories need regeneration
func BubbleUpParents(dir, root string, needs map[string]bool) {
	for {
		parent := filepath.Dir(dir)

		// Stop if we've reached the top directory
		// or if we've gone above the root directory
		if parent == dir || len(parent) < len(root) {
			break
		}

		// Mark this parent directory as needing regeneration, but only if it's not the root
		if parent != root {
			needs[parent] = true
		}

		// Move up to the next parent
		dir = parent
	}
}

// ErrPathOutsideBase indicates a path is outside the allowed base directory
var ErrPathOutsideBase = errors.New("path is outside of allowed base directory")

// ErrInvalidPath indicates a general path validation error
var ErrInvalidPath = errors.New("invalid path")

// ErrNotDirectory indicates a path exists but is not a directory
var ErrNotDirectory = errors.New("path is not a directory")

// ErrNotFile indicates a path exists but is not a file
var ErrNotFile = errors.New("path is not a file")

// ValidatePathWithinBase checks if a path is strictly contained within a base directory.
// It normalizes and absolutizes the path, then verifies it doesn't escape the base directory.
//
// Parameters:
//   - path: The path to validate
//   - baseDir: The base directory that the path must be contained within. MUST be non-empty.
//   - allowBaseDir: Whether the base directory itself is an acceptable path (true) or only subdirectories (false)
//
// Returns:
//   - The cleaned, absolute path if valid
//   - An error if the path is invalid or outside the base directory
func ValidatePathWithinBase(path, baseDir string, allowBaseDir bool) (string, error) {
	// Step 1: Clean the path to normalize it
	cleanPath := filepath.Clean(path)

	// Step 2: Convert to absolute path
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidPath, err)
	}

	// Require a non-empty baseDir for proper validation
	if baseDir == "" {
		return "", errors.New("baseDir cannot be empty for validation")
	}

	// Get absolute base directory if it's not already
	absBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		return "", fmt.Errorf("invalid base directory: %w", err)
	}

	// Step 3: Check path is within allowed boundary
	if !allowBaseDir && absPath == absBaseDir {
		return "", fmt.Errorf("%w: path %q cannot be the base directory %q",
			ErrPathOutsideBase, path, baseDir)
	}

	// Check if the path starts with the base directory
	if !strings.HasPrefix(absPath, absBaseDir+string(os.PathSeparator)) && absPath != absBaseDir {
		return "", fmt.Errorf("%w: path %q is outside of allowed directory %q",
			ErrPathOutsideBase, path, baseDir)
	}

	return absPath, nil
}

// ValidateFilePath checks if a path exists, is a file (not a directory), and is under the base directory.
// It fully validates the path, including normalization, absolutization, and containment verification.
//
// Parameters:
//   - path: The file path to validate
//   - baseDir: The base directory that the path must be contained within. MUST be non-empty.
//   - allowBaseDir: Whether the base directory itself is an acceptable path
//   - mustExist: Whether the file must exist (true) or not (false)
//
// Returns:
//   - The cleaned, absolute path if valid
//   - An error if the path is invalid, outside the base directory, or not a file
func ValidateFilePath(path, baseDir string, allowBaseDir, mustExist bool) (string, error) {
	// First validate the general path constraints
	absPath, err := ValidatePathWithinBase(path, baseDir, allowBaseDir)
	if err != nil {
		return "", err
	}

	// If the file must exist, check it exists and is a file
	if mustExist {
		info, err := os.Stat(absPath)
		if err != nil {
			return "", fmt.Errorf("cannot access path %q: %w", path, err)
		}

		if info.IsDir() {
			return "", fmt.Errorf("%w: path %q is a directory, expected a file",
				ErrNotFile, path)
		}
	}

	return absPath, nil
}

// ValidateDirPath checks if a path exists, is a directory, and is under the base directory.
// It fully validates the path, including normalization, absolutization, and containment verification.
//
// Parameters:
//   - path: The directory path to validate
//   - baseDir: The base directory that the path must be contained within. MUST be non-empty.
//   - allowBaseDir: Whether the base directory itself is an acceptable path
//   - mustExist: Whether the directory must exist (true) or not (false)
//
// Returns:
//   - The cleaned, absolute path if valid
//   - An error if the path is invalid, outside the base directory, or not a directory
func ValidateDirPath(path, baseDir string, allowBaseDir, mustExist bool) (string, error) {
	// First validate the general path constraints
	absPath, err := ValidatePathWithinBase(path, baseDir, allowBaseDir)
	if err != nil {
		return "", err
	}

	// If the directory must exist, check it exists and is a directory
	if mustExist {
		info, err := os.Stat(absPath)
		if err != nil {
			return "", fmt.Errorf("cannot access path %q: %w", path, err)
		}

		if !info.IsDir() {
			return "", fmt.Errorf("%w: path %q is not a directory",
				ErrNotDirectory, path)
		}
	}

	return absPath, nil
}
