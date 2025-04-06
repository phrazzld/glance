package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// directoryChecker defines an interface for checking directory existence
// This allows for easier testing by substituting a mock implementation
type directoryChecker interface {
	CheckDirectory(path string) error
}

// defaultChecker implements the directoryChecker interface using the real filesystem
type defaultChecker struct{}

// CheckDirectory verifies the path exists and is a directory
func (d *defaultChecker) CheckDirectory(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("cannot access directory %q: %w", path, err)
	}
	if !stat.IsDir() {
		return fmt.Errorf("path %q is a file, not a directory", path)
	}
	return nil
}

// Global variable to allow tests to override the directory checker
var dirChecker directoryChecker = &defaultChecker{}

// LoadConfig parses command-line flags, loads environment variables,
// and initializes the application configuration.
//
// It handles:
// - Command-line flag parsing
// - Loading environment variables from .env file
// - Reading the prompt template
// - Validating required settings
//
// The args parameter should contain the full command-line arguments
// (including the program name in args[0]).
func LoadConfig(args []string) (*Config, error) {
	// Start with a default configuration
	cfg := NewDefaultConfig()

	// Define flags
	cmdFlags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	var (
		force      bool
		verbose    bool
		promptFile string
	)

	cmdFlags.BoolVar(&force, "force", false, "regenerate GLANCE.md even if it already exists")
	cmdFlags.BoolVar(&verbose, "verbose", false, "enable verbose logging (debug level)")
	cmdFlags.StringVar(&promptFile, "prompt-file", "", "path to custom prompt file (overrides default)")

	// Parse flags
	if err := cmdFlags.Parse(args[1:]); err != nil {
		return nil, fmt.Errorf("failed to parse command-line arguments: %w", err)
	}

	// Validate target directory
	if cmdFlags.NArg() != 1 {
		return nil, errors.New("missing target directory: exactly one directory must be specified")
	}

	// Get target directory
	targetDir := cmdFlags.Arg(0)
	absDir, err := filepath.Abs(targetDir)
	if err != nil {
		return nil, fmt.Errorf("invalid target directory: %w", err)
	}

	// Check if directory exists and is actually a directory
	if err := dirChecker.CheckDirectory(absDir); err != nil {
		return nil, err
	}

	// Load .env if present (but don't fail if not found)
	if err := godotenv.Load(); err != nil {
		logrus.Warn("📝 No .env file found or couldn't load it. Using system environment variables instead.")
	}

	// Get API key from environment
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, errors.New("GEMINI_API_KEY is missing: please set this environment variable or add it to your .env file")
	}

	// Load prompt template
	promptTemplate, err := loadPromptTemplate(promptFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load prompt template: %w", err)
	}

	// Apply all configuration settings using the builder pattern
	cfg = cfg.
		WithAPIKey(apiKey).
		WithTargetDir(absDir).
		WithForce(force).
		WithVerbose(verbose).
		WithPromptTemplate(promptTemplate)

	return cfg, nil
}

// loadPromptTemplate tries to read from the specified file path, then "prompt.txt",
// and falls back to the default prompt template if neither is available.
func loadPromptTemplate(path string) (string, error) {
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("failed to read custom prompt template from '%s': %w", path, err)
		}
		return string(data), nil
	}

	if data, err := os.ReadFile("prompt.txt"); err == nil {
		return string(data), nil
	}

	return defaultPromptTemplate, nil
}
