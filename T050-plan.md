# T050 Plan: Add Tests for Custom Template Handling

## Task Description
- **T050 · test · p2: add tests for custom template handling**
- **Context:** CR-06: Fix LLM template handling
- **Action:**
  1. Test service with custom template
  2. Verify generated content reflects template
- **Done when:**
  1. Tests pass for custom templates via `--prompt-file`

## Implementation Approach

This task involves testing the end-to-end functionality of custom template handling, focusing on the ability to specify custom prompt templates via the `--prompt-file` flag.

The testing will focus on two main aspects:
1. Integration testing of the `config.LoadConfig` function to ensure it correctly loads and handles the prompt template from a file
2. Testing the LLM service to verify that the template content is properly passed to the prompt generation process

### Steps

1. Create integration tests that:
   - Create a temporary prompt template file with distinctive content
   - Configure the application to use this custom template file
   - Verify that the generated content incorporates the custom template

2. Ensure tests cover:
   - Happy path (valid template file)
   - Error cases (invalid file path, invalid template format)
   - Default behavior (no template file specified)

## Files to Modify
- `/Users/phaedrus/Development/glance/integration_test.go`: Add integration tests for custom template

## Test Approach
Since this is primarily a test task, we'll write comprehensive tests that verify:
1. The config loading process correctly reads and loads the template
2. The template content is correctly used when generating prompts
3. The LLM service correctly uses the template from options
