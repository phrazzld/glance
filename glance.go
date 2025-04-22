package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/joho/godotenv" // Used by the config package for loading environment variables
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

// No need for queueItem anymore as we're using the filesystem package for directory scanning

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

	// Set up logging with debug level
	setupLogging()

	// Set up the LLM client and service using the function variable
	llmClient, llmService, err := setupLLMService(cfg)
	if err != nil {
		logrus.Fatalf("Failed to initialize LLM service: %v", err)
	}
	defer llmClient.Close()

	// Scan directories and process them to generate glance.md files
	dirs, ignoreChains, err := scanDirectories(cfg)
	if err != nil {
		logrus.Fatalf("Directory scan failed: %v - Check file permissions and disk space", err)
	}

	// Process directories and generate glance.md files
	results := processDirectories(dirs, ignoreChains, cfg, llmService)

	// Print summary of results
	printDebrief(results)
}

// -----------------------------------------------------------------------------
// Main function components
// -----------------------------------------------------------------------------

// setupLogging configures the logger with level based on environment variable
// and initializes the package-level loggers in other packages
func setupLogging() {
	// Get logging level from environment variable, default to info level
	logLevelStr := os.Getenv("GLANCE_LOG_LEVEL")

	// Parse log level string to logrus.Level
	var logLevel logrus.Level
	switch strings.ToLower(logLevelStr) {
	case "debug":
		logLevel = logrus.DebugLevel
	case "info", "":
		logLevel = logrus.InfoLevel
	case "warn", "warning":
		logLevel = logrus.WarnLevel
	case "error":
		logLevel = logrus.ErrorLevel
	default:
		// Invalid level defaults to info and logs a warning
		logLevel = logrus.InfoLevel
		fmt.Printf("Invalid log level: %s. Using default (info) instead.\n", logLevelStr)
	}

	// Set the configured log level
	logrus.SetLevel(logLevel)

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

	// Initialize package-level loggers in other packages
	filesystem.SetLogger(logrus.StandardLogger())
}

// SetupLLMServiceFunc is a function type for creating LLM clients and services.
// This allows for easier mocking in tests without the complexity of a full factory interface.
type SetupLLMServiceFunc func(cfg *config.Config) (llm.Client, *llm.Service, error)

// The implementation to use - can be swapped in tests
var setupLLMServiceFunc SetupLLMServiceFunc = createLLMService

// setupLLMService creates a client and service
func setupLLMService(cfg *config.Config) (llm.Client, *llm.Service, error) {
	return setupLLMServiceFunc(cfg)
}

// createLLMService is the actual implementation for initializing the LLM client and service
func createLLMService(cfg *config.Config) (llm.Client, *llm.Service, error) {
	// Create the client with functional options
	client, err := llm.NewGeminiClient(
		cfg.APIKey,
		llm.WithModelName("gemini-2.0-flash"),
		llm.WithMaxRetries(cfg.MaxRetries),
		llm.WithTimeout(60),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create LLM client: %w", err)
	}

	// Create the service with functional options
	service, err := llm.NewService(
		client,
		llm.WithServiceMaxRetries(cfg.MaxRetries),
		llm.WithPromptTemplate(cfg.PromptTemplate),
	)
	if err != nil {
		client.Close()
		return nil, nil, fmt.Errorf("failed to create LLM service: %w", err)
	}

	return client, service, nil
}

