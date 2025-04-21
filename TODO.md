# TODO.md - Migration to google.golang.org/genai

This document outlines the specific tasks required to migrate the Glance project from `github.com/google/generative-ai-go/genai` to `google.golang.org/genai`, following the implementation plan.

## [x] [T001] Add New Dependency: `google.golang.org/genai`

**Description:**
Add the official `google.golang.org/genai` package to the project's `go.mod` file.

**Acceptance Criteria:**
- `google.golang.org/genai` is listed as a requirement in `go.mod`.
- Running `go mod download google.golang.org/genai` succeeds.
- Running `go mod tidy` retains the new dependency.

**Depends On:** None

**Estimated Effort:** Low

## [x] [T002] Update Imports and Client Initialization in `llm/client.go`

**Description:**
Replace imports of the old `genai` package with `google.golang.org/genai` in `llm/client.go`. Update the `NewClient` function signature and internal logic to use the new package's `genai.NewClient` function and associated options/types.

**Acceptance Criteria:**
- `llm/client.go` imports `google.golang.org/genai`.
- `NewClient` function compiles successfully.
- Client initialization logic uses the new package's API (`genai.NewClient`).
- Basic configuration (like API key handling) is correctly mapped to the new client initialization.

**Depends On:** T001

**Estimated Effort:** Medium

## [x] [T003] Update Model Selection Logic in `llm/client.go`

**Description:**
Modify the code within the LLM client implementation (`llm/client.go`) that selects or retrieves the generative model instance. This typically involves replacing calls like `client.GetModel` or similar with the new `client.GenerativeModel` pattern.

**Acceptance Criteria:**
- Code responsible for obtaining a model handle (e.g., `client.GenerativeModel("gemini-pro")`) compiles.
- Model name configuration is correctly passed to the new API.

**Depends On:** T002

**Estimated Effort:** Low

## [x] [T004] Refactor Non-Streaming API Calls in `llm/client.go`

**Description:**
Update the client methods responsible for making non-streaming (synchronous) API calls (e.g., `GenerateContent`). Adapt the request construction (prompt, parts, safety settings) and response handling (extracting text content, checking errors) to match the new `GenerativeModel.GenerateContent` API.

**Acceptance Criteria:**
- Non-streaming API call methods compile.
- Request structures (`genai.Content`, `genai.Part`) are correctly populated using the new types.
- Response data is correctly extracted from the new response types (`genai.GenerateContentResponse`).
- Basic error handling for non-streaming calls uses the new API's error patterns.

**Depends On:** T003

**Estimated Effort:** Medium

## [x] [T005] Refactor Streaming API Calls in `llm/client.go`

**Description:**
Update the client methods responsible for making streaming API calls (e.g., `GenerateContentStream`). Adapt the request construction and response handling logic to use the new `GenerativeModel.GenerateContentStream` API, including iterator patterns and handling of streamed `GenerateContentResponse` chunks.

**Acceptance Criteria:**
- Streaming API call methods compile.
- Request structures are correctly populated for streaming calls.
- The response stream iterator (`genai.GenerateContentResponseIterator`) is used correctly.
- Content is correctly extracted from streamed response chunks.
- Stream termination and error handling (`iterator.Stop`, error checks) are correctly implemented.

**Depends On:** T003

**Estimated Effort:** Medium

## [x] [T006] Update Token Counting Logic in `llm/client.go`

**Description:**
Update the code responsible for counting tokens to use the corresponding functionality in the new `google.golang.org/genai` package (e.g., `client.CountTokens`). Adapt the request structure as needed for this specific API call.

**Acceptance Criteria:**
- Token counting methods compile.
- Requests to the `CountTokens` API use the correct types and structure.
- Token count is correctly extracted from the response.

**Depends On:** T003

**Estimated Effort:** Medium

## [x] [T007] Update Error Handling Logic in `llm/client.go`

**Description:**
Review and adapt all error checking and handling within `llm/client.go` to match the error types, patterns, and specific error conditions returned by the new `google.golang.org/genai` package. Ensure errors are appropriately wrapped or propagated.

**Acceptance Criteria:**
- Error handling code compiles.
- Error checks correctly identify and handle errors specific to the new API (e.g., API key errors, content filtering, rate limits).
- Errors are logged or returned consistently with the rest of the application.

**Depends On:** T004, T005, T006

**Estimated Effort:** Medium

## [x] [T008] Update Functional Options and Configuration in `llm/client.go`

**Description:**
Review and update any functional options used during client/model setup or API calls (e.g., for setting temperature, topP, topK, safety settings, timeouts, retries). Map existing configuration values to the equivalent options in the new package.

