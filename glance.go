package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

var (
	force   = flag.Bool("force", false, "regenerate glance.md even if it already exists")
	verbose = flag.Bool("verbose", false, "enable verbose logging (debug level)")
)

const (
	maxRetries   = 3
	maxFileBytes = 5 * 1024 * 1024
)

type result struct {
	dir      string
	attempts int
	success  bool
	err      error
}

func init() {
	if err := godotenv.Load(); err != nil {
		logrus.Warn("no .env file found (or error loading it), continuing with system environment")
	}
}

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, `usage:
  glance [--force] [--verbose] /path/to/dir

options:
  --force    regenerate glance.md even if it already exists
  --verbose  enable verbose logging (debug level)
`)
	}

	flag.Parse()

	if *verbose {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	targetDir := flag.Arg(0)
	absDir, err := filepath.Abs(targetDir)
	if err != nil {
		logrus.Fatalf("error: invalid target directory: %v", err)
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

	// load the top-level .gitignore (if any)
	var topIgnore *gitignore.GitIgnore
	topIgnore, _ = loadGitignore(absDir)

	logrus.Info("fabulous! scanning directories now...")

	s := spinner.New(spinner.CharSets[14], 120*time.Millisecond)
	s.Suffix = " scanning directories..."
	s.FinalMSG = "scan complete!\n"
	s.Start()
	dirsList, err := listAllDirs(absDir, topIgnore)
	s.Stop()
	if err != nil {
		logrus.Fatalf("directory scan failed: %v", err)
	}

	// process from the deepest subdirectories upward so sub-glances exist first
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

	// this map will track directories that must be forced to regenerate
	needsRegen := make(map[string]bool)

	var finalResults []result

	for _, d := range dirsList {
		// decide if we should forcibly regenerate glance.md
		forceThisDir, errCheck := shouldRegenerate(d, *force, topIgnore)
		if errCheck != nil {
			// if we can't check mod times, log an error but attempt to proceed
			if *verbose {
				logrus.Warnf("mod-time check failed for %s: %v", d, errCheck)
			}
		}

		// combine logic: if we've previously flagged this directory or parent's
		// changes bubble up, that also triggers a force
		forceThisDir = forceThisDir || needsRegen[d]

		r := processDirWithRetry(d, forceThisDir, apiKey, topIgnore)
		finalResults = append(finalResults, r)
		_ = bar.Add(1)

		if r.success && r.attempts > 0 && forceThisDir {
			// if we just regenerated this directory,
			// bubble up to all parent directories
			bubbleUpParents(d, absDir, needsRegen)
		}
	}

	fmt.Println()
	logrus.Infof("done! a glance.md file has been generated recursively up to: %s", absDir)

	printDebrief(finalResults)
}

// shouldRegenerate checks if we need to regenerate glance.md in `dir` due to --force,
// no existing file, or the presence of any changes (added/deleted/modified files) since glance.md was last modified.
func shouldRegenerate(dir string, globalForce bool, topIgnore *gitignore.GitIgnore) (bool, error) {
	if globalForce {
		return true, nil
	}

	glancePath := filepath.Join(dir, "glance.md")
	glanceInfo, err := os.Stat(glancePath)
	if err != nil {
		// if no glance.md or can't stat it, definitely regenerate
		return true, nil
	}

	// get the newest mod time of all files (recursively) in this directory
	latest, err := latestModTime(dir, topIgnore)
	if err != nil {
		return false, err
	}

	// if any file in the subtree is newer than glance.md, we regenerate
	if latest.After(glanceInfo.ModTime()) {
		return true, nil
	}
	return false, nil
}

// latestModTime scans a directory recursively to find the newest mod time of all files.
// respects .gitignore for subfolders if topIgnore is specified.
func latestModTime(dir string, topIgnore *gitignore.GitIgnore) (time.Time, error) {
	var latest time.Time
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, werr error) error {
		if werr != nil {
			return werr
		}
		// skip .git, node_modules, hidden, or anything the top-level .gitignore matches
		if d.IsDir() && path != dir {
			base := d.Name()
			if base == ".git" || base == "node_modules" || strings.HasPrefix(base, ".") {
				return filepath.SkipDir
			}
			rel, errRel := filepath.Rel(dir, path)
			if errRel == nil && shouldIgnore(topIgnore, rel) {
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

// bubbleUpParents marks all ancestors of `dir` (up to `root`) to be regenerated
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

// listAllDirs enumerates directories BFS style, skipping hidden, .git, and anything matched by the top-level .gitignore.
func listAllDirs(start string, topIgnore *gitignore.GitIgnore) ([]string, error) {
	var queue []string
	queue = append(queue, start)
	var results []string

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		results = append(results, current)

		entries, err := os.ReadDir(current)
		if err != nil {
			return nil, err
		}

		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			name := e.Name()

			// skip hidden folders
			if strings.HasPrefix(name, ".") {
				continue
			}
			// skip .git or node_modules specifically
			if name == ".git" || name == "node_modules" {
				continue
			}
			// if the top-level .gitignore says to skip
			rel, err := filepath.Rel(start, filepath.Join(current, name))
			if err == nil && shouldIgnore(topIgnore, rel) {
				if *verbose {
					logrus.Debugf("skipping %s due to .gitignore match", rel)
				}
				continue
			}
			queue = append(queue, filepath.Join(current, name))
		}
	}
	return results, nil
}

// processDirWithRetry tries up to maxRetries to generate a glance for "dir" if forceThisDir is true,
// or if no glance.md exists. if not forced and glance.md exists, we skip.
func processDirWithRetry(dir string, forceThisDir bool, apiKey string, topIgnore *gitignore.GitIgnore) result {
	r := result{dir: dir}
	glancePath := filepath.Join(dir, "glance.md")

	// if not forced and glance.md already exists, skip
	if !forceThisDir {
		if _, err := os.Stat(glancePath); err == nil {
			if *verbose {
				logrus.Debugf("skipping %s because glance.md already exists and no forced regeneration", dir)
			}
			r.success = true
			r.attempts = 0
			return r
		}
	}

	subdirs, err := readSubdirectories(dir, topIgnore)
	if err != nil {
		r.err = err
		return r
	}
	subGlances, err := gatherSubGlances(subdirs)
	if err != nil {
		r.err = fmt.Errorf("gatherSubGlances failed: %w", err)
		return r
	}
	fileContents, err := gatherLocalFiles(dir, topIgnore)
	if err != nil {
		r.err = fmt.Errorf("gatherLocalFiles failed: %w", err)
		return r
	}

	for attempt := 1; attempt <= maxRetries; attempt++ {
		r.attempts = attempt

		if *verbose {
			logrus.Debugf("attempt #%d for %s -> subdirs=%d subGlancesLen=%d localFiles=%d",
				attempt, dir, len(subdirs), len(subGlances), len(fileContents))
		}

		summary, llmErr := generateGlanceText(dir, fileContents, subGlances, apiKey)
		if llmErr == nil {
			if werr := os.WriteFile(glancePath, []byte(summary), 0o644); werr != nil {
				r.err = fmt.Errorf("failed writing glance.md to %s: %w", dir, werr)
				return r
			}
			r.success = true
			r.err = nil
			return r
		}
		if *verbose {
			logrus.Debugf("attempt %d for %s failed: %v", attempt, dir, llmErr)
		}
		r.err = llmErr
	}

	return r
}

func generateGlanceText(dir string, fileMap map[string]string, subGlances string, apiKey string) (string, error) {
	var promptBuilder strings.Builder
	promptBuilder.WriteString("you are an expert code reviewer and technical writer.\n\n")
	promptBuilder.WriteString("generate a concise, purely descriptive technical overview of this directory:\n")
	promptBuilder.WriteString("- highlight purpose, architecture, key file roles\n")
	promptBuilder.WriteString("- mention important dependencies or gotchas\n")
	promptBuilder.WriteString("- do NOT provide recommendations, suggestions, or future changes\n")
	promptBuilder.WriteString("- do NOT include a recommendations or next steps section\n\n")
	promptBuilder.WriteString(fmt.Sprintf("target directory: %s\n\n", dir))

	promptBuilder.WriteString("subdirectory summaries:\n")
	promptBuilder.WriteString(subGlances)
	promptBuilder.WriteString("\n\nlocal file contents:\n\n")

	for filename, content := range fileMap {
		finalContent := content
		if len(finalContent) > maxFileBytes {
			finalContent = finalContent[:maxFileBytes] + "...(truncated)"
		}
		promptBuilder.WriteString(fmt.Sprintf("=== file: %s ===\n%s\n\n", filename, finalContent))
	}

	promptStr := promptBuilder.String()

	if *verbose {
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
	if *verbose {
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

// gatherSubGlances merges existing subdirectory glance.md contents only
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

// gatherLocalFiles enumerates only the immediate files in `dir`, ignoring topIgnore, etc.
func gatherLocalFiles(dir string, topIgnore *gitignore.GitIgnore) (map[string]string, error) {
	files := make(map[string]string)
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, werr error) error {
		if werr != nil {
			return werr
		}
		// if subdirectory (and not the dir itself), skip
		if d.IsDir() && path != dir {
			return fs.SkipDir
		}
		if d.IsDir() {
			return nil
		}
		base := d.Name()
		if base == "glance.md" || strings.HasPrefix(base, ".") {
			return nil
		}
		// if .gitignore says skip
		rel, _ := filepath.Rel(dir, path)
		if shouldIgnore(topIgnore, rel) {
			if *verbose {
				logrus.Debugf("ignoring file via .gitignore: %s", rel)
			}
			return nil
		}

		isText, err := isTextFile(path)
		if err != nil {
			if *verbose {
				logrus.Debugf("error checking if file is text: %s => %v", path, err)
			}
			return nil
		}
		if !isText {
			if *verbose {
				logrus.Debugf("skipping non-text file: %s", path)
			}
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		contentStr := strings.ToValidUTF8(string(content), "�")
		files[rel] = contentStr
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

// readSubdirectories enumerates immediate subdirectories in `dir`, skipping node_modules, hidden, topIgnore, etc.
func readSubdirectories(dir string, topIgnore *gitignore.GitIgnore) ([]string, error) {
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
		if name == ".git" || name == "node_modules" {
			continue
		}
		if strings.HasPrefix(name, ".") {
			continue
		}
		rel, _ := filepath.Rel(dir, filepath.Join(dir, name))
		if shouldIgnore(topIgnore, rel) {
			if *verbose {
				logrus.Debugf("ignoring subdir via .gitignore: %s", rel)
			}
			continue
		}
		subdirs = append(subdirs, filepath.Join(dir, name))
	}
	return subdirs, nil
}

// loadGitignore tries to parse .gitignore in the given dir
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

func shouldIgnore(matcher *gitignore.GitIgnore, name string) bool {
	if matcher == nil {
		return false
	}
	return matcher.MatchesPath(name)
}

func reverseSlice(s []string) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

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

// isTextFile checks up to 512 bytes with http.DetectContentType
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

