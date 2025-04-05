# Implement LoadConfig Function

## Goal
Move the flag parsing, environment variable loading, and prompt template loading logic from the main function into a dedicated LoadConfig function in the config package.

## Implementation Approach
I will create a LoadConfig function in the config package that combines all configuration initialization tasks. The approach will:

1. Create a `LoadConfig` function that:
   - Parses command-line flags and validates them
   - Loads environment variables from the .env file if available
   - Loads the prompt template from the file system
   - Returns a fully initialized Config struct

2. Implement a clear error handling strategy where:
   - Flag parsing errors are returned with context
   - Environment loading errors are logged but don't prevent continuing
   - Missing API key results in an error
   - Invalid target directory results in an error
   - Prompt template loading errors are returned with context

3. Make the LoadConfig function flexible:
   - Accept command-line arguments as a parameter
   - Return a properly initialized Config object with all settings
   - Return any errors with appropriate context

4. Move the existing `loadPromptTemplate` function:
   - Move it to the config package
   - Make it unexported (lowercase first letter)
   - Adapt it to use the defaultPromptTemplate from the config package

## Reasoning
I chose this approach because:

1. **Clean Separation of Concerns**: Moves all configuration-related code to the config package, keeping the main function focused on high-level application flow.

2. **Proper Error Handling**: Returns errors instead of calling logrus.Fatal directly, allowing the caller to decide how to handle different error conditions.

3. **Flexible Design**: Accepting arguments as a parameter allows for easier testing and potential future flexibility (e.g., reading configuration from a file).

4. **Consistency with Existing Pattern**: Follows the immutable builder pattern already established in the Config struct.

5. **Code Reuse**: Leverages the existing Config struct and its builder methods, avoiding duplication.

Alternative approaches considered:
- Having LoadConfig call os.Exit directly (rejected as it makes testing difficult and violates separation of concerns)
- Returning raw flag values instead of a Config object (rejected as it loses type safety and the benefits of the Config struct)
- Adding additional validation in LoadConfig beyond basic flag parsing (rejected as validation logic should generally be separate from configuration loading)