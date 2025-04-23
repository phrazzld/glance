# Todo

## Logging Refactor and Cleanup

- [x] **T001 · Chore · P1: remove backup files from version control**
    - **Context:** cr-01 Remove Backup Files from Repository
    - **Action:**
        1. Run `git rm glance.go.bak integration_test.go.bak`.
        2. Add `*.bak` to the `.gitignore` file.
        3. Commit the removal and the `.gitignore` update.
    - **Done‑when:**
        1. `glance.go.bak` and `integration_test.go.bak` are no longer tracked by git.
        2. `.gitignore` prevents future tracking of `*.bak` files.
    - **Depends‑on:** none

- [x] **T002 · Feature · P1: make logging level configurable via environment variable**
    - **Context:** cr-02 Make Logging Level Configurable
    - **Action:**
        1. Modify logging setup (`setupLogging` or `config.LoadConfig`) to read `GLANCE_LOG_LEVEL` env var.
        2. Parse var ("debug", "info", "warn", "error") mapping to `logrus.Level`, defaulting to `logrus.InfoLevel`.
        3. Call `logrus.SetLevel()` with the determined level.
    - **Done‑when:**
        1. Running the application with `GLANCE_LOG_LEVEL=debug` shows debug logs.
        2. Running the application with `GLANCE_LOG_LEVEL=warn` (or unset) hides debug and info logs.
        3. Running with an invalid value defaults to info level.
    - **Depends‑on:** none

- [x] **T003 · Refactor · P1: decouple filesystem functions from global logger state**
    - **Context:** cr-03 Decouple Filesystem Package from Global Logger State
    - **Action:**
        1. Identify functions/structs in `filesystem` package using global `logrus`.
        2. Modify identified functions/structs to accept a `logger logrus.FieldLogger` parameter/field.
        3. Update all call sites to pass a logger instance and replace global `logrus` calls with injected logger calls.
    - **Done‑when:**
        1. No functions in the `filesystem` package directly reference the global `logrus` instance or its state.
        2. Dependencies on a logger are explicit via parameters or struct fields.
        3. All existing tests for the `filesystem` package pass.
    - **Depends‑on:** none

- [x] **T004 · Test · P1: fix skipped configuration and template tests**
    - **Context:** cr-04 Fix Skipped Configuration Tests
    - **Action:**
        1. Identify skipped tests in `config/loadconfig_test.go` and `config/template_test.go`.
        2. Refactor test setup to use `t.TempDir()` for temporary file/directory creation.
        3. Ensure paths used in tests conform to `filesystem.ValidateFilePath` requirements.
        4. Remove `t.Skip()` calls.
    - **Done‑when:**
        1. All previously skipped tests in `config/loadconfig_test.go` now pass.
        2. All previously skipped tests in `config/template_test.go` now pass.
    - **Depends‑on:** none

- [x] **T005 · Refactor · P2: remove redundant log level checks**
    - **Context:** cr-05 Remove Redundant Log Level Checks
    - **Action:**
        1. Find all instances of `if logrus.IsLevelEnabled(...) { logger.Level(...) }`.
        2. Remove the explicit `if logrus.IsLevelEnabled(...)` check, relying on the logger method's internal check.
    - **Done‑when:**
        1. Explicit `IsLevelEnabled` checks preceding logging calls are removed.
        2. Logging behavior remains unchanged (correct level filtering still occurs).
    - **Depends‑on:** [T003]

- [x] **T006 · Refactor · P2: standardize log messages to structured format**
    - **Context:** cr-06 Standardize Log Message Formatting
    - **Action:**
        1. Identify all logging calls using simple string formatting (e.g., `logrus.Infof("User %s logged in", user)`).
        2. Convert these calls to use structured logging (e.g., `logrus.WithField("user", user).Info("User logged in")`).
        3. Ensure consistent field naming conventions are used.
    - **Done‑when:**
        1. All application log entries use structured formatting (`WithFields`).
        2. Log output is consistent and easily parsable.
    - **Depends‑on:** none

