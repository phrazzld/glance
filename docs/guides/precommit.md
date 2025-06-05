# Pre-commit Hooks for Glance

This document describes the pre-commit hook setup for the Glance project to ensure code quality and consistency.

## Overview

The Glance project uses the pre-commit framework to run automated checks before each commit. These checks help maintain code quality by:

- Ensuring consistent code formatting with `go fmt`
- Running static analysis with `go vet` and `golangci-lint`
- Verifying tests pass with `go test`
- Checking for common issues like trailing whitespace and merge conflicts
- Detecting accidentally committed secrets or credentials

## Installation

1. Install pre-commit:

```bash
# Using pip
pip install pre-commit

# Using Homebrew
brew install pre-commit
```

2. **Note about golangci-lint installation:**

The pre-commit configuration now uses `language: golang` for golangci-lint, which means:

- You don't need to manually install golangci-lint; pre-commit will manage it for you
- This ensures the exact version specified in `.pre-commit-config.yaml` is used
- The installation is isolated from your system-wide Go tools

If you previously installed golangci-lint manually, you can continue to use it for direct invocation, but pre-commit will use its own managed version:

```bash
# Optional - only if you want to run golangci-lint outside of pre-commit
# Using Homebrew
brew install golangci-lint

# Using Go (make sure to use the version specified in .pre-commit-config.yaml)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.1.2
```

> **Note:** For version consistency, always check the current version in `.pre-commit-config.yaml`
> and use the same version if you install it manually. The `rev:` field under the golangci-lint
> repo configuration is the single source of truth for the version.

3. Install the git hook scripts:

```bash
cd /path/to/glance
pre-commit install
```

## Configuration Files

The pre-commit configuration is defined in two main files:

1. `.pre-commit-config.yaml` - Defines the hooks to run
2. `.golangci.yml` - Configures the golangci-lint behavior

## Usage

### Normal Workflow

After installation, the hooks will run automatically when you commit changes:

```bash
git add .
git commit -m "Your commit message"
```

If any hooks fail, the commit will be aborted with an error message.

### Running Manually

You can run all hooks manually on all files:

```bash
pre-commit run --all-files
```

Or run a specific hook:

```bash
pre-commit run go-fmt --all-files
```

### Temporarily Bypassing Hooks

In rare cases, you may need to bypass the hooks (not recommended):

```bash
git commit -m "Your commit message" --no-verify
```

## Included Hooks

The Glance project includes the following hooks:

### Code Quality & Formatting
- `go-fmt`: Ensures code follows Go formatting standards
- `go-imports`: Fixes import ordering and formatting
- `go-vet`: Examines code for suspicious constructs
- `golangci-lint`: Comprehensive linter that combines many Go linters
  - Uses the modern configuration format with explicit `version: "2"` in `.golangci.yml`
  - Pre-commit manages installation with `language: golang` for better reproducibility
  - Timeout and configuration are synchronized between pre-commit and CI environments
  - See [LINTING.md](LINTING.md) for complete standardization details

### Testing
- `go-unit-tests`: Runs unit tests to ensure they pass (including long-running tests)
- `go-mod-tidy`: Ensures the go.mod file is up to date

### File Hygiene
- `trailing-whitespace`: Trims trailing whitespace
- `end-of-file-fixer`: Ensures files end with a newline
- `check-yaml`: Validates YAML syntax
- `check-added-large-files`: Prevents large files (>500KB) from being committed
- `mixed-line-ending`: Normalizes line endings to LF
- `check-merge-conflict`: Ensures merge conflict markers aren't committed

### Security
- `detect-secrets`: Scans for potential secrets/credentials
- `detect-private-key`: Prevents private keys from being committed

### Code Organization
- `check-file-length`: Ensures files don't exceed recommended line count limits
  - Warns for files over 500 lines
  - Fails for files over 1000 lines

## Troubleshooting

### Common Issues

1. **Hook installation failed**: Make sure you have the latest pre-commit version
2. **golangci-lint errors**: Check that you're using a version compatible with Go 1.23+
3. **Slow performance**: Consider adjusting timeout settings in `.golangci.yml`

## Testing Strategy

### Consistent Test Environments

The `go-unit-tests` hook runs all tests, including long-running tests. Previously, this hook used the `-short` flag which skipped tests that checked themselves with `testing.Short()`. This flag has been removed to ensure:

1. **Environment consistency**: The same tests run in both local pre-commit checks and CI environments
2. **Preventing CI surprises**: Tests that pass locally will also pass in CI, avoiding situations where developers push code that passes pre-commit but fails in CI
3. **Complete validation**: All tests validate your changes before commit, not just a subset

While removing the `-short` flag may increase the time it takes for pre-commit hooks to run, the trade-off for reliability and consistency is worthwhile. The `go-unit-tests` hook still uses the `-race` flag to detect race conditions.

### Advanced Testing Options

If you need to skip pre-commit hooks for a quick commit in an emergency, you can use:

```bash
git commit -m "Your commit message" --no-verify
```

However, this should be used sparingly and followed by a proper commit that passes all hooks.

### Getting Help

If you encounter issues with the pre-commit hooks, please:

1. Check the error message for specific details
2. Refer to the [pre-commit documentation](https://pre-commit.com/)
3. Open an issue in the Glance repository
