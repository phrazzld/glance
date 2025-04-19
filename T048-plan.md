# T048 Plan: Add WithPromptTemplate Option Function

## Context
- Current Ticket: T048 - add WithPromptTemplate option function
- Related to Change Request CR-06: Fix LLM template handling
- This is a follow-up to T047, which added the PromptTemplate field to ServiceOptions

## Current Implementation
- In `llm/service.go`, the `ServiceOptions` struct now has a `PromptTemplate` field
- There is a method on the `ServiceOptions` struct: `WithPromptTemplate(template string) *ServiceOptions`
- However, there is no standalone option function like:
  ```go
  func WithPromptTemplate(template string) ServiceOption {
      // ...
  }
  ```
- This standalone function is needed to configure the service during creation with `NewService`

## Changes Needed
1. **Add the `WithPromptTemplate` option function**:
   - Add a function that matches the pattern of existing option functions
   - Return a `ServiceOption` that sets the `PromptTemplate` field
   - Follow the same pattern as the existing `WithMaxRetries`, `WithModelName`, and `WithVerbose` functions

## Verification
- Make sure all tests still pass
- Verify code compiles correctly

## Limitations
- This ticket only adds the option function
- Using the template value to update template loading behavior will be done in T049