- [x] **T007 · Refactor · P2: remove correlation id functionality**
    - **Context:** cr-08 Re-evaluate Correlation ID Necessity
    - **Action:**
        1. Remove the `generateCorrelationID` function (likely in `llm/service.go`).
        2. Remove all `correlation_id` fields from `logrus.WithFields` calls throughout the codebase.
        3. Update any tests asserting the presence of `correlation_id` in logs.
    - **Done‑when:**
        1. Correlation ID generation code is removed.
        2. No log entries contain the `correlation_id` field.
        3. Relevant tests pass.
    - **Depends‑on:** none

- [x] **T008 · Chore · P2: clarify readme logging description**
    - **Context:** cr-11 Clarify README Logging Description
    - **Action:**
        1. Update the "Logging" section in `README.md` to explain the default level (`info`).
        2. Document the `GLANCE_LOG_LEVEL` environment variable, its purpose, and valid values ("debug", "info", "warn", "error").
        3. Add examples showing how to set the log level.
    - **Done‑when:**
        1. `README.md` accurately describes default logging behavior.
        2. `README.md` clearly explains how to configure the log level using `GLANCE_LOG_LEVEL`.
    - **Depends‑on:** [T002]

## Glance Binary Crash Resolution

- [x] **T010 · Bug · P1: fix API finish reason handling for "STOP" response**
    - **Context:** Consultation on binary crash issue
    - **Action:**
        1. Modify `llm/client.go` to accept "STOP" as a valid finish reason from the Gemini API.
        2. Update the condition that checks for successful completion to include "STOP" alongside "FINISHED".
        3. Run tests to ensure proper handling of the modified condition.
    - **Done‑when:**
        1. The `client.go` code considers both "FINISHED" and "STOP" as valid completion reasons.
        2. The binary no longer crashes when running on directories.
        3. All tests continue to pass.
    - **Depends‑on:** none

- [ ] **T011 · Refactor · P1: remove remaining emojis from UI components**
    - **Context:** Continuation of emoji removal from UI feedback
    - **Action:**
        1. Remove all remaining emoji characters from spinner messages in `ui/feedback.go`.
        2. Remove all remaining emoji characters from progress bar descriptions in `ui/feedback.go`.
        3. Replace emojis with plain text alternatives throughout the file.
    - **Done‑when:**
        1. No emoji characters remain in the `ui/feedback.go` file.
        2. Running the application shows no emojis in output.
        3. All UI components (spinners, progress bars) continue to function correctly.
    - **Depends‑on:** none

- [ ] **T012 · Chore · P2: verify Gemini API key and model compatibility**
    - **Context:** Ensure proper configuration for API access
    - **Action:**
        1. Verify that the API key in `.env` has the proper permissions.
        2. Confirm that the "gemini-2.0-flash" model is correctly specified and accessible.
        3. Check if any model-specific settings need adjustment.
    - **Done‑when:**
        1. The application successfully connects to the Gemini API.
        2. Content generation completes without errors.
        3. Proper error messages are displayed if there are issues with the API key or model.
    - **Depends‑on:** none

- [ ] **T013 · Test · P2: add test for API finish reason handling**
    - **Context:** Test the fix for "STOP" finish reason
    - **Action:**
        1. Create or update tests in `llm/client_test.go` to verify that "STOP" finish reason is handled correctly.
        2. Ensure the test covers scenarios where the API returns different finish reasons.
        3. Verify that the correct response is returned when finish reason is "STOP".
    - **Done‑when:**
        1. Test for "STOP" finish reason handling exists.
        2. Test passes successfully.
        3. Test provides adequate coverage for the modified condition.
    - **Depends‑on:** [T010]

- [ ] **T014 · Chore · P3: finalize and verify emoji removal across codebase**
    - **Context:** Ensure consistent approach to UI output
    - **Action:**
        1. Perform a global search for any remaining emoji characters.
        2. Verify all user-facing text for consistent, professional tone.
        3. Ensure all string literals in logging and UI follow the same style.
    - **Done‑when:**
        1. No emojis remain in any source files.
        2. Application output is consistent across all components.
        3. User experience maintains clarity without emoji decorations.
    - **Depends‑on:** [T011]
