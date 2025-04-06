// Package filesystem provides functionality for scanning, reading, and managing
// filesystem operations in the glance application.
package filesystem

import (
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

// LatestModTime finds the most recent modification time of any file or directory
// in the specified directory (recursively searched).
//
// Parameters:
//   - dir: The directory to search for the latest modification time
//   - ignoreChain: A chain of gitignore matchers to check for ignored files/directories
//   - verbose: Whether to log verbose debug information
//
// Returns:
//   - The most recent modification time found
//   - An error, if any occurred during the search
func LatestModTime(dir string, ignoreChain IgnoreChain, verbose bool) (time.Time, error) {
	var latest time.Time

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, werr error) error {
		if werr != nil {
			return werr
		}

		// For directories (except the root dir), check if we should skip them
		if d.IsDir() && path != dir {
			// Check if the directory should be ignored
			if ShouldIgnoreDir(path, dir, ignoreChain, verbose) {
				return fs.SkipDir
			}
		}

		// Get file info for modification time
		info, errStat := d.Info()
		if errStat != nil {
			if verbose && logrus.IsLevelEnabled(logrus.DebugLevel) {
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

// ShouldRegenerate determines if a GLANCE.md file in a directory needs to be regenerated.
// Regeneration is needed if:
// - Force is true
// - GLANCE.md doesn't exist
// - Any file in the directory is newer than GLANCE.md
//
// Parameters:
//   - dir: The directory to check for regeneration need
//   - globalForce: Whether regeneration is forced globally
//   - ignoreChain: A chain of gitignore matchers to check for ignored files/directories
//   - verbose: Whether to log verbose debug information
//
// Returns:
//   - true if regeneration is needed, false otherwise
//   - an error, if any occurred during the check
func ShouldRegenerate(dir string, globalForce bool, ignoreChain IgnoreChain, verbose bool) (bool, error) {
	// Always regenerate if force is true
	if globalForce {
		if verbose && logrus.IsLevelEnabled(logrus.DebugLevel) {
			logrus.Debugf("Force regeneration for %s", dir)
		}
		return true, nil
	}

	// Check if GLANCE.md exists
	glancePath := filepath.Join(dir, GlanceFilename)
	glanceInfo, err := os.Stat(glancePath)
	if err != nil {
		if verbose && logrus.IsLevelEnabled(logrus.DebugLevel) {
			logrus.Debugf("GLANCE.md not found in %s, will generate", dir)
		}
		return true, nil
	}

	// Check if any file is newer than GLANCE.md
	latest, err := LatestModTime(dir, ignoreChain, verbose)
	if err != nil {
		return false, err
	}

	if latest.After(glanceInfo.ModTime()) {
		if verbose && logrus.IsLevelEnabled(logrus.DebugLevel) {
			logrus.Debugf("Found newer files in %s, will regenerate GLANCE.md", dir)
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
