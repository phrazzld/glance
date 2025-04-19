package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/joho/godotenv" // Used by the config package for loading environment variables
	gitignore "github.com/sabhiram/go-gitignore"
	"github.com/sirupsen/logrus"

	"glance/config"
	"glance/filesystem"
	"glance/llm"
	"glance/ui"
)

// -----------------------------------------------------------------------------
// type definitions
// -----------------------------------------------------------------------------

// result tracks per-directory summarization outcomes.
type result struct {
	dir      string
	attempts int
	success  bool
	err      error
}

// queueItem is used for BFS directory scanning.
type queueItem struct {
	path        string
	ignoreChain []*gitignore.GitIgnore
}

// Removed the promptData struct - this is now part of the llm package

// -----------------------------------------------------------------------------
// main
// -----------------------------------------------------------------------------

func main() {
	// Load configuration from command-line flags, environment variables, etc.
	cfg, err := config.LoadConfig(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Set up logging based on the verbose flag
	setupLogging(cfg.Verbose)

	// Set up the LLM client and service
	llmClient, llmService, err := setupLLMService(cfg)
	if err != nil {
		logrus.Fatalf("🚫 Failed to initialize LLM service: %v", err)
	}
	defer llmClient.Close()

	// Scan directories and process them to generate glance.md files
	dirs, ignoreChains, err := scanDirectories(cfg)
	if err != nil {
		logrus.Fatalf("🚫 Directory scan failed: %v - Check file permissions and disk space", err)
	}

	// Process directories and generate glance.md files
	results := processDirectories(dirs, ignoreChains, cfg, llmService)

	// Print summary of results
	printDebrief(results)
}

// -----------------------------------------------------------------------------
// Main function components
// -----------------------------------------------------------------------------

// setupLogging configures the logger based on the verbose flag
func setupLogging(verbose bool) {
	// Set log level based on verbose flag
	if verbose {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	// Configure formatter with custom settings
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:    true,
		ForceColors:      true,
		TimestampFormat:  "2006-01-02 15:04:05",
		DisableTimestamp: false,
		PadLevelText:     true,
		ForceQuote:       false,
		DisableSorting:   true,
		DisableColors:    false,
	})
}

// setupLLMService initializes the LLM client and service
func setupLLMService(cfg *config.Config) (llm.Client, *llm.Service, error) {
	// Create client options
	clientOptions := llm.DefaultClientOptions().
		WithModelName("gemini-2.5-flash-preview-04-17").
		WithMaxRetries(cfg.MaxRetries).
		WithTimeout(60)

	// Create the client
	client, err := llm.NewGeminiClient(cfg.APIKey, clientOptions)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create LLM client: %w", err)
	}

	// Create service options
	serviceOptions := []llm.ServiceOption{
		llm.WithMaxRetries(cfg.MaxRetries),
		llm.WithVerbose(cfg.Verbose),
	}

	// Create the service
	service, err := llm.NewService(client, serviceOptions...)
	if err != nil {
		client.Close()
		return nil, nil, fmt.Errorf("failed to create LLM service: %w", err)
	}

	return client, service, nil
}

// scanDirectories performs BFS scanning and gathers .gitignore chain info per directory
func scanDirectories(cfg *config.Config) ([]string, map[string][]*gitignore.GitIgnore, error) {
	logrus.Info("✨ Excellent! Scanning directories now... Let's explore your code!")

	// Show a spinner while scanning
	scanner := ui.NewScanner()
	scanner.Start()
	defer scanner.Stop()

	// Perform BFS scanning and gather .gitignore chain info per directory
	dirsList, dirToIgnoreChain, err := listAllDirsWithIgnores(cfg.TargetDir)
	if err != nil {
		return nil, nil, err
	}

	// Process from deepest subdirectories upward
	reverseSlice(dirsList)

	return dirsList, dirToIgnoreChain, nil
}

