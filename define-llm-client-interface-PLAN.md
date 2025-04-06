# Define LLM Client Interface

## Task Goal
Create llm/client.go with a Client interface and a GeminiClient implementation. This will provide a clean abstraction for interaction with the LLM API, making it easier to test, maintain, and potentially support alternative LLM providers in the future.

## Implementation Approach
I'll create a new llm package with a client.go file that contains:

1. **Client Interface**: An interface defining the core methods needed for interacting with LLM services:
   - `Generate(ctx context.Context, prompt string) (string, error)` - For synchronous text generation
   - `CountTokens(ctx context.Context, prompt string) (int, error)` - For token counting
   - `Close()` - For proper resource cleanup

2. **GeminiClient Implementation**: A concrete implementation of the Client interface that uses Google's Gemini API:
   - Encapsulate the genai.Client instance
   - Implement all interface methods using the underlying Google Gemini API
   - Handle stream processing, error handling, and context management
   - Add logging for debugging and monitoring API calls

3. **ClientOptions**: A configuration structure for clients with options like:
   - ModelName - The specific model to use (e.g., "gemini-2.0-flash")
   - MaxRetries - Number of retries for failed API calls
   - Timeout - Context timeout for API requests

4. **NewGeminiClient**: A constructor function that creates a properly configured GeminiClient:
   - Takes an API key and options as parameters
   - Sets up the underlying genai.Client
   - Configures appropriate defaults

## Key Reasoning
This approach is best because:

1. **Abstraction**: The interface provides a clean abstraction over the underlying LLM API, which makes it easier to switch providers or update the implementation without affecting the rest of the codebase.

2. **Testability**: With a proper interface, it becomes straightforward to create mock implementations for testing without making actual API calls.

3. **Flexibility**: The approach allows for easily supporting different LLM providers in the future, as each would just need to implement the Client interface.

4. **Centralized Error Handling**: By encapsulating API interactions in a dedicated package, error handling can be standardized and improved across the application.

5. **Configuration Clarity**: The options structure makes it clear what can be configured, with sensible defaults and proper documentation.

This implementation aligns with Go best practices for interface design (small, focused interfaces) and follows the existing project patterns for package organization and function naming.