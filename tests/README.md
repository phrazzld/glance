# Testing Framework for Glance

This document provides an overview of the testing framework for the Glance project.

## Test Structure

The testing framework consists of several test files:

- `glance_test.go`: Unit tests for core functionality
- `main_test.go`: End-to-end tests that check the CLI execution
- `mock_test.go`: Demonstrates the use of testify/mock for mocking dependencies

## Dependencies

The testing framework uses the following external dependencies:

- `github.com/stretchr/testify/assert`: For assertions and validations
- `github.com/stretchr/testify/require`: For failing tests immediately on critical errors
- `github.com/stretchr/testify/mock`: For creating mock implementations

## Running Tests

### Basic Test Execution

To run all tests:

```bash
go test ./...
```

To run tests with verbose output:

```bash
go test -v ./...
```

### Running CLI Execution Tests

The CLI execution tests in `main_test.go` require the compiled binary and a valid GEMINI_API_KEY. To run these tests:

1. Build the binary:
   ```bash
   go build -o glance
   ```

2. Set your GEMINI_API_KEY environment variable:
   ```bash
   export GEMINI_API_KEY=your-api-key
   ```

3. Run the tests with the appropriate environment variable:
   ```bash
   TEST_WITH_COMPILED_BINARY=true go test -v
   ```

### End-to-End Test Coverage

The end-to-end tests in `main_test.go` cover the following key functionality:

1. **Basic CLI Execution**: Verifies that the CLI can be executed without errors
2. **Usage Information**: Checks that proper usage information is displayed when run without arguments
3. **GLANCE.md Generation**: Tests the basic GLANCE.md file generation in a nested directory structure
4. **Force Flag**: Verifies that the `--force` flag properly regenerates existing GLANCE.md files
5. **Modified Files Detection**: Tests that GLANCE.md is regenerated when files in the directory are modified
6. **Verbose Output**: Checks that the `--verbose` flag provides additional debug information
7. **Custom Prompt Files**: Tests using a custom prompt template file
8. **Change Propagation**: Verifies that changes in subdirectories trigger regeneration in parent directories
9. **Binary File Handling**: Tests that binary files are properly detected and excluded

## Adding New Tests

When adding new tests, follow these guidelines:

1. **Unit Tests**:
   - Place in the appropriate `_test.go` file based on what you're testing
   - Use `github.com/stretchr/testify/assert` for assertions
   - Focus on testing individual functions and components

2. **Integration Tests**:
   - Place in `main_test.go`
   - Use the `TEST_WITH_COMPILED_BINARY` environment variable to conditionally skip tests
   - Check for `GEMINI_API_KEY` when needed

3. **Mocking Dependencies**:
   - Create mock implementations using `github.com/stretchr/testify/mock`
   - See `mock_test.go` for examples of how to use mocks

## Test Helper Functions

- `setupTestDir(t *testing.T, prefix string)`: Creates a temporary test directory and returns a cleanup function
- `setupTestProjectStructure(t *testing.T)`: Creates a complete test project structure with multiple directories, files, and a .gitignore configuration
- More helpers can be added as needed to reduce test boilerplate

## Code Coverage

To run tests with code coverage:

```bash
go test -cover ./...
```

To generate a coverage report:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```