// processDirectories generates glance.md files for each directory in the list
func processDirectories(dirsList []string, dirToIgnoreChain map[string][]*gitignore.GitIgnore, cfg *config.Config, llmService *llm.Service) []result {
	logrus.Info("🧠 Preparing to generate all glance.md files... Getting ready to make your code shine!")

	// Create progress bar
	bar := ui.NewProcessor(len(dirsList))

	needsRegen := make(map[string]bool)
	var finalResults []result

	// Process each directory
	for _, d := range dirsList {
		ignoreChain := dirToIgnoreChain[d]

		// Check if we need to regenerate the glance.md file
		forceDir, errCheck := shouldRegenerate(d, cfg.Force, ignoreChain)
		if errCheck != nil && cfg.Verbose {
			logrus.Warnf("⏱️ Couldn't check modification time for %s: %v", d, errCheck)
		}

		forceDir = forceDir || needsRegen[d]

		// Process the directory with retry logic
		r := processDirectory(d, forceDir, ignoreChain, cfg, llmService)
		finalResults = append(finalResults, r)

		if err := bar.Increment(); err != nil {
			logrus.Warnf("Failed to increment progress bar: %v", err)
		}

		// Bubble up parent's regeneration flag if needed
		if r.success && r.attempts > 0 && forceDir {
			bubbleUpParents(d, cfg.TargetDir, needsRegen)
		}
	}

	fmt.Println()
	logrus.Infof("🎯 All done! glance.md files have been generated for your codebase up to: %s", cfg.TargetDir)

	return finalResults
}

// processDirectory processes a single directory with retry logic
func processDirectory(dir string, forceDir bool, ignoreChain []*gitignore.GitIgnore, cfg *config.Config, llmService *llm.Service) result {
	r := result{dir: dir}

	glancePath := filepath.Join(dir, "glance.md")
	fileExists := false

	// Check if the file exists (and remember the result)
	if _, err := os.Stat(glancePath); err == nil {
		fileExists = true
	}

	// Skip if glance.md exists and we're not forcing regeneration
	if fileExists && !forceDir && !cfg.Force {
		if cfg.Verbose {
			logrus.Debugf("⏩ Skipping %s (glance.md already exists and looks fresh)", dir)
		}
		r.success = true
		r.attempts = 0 // Explicitly mark that we didn't attempt to regenerate
		return r
	}

	// Gather data for glance.md generation
	subdirs, err := readSubdirectories(dir, ignoreChain)
	if err != nil {
		r.err = err
		return r
	}
	subGlances, err := gatherSubGlances(subdirs)
	if err != nil {
		r.err = fmt.Errorf("gatherSubGlances failed: %w", err)
		return r
	}
	fileContents, err := gatherLocalFiles(dir, ignoreChain, cfg.MaxFileBytes, cfg.Verbose)
	if err != nil {
		r.err = fmt.Errorf("gatherLocalFiles failed: %w", err)
		return r
	}

	if cfg.Verbose {
		logrus.Debugf("📊 Processing %s → Found %d subdirs, %d sub-glances, %d local files",
			dir, len(subdirs), len(subGlances), len(fileContents))
	}

	// Create context for LLM operations
	ctx := context.Background()

	// Generate markdown content using the LLM service
	summary, llmErr := llmService.GenerateGlanceMarkdown(ctx, dir, fileContents, subGlances)
	if llmErr != nil {
		r.attempts = 1 // Service already handles retries internally
		r.err = llmErr
		return r
	}

	// Validate the glance.md path before writing
	validatedPath, pathErr := filesystem.ValidateFilePath(glancePath, dir, true, false)
	if pathErr != nil {
		r.err = fmt.Errorf("invalid glance.md path for %s: %w", dir, pathErr)
		return r
	}
	
	// Write the generated content to file using the validated path
	if werr := os.WriteFile(validatedPath, []byte(summary), 0o600); werr != nil { // #nosec G306 -- Changed to 0600 for security & path validated
		r.err = fmt.Errorf("failed writing glance.md to %s: %w", dir, werr)
		return r
	}

	r.success = true
	r.attempts = 1 // Service already handles retries internally
	r.err = nil
	return r
}

// Removed the generateMarkdown function - this functionality is now handled by the LLM service

// -----------------------------------------------------------------------------
// .gitignore scanning and BFS
// -----------------------------------------------------------------------------

