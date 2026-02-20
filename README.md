# Glance README

[![Pre-commit Checks](https://github.com/phrazzld/glance/actions/workflows/precommit.yml/badge.svg)](https://github.com/phrazzld/glance/actions/workflows/precommit.yml)
[![Go Tests](https://github.com/phrazzld/glance/actions/workflows/test.yml/badge.svg)](https://github.com/phrazzld/glance/actions/workflows/test.yml)
[![Go Linting](https://github.com/phrazzld/glance/actions/workflows/lint.yml/badge.svg)](https://github.com/phrazzld/glance/actions/workflows/lint.yml)
[![Go Build](https://github.com/phrazzld/glance/actions/workflows/build.yml/badge.svg)](https://github.com/phrazzld/glance/actions/workflows/build.yml)

## What is it?

Glance is a command-line tool that recursively scans a directory tree and generates a `glance.md` file in each directory. This file provides a high-level summary of the directory's contents, generated using Google's Generative AI API (Gemini).

## Quick Usage

1. **Build or Run:**
   Clone the repository and run the tool with:

       go build -o glance && ./glance [--force] /path/to/directory

   or run directly:

       go run . [--force] /path/to/directory

2. **Set Up Environment:**
   Ensure you have a valid `GEMINI_API_KEY` set in your environment or in a `.env` file.

3. **Flags:**
   - `--force` will regenerate `glance.md` even if it already exists.
   - `--prompt-file` allows specifying a custom prompt template file.

## Environment Variables

- **GEMINI_API_KEY:**
  Your Google Generative AI API key. This must be valid for the Gemini calls to succeed.

- **GLANCE_LOG_LEVEL:**
  Controls the verbosity of logging. Valid values: `debug`, `info` (default), `warn`, `error`.

## LLM Configuration

Glance uses Google's Gemini AI model for generating summaries:

- **Default Model:** `gemini-3-flash-preview`
- **Token Management:** Automatically truncates large files to avoid token limits
- **Error Handling:** Includes automatic retries with backoff for API failures
- **Upgrade Path:** New Gemini models can be supported by updating the default model name in the configuration (no code changes required)

## .env File

Optionally, create a `.env` file in the same directory as the tool to automatically load your environment variables. For example:

```
GEMINI_API_KEY=your_api_key_here
GLANCE_LOG_LEVEL=debug
```

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

- **File Permissions:**
  Glance uses restrictive file permissions (0600 / rw-------) for all generated files to protect potentially sensitive information. This means only the user who ran Glance can read or modify the generated glance.md files.

## Logging

Glance uses [logrus](https://github.com/sirupsen/logrus) for logging:

- **Default Log Level:** Info level (`logrus.InfoLevel`) is set by default.
- **Configurable Log Level:** You can change the log level using the `GLANCE_LOG_LEVEL` environment variable.
- **Structured Logging:** Uses logrus fields to provide contextual information in logs.
- **Visual Feedback:** Features a spinner and a progress bar during scanning and generation.

### Configuring Log Level

Set the `GLANCE_LOG_LEVEL` environment variable to one of these values:

- `debug` - Most verbose, shows all diagnostic information.
- `info` - Default level, shows general operational information.
- `warn` - Only warnings and errors.
- `error` - Only serious errors.

#### Examples:

```bash
# Set to debug level for maximum verbosity
GLANCE_LOG_LEVEL=debug ./glance /path/to/directory

# Set to error level for minimal output
GLANCE_LOG_LEVEL=error ./glance /path/to/directory

# In your .env file
GLANCE_LOG_LEVEL=warn
```

If an invalid level is specified, Glance will default to `info` level.

## Package Structure

Glance is organized into several packages:

- **config:** Configuration management for API keys, directories, and other settings
- **errors:** Custom error types and error handling utilities
- **filesystem:** Directory scanning, file reading, and gitignore handling
- **llm:** Abstractions for interacting with the Gemini API
- **ui:** User interface components for feedback, including spinners and progress bars
- **internal/mocks:** Shared mock implementations for testing

## Developer Setup

### Quick Setup (Recommended)

We provide a comprehensive setup script that will configure your entire development environment:

```bash
./scripts/setup-dev-environment.sh
```

This script will:
- Verify Go and Git installations
- Configure Git settings if needed
- Install and configure pre-commit hooks
- Install the GitHub CLI (optional)
- Set up a local environment file
- Verify Go modules
- Build the project
- Provide next steps

### Pre-commit Hooks

Glance uses pre-commit hooks to ensure code quality and consistency. These hooks automatically check your code before each commit to catch issues early.

#### What Pre-commit Hooks Do

- Ensure code follows Go formatting standards (`go fmt`, `go imports`)
- Run static analysis to catch potential bugs (`go vet`, `golangci-lint` - see [docs/LINTING.md](docs/LINTING.md))
- Verify tests pass before committing (`go test`)
- Fix common issues like trailing whitespace and line endings
- Prevent accidentally committing secrets or sensitive data
- Block large files and other unwanted content from the repository

#### Manual Installation

If you prefer to set up only the pre-commit hooks manually:

**Option 1: Use our pre-commit setup script**
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

# Using Go (use the version specified in .pre-commit-config.yaml)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.1.2
```

> **Note:** For version consistency, check the current version in `.pre-commit-config.yaml`
> and ensure you're using the same version as specified in the `rev:` field under the
> golangci-lint repo configuration.

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

### GitHub Actions Workflows

Glance uses GitHub Actions for continuous integration and deployment. Our workflows automatically test, lint, and build the project on every push and pull request.

For detailed information about our GitHub Actions setup, including workflow configurations, triggers, and troubleshooting tips, see [docs/GITHUB_ACTIONS.md](/docs/GITHUB_ACTIONS.md).

### Testing and Mocking Strategy

Glance follows a balanced approach to testing and mocking:

- Interface-based mocking at true API boundaries
- Function variable mocking for internal implementation details
- Shared mock implementations in the internal/mocks package

For more details on our testing and mocking approach, including guidelines and examples, see [docs/MOCKING_APPROACH.md](/docs/MOCKING_APPROACH.md).

## Dependencies

- [github.com/briandowns/spinner](https://github.com/briandowns/spinner) – Spinner animation.
- [google.golang.org/genai](https://pkg.go.dev/google.golang.org/genai) – Gemini API client.
- [github.com/joho/godotenv](https://github.com/joho/godotenv) – Loads environment variables from a `.env` file.
- [github.com/sabhiram/go-gitignore](https://github.com/sabhiram/go-gitignore) – Parses `.gitignore` files.
- [github.com/schollz/progressbar/v3](https://github.com/schollz/progressbar) – Displays a progress bar.
- [github.com/sirupsen/logrus](https://github.com/sirupsen/logrus) – Provides structured logging.
- [github.com/stretchr/testify](https://github.com/stretchr/testify) – Testing toolkit.

## License

This repository is provided under the MIT License. See the [LICENSE](LICENSE) file for details.

## Enjoy!

Use Glance to quickly generate summaries of your projects and make your directories easier to understand. Happy coding!
