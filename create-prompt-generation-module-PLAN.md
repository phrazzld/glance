# Create Prompt Generation Module

## Task Goal
Create llm/prompt.go to handle prompt template handling and generation, moving this functionality from the glance.go and config packages to a dedicated module in the LLM package. This will centralize prompt-related code, making it more maintainable and testable.

## Implementation Approach
I'll create a new file `llm/prompt.go` that contains:

1. **PromptData Structure**: 
   - Define a structure similar to the existing promptData but in the llm package
   - Include fields for Directory, SubGlances, and FileContents

2. **Template Management Functions**:
   - `LoadTemplate(path string) (string, error)` - Load a template from a file or use the default
   - `DefaultTemplate() string` - Return the default prompt template

3. **Prompt Generation Function**:
   - `GeneratePrompt(data *PromptData, templateStr string) (string, error)` - Generate a prompt by filling the template with data

4. **Helper Functions**:
   - Functions to help format the file contents and other data for the prompt

The implementation will use Go's text/template package for template processing and will handle errors appropriately.

## Key Reasoning
This approach is best because:

1. **Package Organization**: Moving prompt generation to the llm package aligns with the project's ongoing refactoring to organize functionality into dedicated packages. Prompt generation is logically part of the LLM interaction flow.

2. **Reusability and Configurability**: By creating dedicated functions for template loading and prompt generation, we make it easier to reuse this code and configure different templates for different purposes in the future.

3. **Testing**: Separating prompt generation from API interaction makes it easier to test template rendering without making API calls.

4. **Reduced Duplication**: Currently, prompt loading and template code appears in both glance.go and config/loadconfig.go. Centralizing it eliminates this duplication.

5. **Cleaner Abstractions**: This creates a cleaner separation between template management (llm/prompt.go), LLM client interaction (llm/client.go), and application logic (glance.go).

This implementation supports the future task "Refactor API Interaction" by setting up the necessary prompt generation functionality that will be used when the API interaction code is moved to the llm package.