# GitHub Actions Workflows

This document provides comprehensive details about the GitHub Actions workflows configured for the Glance project. These workflows automate testing, linting, building, and quality checks to ensure code reliability and maintainability.

## Table of Contents

1. [Overview](#overview)
2. [Workflow Files](#workflow-files)
3. [Go Tests Workflow](#go-tests-workflow)
4. [Go Linting Workflow](#go-linting-workflow)
5. [Go Build Workflow](#go-build-workflow)
6. [Pre-commit Checks Workflow](#pre-commit-checks-workflow)
7. [Workflow Triggers](#workflow-triggers)
8. [Troubleshooting](#troubleshooting)
9. [Advanced Usage](#advanced-usage)
    - [Secure Tool Installation](#secure-tool-installation)

## Overview

GitHub Actions is a CI/CD platform that automates software development workflows directly within GitHub. For the Glance project, we've configured four primary workflows:

1. **Go Tests**: Runs the test suite across multiple Go versions
2. **Go Linting**: Performs static code analysis and style checks
3. **Go Build**: Verifies builds on multiple operating systems and Go versions
4. **Pre-commit Checks**: Enforces code quality standards via pre-commit hooks

All workflows run on:
- Push to the master branch (excluding documentation changes)
- Pull requests targeting the master branch
- Some workflows have additional scheduled runs for security scans and dependency checks

## Workflow Files

All GitHub Actions workflow files are located in the `.github/workflows/` directory:

- `test.yml`: Go test suite runner
- `lint.yml`: Linting and static analysis
- `build.yml`: Cross-platform build verification
- `precommit.yml`: Pre-commit hooks verification

## Go Tests Workflow

**File:** `.github/workflows/test.yml`

This workflow runs the Glance test suite with race detection enabled.

### Configuration

- **Name:** Go Tests
- **Timeout:** 10 minutes per job
- **Environment:**
  - Matrix of Go versions: 1.23 and 1.24
  - Ubuntu latest

### Steps

1. Set up Go environment with specified version
2. Check out the repository code
3. Verify Go modules
4. Run tests with race detection enabled
5. Generate and upload code coverage reports

### Example Output

When successful, all tests pass and coverage reports are generated:

```
ok  	github.com/phrazzld/glance/config	0.253s	coverage: 92.5% of statements
ok  	github.com/phrazzld/glance/errors	0.157s	coverage: 96.3% of statements
ok  	github.com/phrazzld/glance/filesystem	0.328s	coverage: 89.1% of statements
ok  	github.com/phrazzld/glance/llm	0.421s	coverage: 86.2% of statements
ok  	github.com/phrazzld/glance/ui	0.189s	coverage: 91.7% of statements
ok  	github.com/phrazzld/glance	0.635s	coverage: 88.4% of statements
```

### What It Catches

- Unit test failures
- Race conditions in concurrent code
- Integration test failures
- Insufficient test coverage

### Artifacts

- Coverage profiles (saved for 7 days)
- Coverage reports for each Go version

## Go Linting Workflow

**File:** `.github/workflows/lint.yml`

This workflow performs static code analysis to identify potential bugs, style issues, and other code quality problems.

### Configuration

- **Name:** Go Linting
- **Timeout:** 5 minutes per job
- **Environment:**
  - Go 1.24
  - Ubuntu latest

### Steps

1. Set up Go environment
2. Check out repository code
3. Run golangci-lint with configuration from `.golangci.yml`
4. Run additional checks:
   - Go vet
   - Go mod verify
   - Go mod tidy check
   - Spelling check (US locale)

### Example Output

When issues are found, golangci-lint provides detailed output:

```
filesystem/scanner.go:106:6: ineffectual assignment to err (ineffassign)
	if err := scanner.processPath(path); err != nil {
	     ^
llm/client.go:75:9: S1005: unnecessary assignment to the blank identifier (gosimple)
	if _ = resp.Clean(); err != nil {
	        ^
```

### What It Catches

- Unused variables and imports
- Inefficient code patterns
- Potential bugs and logical errors
- Style inconsistencies
- Untidy go.mod files
- Spelling errors in code
- Possible race conditions

### Schedule

In addition to the standard triggers, this workflow runs monthly to check dependencies:
- Scheduled on the 1st of each month at 01:00 UTC

## Go Build Workflow

**File:** `.github/workflows/build.yml`

This workflow verifies that Glance builds successfully across multiple operating systems and Go versions.

### Configuration

- **Name:** Go Build
- **Timeout:** 5 minutes per job
- **Environment:**
  - Matrix of:
    - Go versions: 1.23 and 1.24
    - OS: Ubuntu, macOS, Windows
  - Bash shell on all platforms

### Steps

1. Set up Go environment with specified version
2. Check out repository code
3. Verify Go modules
4. Build Glance with optimized flags (`-ldflags="-s -w"`)
5. Verify the binary exists
6. Upload built binaries as artifacts

### Example Output

Successful builds will show:

```
Run go build -ldflags="-s -w" -o glance
```

### What It Catches

- Build errors on different operating systems
- OS-specific compatibility issues
- Dependency resolution problems
- Compatibility issues with different Go versions

### Artifacts

- Built binaries for each OS and Go version combination
- Named format: `glance-{os}-{go-version}` (e.g., `glance-ubuntu-1.24`)

## Pre-commit Checks Workflow

**File:** `.github/workflows/precommit.yml`

This workflow runs all configured pre-commit hooks to enforce code quality standards.

### Configuration

- **Name:** Pre-commit Checks
- **Environment:**
  - Go 1.24
  - Python 3.10
  - golangci-lint v1.57.0 (installed via pre-commit)

### Steps

1. Set up Go and Python environments
2. Check out repository code
3. Install pre-commit
4. Run pre-commit hooks on all files (including golangci-lint hook)

### Example Output

Pre-commit output shows the status of each hook:

```
Check Yaml..........................................Passed
Fix End of Files...................................Passed
Trim Trailing Whitespace...........................Passed
golangci-lint......................................Failed
- hook id: golangci-lint
- exit code: 1

go-fmt............................................Passed
go-imports........................................Passed
go-vet............................................Passed
go-test...........................................Passed
detect-secrets....................................Passed
```

### What It Catches

- Code formatting issues
- Trailing whitespace
- Missing newlines at end of files
- Go linting issues
- Failed tests
- Potential secrets in code
- Large file additions
- Many other issues based on the configured hooks

### Schedule

In addition to standard triggers, this workflow runs weekly for security scans:
- Scheduled every Sunday at 00:00 UTC

## Workflow Triggers

All workflows use selective trigger conditions to balance thoroughness with performance:

### Common Triggers

- **Push to master branch**
  - Ignores documentation changes (`**.md`, `docs/**`, etc.)
  - Only runs workflows relevant to the changed files

- **Pull requests to master branch**
  - Runs on all pull requests to ensure code quality before merging

### Path Ignore Configuration

All workflows ignore changes to:
- Markdown files (`**.md`)
- Documentation directory (`docs/**`)
- License files
- Issue templates

### Concurrency Control

All workflows implement concurrency control to prevent redundant runs:
```yaml
concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true
```

This configuration:
- Groups workflow runs by workflow name and git ref
- Cancels in-progress workflows when new commits are pushed

## Troubleshooting

### Common Issues

1. **Workflow Timeouts**
   - Check for long-running tests or builds
   - Consider optimizing test execution
   - Increase timeout limits if necessary

2. **Failed Linting Checks**
   - Run `golangci-lint run` locally to identify issues
   - Apply `go fmt` and `go imports` to fix formatting
   - Address issues before pushing again

3. **Build Failures on Specific OS**
   - Check for OS-specific code that may not be compatible
   - Use conditional compilation with build tags if needed
   - Test locally with the specific Go version

4. **Pre-commit Failures**
   - Run `pre-commit run --all-files` locally to identify issues
   - Fix the issues identified by the hooks
   - Commit again after resolving the issues

### Debugging Workflow Runs

To get more detailed information from workflow runs:

1. Enable debug logging by setting a secret:
   ```
   ACTIONS_RUNNER_DEBUG: true
   ACTIONS_STEP_DEBUG: true
   ```

2. Use workflow run annotations to identify specific issues
3. Check the "Annotations" tab in the GitHub Actions workflow run

## Advanced Usage

### Secure Tool Installation

We follow security best practices for installing tools in our GitHub Actions workflows:

#### golangci-lint Installation Methods

We have standardized on specific methods for installing and running golangci-lint. See [LINTING.md](LINTING.md) for the complete standardization policy. In summary:

1. **Pre-commit Hook Installation (Standard for Local Development)**

   The pre-commit framework automatically installs the golangci-lint tool when the hook runs. This method:
   - Uses pre-commit's secure installation mechanism
   - Ensures version consistency with the `.pre-commit-config.yaml` file
   - Eliminates the need for manual installation steps

   Example configuration in `.pre-commit-config.yaml`:
   ```yaml
   - repo: https://github.com/golangci/golangci-lint
     rev: v1.57.0
     hooks:
       - id: golangci-lint
         name: golangci-lint
         entry: golangci-lint run
         types: [go]
         language: system
         pass_filenames: false
         args: ["--config=.golangci.yml", "--timeout=2m"]
   ```

2. **Official GitHub Action (Standard for CI Workflows)**

   For dedicated linting workflows, we use the official golangci-lint GitHub Action:
   ```yaml
   - name: Install golangci-lint
     uses: golangci/golangci-lint-action@v4
     with:
       version: v1.57.0
       args: --config=.golangci.yml
       only-new-issues: true
   ```

   This method:
   - Uses the officially maintained action
   - Has built-in caching for faster execution
   - Provides detailed output formatting
   - Offers additional features like "only-new-issues"

#### Security Considerations

We avoid using `curl | sh` patterns for tool installation because:
- They execute arbitrary code from the internet
- They bypass security checks and verification
- They introduce potential supply chain attack vectors
- They may download unexpected or compromised code if the source is compromised

Using the methods above helps maintain a secure CI/CD pipeline while ensuring consistent tool behavior across workflows.

### Skipping Workflows

To skip CI workflows for minor changes, include `[skip ci]` or `[ci skip]` in your commit message.

### Manual Workflow Dispatch

Most workflows can be manually triggered from the Actions tab in GitHub:
1. Go to the "Actions" tab
2. Select the workflow
3. Click "Run workflow"
4. Select branch and click "Run workflow"

### Workflow Dependencies

Some workflows produce artifacts used by others:
- Test workflow produces coverage reports
- Build workflow produces binaries

Access these artifacts from the workflow run page under the "Artifacts" section.

### Local Workflow Testing

To test workflows locally before pushing:
1. Install [act](https://github.com/nektos/act)
2. Run `act push` to simulate a push event
3. Run `act pull_request` to simulate a pull request event