// listAllDirsWithIgnores performs a BFS from `root`, collecting subdirectories
// and merging each directory's .gitignore with its parent's chain.
func listAllDirsWithIgnores(root string) ([]string, map[string][]*gitignore.GitIgnore, error) {
	var dirsList []string

	// BFS queue
	queue := []queueItem{
		{path: root, ignoreChain: nil},
	}

	// map of directory -> chain of .gitignore objects
	dirToChain := make(map[string][]*gitignore.GitIgnore)
	dirToChain[root] = nil

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		dirsList = append(dirsList, current.path)

		// load .gitignore in current.path, if exists
		localIgnore, _ := loadGitignore(current.path)

		// Create a copy of the parent chain to avoid modifying it
		var combinedChain []*gitignore.GitIgnore
		if len(current.ignoreChain) > 0 {
			combinedChain = make([]*gitignore.GitIgnore, len(current.ignoreChain))
			copy(combinedChain, current.ignoreChain)
		}

		// Add the local .gitignore if it exists
		if localIgnore != nil {
			combinedChain = append(combinedChain, localIgnore)
		}

		dirToChain[current.path] = combinedChain

		entries, err := os.ReadDir(current.path)
		if err != nil {
			return nil, nil, err
		}

		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			name := e.Name()

			// skip hidden or heavy directories
			if strings.HasPrefix(name, ".") || name == "node_modules" {
				continue
			}

			fullChildPath := filepath.Join(current.path, name)
			rel, _ := filepath.Rel(root, fullChildPath)

			// Check if this directory should be ignored by any gitignore rules
			if isIgnored(rel, combinedChain) {
				if logrus.IsLevelEnabled(logrus.DebugLevel) {
					logrus.Debugf("skipping %s because of .gitignore match", rel)
				}
				continue
			}

			queue = append(queue, queueItem{
				path:        fullChildPath,
				ignoreChain: combinedChain,
			})
		}
	}

	return dirsList, dirToChain, nil
}

// loadGitignore parses the .gitignore file in a directory.
func loadGitignore(dir string) (*gitignore.GitIgnore, error) {
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

// isIgnored checks a path against a chain of .gitignore patterns.
func isIgnored(rel string, chain []*gitignore.GitIgnore) bool {
	for _, ig := range chain {
		if ig == nil {
			continue
		}
		// For directories, test both with and without trailing slash
		// as gitignore patterns like "dir/" only match "dir/" and not "dir"
		if ig.MatchesPath(rel) || ig.MatchesPath(rel+"/") {
			return true
		}
	}
	return false
}

// reverseSlice reverses a slice of directory paths in-place.
func reverseSlice(s []string) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

// -----------------------------------------------------------------------------
// file collection and processing
// -----------------------------------------------------------------------------

// gatherSubGlances merges the contents of existing subdirectory glance.md files.
// It validates all file paths to prevent path traversal vulnerabilities.
func gatherSubGlances(subdirs []string) (string, error) {
	var combined []string
	for _, sd := range subdirs {
		// Validate the subdirectory first
		validDir, err := filesystem.ValidateDirPath(sd, "", true, true)
		if err != nil {
			logrus.Warnf("Skipping invalid subdirectory for glance.md collection: %v", err)
			continue
		}
		
		// Construct and validate the glance.md path
		glancePath := filepath.Join(validDir, "glance.md")
		validPath, err := filesystem.ValidateFilePath(glancePath, validDir, true, true)
		if err != nil {
			logrus.Debugf("Skipping invalid glance.md path: %v", err)
			continue
		}
		
		// Read the validated file path
		// #nosec G304 -- Reading validated glance.md files from subdirectories is core functionality
		data, err := os.ReadFile(validPath)
		if err == nil {
			combined = append(combined, strings.ToValidUTF8(string(data), "�"))
		}
	}
	return strings.Join(combined, "\n\n"), nil
}

// readSubdirectories lists immediate subdirectories in a directory, skipping hidden or ignored ones.
// It validates the directory path to prevent path traversal vulnerabilities.
func readSubdirectories(dir string, ignoreChain []*gitignore.GitIgnore) ([]string, error) {
	// Validate the directory path first
	validDir, err := filesystem.ValidateDirPath(dir, "", true, true)
	if err != nil {
		return nil, fmt.Errorf("invalid directory path: %w", err)
	}
	
	entries, err := os.ReadDir(validDir)
	if err != nil {
		return nil, err
	}
	var subdirs []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasPrefix(name, ".") || name == "node_modules" {
			continue
		}
		fullPath := filepath.Join(validDir, name)
		
		// Validate the subdirectory path
		validPath, err := filesystem.ValidateDirPath(fullPath, validDir, true, true)
		if err != nil {
			logrus.Debugf("Skipping invalid subdirectory: %v", err)
			continue
		}
		
		rel, _ := filepath.Rel(validDir, validPath)
		if isIgnored(rel, ignoreChain) {
			if logrus.IsLevelEnabled(logrus.DebugLevel) {
				logrus.Debugf("🙈 Ignoring subdirectory (matched .gitignore pattern): %s", rel)
			}
			continue
		}
		subdirs = append(subdirs, fullPath)
	}
	return subdirs, nil
}

