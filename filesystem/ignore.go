// Package filesystem provides functionality for scanning, reading, and managing
// filesystem operations in the glance application.
package filesystem

import (
	"path/filepath"
	"strings"
)

// Constants for default ignore patterns
const (
	// GlanceFilename is the standard filename for glance summaries
	GlanceFilename = "glance.md"

	// NodeModulesDir is a heavy directory that should be skipped by default
	NodeModulesDir = "node_modules"
)

// ShouldIgnoreFile determines if a file should be ignored during processing.
// A file is ignored if:
// - It's a hidden file (name starts with ".")
// - It's named glance.md (to avoid processing our own output files)
// - It matches any gitignore rule in the provided chain
//
// Parameters:
//   - path: The absolute path to the file
//   - baseDir: The base directory relative to which the file is being evaluated
//   - ignoreChain: A chain of gitignore matchers to check for ignored files
//
// Returns:
//   - true if the file should be ignored, false otherwise
func ShouldIgnoreFile(path string, baseDir string, ignoreChain IgnoreChain) bool {
	// Get the file name without the path
	filename := filepath.Base(path)

	// Always ignore hidden files
	if strings.HasPrefix(filename, ".") {
		log.Debugf("Ignoring hidden file: %s", path)
		return true
	}

	// Always ignore glance.md files (our output files)
	if filename == GlanceFilename {
		log.Debugf("Ignoring glance.md file: %s", path)
		return true
	}

	// Check gitignore rules
	if MatchesGitignore(path, baseDir, ignoreChain, false) {
		return true
	}

	return false
}

// ShouldIgnoreDir determines if a directory should be ignored during processing.
// A directory is ignored if:
// - It's a hidden directory (name starts with ".")
// - It's a node_modules directory
// - It matches any gitignore rule in the provided chain
//
// Parameters:
//   - path: The absolute path to the directory
//   - baseDir: The base directory relative to which the directory is being evaluated
//   - ignoreChain: A chain of gitignore matchers to check for ignored directories
//
// Returns:
//   - true if the directory should be ignored, false otherwise
func ShouldIgnoreDir(path string, baseDir string, ignoreChain IgnoreChain) bool {
	// Get the directory name without the path
	dirname := filepath.Base(path)

	// Always ignore hidden directories
	if strings.HasPrefix(dirname, ".") {
		log.Debugf("Ignoring hidden directory: %s", path)
		return true
	}

	// Always ignore node_modules
	if dirname == NodeModulesDir {
		log.Debugf("Ignoring node_modules directory: %s", path)
		return true
	}

	// Check gitignore rules
	if MatchesGitignore(path, baseDir, ignoreChain, true) {
		return true
	}

	return false
}

// MatchesGitignore checks if a path matches any gitignore rule in the provided chain.
//
// Parameters:
//   - path: The absolute path to check
//   - baseDir: The base directory relative to which the path is being evaluated
//   - ignoreChain: A chain of gitignore matchers to check for ignored paths
//   - isDir: Whether the path is a directory (affects matching for patterns with trailing slashes)
//
// Returns:
//   - true if the path matches any gitignore rule, false otherwise
func MatchesGitignore(path string, baseDir string, ignoreChain IgnoreChain, isDir bool) bool {
	// Check if the path matches any gitignore rule in the chain
	for _, rule := range ignoreChain {
		// Skip rules from directories that are not ancestors of the current path
		if !strings.HasPrefix(baseDir, rule.OriginDir) {
			continue
		}

		// Get the path relative to the rule's origin
		relPath, err := filepath.Rel(rule.OriginDir, path)
		if err != nil {
			log.Debugf("Error calculating relative path for %s from %s: %v",
				path, rule.OriginDir, err)
			continue
		}

		// Convert to slash path for consistent matching
		relPath = filepath.ToSlash(relPath)

		// For directories, we need to test both with and without trailing slash
		// because gitignore patterns like "dir/" only match "dir/" and not "dir"
		if isDir {
			if rule.Matcher.MatchesPath(relPath) || rule.Matcher.MatchesPath(relPath+"/") {
				log.Debugf("Path %s matched by gitignore rule from %s", path, rule.OriginDir)
				return true
			}
		} else {
			if rule.Matcher.MatchesPath(relPath) {
				log.Debugf("Path %s matched by gitignore rule from %s", path, rule.OriginDir)
				return true
			}
		}
	}

	return false
}
