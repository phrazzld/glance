package main

import (
	"context"
	"fmt"
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
// This function now uses filesystem.ListDirsWithIgnores with type conversion for backward compatibility.
func listAllDirsWithIgnores(root string) ([]string, map[string][]*gitignore.GitIgnore, error) {
	// Use the filesystem package function to get the directories and ignore chains
	dirsList, dirToIgnoreChain, err := filesystem.ListDirsWithIgnores(root)
	if err != nil {
		return nil, nil, err
	}
	
	// Convert IgnoreChain to []*gitignore.GitIgnore for backward compatibility
	dirToChain := make(map[string][]*gitignore.GitIgnore, len(dirToIgnoreChain))
	for dir, ignoreChain := range dirToIgnoreChain {
		dirToChain[dir] = filesystem.ExtractGitignoreMatchers(ignoreChain)
	}
	
	return dirsList, dirToChain, nil
}

// loadGitignore parses the .gitignore file in a directory.
// This is now a wrapper around filesystem.LoadGitignore to maintain backward compatibility.
func loadGitignore(dir string) (*gitignore.GitIgnore, error) {
	return filesystem.LoadGitignore(dir)
}

// isIgnored checks a path against a chain of .gitignore patterns.
// This function now uses filesystem.MatchesGitignore with appropriate type conversion.
func isIgnored(rel string, chain []*gitignore.GitIgnore) bool {
	// Convert raw gitignore chain to IgnoreChain
	ignoreChain := filesystem.CreateIgnoreChain(chain, "")
	
	// Use the filesystem package function that provides more comprehensive matching
	return filesystem.MatchesGitignore(rel, "", ignoreChain, true)
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
// This implementation enhances the original by using filesystem.ReadTextFile.
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

		// Use filesystem.ReadTextFile instead of os.ReadFile
		// This provides better validation and UTF-8 handling
		content, err := filesystem.ReadTextFile(validPath, 0, validDir)
		if err == nil {
			combined = append(combined, content)
		}
	}
	return strings.Join(combined, "\n\n"), nil
}

// readSubdirectories lists immediate subdirectories in a directory, skipping hidden or ignored ones.
// This implementation uses filesystem package functions with appropriate filtering.
func readSubdirectories(dir string, ignoreChain []*gitignore.GitIgnore) ([]string, error) {
	// Validate the directory path first
	validDir, err := filesystem.ValidateDirPath(dir, "", true, true)
	if err != nil {
		return nil, fmt.Errorf("invalid directory path: %w", err)
	}
	
	// Convert raw gitignore chain to IgnoreChain
	fsIgnoreChain := filesystem.CreateIgnoreChain(ignoreChain, dir)
	
	// Read directory entries
	entries, err := os.ReadDir(validDir)
	if err != nil {
		return nil, err
	}
	
	// Filter for immediate subdirectories only
	var subdirs []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		
		name := e.Name()
		fullPath := filepath.Join(validDir, name)
		
		// Use the filesystem package for directory filtering
		if filesystem.ShouldIgnoreDir(fullPath, validDir, fsIgnoreChain, logrus.IsLevelEnabled(logrus.DebugLevel)) {
			continue
		}
		
		// Validate the subdirectory path
		validPath, err := filesystem.ValidateDirPath(fullPath, validDir, true, true)
		if err != nil {
			logrus.Debugf("Skipping invalid subdirectory: %v", err)
			continue
		}
		
		subdirs = append(subdirs, validPath)
	}
	return subdirs, nil
}

// gatherLocalFiles reads immediate files in a directory (excluding glance.md, hidden files, etc.).
// This function now uses filesystem.GatherLocalFiles with appropriate type conversion.
func gatherLocalFiles(dir string, ignoreChain []*gitignore.GitIgnore, maxFileBytes int64, verbose bool) (map[string]string, error) {
	// Convert raw gitignore chain to IgnoreChain
	// We need to set the origin directory correctly for proper relative path calculations
	fsIgnoreChain := filesystem.CreateIgnoreChain(ignoreChain, dir)
	
	// Use the filesystem package function that provides more comprehensive validation and handling
	return filesystem.GatherLocalFiles(dir, fsIgnoreChain, maxFileBytes, verbose)
}

// Note: We now use filesystem.IsTextFile instead of this local function
// which provides path validation

// -----------------------------------------------------------------------------
// regeneration logic and utilities
// -----------------------------------------------------------------------------

// shouldRegenerate determines if a glance.md file needs to be regenerated.
// This function now uses filesystem.ShouldRegenerate with appropriate type conversion.
func shouldRegenerate(dir string, globalForce bool, ignoreChain []*gitignore.GitIgnore) (bool, error) {
	// Convert raw gitignore chain to IgnoreChain
	fsIgnoreChain := filesystem.CreateIgnoreChain(ignoreChain, "")
	
	// Use the filesystem package function that provides more comprehensive handling
	return filesystem.ShouldRegenerate(dir, globalForce, fsIgnoreChain, logrus.IsLevelEnabled(logrus.DebugLevel))
}

// latestModTime finds the most recent modification time in a directory.
// This function now uses filesystem.LatestModTime with appropriate type conversion.
func latestModTime(dir string, ignoreChain []*gitignore.GitIgnore) (time.Time, error) {
	// Convert raw gitignore chain to IgnoreChain
	fsIgnoreChain := filesystem.CreateIgnoreChain(ignoreChain, "")
	
	// Use the filesystem package function that provides more comprehensive handling
	return filesystem.LatestModTime(dir, fsIgnoreChain, logrus.IsLevelEnabled(logrus.DebugLevel))
}

// bubbleUpParents marks parent directories for regeneration.
// This is now a wrapper around filesystem.BubbleUpParents to maintain backward compatibility.
func bubbleUpParents(dir, root string, needs map[string]bool) {
	filesystem.BubbleUpParents(dir, root, needs)
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
