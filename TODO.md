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

- [ ] **T004 · Test · P1: fix skipped configuration and template tests**
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

- [ ] **T005 · Refactor · P2: remove redundant log level checks**
    - **Context:** cr-05 Remove Redundant Log Level Checks
    - **Action:**
        1. Find all instances of `if logrus.IsLevelEnabled(...) { logger.Level(...) }`.
        2. Remove the explicit `if logrus.IsLevelEnabled(...)` check, relying on the logger method's internal check.
    - **Done‑when:**
        1. Explicit `IsLevelEnabled` checks preceding logging calls are removed.
        2. Logging behavior remains unchanged (correct level filtering still occurs).
    - **Depends‑on:** [T003]

- [ ] **T006 · Refactor · P2: standardize log messages to structured format**
    - **Context:** cr-06 Standardize Log Message Formatting
    - **Action:**
        1. Identify all logging calls using simple string formatting (e.g., `logrus.Infof("User %s logged in", user)`).
        2. Convert these calls to use structured logging (e.g., `logrus.WithField("user", user).Info("User logged in")`).
        3. Ensure consistent field naming conventions are used.
    - **Done‑when:**
        1. All application log entries use structured formatting (`WithFields`).
        2. Log output is consistent and easily parsable.
    - **Depends‑on:** none

- [ ] **T007 · Refactor · P2: remove correlation id functionality**
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

- [ ] **T008 · Chore · P2: clarify readme logging description**
    - **Context:** cr-11 Clarify README Logging Description
    - **Action:**
        1. Update the "Logging" section in `README.md` to explain the default level (`info`).
        2. Document the `GLANCE_LOG_LEVEL` environment variable, its purpose, and valid values ("debug", "info", "warn", "error").
        3. Add examples showing how to set the log level.
    - **Done‑when:**
        1. `README.md` accurately describes default logging behavior.
        2. `README.md` clearly explains how to configure the log level using `GLANCE_LOG_LEVEL`.
    - **Depends‑on:** [T002]

- [ ] **T009 · Chore · P2: add context for todo.md removal in backlog.md**
    - **Context:** cr-10 Add Context for TODO.md Removal
    - **Action:**
        1. Add a note to `BACKLOG.md` stating it supersedes `TODO.md` for task tracking.
        2. Briefly summarize the major logging refactor tasks (from this plan) that were completed, referencing the original `TODO.md` if applicable/possible.
    - **Done‑when:**
        1. `BACKLOG.md` contains a note about replacing `TODO.md`.
        2. `BACKLOG.md` includes context about the completion of previous logging-related tasks.
    - **Depends‑on:** none