**Acceptance Criteria:**
- Configuration code related to generation parameters compiles.
- Functional options or request fields for parameters like temperature, safety settings, etc., use the new API's types and patterns.
- Existing configuration mechanisms (e.g., env vars, flags) correctly influence the settings applied to the new API calls.

**Depends On:** T002, T004, T005

**Estimated Effort:** Medium

## [x] [T009] Update `llm.Client` Interface

**Description:**
Review the `llm.Client` interface definition in `llm/client.go` (or wherever it's defined). Update method signatures (parameters, return types) if they have changed as a result of the refactoring in T002-T008 to accurately reflect the capabilities and types of the new implementation.

**Acceptance Criteria:**
- The `llm.Client` interface definition compiles.
- Method signatures in the interface match the refactored implementation in `llm/client.go`.
- Code consuming this interface (like `llm/service.go`) may show compile errors, which is expected at this stage.

**Depends On:** T007, T008

**Estimated Effort:** Low

## [x] [T010] Update Mocks for `llm.Client`

**Description:**
Regenerate or manually update any mock implementations of the `llm.Client` interface (e.g., using `gomock` or similar, likely in `internal/mocks`). Ensure the mocks align with the updated interface definition from T009.

**Acceptance Criteria:**
- Mock generation commands succeed (if applicable).
- Mock implementation files compile successfully.
- Mock methods match the signatures defined in the updated `llm.Client` interface.

**Depends On:** T009

**Estimated Effort:** Medium

## [x] [T011] Update `llm/service.go` to Use Refactored Client

**Description:**
Modify the LLM service layer (`llm/service.go`) to work with the updated `llm.Client` interface (T009). Adjust how service methods call the client, handle returned data types, and manage errors based on the changes introduced by the new client implementation.

**Acceptance Criteria:**
- `llm/service.go` compiles successfully.
- Service methods correctly call the updated `llm.Client` interface methods.
- Type conversions or handling logic for new request/response types are implemented correctly within the service.
- Error handling within the service layer correctly interprets errors propagated from the updated client.

**Depends On:** T009

**Estimated Effort:** Medium

## [x] [T012] Update Main Application/CLI Code and Call Sites

**Description:**
Review and update code in the main application entry points (`cmd/`, `main.go`, etc.) or any other locations that directly instantiate or interact with the `llm.Client` or `llm.Service`. Ensure configuration is passed correctly, context handling is appropriate, and any direct usage of old types/methods is removed or updated.

**Acceptance Criteria:**
- Main application code compiles successfully.
- Instantiation of LLM client/service uses the updated constructors or methods.
- Configuration values (e.g., model name, API key) are correctly passed down to the LLM components.
- Any code directly handling LLM responses or errors is updated for the new types/patterns.

**Depends On:** T011

**Estimated Effort:** Medium

## [ ] [T013] Update Unit Tests for `llm/client_test.go`

**Description:**
Modify the unit tests in `llm/client_test.go` to align with the refactored client implementation (T002-T008). Update test setup, assertions, and potentially use updated mocks (T010 is implicitly needed, but tests depend on the *implementation*). Ensure tests cover the behavior of the client using the new `google.golang.org/genai` API.

**Acceptance Criteria:**
- All unit tests in `llm/client_test.go` pass (`go test ./llm/...`).
- Tests correctly mock the underlying `google.golang.org/genai` interactions where necessary (or use the updated `llm.Client` mock).
- Test assertions are updated to reflect the expected behavior and response types of the new API.
- Test coverage for `llm/client.go` is maintained or improved.

**Depends On:** T007, T008, T010 (mock availability)

**Estimated Effort:** High

## [ ] [T014] Update Unit Tests for `llm/service_test.go`

**Description:**
Modify the unit tests in `llm/service_test.go` to align with the refactored service implementation (T011). Update test setup, assertions, and ensure the tests correctly use the updated `llm.Client` mock (T010).

**Acceptance Criteria:**
- All unit tests in `llm/service_test.go` pass (`go test ./llm/...`).
- Tests correctly use the updated `llm.Client` mock (T010) to simulate client behavior.
- Test assertions are updated to reflect any changes in the service layer's logic or return types.

**Depends On:** T011, T010

**Estimated Effort:** Medium

## [ ] [T015] Update Other Relevant Unit Tests

**Description:**
Review and update any other unit tests throughout the codebase (e.g., in `cmd/` tests) that might be indirectly affected by the changes in the LLM client/service interfaces, types, or behavior.

**Acceptance Criteria:**
- All potentially affected unit tests across the project pass (`go test ./...`).

**Depends On:** T012, T013, T014

**Estimated Effort:** Medium

## [ ] [T016] Run and Update Integration/E2E Tests

**Description:**
Execute the project's full suite of integration and end-to-end (E2E) tests. Identify any failures caused by the migration to the new `genai` package. Update test setup, execution steps, or assertions as needed to ensure they pass with the refactored code. This may involve running against a real API endpoint or carefully mocked integration environment.

**Acceptance Criteria:**
- All integration tests pass.
- All E2E tests (e.g., CLI tests) pass.
- Tests accurately validate the application's behavior using the new `google.golang.org/genai` client.

**Depends On:** T012, T015

**Estimated Effort:** High

## [ ] [T017] Update README and Documentation Files

**Description:**
Search and replace all references to the old package path (`github.com/google/generative-ai-go/genai`) with the new path (`google.golang.org/genai`) in the main `README.md` file and any other documentation files (e.g., under a `docs/` directory). Update any setup instructions, environment variable mentions, or usage examples that might have changed due to the migration.

**Acceptance Criteria:**
- README and all files in `docs/` reference `google.golang.org/genai`.
- Setup instructions (dependency installation, API key setup) are accurate for the new package.
- Usage examples reflect any changes in API or configuration.

**Depends On:** T016 (ensures functionality is stable before documenting)

**Estimated Effort:** Medium

## [ ] [T018] Update Inline Code Comments

**Description:**
Review code comments, particularly within the `llm` package and related areas. Update any comments that specifically mention the old package path, its types, or methods to accurately reflect the new `google.golang.org/genai` implementation.

**Acceptance Criteria:**
- Inline code comments accurately reference the new package and its concepts where applicable.
- Outdated comments referring to the old API are removed or updated.

**Depends On:** T016

**Estimated Effort:** Low

## [ ] [T019] Review and Update Developer Scripts

**Description:**
Check any scripts used for development (e.g., build scripts, test runners, setup scripts in `scripts/` or `Makefile`) for hardcoded references to the old package path (`github.com/google/generative-ai-go/genai`) and update them if necessary.

**Acceptance Criteria:**
- Developer scripts execute correctly.
- Scripts do not contain outdated references to the old dependency path.

**Depends On:** T016

**Estimated Effort:** Low

## [ ] [T020] Remove Old Dependency and Tidy `go.mod`

**Description:**
Remove the line requiring `github.com/google/generative-ai-go/genai` from the `go.mod` file. Run `go mod tidy` to clean up the module file and ensure consistency.

**Acceptance Criteria:**
- The `github.com/google/generative-ai-go/genai` dependency is no longer present in `go.mod` or `go.sum`.
- `go mod tidy` runs without errors and does not remove `google.golang.org/genai`.
- The project still builds and all tests pass after this change.

**Depends On:** T016 (ensures all code usages are removed and tests confirm functionality)

**Estimated Effort:** Low

## [ ] [T021] Remove Dead Code Related to Old API

**Description:**
Identify and remove any helper functions, types, constants, or variables that were specifically created to support or workaround aspects of the old `github.com/google/generative-ai-go/genai` package and are no longer needed with the new `google.golang.org/genai` package.

**Acceptance Criteria:**
- Unused code artifacts related solely to the old API are removed.
- The project still builds and all tests pass after removal.
- Codebase is cleaner and free of legacy helpers for the old API.

**Depends On:** T020

**Estimated Effort:** Low

## [ ] [T022] Run Code Formatting and Quality Checks

**Description:**
Run all standard code formatting (`go fmt ./...`), linting (`golangci-lint run`), and any other configured pre-commit hooks or CI quality checks across the entire codebase. Fix any issues reported by these tools.

**Acceptance Criteria:**
- `go fmt ./...` reports no changes.
- `golangci-lint run` (or equivalent) passes without errors.
- All pre-commit hooks / CI quality checks pass.

**Depends On:** T021

**Estimated Effort:** Low

## [ ] [T023] Perform Final Code Review and Manual Test

**Description:**
Conduct a final, holistic code review focusing on the changes made during the migration. Ensure consistency, correctness, and adherence to project standards. Perform a manual test run of the application's core functionality (e.g., running the Glance CLI against a sample directory with a valid API key) to provide a final sanity check.

**Acceptance Criteria:**
- Code review comments (if any) are addressed.
- Manual execution of the application's primary use case(s) succeeds and produces the expected output.
- No remaining references to the old package are found via search.
- The changes adhere to the project's development philosophy (simplicity, modularity).

**Depends On:** T016, T022

**Estimated Effort:** Medium
