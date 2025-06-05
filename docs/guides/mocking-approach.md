# Mocking Strategy Guidelines

This document outlines the balanced approach to mocking in the Glance project, ensuring we maintain a good balance between testability and simplicity.

## Core Principles

1. **Interface-Based Mocking at True Boundaries**
   - Use interfaces for true API boundaries where multiple implementations exist
   - Examples: `llm.Client`, filesystem operations
   - Implement these using testify/mock

2. **Function Variable Mocking for Internal Operations**
   - For single-function helpers or internal implementation details
   - Replace the function variable during tests and restore in a defer statement
   - Example: `var createGeminiClient = func(...) { ... }`

3. **Keep Mocks Minimal and Focused**
   - Only mock what's needed for the test
   - Avoid over-abstraction and unnecessary layers
   - Mock at the most appropriate level of abstraction

## When to Use Interface-Based Mocking

- External service boundaries (LLM APIs, file system)
- Components where you have/expect multiple implementations
- When behavior is complex and benefits from the structured approach of mock expectations

## When to Use Function Variable Mocking

- Internal helpers with single implementations
- Factory functions whose implementation details aren't important to the test
- Functions where the implementation is likely to change but the interface remains stable

## Example: Interface-Based Mocking

```go
// Define the interface
type Client interface {
    Generate(ctx context.Context, prompt string) (string, error)
    CountTokens(ctx context.Context, prompt string) (int, error)
    Close()
}

// In tests
mockClient := new(mocks.LLMClient)
mockClient.On("Generate", ctx, "test prompt").Return("response", nil)
mockClient.AssertExpectations(t)
```

## Example: Function Variable Mocking

```go
// Define the function variable
var createGeminiClient = func(apiKey string, options ...Option) (Client, error) {
    return newGeminiClient(apiKey, options...)
}

// In tests
origFunc := createGeminiClient
defer func() { createGeminiClient = origFunc }()

createGeminiClient = func(apiKey string, options ...Option) (Client, error) {
    return mockClient, nil
}
```

## Implementation Guidelines

1. **Keep Mocking in its Place**:
   - Place mocks in the `internal/mocks` package when they're used across multiple tests
   - Keep test-specific mocks within the test file

2. **Avoid Factory Interfaces**:
   - Prefer function variables for factory patterns rather than creating factory interfaces
   - This simplifies the codebase while maintaining testability

3. **Clean Up After Tests**:
   - Always restore original function implementations in defer statements
   - Clean up any resources created during tests

By following these guidelines, we maintain a balance between clean architecture and practical testability in a CLI tool context.
