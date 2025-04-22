// Package filesystem provides functionality for scanning, reading, and managing
// filesystem operations in the glance application.
package filesystem

import (
	"os"
	"path/filepath"
	"strings"

	gitignore "github.com/sabhiram/go-gitignore"
	"github.com/sirupsen/logrus"
)

// IgnoreRule stores a compiled gitignore matcher and its origin directory.
type IgnoreRule struct {
	OriginDir string // Absolute path to the directory containing the .gitignore file
	Matcher   *gitignore.GitIgnore
}

// IgnoreChain represents the cumulative list of ignore rules applicable to a directory.
type IgnoreChain []IgnoreRule

// queueItem is used for BFS directory scanning.
type queueItem struct {
	path        string
	ignoreChain IgnoreChain
}

// ListDirsWithIgnores performs a BFS from the root directory, collecting subdirectories
// and merging each directory's .gitignore with its parent's chain.
//
// This is the consolidated BFS implementation that handles all directory scanning use cases
// in the application. It uses the shared ignore functions to determine which directories
// should be included or excluded during traversal.
//
// Parameters:
//   - root: The starting directory for the BFS traversal
//
// Returns:
//   - A slice of directory paths
//   - A map of directory path -> chain of ignore rules
//   - An error, if any occurred during directory traversal
func ListDirsWithIgnores(root string) ([]string, map[string]IgnoreChain, error) {
	var dirsList []string

	// BFS queue
	queue := []queueItem{
		{path: root, ignoreChain: IgnoreChain{}},
	}

	// map of directory -> chain of ignore rules
	dirToChain := make(map[string]IgnoreChain)
	dirToChain[root] = IgnoreChain{}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// We always add the root directory
		if current.path == root {
			dirsList = append(dirsList, current.path)
		} else {
			// For non-root directories, use the shared ignore functions to check
			// if the directory should be included
			if !ShouldIgnoreDir(current.path, filepath.Dir(current.path), current.ignoreChain) {
				dirsList = append(dirsList, current.path)
			} else {
				// Skip this directory - don't process its children
				if IsLevelEnabled(logrus.DebugLevel) {
					log.Debugf("Skipping directory %s: matched by ignore rules", current.path)
				}
				continue
			}
		}

		// Load .gitignore in the current directory, if it exists
		localIgnore, err := LoadGitignore(current.path)
		if err != nil && IsLevelEnabled(logrus.DebugLevel) {
			log.Debugf("Error loading .gitignore from %s: %v", current.path, err)
		}

		// Build the combined chain for this directory's children
		// First, copy the parent chain to avoid modifying it
		combinedChain := make(IgnoreChain, len(current.ignoreChain))
		copy(combinedChain, current.ignoreChain)

		// Add the local .gitignore rule if one exists
		if localIgnore != nil {
			newRule := IgnoreRule{
				OriginDir: current.path,
				Matcher:   localIgnore,
			}
			combinedChain = append(combinedChain, newRule)
		}

		// Store the applicable ignore chain for this directory
		dirToChain[current.path] = combinedChain

		// Read and process child directories
		entries, err := os.ReadDir(current.path)
		if err != nil {
			return nil, nil, err
		}

		for _, e := range entries {
			// Skip non-directories
			if !e.IsDir() {
				continue
			}

			name := e.Name()
			fullChildPath := filepath.Join(current.path, name)

			// Use the helper function to check for hidden dirs and node_modules
			// This is an optimization to avoid creating queue items for directories
			// we know will be excluded
			if strings.HasPrefix(name, ".") || name == NodeModulesDir {
				if IsLevelEnabled(logrus.DebugLevel) {
					log.Debugf("Skipping hidden/node_modules directory: %s", fullChildPath)
				}
				continue
			}

			// Queue the directory for processing
			// It will be checked against ignore rules in the next iteration
			queue = append(queue, queueItem{
				path:        fullChildPath,
				ignoreChain: combinedChain,
			})
		}
	}

	return dirsList, dirToChain, nil
}

// LoadGitignore parses the .gitignore file in a directory and returns a GitIgnore object.
// If no .gitignore file exists, it returns nil for both the GitIgnore object and the error.
//
// Parameters:
//   - dir: The directory to check for a .gitignore file
//
// Returns:
//   - A pointer to a GitIgnore object, or nil if no .gitignore file exists
//   - An error, if any occurred during parsing
func LoadGitignore(dir string) (*gitignore.GitIgnore, error) {
	path := filepath.Join(dir, ".gitignore")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}
	g, err := gitignore.CompileIgnoreFile(path)
	if err != nil {
		return nil, err
	}
	return g, nil
}

// The compatibility functions ExtractGitignoreMatchers and CreateIgnoreChain
// have been removed as part of the migration to use IgnoreChain consistently
// throughout the codebase.
