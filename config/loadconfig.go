package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	"glance/llm"
)

// LoadPromptTemplateFunc defines a function type for loading prompt templates
// This allows us to replace it in tests
type LoadPromptTemplateFunc func(path string) (string, error)

// loadPromptTemplate is the function to use for loading prompt templates
var loadPromptTemplate LoadPromptTemplateFunc = LoadPromptTemplate

// directoryChecker defines an interface for checking directory existence
// This allows for easier testing by substituting a mock implementation
type directoryChecker interface {
	// CheckDirectory verifies the path exists and is a directory
	// Returns the validated path (can be relative or absolute) and any error
	CheckDirectory(path string) (string, error)
}

// defaultChecker implements the directoryChecker interface using the real filesystem
type defaultChecker struct{}

// CheckDirectory verifies the path exists and is a directory
// It accepts and preserves relative paths rather than forcing absolute conversion
func (d *defaultChecker) CheckDirectory(path string) (string, error) {
	// Clean the path to normalize it, but preserve its relative/absolute state
	cleanPath := filepath.Clean(path)

	stat, err := os.Stat(cleanPath)
	if err != nil {
		return "", fmt.Errorf("cannot access directory %q: %w", cleanPath, err)
	}
	if !stat.IsDir() {
		return "", fmt.Errorf("path %q is a file, not a directory", cleanPath)
	}
	return cleanPath, nil
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
		promptFile string
	)

	cmdFlags.BoolVar(&force, "force", false, "regenerate glance.md even if it already exists")
	cmdFlags.StringVar(&promptFile, "prompt-file", "", "path to custom prompt file (overrides default)")

	// Parse flags
	if err := cmdFlags.Parse(args[1:]); err != nil {
		return nil, fmt.Errorf("failed to parse command-line arguments: %w", err)
	}

	// Validate target directory
	if cmdFlags.NArg() != 1 {
		return nil, errors.New("missing target directory: exactly one directory must be specified")
	}

	// Get target directory and validate it
	targetDir := cmdFlags.Arg(0)

	// Check if directory exists and is actually a directory
	// The directoryChecker will clean the path and verify it's a directory
	validatedDir, err := dirChecker.CheckDirectory(targetDir)
	if err != nil {
		return nil, err
	}

	// Convert to absolute path only if needed for validation boundary
	// This is only needed when we need absolute paths for security validation
	absDir := validatedDir
	if !filepath.IsAbs(validatedDir) {
		absDir, err = filepath.Abs(validatedDir)
		if err != nil {
			return nil, fmt.Errorf("invalid target directory: %w", err)
		}
	}

	// Store the validated directory as our trusted root
	// This is safe since we've already verified it exists and is a directory

	// Load .env if present (but don't fail if not found)
	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found or couldn't load it. Using system environment variables instead.")
	}

	// Get API key from environment
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, errors.New("GEMINI_API_KEY is missing: please set this environment variable or add it to your .env file")
	}

	// Load prompt template using the centralized function
	promptTemplate, err := loadPromptTemplate(promptFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load prompt template: %w", err)
	}

	// If no template was found, use the default from llm package
	if promptTemplate == "" {
		promptTemplate = llm.DefaultTemplate()
	}

	// Apply all configuration settings using the builder pattern
	cfg = cfg.
		WithAPIKey(apiKey).
		WithTargetDir(absDir).
		WithForce(force).
		WithPromptTemplate(promptTemplate)

	return cfg, nil
}
