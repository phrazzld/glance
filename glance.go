package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/briandowns/spinner"
	"github.com/google/generative-ai-go/genai"
	_ "github.com/joho/godotenv" // Used by the config package for loading environment variables
	gitignore "github.com/sabhiram/go-gitignore"
	"github.com/schollz/progressbar/v3"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	
	"glance/config"
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

// promptData is used for filling the text/template.
type promptData struct {
	Directory    string
	SubGlances   string
	FileContents string
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

	// Set up logging based on the verbose flag
	setupLogging(cfg.Verbose)

	// Scan directories and process them to generate GLANCE.md files
	dirs, ignoreChains, err := scanDirectories(cfg)
	if err != nil {
		logrus.Fatalf("🚫 Directory scan failed: %v - Check file permissions and disk space", err)
	}

	// Process directories and generate GLANCE.md files
	results := processDirectories(dirs, ignoreChains, cfg)

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

// scanDirectories performs BFS scanning and gathers .gitignore chain info per directory
func scanDirectories(cfg *config.Config) ([]string, map[string][]*gitignore.GitIgnore, error) {
	logrus.Info("✨ Excellent! Scanning directories now... Let's explore your code!")

	// Show a spinner while scanning
	s := spinner.New(spinner.CharSets[14], 120*time.Millisecond)
	s.Suffix = " 🔍 Scanning directories and loading .gitignore files..."
	s.FinalMSG = "🎉 Scan complete! Found all the good stuff!\n"
	s.Start()
	defer s.Stop()

	// Perform BFS scanning and gather .gitignore chain info per directory
	dirsList, dirToIgnoreChain, err := listAllDirsWithIgnores(cfg.TargetDir)
	if err != nil {
		return nil, nil, err
	}

	// Process from deepest subdirectories upward
	reverseSlice(dirsList)

	return dirsList, dirToIgnoreChain, nil
}

// processDirectories generates GLANCE.md files for each directory in the list
func processDirectories(dirsList []string, dirToIgnoreChain map[string][]*gitignore.GitIgnore, cfg *config.Config) []result {
	logrus.Info("🧠 Preparing to generate all GLANCE.md files... Getting ready to make your code shine!")

	// Create progress bar
	bar := progressbar.NewOptions(len(dirsList),
		progressbar.OptionSetDescription("✍️ Creating GLANCE files"),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(40),
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerPadding: "░",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	needsRegen := make(map[string]bool)
	var finalResults []result

	// Process each directory
	for _, d := range dirsList {
		ignoreChain := dirToIgnoreChain[d]

		// Check if we need to regenerate the GLANCE.md file
		forceDir, errCheck := shouldRegenerate(d, cfg.Force, ignoreChain)
		if errCheck != nil && cfg.Verbose {
			logrus.Warnf("⏱️ Couldn't check modification time for %s: %v", d, errCheck)
		}

		forceDir = forceDir || needsRegen[d]

		// Process the directory with retry logic
		r := processDirectory(d, forceDir, ignoreChain, cfg)
		finalResults = append(finalResults, r)

		_ = bar.Add(1)

		// Bubble up parent's regeneration flag if needed
		if r.success && r.attempts > 0 && forceDir {
			bubbleUpParents(d, cfg.TargetDir, needsRegen)
		}
	}

	fmt.Println()
	logrus.Infof("🎯 All done! GLANCE.md files have been generated for your codebase up to: %s", cfg.TargetDir)

	return finalResults
}

// processDirectory processes a single directory with retry logic
func processDirectory(dir string, forceDir bool, ignoreChain []*gitignore.GitIgnore, cfg *config.Config) result {
	r := result{dir: dir}

	glancePath := filepath.Join(dir, "GLANCE.md")
	if !forceDir {
		// Skip if GLANCE.md exists and not forcing regeneration
		if _, err := os.Stat(glancePath); err == nil {
			if cfg.Verbose {
				logrus.Debugf("⏩ Skipping %s (GLANCE.md already exists and looks fresh)", dir)
			}
			r.success = true
			return r
		}
	}

	// Gather data for GLANCE.md generation
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

	// Attempt to generate GLANCE.md with retries
	for attempt := 1; attempt <= cfg.MaxRetries; attempt++ {
		r.attempts = attempt

		if cfg.Verbose {
			logrus.Debugf("🔄 Attempt #%d for %s → Found %d subdirs, %d sub-glances, %d local files",
				attempt, dir, len(subdirs), len(subGlances), len(fileContents))
		}

		// Generate markdown content
		summary, llmErr := generateMarkdown(dir, fileContents, subGlances, cfg)
		if llmErr == nil {
			if werr := os.WriteFile(glancePath, []byte(summary), 0o644); werr != nil {
				r.err = fmt.Errorf("failed writing GLANCE.md to %s: %w", dir, werr)
				return r
			}
			r.success = true
			r.err = nil
			return r
		}
		if cfg.Verbose {
			logrus.Debugf("❌ Attempt %d for %s failed: %v - Will retry if attempts remain", attempt, dir, llmErr)
		}
		r.err = llmErr
	}

	return r
}

// generateMarkdown generates the content for a GLANCE.md file
func generateMarkdown(dir string, fileMap map[string]string, subGlances string, cfg *config.Config) (string, error) {
	// Build file contents chunk
	var fileContentsBuilder strings.Builder
	for filename, content := range fileMap {
		fileContentsBuilder.WriteString(fmt.Sprintf("=== file: %s ===\n%s\n\n", filename, content))
	}

	// Fill promptData struct for the template
	data := promptData{
		Directory:    dir,
		SubGlances:   subGlances,
		FileContents: fileContentsBuilder.String(),
	}

	// Parse and execute the prompt template
	tmpl, err := template.New("prompt").Parse(cfg.PromptTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse prompt template: %w", err)
	}
	var rendered bytes.Buffer
	if err = tmpl.Execute(&rendered, data); err != nil {
		return "", fmt.Errorf("failed to execute prompt template: %w", err)
	}

	promptStr := rendered.String()
	if cfg.Verbose {
		logrus.Debugf("[generateMarkdown] directory=%s, prompt length in bytes=%d", dir, len(promptStr))
	}

	// Set up the Gemini API client
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.APIKey))
	if err != nil {
		return "", err
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.0-flash")

	// Optional token debug
	if cfg.Verbose {
		tokenResp, tokenErr := model.CountTokens(ctx, genai.Text(promptStr))
		if tokenErr == nil {
			logrus.Debugf("🔤 Token count for %s: %d tokens in prompt", dir, tokenResp.TotalTokens)
		} else {
			logrus.Debugf("⚠️ Couldn't count tokens for %s: %v", dir, tokenErr)
		}
	}

	// Generate content using the Gemini API
	stream := model.GenerateContentStream(ctx, genai.Text(promptStr))

	var result strings.Builder
	for {
		resp, err := stream.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return "", err
		}
		for _, c := range resp.Candidates {
			if c.Content == nil {
				continue
			}
			for _, p := range c.Content.Parts {
				if txt, ok := p.(genai.Text); ok {
					result.WriteString(string(txt))
				}
			}
		}
	}
	return result.String(), nil
}

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
		var newChain []*gitignore.GitIgnore
		if localIgnore != nil {
			newChain = append(newChain, localIgnore)
		}
		combinedChain := append(current.ignoreChain, newChain...)

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
		if ig.MatchesPath(rel) {
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

// gatherSubGlances merges the contents of existing subdirectory GLANCE.md files.
func gatherSubGlances(subdirs []string) (string, error) {
	var combined []string
	for _, sd := range subdirs {
		data, err := os.ReadFile(filepath.Join(sd, "GLANCE.md"))
		if err == nil {
			combined = append(combined, strings.ToValidUTF8(string(data), "�"))
		}
	}
	return strings.Join(combined, "\n\n"), nil
}

// readSubdirectories lists immediate subdirectories in a directory, skipping hidden or ignored ones.
func readSubdirectories(dir string, ignoreChain []*gitignore.GitIgnore) ([]string, error) {
	entries, err := os.ReadDir(dir)
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
		fullPath := filepath.Join(dir, name)
		rel, _ := filepath.Rel(dir, fullPath)
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

// gatherLocalFiles reads immediate files in a directory (excluding GLANCE.md, hidden files, etc.).
func gatherLocalFiles(dir string, ignoreChain []*gitignore.GitIgnore, maxFileBytes int64, verbose bool) (map[string]string, error) {
	files := make(map[string]string)
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, werr error) error {
		if werr != nil {
			return werr
		}
		// skip subdirectories (beyond the current dir)
		if d.IsDir() && path != dir {
			return fs.SkipDir
		}
		if d.IsDir() || d.Name() == "GLANCE.md" || strings.HasPrefix(d.Name(), ".") {
			return nil
		}

		rel, _ := filepath.Rel(dir, path)
		if isIgnored(rel, ignoreChain) {
			if verbose {
				logrus.Debugf("ignoring file via .gitignore chain: %s", rel)
			}
			return nil
		}

		isText, errCheck := isTextFile(path)
		if errCheck != nil && verbose {
			logrus.Debugf("error checking if file is text: %s => %v", path, errCheck)
		}
		if !isText {
			if verbose {
				logrus.Debugf("📊 Skipping binary/non-text file: %s", path)
			}
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		contentStr := strings.ToValidUTF8(string(content), "�")
		if len(contentStr) > int(maxFileBytes) {
			contentStr = contentStr[:maxFileBytes] + "...(truncated)"
		}
		files[rel] = contentStr
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

// isTextFile checks a file's content type by reading its first 512 bytes.
func isTextFile(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return false, err
	}
	ctype := http.DetectContentType(buf[:n])
	if strings.HasPrefix(ctype, "text/") ||
		strings.HasPrefix(ctype, "application/json") ||
		strings.HasPrefix(ctype, "application/xml") ||
		strings.Contains(ctype, "yaml") {
		return true, nil
	}
	return false, nil
}

// -----------------------------------------------------------------------------
// regeneration logic and utilities
// -----------------------------------------------------------------------------

func shouldRegenerate(dir string, globalForce bool, ignoreChain []*gitignore.GitIgnore) (bool, error) {
	if globalForce {
		return true, nil
	}

	glancePath := filepath.Join(dir, "GLANCE.md")
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

// loadPromptTemplate tries to read from the specified file path, then "prompt.txt",
// and falls back to the default prompt template.
// This function is retained for test compatibility and is used by the config package.
func loadPromptTemplate(path string) (string, error) {
	// Import statement for godotenv is kept for compatibility with the existing tests
	// In the actual application, godotenv is imported and used by the config package
	defaultPrompt := `you are an expert code reviewer and technical writer.
generate a descriptive technical overview of this directory:
- highlight purpose, architecture, and key file roles
- mention important dependencies or gotchas
- do NOT provide recommendations or next steps

directory: {{.Directory}}

subdirectory summaries:
{{.SubGlances}}

local file contents:
{{.FileContents}}
`

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
	return defaultPrompt, nil
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
	logrus.Info("📊 === FINAL SUMMARY === 📊")
	logrus.Infof("🔢 Processed %d directories → %d successes, %d failures", len(results), totalSuccess, totalFailed)

	if totalFailed == 0 {
		logrus.Info("🌟 Perfect run! No failures detected. Your codebase is now well-documented!")
		return
	}

	logrus.Info("⚠️ Some directories couldn't be processed:")
	for _, r := range results {
		if !r.success {
			logrus.Warnf("❌ %s: Attempts=%d Error=%v", r.dir, r.attempts, r.err)
		}
	}
	logrus.Info("📊 ===================== 📊")
}

