package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/joho/godotenv" // Used by the config package for loading environment variables
	progressbar "github.com/schollz/progressbar/v3"
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
	results, _ := processDirectories(dirs, ignoreChains, cfg, llmService, os.Stderr)

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
	primaryClient, err := llm.NewGeminiClient(
		cfg.APIKey,
		llm.WithModelName("gemini-3-flash-preview"),
		llm.WithMaxRetries(0), // Single attempt per tier; FallbackClient handles retries.
		llm.WithMaxOutputTokens(4096),
		llm.WithTimeout(60),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create primary Gemini client: %w", err)
	}

	stableClient, err := llm.NewGeminiClient(
		cfg.APIKey,
		llm.WithModelName("gemini-2.5-flash"),
		llm.WithMaxRetries(0), // Single attempt per tier; FallbackClient handles retries.
		llm.WithMaxOutputTokens(4096),
		llm.WithTimeout(60),
	)
	if err != nil {
		primaryClient.Close()
		return nil, nil, fmt.Errorf("failed to create stable Gemini fallback client: %w", err)
	}

	tiers := []llm.FallbackTier{
		{Name: "gemini-3-flash-preview", Client: primaryClient},
		{Name: "gemini-2.5-flash", Client: stableClient},
	}

	openRouterKey := strings.TrimSpace(os.Getenv("OPENROUTER_API_KEY"))
	if openRouterKey == "" {
		logrus.Warn("OPENROUTER_API_KEY is not set; cross-provider fallback (x-ai/grok-4.1-fast) is disabled")
	} else {
		grokFallbackClient, grokErr := llm.NewOpenRouterClient(
			openRouterKey,
			llm.WithModelName("x-ai/grok-4.1-fast"),
			llm.WithMaxRetries(0), // Single attempt per tier; FallbackClient handles retries.
			llm.WithMaxOutputTokens(4096),
			llm.WithTimeout(60),
		)
		if grokErr != nil {
			primaryClient.Close()
			stableClient.Close()
			return nil, nil, fmt.Errorf("failed to create OpenRouter Grok fallback client: %w", grokErr)
		}

		tiers = append(tiers, llm.FallbackTier{
			Name:   "x-ai/grok-4.1-fast",
			Client: grokFallbackClient,
		})
	}

	client, err := llm.NewFallbackClient(tiers, cfg.MaxRetries)
	if err != nil {
		for _, tier := range tiers {
			tier.Client.Close()
		}
		return nil, nil, fmt.Errorf("failed to create fallback client chain: %w", err)
	}

	tierNames := make([]string, len(tiers))
	for i, tier := range tiers {
		tierNames[i] = tier.Name
	}
	compositeModelName := "fallback(" + strings.Join(tierNames, "->") + ")"

	// Create the service with functional options
	service, err := llm.NewService(
		client,
		llm.WithServiceModelName(compositeModelName),
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

// processDirectories generates glance.md files for each directory in the list and returns the map of directories
// needing regeneration. progressOut controls where progress bar output is written; pass io.Discard to suppress it.
func processDirectories(
	dirsList []string,
	dirToIgnoreChain map[string]filesystem.IgnoreChain,
	cfg *config.Config,
	llmService *llm.Service,
	progressOut io.Writer,
) ([]result, map[string]bool) {
	logrus.Info("Preparing to generate glance output files...")

	// Set up options for the progress bar
	options := []progressbar.Option{
		progressbar.OptionSetDescription("Creating glance files"),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(40),
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionSetWriter(progressOut),
	}

	// Create progress bar with the configured options
	bar := progressbar.NewOptions(len(dirsList), options...)

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

		// Ignore error for non-critical UI
		_ = bar.Add(1)

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

	// Finish the progress bar (ignore errors for non-critical UI)
	_ = bar.Finish()

	logrus.WithField("target_dir", cfg.TargetDir).Info("All done! glance output files have been generated for your codebase")

	return finalResults, needsRegen
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
		// The forceDir variable comes from ShouldRegenerate or parent propagation
		// We don't try to distinguish the exact reason, as it's correctly derived from
		// ShouldRegenerate or the parent propagation mechanism
		fields["reason"] = "local_changes_or_child_regenerated"
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

	// Directories with no analyzable content have nothing for the LLM to work with.
	// Calling the LLM with an empty prompt causes hallucination based on the
	// directory path name alone (e.g., inventing Rails framework details for
	// a Next.js project's /lib/assets). Write a minimal stub instead.
	if len(fileContents) == 0 && strings.TrimSpace(subGlances) == "" {
		stubDesc := stubDescription(dir, subdirs)
		logrus.WithField("directory", dir).Debug("Skipping LLM for directory with no analyzable content — writing minimal stub")
		// Base(dir) is intentional: stub heading is a display label, not a path reference.
		stub := fmt.Sprintf("# %s\n\n%s\n", filepath.Base(dir), stubDesc)
		glancePath := filepath.Join(dir, filesystem.GlanceFilename)
		validatedPath, pathErr := filesystem.ValidateFilePath(glancePath, dir, true, false)
		if pathErr != nil {
			r.err = fmt.Errorf("invalid glance.md path for %s: %w", dir, pathErr)
			return r
		}
		// #nosec G306 -- Using filesystem.DefaultFileMode (0600) for security & path validated
		if werr := os.WriteFile(validatedPath, []byte(stub), filesystem.DefaultFileMode); werr != nil {
			r.err = fmt.Errorf("failed writing stub glance.md to %s: %w", dir, werr)
			return r
		}
		r.success = true
		r.attempts = 1 // Counts as processed: triggers BubbleUpParents for parent regen
		return r
	}

	// Create context for LLM operations
	ctx := context.Background()

	// Use relative path in the LLM prompt to avoid leaking machine-specific paths.
	// Both cfg.TargetDir and dir are absolute (enforced by LoadConfig + scanning),
	// so Rel should never fail; the fallback is a safeguard, not an expected code path.
	relDir, relErr := filepath.Rel(cfg.TargetDir, dir)
	if relErr != nil {
		logrus.WithFields(logrus.Fields{
			"root":  cfg.TargetDir,
			"dir":   dir,
			"error": relErr,
		}).Warn("filepath.Rel failed; falling back to Base — absolute path may appear in LLM prompt")
		relDir = filepath.Base(dir)
	}

	logrus.WithFields(logrus.Fields{
		"directory": dir,
		"stage":     "llm_generation",
	}).Debug("Generating markdown content using LLM service")

	summary, llmErr := llmService.GenerateGlanceMarkdown(ctx, relDir, fileContents, subGlances)
	if llmErr != nil {
		logrus.WithFields(logrus.Fields{
			"directory": dir,
			"error":     llmErr,
			"stage":     "llm_generation",
		}).Error("Failed to generate markdown with LLM service")
		r.attempts = 1
		r.err = llmErr
		return r
	}

	// Validate the glance output path before writing
	glancePath := filepath.Join(dir, filesystem.GlanceFilename)
	logrus.WithFields(logrus.Fields{
		"directory": dir,
		"path":      glancePath,
		"stage":     "path_validation",
	}).Debug("Validating glance output path")

	validatedPath, pathErr := filesystem.ValidateFilePath(glancePath, dir, true, false)
	if pathErr != nil {
		logrus.WithFields(logrus.Fields{
			"directory": dir,
			"path":      glancePath,
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
	r.attempts = 1
	r.err = nil
	return r
}

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

// reverseSlice reverses a slice of directory paths in-place.
func reverseSlice(s []string) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

// -----------------------------------------------------------------------------
// file collection and processing
// -----------------------------------------------------------------------------

// gatherSubGlances merges the contents of existing subdirectory glance output files.
// Falls back to the legacy filename (glance.md) when the current filename (.glance.md)
// is absent, so parent summaries remain complete during the upgrade migration window.
// The baseDir parameter defines the security boundary for path validations within the function.
func gatherSubGlances(baseDir string, subdirs []string) (string, error) {
	var combined []string
	for _, sd := range subdirs {
		// Validate the subdirectory using the provided baseDir for consistent security boundary
		validDir, err := filesystem.ValidateDirPath(sd, baseDir, true, true)
		if err != nil {
			logrus.Warnf("Skipping invalid subdirectory for glance output collection: %v", err)
			continue
		}

		// Resolve the glance output path: prefer current filename, fall back to legacy.
		candidateNames := []string{filesystem.GlanceFilename, filesystem.LegacyGlanceFilename}
		var validPath string
		for _, name := range candidateNames {
			p := filepath.Join(validDir, name)
			vp, vpErr := filesystem.ValidateFilePath(p, validDir, true, true)
			if vpErr == nil {
				validPath = vp
				break
			}
		}
		if validPath == "" {
			logrus.Debugf("Skipping invalid glance output path for subdirectory: %s", validDir)
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

// stubDescription returns the body text for a minimal stub when no LLM-analyzable content
// exists. It distinguishes truly empty directories from directories that have files the LLM
// cannot process (binary, hidden, oversized, or gitignored files).
func stubDescription(dir string, subdirs []string) string {
	if len(subdirs) > 0 {
		// Has subdirectories (whose own summaries were also empty) — not truly empty.
		return "No analyzable text content."
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "Empty directory."
	}
	for _, e := range entries {
		name := e.Name()
		if !e.IsDir() && name != filesystem.GlanceFilename && name != filesystem.LegacyGlanceFilename {
			// At least one real file exists that GatherLocalFiles filtered out.
			return "No analyzable text content."
		}
	}
	return "Empty directory."
}

// gatherLocalFiles reads immediate files in a directory (excluding glance.md, hidden files, etc.).
// This function now uses filesystem.GatherLocalFiles directly with the IgnoreChain.
func gatherLocalFiles(dir string, ignoreChain filesystem.IgnoreChain, maxFileBytes int64) (map[string]string, error) {
	// Use the filesystem package function that provides comprehensive validation and handling
	return filesystem.GatherLocalFiles(dir, ignoreChain, maxFileBytes)
}

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
