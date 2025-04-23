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
		logrus.WithField("error", err).Fatal("Failed to initialize LLM service")
	}
	defer llmClient.Close()

	// Scan directories and process them to generate glance.md files
	dirs, ignoreChains, err := scanDirectories(cfg)
	if err != nil {
		logrus.WithField("error", err).Fatal("Directory scan failed - Check file permissions and disk space")
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
	logrus.Info("Scanning directories...")

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
	logrus.Info("Preparing to generate glance.md files...")

	// Create progress bar
	bar := ui.NewProcessor(len(dirsList))

	// Create map to track directories needing regeneration due to child changes
	needsRegen := make(map[string]bool)
	var finalResults []result

	// Process each directory
	for _, d := range dirsList {
		ignoreChain := dirToIgnoreChain[d]

		// Check if we need to regenerate the glance.md file based on local file changes
		forceDir, errCheck := filesystem.ShouldRegenerate(d, cfg.Force, ignoreChain)
		if errCheck != nil {
			logrus.WithFields(logrus.Fields{
				"directory": d,
				"error":     errCheck,
			}).Warn("Couldn't check modification time")
		}

		// Also check if this directory needs regeneration due to child directory changes
		forceDir = forceDir || needsRegen[d]

		if needsRegen[d] {
			logrus.WithFields(logrus.Fields{
				"directory": d,
				"reason":    "child directory regenerated",
			}).Debug("Directory marked for regeneration due to child changes")
		}

		// Process the directory with retry logic
		r := processDirectory(d, forceDir, ignoreChain, cfg, llmService)
		finalResults = append(finalResults, r)

		if err := bar.Increment(); err != nil {
			logrus.WithField("error", err).Warn("Failed to increment progress bar")
		}

		// Bubble up parent's regeneration flag if needed - only when regeneration was
		// successful and actually attempted (not skipped)
		if r.success && r.attempts > 0 && forceDir {
			logrus.WithFields(logrus.Fields{
				"directory": d,
				"reason":    "successfully regenerated",
			}).Debug("Marking parent directories for regeneration")
			filesystem.BubbleUpParents(d, cfg.TargetDir, needsRegen)
		}
	}

	fmt.Println()
	logrus.WithField("target_dir", cfg.TargetDir).Info("All done! glance.md files have been generated for your codebase")

	return finalResults
}

