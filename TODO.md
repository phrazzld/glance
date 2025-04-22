# Todo

## Config
- [x] **T001 · Refactor · P1: remove verbose flag from config package**
    - **Context:** Plan Step 1: Remove Verbose Flag from Config
    - **Action:**
        1. Delete `Verbose bool` field and `WithVerbose` method from `config.Config` (`config/config.go`).
        2. Remove `--verbose` flag definition and application from `config/loadconfig.go`.
        3. Remove related tests for `Verbose` field, `WithVerbose` method, and `--verbose` flag (`config/*_test.go`).
    - **Done‑when:**
        1. `Verbose` field and related logic are removed from the `config` package.
        2. Related unit tests are removed or updated and pass.
        3. `go test ./config/...` passes.
    - **Depends‑on:** none

## Glance (Main)
- [x] **T002 · Refactor · P1: set default log level to debug in setupLogging**
    - **Context:** Plan Step 2: Set Default Log Level to Debug
    - **Action:**
        1. Modify `setupLogging` in `glance.go` to remove the `verbose` parameter.
        2. Unconditionally call `logrus.SetLevel(logrus.DebugLevel)` inside `setupLogging`.
        3. Update all calls to `setupLogging` in `glance.go` to remove the verbose argument.
    - **Done‑when:**
        1. `setupLogging` function signature is updated.
        2. Log level is always set to `logrus.DebugLevel`.
        3. Code compiles and related tests (if any) pass.
    - **Depends‑on:** none

## Filesystem
- [x] **T003 · Refactor · P1: remove verbose parameter and checks from filesystem package**
    - **Context:** Plan Step 3: Update Filesystem Package
    - **Action:**
        1. Remove `verbose bool` parameter from `ShouldIgnoreFile`, `ShouldIgnoreDir`, `LatestModTime`, `ShouldRegenerate`, `GatherLocalFiles`.
        2. Replace `if verbose && logrus.IsLevelEnabled(logrus.DebugLevel)` checks with `if logrus.IsLevelEnabled(logrus.DebugLevel)`.
        3. Remove emojis (e.g., 📊) from log messages within the `filesystem` package.
    - **Done‑when:**
        1. Function signatures in the `filesystem` package are updated.
        2. Conditional verbose logging checks are removed or simplified.
        3. Emojis are removed from logs in this package.
        4. `go test ./filesystem/...` passes.
    - **Depends‑on:** none

## LLM
- [x] **T004 · Refactor · P1: remove verbose parameter/field and checks from llm package**
    - **Context:** Plan Step 4: Update LLM Package
    - **Action:**
        1. Remove `Verbose bool` field and `WithVerbose` function from `llm.ClientOptions` and `llm.ServiceConfig`.
        2. Replace verbose checks (e.g., `if c.options.Verbose`, `if s.verbose`) with `if logrus.IsLevelEnabled(logrus.DebugLevel)`.
        3. Remove emojis (e.g., 🚀, 🔄, ❌, 🔤, ⚠️) from log messages within the `llm` package.
    - **Done‑when:**
        1. Structs and functions in the `llm` package are updated.
        2. Conditional verbose logging checks are removed or simplified.
        3. Emojis are removed from logs in this package.
        4. `go test ./llm/...` passes.
    - **Depends‑on:** none

## UI
- [x] **T005 · Refactor · P1: remove verbose parameter and emoji from ui.ReportError**
    - **Context:** Plan Step 5: Update UI Package
    - **Action:**
        1. Remove the `verbose bool` parameter from `ui.ReportError` (`ui/feedback.go`).
        2. Remove the emoji (e.g., ❌) from the `logrus.Errorf` call within `ReportError`.
    - **Done‑when:**
        1. `ui.ReportError` function signature is updated.
        2. Emoji is removed from the error log message.
        3. `go test ./ui/...` passes.
    - **Depends‑on:** none

## Glance (Main)
- [ ] **T006 · Refactor · P1: update glance main code to remove verbose args and emojis**
    - **Context:** Plan Step 6: Update Main Glance Code
    - **Action:**
        1. Update all calls to `filesystem`, `llm`, and `ui` functions in `glance.go` to remove the `verbose` arguments.
        2. Remove emojis (e.g., ✨, 🚫, 🧠, 🎯, 📊, 🔢, 🌟, ⚠️) from all `logrus` calls in `glance.go`.
        3. Replace conditional debug logging checks (e.g., `if cfg.Verbose { logrus.Debugf(...) }`) with direct `logrus.Debugf(...)` calls.
    - **Done‑when:**
        1. All calls to modified functions in `filesystem`, `llm`, `ui` packages are updated.
        2. Emojis are removed from log messages in `glance.go`.
        3. Direct debug logging is used instead of conditional checks.
        4. Code compiles and relevant tests pass.
    - **Depends‑on:** [T001, T002, T003, T004, T005]

## Logging
- [ ] **T007 · Feature · P2: implement structured logging using logrus fields**
    - **Context:** Plan Step 7: Introduce Structured Logging; Logging & Observability Section
    - **Action:**
        1. Identify key logging points (e.g., directory processing, file reads, LLM calls, errors) in `glance.go`, `filesystem`, `llm`.
        2. Modify relevant `logrus.*` calls to use `logrus.WithField` or `logrus.WithFields` adding context (e.g., `directory`, `file`, `error`).
        3. Implement specific structured log events listed in the plan (e.g., DirectoryScanStarted, FileIgnored, LLMRequestSent).
    - **Done‑when:**
        1. Key log messages include structured fields (e.g., `directory`, `file`, `error`).
        2. Log output is demonstrably more structured and informative.
        3. Code compiles and tests pass.
    - **Depends‑on:** [T006]

## Testing
- [ ] **T008 · Test · P1: update tests to reflect verbose removal and structured logging**
    - **Context:** Plan Step 8: Update Tests; Testing Strategy
    - **Action:**
        1. Update `TestSetupLogging` to assert `logrus.DebugLevel` is always set.
        2. Update unit/integration tests calling `filesystem`, `llm`, `ui` functions to remove the `verbose` parameter.
        3. Review/update/remove tests related to the `--verbose` flag (e.g., `TestGlanceVerboseFlag`).
        4. Update any tests asserting specific log message content to account for removed emojis and added structured fields.
    - **Done‑when:**
        1. All existing unit and integration tests pass after the refactoring.
        2. Tests accurately reflect the new logging behavior (always debug, no emojis, structured fields).
        3. Test coverage is maintained or improved.
    - **Depends‑on:** [T001, T002, T003, T004, T005, T006, T007]

## Documentation
- [ ] **T009 · Chore · P2: update documentation to remove verbose flag references**
    - **Context:** Plan Step 9: Update Documentation; Documentation Section
    - **Action:**
        1. Remove description and usage of the `--verbose` flag from `README.md`.
        2. Update `README.md` and any relevant files in `docs/` to state that logging is detailed (debug level) by default.
        3. Update code comments (functions, structs) in `config`, `filesystem`, `llm`, `ui` to reflect removed `verbose` parameters/fields.
    - **Done‑when:**
        1. `README.md` accurately reflects the removal of the `--verbose` flag and default debug logging.
        2. Code comments are updated for modified functions/structs.
        3. No references to the `--verbose` flag remain in user-facing documentation.
    - **Depends‑on:** [T001]
