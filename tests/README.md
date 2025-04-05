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

The CLI execution tests in `main_test.go` require the compiled binary. To run these tests:

1. Build the binary:
   ```bash
   go build -o glance
   ```

2. Run the tests with the appropriate environment variable:
   ```bash
   TEST_WITH_COMPILED_BINARY=true go test -v
   ```

## Adding New Tests

When adding new tests, follow these guidelines:

1. **Unit Tests**:
   - Place in the appropriate `_test.go` file based on what you're testing
   - Use `github.com/stretchr/testify/assert` for assertions
   - Focus on testing individual functions and components

2. **Integration Tests**:
   - Place in `main_test.go`
   - Use the `TestWithCompiledBinary` approach for tests that need the binary

3. **Mocking Dependencies**:
   - Create mock implementations using `github.com/stretchr/testify/mock`
   - See `mock_test.go` for examples of how to use mocks

## Test Helper Functions

- `setupTestDir(t *testing.T, prefix string)`: Creates a temporary test directory and returns a cleanup function
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