// processDirectory processes a single directory with retry logic
func processDirectory(dir string, forceDir bool, ignoreChain filesystem.IgnoreChain, cfg *config.Config, llmService *llm.Service) result {
	r := result{dir: dir}

	// forceDir already indicates if regeneration is needed based on filesystem.ShouldRegenerate
	// or parent propagation in processDirectories
	if !forceDir && !cfg.Force {
		logrus.WithFields(logrus.Fields{
			"directory": dir,
			"reason":    "up-to-date",
			"action":    "skip",
		}).Debug("Skipping directory - glance.md already exists and looks fresh, no child changes detected")
		r.success = true
		r.attempts = 0 // Explicitly mark that we didn't attempt to regenerate
		return r
	}

	// Log the reason for processing this directory with additional context
	fields := logrus.Fields{
		"directory": dir,
		"action":    "regenerate",
	}

	if cfg.Force {
		fields["reason"] = "global_force_flag"
		logrus.WithFields(fields).Debug("Processing directory - global force flag is set")
	} else if forceDir {
		// Try to be more specific about the reason
		if len(dir) > 0 && filepath.Base(dir) == "glance.md" {
			fields["reason"] = "glance_file_missing"
		} else {
			fields["reason"] = "local_changes_or_child_regenerated"
		}
		logrus.WithFields(fields).Debug("Processing directory - local changes or child directory regenerated")
	}

	// Gather data for glance.md generation
	logrus.WithFields(logrus.Fields{
		"directory": dir,
		"stage":     "gather_subdirectories",
	}).Debug("Reading subdirectories")

	subdirs, err := readSubdirectories(dir, ignoreChain)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"directory": dir,
			"error":     err,
			"stage":     "gather_subdirectories",
		}).Error("Failed to read subdirectories")
		r.err = err
		return r
	}

	logrus.WithFields(logrus.Fields{
		"directory":     dir,
		"subdirs_count": len(subdirs),
		"stage":         "gather_subglances",
	}).Debug("Gathering glance files from subdirectories")

	subGlances, err := gatherSubGlances(dir, subdirs)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"directory": dir,
			"error":     err,
			"stage":     "gather_subglances",
		}).Error("Failed to gather glance files from subdirectories")
		r.err = fmt.Errorf("gatherSubGlances failed: %w", err)
		return r
	}

	logrus.WithFields(logrus.Fields{
		"directory": dir,
		"stage":     "gather_local_files",
	}).Debug("Gathering local files")

	fileContents, err := gatherLocalFiles(dir, ignoreChain, cfg.MaxFileBytes)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"directory": dir,
			"error":     err,
			"stage":     "gather_local_files",
		}).Error("Failed to gather local files")
		r.err = fmt.Errorf("gatherLocalFiles failed: %w", err)
		return r
	}

	logrus.WithFields(logrus.Fields{
		"directory":        dir,
		"subdirs_count":    len(subdirs),
		"subglances_count": len(subGlances),
		"files_count":      len(fileContents),
		"stage":            "data_gathering_complete",
	}).Debug("Directory data gathering complete")

	// Create context for LLM operations
	ctx := context.Background()

	// Generate markdown content using the LLM service
	logrus.WithFields(logrus.Fields{
		"directory": dir,
		"stage":     "llm_generation",
	}).Debug("Generating markdown content using LLM service")

	summary, llmErr := llmService.GenerateGlanceMarkdown(ctx, dir, fileContents, subGlances)
	if llmErr != nil {
		logrus.WithFields(logrus.Fields{
			"directory": dir,
			"error":     llmErr,
			"stage":     "llm_generation",
		}).Error("Failed to generate markdown with LLM service")
		r.attempts = 1 // Service already handles retries internally
		r.err = llmErr
		return r
	}

	// Validate the glance.md path before writing
	glancePath := filepath.Join(dir, "glance.md")
	logrus.WithFields(logrus.Fields{
		"directory": dir,
		"file_path": glancePath,
		"stage":     "path_validation",
	}).Debug("Validating glance.md path")

	validatedPath, pathErr := filesystem.ValidateFilePath(glancePath, dir, true, false)
	if pathErr != nil {
		logrus.WithFields(logrus.Fields{
			"directory": dir,
			"file_path": glancePath,
			"error":     pathErr,
			"stage":     "path_validation",
		}).Error("Invalid glance.md path")
		r.err = fmt.Errorf("invalid glance.md path for %s: %w", dir, pathErr)
		return r
	}

	// Write the generated content to file using the validated path
	// #nosec G306 -- Using filesystem.DefaultFileMode (0600) for security & path validated
	if werr := os.WriteFile(validatedPath, []byte(summary), filesystem.DefaultFileMode); werr != nil { // Path validated & using secure permissions
		logrus.WithFields(logrus.Fields{
			"directory": dir,
			"path":      validatedPath,
			"error":     werr,
			"stage":     "file_write",
		}).Error("Failed to write glance.md file")
		r.err = fmt.Errorf("failed writing glance.md to %s: %w", dir, werr)
		return r
	}

	// Log successful generation with content info
	logrus.WithFields(logrus.Fields{
		"directory":   dir,
		"path":        validatedPath,
		"summary_len": len(summary),
		"stage":       "complete",
		"status":      "success",
	}).Debug("Successfully generated and wrote glance.md file")

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
	logrus.WithFields(logrus.Fields{
		"total_dirs":    len(results),
		"success_count": totalSuccess,
		"failure_count": totalFailed,
	}).Info("Directory processing summary")

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
