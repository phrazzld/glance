# T047 Plan: Add PromptTemplate to ServiceOptions

## Context
- Current Ticket: T047 - add PromptTemplate to ServiceOptions
- Related to Change Request CR-06: Fix LLM template handling

## Current Implementation
- In `llm/service.go`, the `ServiceOptions` struct contains fields for:
  - `MaxRetries`
  - `ModelName`
  - `Verbose`
- Currently, the service doesn't store the prompt template
- In `GenerateGlanceMarkdown`, the template is loaded on each call with `LoadTemplate("")`
- The template is either loaded from a file or uses the default template

## Changes Needed
1. **Add a new field to `ServiceOptions` struct**:
   - Add `PromptTemplate` field of type `string` to the `ServiceOptions` struct
   - Add appropriate documentation comment

2. **Update `DefaultServiceOptions` function**:
   - Set the default value of `PromptTemplate` to empty string `""`
   - This maintains current behavior of loading from a file or using default template

3. **Add method to update the template**:
   - Add `WithPromptTemplate` method to the `ServiceOptions` struct
   - This mirrors the existing pattern for other fields

## Verification
- Make sure all tests still pass
- Verify code compiles correctly

## Limitations
- This ticket only adds the field to `ServiceOptions`
- Using the field and adding the option function will be done in subsequent tickets:
  - T048: add WithPromptTemplate option function
  - T049: update service to use stored template
