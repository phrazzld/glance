# Glance README

## What is it?

Glance is a command-line tool that recursively scans a directory tree and generates a `glance.md` file in each directory. This file provides a high-level summary of the directory's contents, generated using Google's Generative AI API (Gemini).

## Quick Usage

1. **Build or Run:**
   Clone the repository and run the tool with:

       go build -o glance && ./glance [--force] [--verbose] /path/to/directory

   or run directly:

       go run . [--force] [--verbose] /path/to/directory

2. **Set Up Environment:**
   Ensure you have a valid `GEMINI_API_KEY` set in your environment or in a `.env` file.

3. **Flags:**
   - `--force` will regenerate `glance.md` even if it already exists.
   - `--verbose` enables detailed logging output.
   - `--prompt-file` allows specifying a custom prompt template file.

## Environment Variables

- **GEMINI_API_KEY:**
  Your Google Generative AI API key. This must be valid for the Gemini calls to succeed.

## .env File

Optionally, create a `.env` file in the same directory as the tool to automatically load your environment variables. For example:

       GEMINI_API_KEY=your_api_key_here

If the `.env` file is absent, Glance will fall back to your system's environment variables.

## What Does It Skip?

- **Hidden Files and Directories:**
  Glance ignores hidden directories (e.g., `.git`) and dotfiles.

- **.gitignore Matches:**
  Files or directories that are listed in a local `.gitignore` are not processed.

- **Existing `glance.md` Files:**
  It won’t overwrite an existing `glance.md` unless you use the `--force` flag.

- **Large Files:**
  Files larger than approximately 5MB are truncated to keep the prompt size manageable.

- **Invalid UTF-8:**
  Any invalid UTF-8 in file contents is sanitized before sending data to the API.

## Logging

Glance uses [logrus](https://github.com/sirupsen/logrus) for logging:
- **Info-level Logging:** Default logging of key actions.
- **Debug-level Logging:** Enabled with the `--verbose` flag for more detailed output.
- Additionally, it features a spinner and a progress bar during scanning and generation.

## Package Structure

Glance is organized into several packages:

- **config:** Configuration management for API keys, directories, and other settings
- **errors:** Custom error types and error handling utilities
- **filesystem:** Directory scanning, file reading, and gitignore handling
- **llm:** Abstractions for interacting with the Gemini API
- **ui:** User interface components for feedback, including spinners and progress bars

## Developer Setup

### Pre-commit Hooks

Glance uses pre-commit hooks to ensure code quality and consistency. These hooks automatically check your code before each commit to catch issues early.

#### What Pre-commit Hooks Do

- Ensure code follows Go formatting standards (`go fmt`, `go imports`)
- Run static analysis to catch potential bugs (`go vet`, `golangci-lint`)
- Verify tests pass before committing (`go test`)
- Fix common issues like trailing whitespace and line endings
- Prevent accidentally committing secrets or sensitive data
- Block large files and other unwanted content from the repository

#### Installation

**Option 1: Use our setup script (recommended)**
```bash
./scripts/setup-precommit.sh
```

**Option 2: Manual installation**

1. Install pre-commit:
```bash
# Using pip (Python)
pip install pre-commit

# Using Homebrew (macOS)
brew install pre-commit

# Using apt (Debian/Ubuntu)
sudo apt update
sudo apt install pre-commit
```

2. Install golangci-lint:
```bash
# Using Homebrew
brew install golangci-lint

# Using Go
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.57.0
```

3. Set up the git hooks:
```bash
cd /path/to/glance
pre-commit install
```

#### Using Pre-commit Hooks

After installation, hooks run automatically on each commit. You can also run them manually:

```bash
# Run all hooks on all files
pre-commit run --all-files

# Run a specific hook
pre-commit run go-fmt --all-files
```

For more details on our pre-commit setup, available hooks, configuration, and troubleshooting, see [docs/PRECOMMIT.md](/docs/PRECOMMIT.md).

## Dependencies

- [github.com/briandowns/spinner](https://github.com/briandowns/spinner) – Spinner animation.
- [github.com/google/generative-ai-go/genai](https://github.com/google/generative-ai-go) – Gemini API client.
- [github.com/joho/godotenv](https://github.com/joho/godotenv) – Loads environment variables from a `.env` file.
- [github.com/sabhiram/go-gitignore](https://github.com/sabhiram/go-gitignore) – Parses `.gitignore` files.
- [github.com/schollz/progressbar/v3](https://github.com/schollz/progressbar) – Displays a progress bar.
- [github.com/sirupsen/logrus](https://github.com/sirupsen/logrus) – Provides structured logging.
- [github.com/stretchr/testify](https://github.com/stretchr/testify) – Testing toolkit.

## License

This repository is provided under an open license (e.g., Apache-2.0). Please refer to the LICENSE file for details.

## Enjoy!

Use Glance to quickly generate summaries of your projects and make your directories easier to understand. Happy coding!
