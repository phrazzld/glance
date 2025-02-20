package main

import (
	"bytes"
	"context"
	"flag"
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
	"github.com/joho/godotenv"
	gitignore "github.com/sabhiram/go-gitignore"
	"github.com/schollz/progressbar/v3"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// -----------------------------------------------------------------------------
// global flags and constants
// -----------------------------------------------------------------------------

var (
	force      bool
	verbose    bool
	promptFile string

	// fallback prompt template
	defaultPrompt = `you are an expert code reviewer and technical writer.
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
)

const (
	maxRetries   = 3
	maxFileBytes = 5 * 1024 * 1024
)

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
	// define cli flags using the standard flag package
	flag.BoolVar(&force, "force", false, "regenerate glance.md even if it already exists")
	flag.BoolVar(&verbose, "verbose", false, "enable verbose logging (debug level)")
	flag.StringVar(&promptFile, "prompt-file", "", "path to custom prompt file (overrides default)")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <directory>\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	// set up logging
	if verbose {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})

	// load .env if present
	if err := godotenv.Load(); err != nil {
		logrus.Warn("no .env file found (or error loading it), continuing with system environment")
	}

	targetDir := flag.Arg(0)
	absDir, err := filepath.Abs(targetDir)
	if err != nil {
		logrus.Fatalf("invalid target directory: %v", err)
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		logrus.Fatal("environment variable GEMINI_API_KEY is not set. set GEMINI_API_KEY or put it in your .env")
	}

	stat, err := os.Stat(absDir)
	if err != nil {
		logrus.Fatalf("cannot read directory %q: %v", absDir, err)
	}
	if !stat.IsDir() {
		logrus.Fatalf("path %q is not a directory", absDir)
	}

	// load prompt template (from --prompt-file, "prompt.txt", or fallback)
	promptTemplate, err := loadPromptTemplate(promptFile)
	if err != nil {
		logrus.Fatalf("failed to load prompt template: %v", err)
	}

	logrus.Info("fabulous! scanning directories now...")

	s := spinner.New(spinner.CharSets[14], 120*time.Millisecond)
	s.Suffix = " scanning directories, loading .gitignore files..."
	s.FinalMSG = "scan complete!\n"
	s.Start()

	// perform BFS scanning and gather .gitignore chain info per directory
	dirsList, dirToIgnoreChain, err := listAllDirsWithIgnores(absDir)
	if err != nil {
		s.Stop()
		logrus.Fatalf("directory scan failed: %v", err)
	}
	s.Stop()

	// process from deepest subdirectories upward
	reverseSlice(dirsList)

	logrus.Info("preparing to generate all glance.md files...")

	bar := progressbar.NewOptions(len(dirsList),
		progressbar.OptionSetDescription("generating glance files"),
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

	for _, d := range dirsList {
		ignoreChain := dirToIgnoreChain[d]

		forceDir, errCheck := shouldRegenerate(d, force, ignoreChain)
		if errCheck != nil && verbose {
			logrus.Warnf("mod-time check failed for %s: %v", d, errCheck)
		}

		forceDir = forceDir || needsRegen[d]

		r := processDirWithRetry(d, forceDir, apiKey, ignoreChain, promptTemplate)
		finalResults = append(finalResults, r)

		_ = bar.Add(1)

		// bubble up parent's regeneration flag if needed
		if r.success && r.attempts > 0 && forceDir {
			bubbleUpParents(d, absDir, needsRegen)
		}
	}

	fmt.Println()
	logrus.Infof("done! a glance.md file has been generated recursively up to: %s", absDir)

	printDebrief(finalResults)
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
				if verbose {
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
// processing directories and generating glance.md
// -----------------------------------------------------------------------------

func processDirWithRetry(dir string, forceDir bool, apiKey string, ignoreChain []*gitignore.GitIgnore, promptTemplate string) result {
	r := result{dir: dir}

	glancePath := filepath.Join(dir, "glance.md")
	if !forceDir {
		// skip if glance.md exists and not forcing regeneration
		if _, err := os.Stat(glancePath); err == nil {
			if verbose {
				logrus.Debugf("skipping %s because glance.md already exists", dir)
			}
			r.success = true
			return r
		}
	}

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
	fileContents, err := gatherLocalFiles(dir, ignoreChain)
	if err != nil {
		r.err = fmt.Errorf("gatherLocalFiles failed: %w", err)
		return r
	}

	for attempt := 1; attempt <= maxRetries; attempt++ {
		r.attempts = attempt

		if verbose {
			logrus.Debugf("attempt #%d for %s -> subdirs=%d, subGlancesLen=%d, localFiles=%d",
				attempt, dir, len(subdirs), len(subGlances), len(fileContents))
		}

		summary, llmErr := generateGlanceText(dir, fileContents, subGlances, apiKey, promptTemplate)
		if llmErr == nil {
			if werr := os.WriteFile(glancePath, []byte(summary), 0o644); werr != nil {
				r.err = fmt.Errorf("failed writing glance.md to %s: %w", dir, werr)
				return r
			}
			r.success = true
			r.err = nil
			return r
		}
		if verbose {
			logrus.Debugf("attempt %d for %s failed: %v", attempt, dir, llmErr)
		}
		r.err = llmErr
	}

	return r
}

// gatherSubGlances merges the contents of existing subdirectory glance.md files.
func gatherSubGlances(subdirs []string) (string, error) {
	var combined []string
	for _, sd := range subdirs {
		data, err := os.ReadFile(filepath.Join(sd, "glance.md"))
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
			if verbose {
				logrus.Debugf("ignoring subdir via .gitignore chain: %s", rel)
			}
			continue
		}
		subdirs = append(subdirs, fullPath)
	}
	return subdirs, nil
}

// gatherLocalFiles reads immediate files in a directory (excluding glance.md, hidden files, etc.).
func gatherLocalFiles(dir string, ignoreChain []*gitignore.GitIgnore) (map[string]string, error) {
	files := make(map[string]string)
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, werr error) error {
		if werr != nil {
			return werr
		}
		// skip subdirectories (beyond the current dir)
		if d.IsDir() && path != dir {
			return fs.SkipDir
		}
		if d.IsDir() || d.Name() == "glance.md" || strings.HasPrefix(d.Name(), ".") {
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
				logrus.Debugf("skipping non-text file: %s", path)
			}
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		contentStr := strings.ToValidUTF8(string(content), "�")
		if len(contentStr) > maxFileBytes {
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
// LLM interaction
// -----------------------------------------------------------------------------

func generateGlanceText(dir string, fileMap map[string]string, subGlances string,
	apiKey string, promptTemplate string) (string, error) {

	// build file contents chunk
	var fileContentsBuilder strings.Builder
	for filename, content := range fileMap {
		fileContentsBuilder.WriteString(fmt.Sprintf("=== file: %s ===\n%s\n\n", filename, content))
	}

	// fill promptData struct for the template
	data := promptData{
		Directory:    dir,
		SubGlances:   subGlances,
		FileContents: fileContentsBuilder.String(),
	}

	// parse and execute the prompt template
	tmpl, err := template.New("prompt").Parse(promptTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse prompt template: %w", err)
	}
	var rendered bytes.Buffer
	if err = tmpl.Execute(&rendered, data); err != nil {
		return "", fmt.Errorf("failed to execute prompt template: %w", err)
	}

	promptStr := rendered.String()
	if verbose {
		logrus.Debugf("[generateGlanceText] directory=%s, prompt length in bytes=%d", dir, len(promptStr))
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return "", err
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-flash")

	// optional token debug
	if verbose {
		tokenResp, tokenErr := model.CountTokens(ctx, genai.Text(promptStr))
		if tokenErr == nil {
			logrus.Debugf("[generateGlanceText] directory=%s, prompt tokens=%d", dir, tokenResp.TotalTokens)
		} else {
			logrus.Debugf("[generateGlanceText] directory=%s, token count check failed: %v", dir, tokenErr)
		}
	}

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

// loadPromptTemplate tries to read from --prompt-file, then "prompt.txt", else returns defaultPrompt.
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
	return defaultPrompt, nil
}

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
	logrus.Info("=== final debrief ===")
	logrus.Infof("processed %d directories -> successes: %d, failures: %d", len(results), totalSuccess, totalFailed)

	if totalFailed == 0 {
		logrus.Info("no failures! we're all set.")
		return
	}

	logrus.Info("some directories failed to generate glance.md:")
	for _, r := range results {
		if !r.success {
			logrus.Warnf(" - %s: attempts=%d err=%v", r.dir, r.attempts, r.err)
		}
	}
	logrus.Info("=====================")
}

