# Pre-commit Hooks for Go Projects (Go 1.23+)

## What is the pre-commit framework?

Pre-commit is a framework for managing and maintaining multi-language pre-commit hooks. It allows developers to define a set of hooks that run before code is committed, ensuring that only quality code is committed to the repository.

### How it works with Go projects:
- Pre-commit hooks run before each commit
- Can prevent commits if hooks fail
- Supports Go-specific tools like go fmt, go vet, golangci-lint
- Configurable via `.pre-commit-config.yaml` file

### Installation:
```bash
# Using pip
pip install pre-commit

# Using homebrew
brew install pre-commit
```

## Go-specific Hooks

### Standard and recommended hooks for Go projects:

1. **Code Formatting**
   - `go-fmt`: Runs `gofmt` to ensure consistent code style
   - `goimports`: Runs `goimports` which formats code and fixes imports

2. **Code Analysis**
   - `go-vet`: Runs `go vet` to examine code for suspicious constructs
   - `go-lint`: Runs `golint` for style mistakes
   - `go-critic`: More thorough static analysis
   - `golangci-lint`: Comprehensive linting tool that aggregates many linters

3. **Testing**
   - `go-unit-tests`: Runs `go test` to ensure tests pass
   - `go-build`: Ensures code compiles without errors

4. **General Purpose Hooks**
   - `trailing-whitespace`: Trims trailing whitespace
   - `end-of-file-fixer`: Ensures files end with a newline
   - `check-yaml`: Validates YAML syntax
   - `check-added-large-files`: Prevents large files from being committed
   - `detect-secrets`: Scans for potential secrets/credentials
   - `check-merge-conflict`: Ensures merge conflict markers aren't committed

## Compatibility with Go 1.23+

Go 1.23+ introduced several changes that affect tooling and linters:

- **Module Awareness**: Hooks should be module-aware and respect go.mod
- **Generics Support**: Linters must support generics (introduced in Go 1.18)
- **Workspace Mode**: Tools should be compatible with workspace mode (go.work)
- **golangci-lint Compatibility**: Use v1.55.0+ to support Go 1.23+ features

### Specific Compatibility Notes

1. **Linter Version Requirements**:
   - golangci-lint v1.55.0 or newer is required for Go 1.23+ compatibility
   - gopls v0.14.0+ for LSP support with Go 1.23+
   - staticcheck v2023.1.6+ for static analysis

2. **Known Issues**:
   - Some older linters like `maligned` are deprecated and may fail
   - Certain custom type analysis tools need updating for generics support
   - Type checking in older linters may fail with newer Go syntax

3. **Go 1.23-specific Features**:
   - Support for range-over-function
   - Improved type inference
   - Loop variable shadowing changes 
   - Error wrapping enhancements

4. **Recommended Linter Settings**:
   - Use `revive` instead of the deprecated `golint`
   - Enable `exportloopref` to catch loop variable reference issues
   - Configure `stylecheck` for consistent styling with newer Go idioms

## Best Practices for Setting Up Pre-commit Hooks

### 1. Sample Configuration

Create a `.pre-commit-config.yaml` file in the root of your project:

```yaml
repos:
- repo: https://github.com/pre-commit/pre-commit-hooks
  rev: v4.5.0
  hooks:
    - id: trailing-whitespace
    - id: end-of-file-fixer
    - id: check-yaml
    - id: check-added-large-files
      args: ['--maxkb=500']

- repo: https://github.com/dnephin/pre-commit-golang
  rev: master
  hooks:
    - id: go-fmt
    - id: go-vet
    - id: go-imports
    - id: go-unit-tests

- repo: https://github.com/golangci/golangci-lint
  rev: v1.57.0  # Use latest version compatible with Go 1.23+
  hooks:
    - id: golangci-lint
      args: [--timeout=5m]

- repo: https://github.com/Yelp/detect-secrets
  rev: v1.4.0
  hooks:
    - id: detect-secrets
```

### 2. Installation and Setup

```bash
# Install pre-commit
pip install pre-commit

# Install the git hook scripts
pre-commit install

# Optional: run against all files
pre-commit run --all-files
```

### 3. CI Integration

Add pre-commit to your CI pipeline to ensure consistency:

```yaml
# In GitHub Actions workflow
- name: Set up pre-commit
  uses: pre-commit/action@v3.0.0
```

### 4. Custom Hook Configuration for Glance

For the Glance project specifically, consider:

- Configure `golangci-lint` with custom rules suitable for the project
- Set appropriate file size limits (usually 500KB-1MB is reasonable)
- Enable specific linters that align with the project's code style
- Create custom hooks for project-specific validation if needed

### 5. Performance Optimization

- Use `golangci-lint` instead of individual linters for better performance
- Consider using `--no-verify` for large changes temporarily, then run hooks separately
- Configure caching to speed up repeated runs

## Implementation for Glance

For the Glance project, the following implementation is recommended:

1. Create `.pre-commit-config.yaml` with the hooks mentioned above
2. Add a `.golangci.yml` configuration file to customize linter settings
3. Document the pre-commit setup in project documentation
4. Add pre-commit badge to README.md to indicate usage
5. Configure CI to run the same checks as pre-commit

This setup will ensure code quality and consistency across the Glance project, with particular attention to the Go 1.23+ compatibility requirements.