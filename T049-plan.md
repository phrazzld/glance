# T049 Plan: Update service to use stored template

## Task Description
- **T049 · refactor · p2: update service to use stored template**
- **Context:** CR-06: Fix LLM template handling
- **Action:**
  1. Modify service to store template from options
  2. Remove fallback template loading in `GenerateGlanceMarkdown`
- **Done when:**
  1. Service uses template from options
  2. Fallback loading is removed

## Implementation Approach

The current implementation loads the template inside the `GenerateGlanceMarkdown` method, ignoring the `PromptTemplate` field that was added in T047. This task will refactor the code to use the stored template from options rather than loading it each time.

Steps:
1. Modify `GenerateGlanceMarkdown` to check the `options.PromptTemplate` field first
2. If the `PromptTemplate` is set, use it directly
3. If the `PromptTemplate` is empty, then call `LoadTemplate` to load a template from file
4. Pass the resulting template to `GeneratePrompt`
5. Add tests to verify the changes

## Files to Modify
- `llm/service.go`: Update the `GenerateGlanceMarkdown` method

## Tests
- Ensure existing tests pass
- Add a test case to verify that a custom template from options is used instead of loading from file