// gatherLocalFiles reads immediate files in a directory (excluding glance.md, hidden files, etc.).
// It validates all file paths to prevent path traversal vulnerabilities.
func gatherLocalFiles(dir string, ignoreChain []*gitignore.GitIgnore, maxFileBytes int64, verbose bool) (map[string]string, error) {
	// Validate the base directory first
	validDir, err := filesystem.ValidateDirPath(dir, "", true, true)
	if err != nil {
		return nil, fmt.Errorf("invalid directory path for file gathering: %w", err)
	}
	
	files := make(map[string]string)
	err = filepath.WalkDir(validDir, func(path string, d fs.DirEntry, werr error) error {
		if werr != nil {
			return werr
		}
		// skip subdirectories (beyond the current dir)
		if d.IsDir() && path != validDir {
			return fs.SkipDir
		}
		if d.IsDir() || d.Name() == "glance.md" || strings.HasPrefix(d.Name(), ".") {
			return nil
		}

		// Validate the file path
		validPath, pathErr := filesystem.ValidateFilePath(path, validDir, true, true)
		if pathErr != nil {
			if verbose {
				logrus.Debugf("Skipping invalid file path: %v", pathErr)
			}
			return nil
		}

		rel, _ := filepath.Rel(validDir, validPath)
		if isIgnored(rel, ignoreChain) {
			if verbose {
				logrus.Debugf("ignoring file via .gitignore chain: %s", rel)
			}
			return nil
		}

		// Now use our filesystem package functions for reading files and checking content type
		isText, errCheck := filesystem.IsTextFile(validPath, validDir)
		if errCheck != nil && verbose {
			logrus.Debugf("error checking if file is text: %s => %v", validPath, errCheck)
		}
		if !isText {
			if verbose {
				logrus.Debugf("📊 Skipping binary/non-text file: %s", validPath)
			}
			return nil
		}
		
		// Read the file with validation
		contentStr, err := filesystem.ReadTextFile(validPath, maxFileBytes, validDir)
		if err != nil {
			if verbose {
				logrus.Debugf("Error reading file %s: %v", validPath, err)
			}
			return nil
		}
		
		files[rel] = contentStr
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

// Note: We now use filesystem.IsTextFile instead of this local function
// which provides path validation

// -----------------------------------------------------------------------------
// regeneration logic and utilities
// -----------------------------------------------------------------------------

func shouldRegenerate(dir string, globalForce bool, ignoreChain []*gitignore.GitIgnore) (bool, error) {
	if globalForce {
		return true, nil
	}

	glancePath := filepath.Join(dir, "glance.md")
	glanceInfo, err := os.Stat(glancePath)
	if err != nil {
		return true, nil
	}

	latest, err := latestModTime(dir, ignoreChain)
	if err != nil {
		return false, err
	}

	if latest.After(glanceInfo.ModTime()) {
		return true, nil
	}
	return false, nil
}

func latestModTime(dir string, ignoreChain []*gitignore.GitIgnore) (time.Time, error) {
	var latest time.Time
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, werr error) error {
		if werr != nil {
			return werr
		}
		if d.IsDir() && path != dir {
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" {
				return filepath.SkipDir
			}
			rel, _ := filepath.Rel(dir, path)
			if isIgnored(rel, ignoreChain) {
				return filepath.SkipDir
			}
		}
		info, errStat := d.Info()
		if errStat != nil {
			return nil
		}
		if info.ModTime().After(latest) {
			latest = info.ModTime()
		}
		return nil
	})
	return latest, err
}

func bubbleUpParents(dir, root string, needs map[string]bool) {
	for {
		parent := filepath.Dir(dir)
		if parent == dir || len(parent) < len(root) {
			break
		}
		needs[parent] = true
		dir = parent
	}
}

// -----------------------------------------------------------------------------
// utility functions
// -----------------------------------------------------------------------------

// Removed loadPromptTemplate - this functionality is now handled by the llm package

// -----------------------------------------------------------------------------
// results reporting
// -----------------------------------------------------------------------------

// printDebrief displays a summary of successes and failures.
func printDebrief(results []result) {
	var totalSuccess, totalFailed int
	for _, r := range results {
		if r.success {
			totalSuccess++
		} else {
			totalFailed++
		}
	}
	logrus.Info("📊 === FINAL SUMMARY === 📊")
	logrus.Infof("🔢 Processed %d directories → %d successes, %d failures", len(results), totalSuccess, totalFailed)

	if totalFailed == 0 {
		logrus.Info("🌟 Perfect run! No failures detected. Your codebase is now well-documented!")
		return
	}

	logrus.Info("⚠️ Some directories couldn't be processed:")
	for _, r := range results {
		if !r.success {
			// Use the UI error reporting
			ui.ReportError(r.err, true, fmt.Sprintf("Failed to process %s (attempts: %d)", r.dir, r.attempts))
		}
	}
	logrus.Info("📊 ===================== 📊")
}
