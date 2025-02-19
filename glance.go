package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
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

func init() {
	// attempt to load env vars from .env (non-fatal if it doesn't exist)
	if err := godotenv.Load(); err != nil {
		logrus.Warn("no .env file found (or error loading it), continuing with system environment")
	}
}

func main() {
	// remove ascii banner as requested, simpler usage
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, `usage:
  glance [--force] [--verbose] /path/to/dir

options:
  --force    regenerate glance.md even if it already exists
  --verbose  enable verbose logging (debug level)
`)
	}

	flag.Parse()

	// set up logging
	if *verbose {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
	// color + full timestamps
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})

	// at least one arg is required
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

	logrus.Info("fabulous! scanning directories now...")

	s := spinner.New(spinner.CharSets[14], 120*time.Millisecond) // a fun spinner
	s.Suffix = " scanning directories..."
	s.FinalMSG = "scan complete!\n"
	s.Start()
	dirsList, err := listAllDirs(absDir)
	s.Stop()
	if err != nil {
		logrus.Fatalf("directory scan failed: %v", err)
	}

	// process from deepest subdirectories up
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

	for _, d := range dirsList {
		if *verbose {
			logrus.Debugf("processing dir: %s", d)
		}
		if err := processDir(d, *force, apiKey); err != nil {
			logrus.Warnf("error processing %q: %v", d, err)
		}
		_ = bar.Add(1)
	}

	fmt.Println()
	logrus.Infof("done! a glance.md file has been generated recursively up to: %s", absDir)
}

func processDir(dir string, force bool, apiKey string) error {
	glancePath := filepath.Join(dir, "glance.md")
	if !force {
		if _, err := os.Stat(glancePath); err == nil {
			if *verbose {
				logrus.Debugf("skipping %s because glance.md already exists", dir)
			}
			return nil
		}
	}

	ignoreMatcher, _ := loadGitignore(dir)
	subdirs, err := readSubdirectories(dir, ignoreMatcher)
	if err != nil {
		return err
	}

	subGlances, err := gatherSubGlances(subdirs)
	if err != nil {
		return fmt.Errorf("gatherSubGlances failed: %w", err)
	}
	fileContents, err := gatherLocalFiles(dir, ignoreMatcher)
	if err != nil {
		return fmt.Errorf("gatherLocalFiles failed: %w", err)
	}

	if *verbose {
		logrus.Debugf("calling gemini model for directory: %s", dir)
	}
	summary, err := generateGlanceText(dir, fileContents, subGlances, apiKey)
	if err != nil {
		return fmt.Errorf("gemini call failed: %w", err)
	}

	if werr := os.WriteFile(glancePath, []byte(summary), 0o644); werr != nil {
		return fmt.Errorf("failed writing glance.md to %s: %w", dir, werr)
	}
	return nil
}

func generateGlanceText(dir string, fileMap map[string]string, subGlances string, apiKey string) (string, error) {
	var promptBuilder strings.Builder
	promptBuilder.WriteString("you are an expert code reviewer and technical writer.\n\n")
	promptBuilder.WriteString("generate a concise yet detailed technical overview of this directory for developers.\n")
	promptBuilder.WriteString("focus on:\n")
	promptBuilder.WriteString("1) high-level purpose and architecture\n")
	promptBuilder.WriteString("2) key file roles\n")
	promptBuilder.WriteString("3) major dependencies or patterns\n")
	promptBuilder.WriteString("4) implementation details\n")
	promptBuilder.WriteString("5) any special gotchas or constraints\n\n")
	promptBuilder.WriteString(fmt.Sprintf("target directory: %s\n", dir))
	promptBuilder.WriteString("\nsubdirectory summaries:\n")
	promptBuilder.WriteString(subGlances)
	promptBuilder.WriteString("\n\nlocal file contents:\n")

	for filename, content := range fileMap {
		finalContent := content
		// only truncate extremely large files (~5mb+)
		if len(finalContent) > 5*1024*1024 {
			finalContent = finalContent[:5*1024*1024] + "...(truncated)"
		}
		promptBuilder.WriteString(fmt.Sprintf("=== file: %s ===\n%s\n\n", filename, finalContent))
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return "", err
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-flash")
	stream := model.GenerateContentStream(ctx, genai.Text(promptBuilder.String()))

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

func gatherSubGlances(subdirs []string) (string, error) {
	var combined []string
	for _, sd := range subdirs {
		data, err := os.ReadFile(filepath.Join(sd, "glance.md"))
		if err == nil {
			// also ensure glance.md is valid utf-8
			combined = append(combined, strings.ToValidUTF8(string(data), "�"))
		}
	}
	return strings.Join(combined, "\n\n"), nil
}

// gatherLocalFiles reads each file's bytes, cleans up invalid utf-8, and returns a map of relPath -> content.
func gatherLocalFiles(dir string, matcher *gitignore.GitIgnore) (map[string]string, error) {
	files := make(map[string]string)
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, werr error) error {
		if werr != nil {
			return werr
		}
		if d.IsDir() {
			return nil
		}
		if path == filepath.Join(dir, "glance.md") {
			return nil
		}
		// skip dotfiles (including .gitignore, etc.) so they don't appear in the prompt
		if strings.HasPrefix(d.Name(), ".") {
			return nil
		}
		// see if .gitignore says to skip
		rel, _ := filepath.Rel(dir, path)
		if shouldIgnore(matcher, rel) {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		// sanitize any invalid utf-8
		contentStr := strings.ToValidUTF8(string(content), "�")
		files[rel] = contentStr
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func readSubdirectories(dir string, matcher *gitignore.GitIgnore) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var subdirs []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		// skip .git or any hidden folder
		if e.Name() == ".git" || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		if shouldIgnore(matcher, e.Name()) {
			continue
		}
		subdirs = append(subdirs, filepath.Join(dir, e.Name()))
	}
	return subdirs, nil
}

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

func listAllDirs(start string) ([]string, error) {
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
		ignoreMatcher, _ := loadGitignore(current)

		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			// skip .git or any hidden folder
			if e.Name() == ".git" || strings.HasPrefix(e.Name(), ".") {
				continue
			}
			if shouldIgnore(ignoreMatcher, e.Name()) {
				continue
			}
			queue = append(queue, filepath.Join(current, e.Name()))
		}
	}
	return results, nil
}

func reverseSlice(s []string) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

