# Create Config Package and Struct

## Goal
Create a dedicated configuration package with a Config struct that consolidates all application settings (API key, target directory, flags, prompt template) into a single organized structure.

## Implementation Approach
I will create a new `config` package with a `config.go` file that defines a `Config` struct to encapsulate all application configuration. This implementation will:

1. Create a `config` directory with a `config.go` file
2. Define a `Config` struct containing all necessary configuration fields:
   - APIKey (string): The Gemini API key
   - TargetDir (string): The directory to process
   - Force (bool): Whether to force regeneration of existing GLANCE.md files
   - Verbose (bool): Whether to enable verbose logging
   - PromptTemplate (string): The content of the prompt template
   - MaxRetries (int): Number of API call retries (currently a constant)
   - MaxFileBytes (int64): Maximum file size to process (currently a constant)

3. Keep implementation of flag parsing and environment loading for later task
4. Include proper documentation with idiomatic Go comments
5. Follow Go's standard code organization patterns

Key decisions:
1. Making both `MaxRetries` and `MaxFileBytes` part of the Config struct to make them configurable in the future
2. Keeping the struct immutable after creation (no setters)

## Reasoning
I chose this approach because:

1. **Clean Separation of Concerns**: Moving configuration to a dedicated package creates a clear separation between configuration management and the rest of the application logic.

2. **Improved Testability**: A configuration structure makes it easier to create and inject test configurations, enhancing test coverage and simplifying test setup.

3. **Future Extensibility**: Using a struct makes it easy to add new configuration options in the future without changing function signatures throughout the codebase.

4. **Reduced Global State**: Replacing global variables with a configuration object passed to functions that need it reduces side effects and makes the code more predictable.

5. **Configuration Consistency**: Having all configuration in one place ensures consistent handling of defaults and reduces the chance of misconfiguration.

Alternative approaches considered:
- Using individual function parameters instead of a struct (rejected due to parameter bloat)
- Using a functional options pattern (more complex than needed for this application)
- Keeping configuration in the main package (limits reusability and testability)