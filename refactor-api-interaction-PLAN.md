# Refactor API Interaction

## Task Goal
Move the Gemini API interaction code from the glance.go file to the llm package and utilize the previously created Client interface and prompt generation module. This will improve code organization, testability, and make the codebase more maintainable by properly separating concerns.

## Implementation Approach
I'll create a new file `llm/service.go` that provides high-level LLM operations specific to the Glance application. The implementation will follow these steps:

1. **Create LLM Service**:
   - Define a `Service` struct that encapsulates LLM-related functionality, composed of:
     - A `Client` interface instance
     - Configuration options

2. **Add Primary Method**:
   - Implement `GenerateGlanceMarkdown(ctx context.Context, dir string, fileMap map[string]string, subGlances string) (string, error)`
   - This method will:
     - Use the prompt module to create and format prompts
     - Use the Client interface to generate content
     - Handle retries and error reporting

3. **Update Glance.go**:
   - Refactor the `generateMarkdown` function to use the new LLM service
   - Remove direct Gemini API calls from glance.go
   - Inject the LLM service during application startup

4. **Implement Factory Function**:
   - Add `NewService(client Client, options ...ServiceOption)` constructor
   - Support optional configuration through functional options pattern

5. **Error Handling**:
   - Centralize error handling for LLM operations
   - Provide meaningful error messages for different failure modes

## Key Reasoning
This approach is best because:

1. **Proper Separation of Concerns**: By moving API interaction code to a dedicated package, we maintain a cleaner architecture where each component has a single responsibility.

2. **Improved Testability**: The service layer can be easily mocked in tests, allowing testing of code that uses LLM functionality without making actual API calls.

3. **Flexibility**: Using the Client interface allows for easy swapping of different LLM providers in the future without changing the rest of the application.

4. **Cleaner Main Package**: The main package (glance.go) becomes focused on application flow and coordination rather than implementation details of API calls.

5. **Reusability**: The service can be potentially reused by other parts of the application that need to interact with LLMs.

This approach aligns with the project's ongoing refactoring efforts to improve code organization and modularity, making the codebase more maintainable in the long term.