// scanDirectories performs BFS scanning and gathers .gitignore chain info per directory
func scanDirectories(cfg *config.Config) ([]string, map[string]filesystem.IgnoreChain, error) {
	logrus.Info("Excellent! Scanning directories now... Let's explore your code!")

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
func processDirectories(dirsList []string, dirToIgnoreChain map[string]filesystem.IgnoreChain, cfg *config.Config, llmService *llm.Service) []result {
	logrus.Info("Preparing to generate all glance.md files... Getting ready to make your code shine!")

	// Create progress bar
	bar := ui.NewProcessor(len(dirsList))

	needsRegen := make(map[string]bool)
	var finalResults []result

	// Process each directory
	for _, d := range dirsList {
		ignoreChain := dirToIgnoreChain[d]

		// Check if we need to regenerate the glance.md file
		forceDir, errCheck := filesystem.ShouldRegenerate(d, cfg.Force, ignoreChain) // Check if regeneration is needed
		if errCheck != nil && filesystem.IsLevelEnabled(logrus.DebugLevel) {
			logrus.Warnf("Couldn't check modification time for %s: %v", d, errCheck)
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
			filesystem.BubbleUpParents(d, cfg.TargetDir, needsRegen)
		}
	}

	fmt.Println()
	logrus.Infof("All done! glance.md files have been generated for your codebase up to: %s", cfg.TargetDir)

	return finalResults
}

// processDirectory processes a single directory with retry logic
func processDirectory(dir string, forceDir bool, ignoreChain filesystem.IgnoreChain, cfg *config.Config, llmService *llm.Service) result {
	r := result{dir: dir}

	// forceDir already indicates if regeneration is needed based on filesystem.ShouldRegenerate
	// called in processDirectories
	if !forceDir && !cfg.Force {
		if filesystem.IsLevelEnabled(logrus.DebugLevel) {
			logrus.Debugf("Skipping %s (glance.md already exists and looks fresh)", dir)
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
	subGlances, err := gatherSubGlances(dir, subdirs)
	if err != nil {
		r.err = fmt.Errorf("gatherSubGlances failed: %w", err)
		return r
	}
	fileContents, err := gatherLocalFiles(dir, ignoreChain, cfg.MaxFileBytes)
	if err != nil {
		r.err = fmt.Errorf("gatherLocalFiles failed: %w", err)
		return r
	}

	if filesystem.IsLevelEnabled(logrus.DebugLevel) {
		logrus.Debugf("Processing %s → Found %d subdirs, %d sub-glances, %d local files",
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
	glancePath := filepath.Join(dir, "glance.md")
	validatedPath, pathErr := filesystem.ValidateFilePath(glancePath, dir, true, false)
	if pathErr != nil {
		r.err = fmt.Errorf("invalid glance.md path for %s: %w", dir, pathErr)
		return r
	}

	// Write the generated content to file using the validated path
	// #nosec G306 -- Using filesystem.DefaultFileMode (0600) for security & path validated
	if werr := os.WriteFile(validatedPath, []byte(summary), filesystem.DefaultFileMode); werr != nil { // Path validated & using secure permissions
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
// This function now uses filesystem.ListDirsWithIgnores directly, returning the native IgnoreChain type.
func listAllDirsWithIgnores(root string) ([]string, map[string]filesystem.IgnoreChain, error) {
	// Use the filesystem package function to get the directories and ignore chains
	return filesystem.ListDirsWithIgnores(root)
}

// Removed loadGitignore and isIgnored functions - now using filesystem package directly

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
// The baseDir parameter defines the security boundary for path validations within the function.
func gatherSubGlances(baseDir string, subdirs []string) (string, error) {
	var combined []string
	for _, sd := range subdirs {
		// Validate the subdirectory using the provided baseDir for consistent security boundary
		validDir, err := filesystem.ValidateDirPath(sd, baseDir, true, true)
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
func readSubdirectories(dir string, ignoreChain filesystem.IgnoreChain) ([]string, error) {
	// Get the parent directory to use as baseDir for validation
	parentDir := filepath.Dir(dir)

	// Validate the directory path using parent as baseDir
	validDir, err := filesystem.ValidateDirPath(dir, parentDir, true, true)
	if err != nil {
		return nil, fmt.Errorf("invalid directory path: %w", err)
	}

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
		if filesystem.ShouldIgnoreDir(fullPath, validDir, ignoreChain) {
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
// This function now uses filesystem.GatherLocalFiles directly with the IgnoreChain.
func gatherLocalFiles(dir string, ignoreChain filesystem.IgnoreChain, maxFileBytes int64) (map[string]string, error) {
	// Use the filesystem package function that provides comprehensive validation and handling
	return filesystem.GatherLocalFiles(dir, ignoreChain, maxFileBytes)
}

// Note: We now use filesystem.IsTextFile instead of this local function
// which provides path validation

// -----------------------------------------------------------------------------
// regeneration logic and utilities
// -----------------------------------------------------------------------------

// Removed shouldRegenerate, latestModTime, and bubbleUpParents functions
// Now using filesystem package functions directly

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
	logrus.Info("=== FINAL SUMMARY ===")
	logrus.Infof("Processed %d directories → %d successes, %d failures", len(results), totalSuccess, totalFailed)

	if totalFailed == 0 {
		logrus.Info("Perfect run! No failures detected. Your codebase is now well-documented!")
		return
	}

	logrus.Info("Some directories couldn't be processed:")
	for _, r := range results {
		if !r.success {
			// Use the UI error reporting
			ui.ReportError(r.err, fmt.Sprintf("Failed to process %s (attempts: %d)", r.dir, r.attempts))
		}
	}
	logrus.Info("=====================")
}
