# GitHub Actions Test

This directory contains files specifically created to test GitHub Actions workflows. Each file is designed to trigger a specific check:

- `lint_test.go`: Contains linting issues to test the lint workflow
  - Unused variable
  - Extra spacing (go fmt issue)
  - Inefficient string concatenation

- `trailing_whitespace.txt`: Contains trailing whitespace to test pre-commit hooks

- `failing_test.go`: Contains a deliberately failing test
  - Assertion failure with clear error message

- `compilation_error_test.go`: Contains a syntax error to test build validation
  - Missing closing parenthesis
  - Reference to undefined variable

- `build_warning.go`: Contains code that should generate build warnings
  - Unused variables
  - Possible misuse of fmt.Println

- `main.go`: A simple file that ties the test files together

- `EXPECTED_RESULTS.md`: Documentation of what we expect to see from each workflow

## Purpose

These files are part of Task T020: Create test pull request to verify GitHub Actions workflows run correctly and identify issues appropriately.

The goal is to ensure:
1. GitHub Actions workflows are correctly configured
2. All checks run on the appropriate files
3. The workflows correctly identify and report issues
4. The triggers work as expected, including path-ignore settings

## How to Use

1. Push these files to the repository
2. Create a pull request targeting the master branch
3. Monitor the GitHub Actions workflow runs
4. Verify the results match the expectations in `EXPECTED_RESULTS.md`

**Note: This README update should be ignored by most workflow triggers due to path-ignore settings for .md files.**