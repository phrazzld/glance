# Todo

## Core Regeneration Logic
- [x] **T002 · Refactor · P1: refactor `processDirectories` for testability**
    - **Context:** cr-02 Refactor Tests to Use Production Code (Step 1)
    - **Action:**
        1. Modify `processDirectories` signature to accept dependencies (like logging, filesystem access) as arguments (interfaces).
        2. Modify `processDirectories` to return the map of directories needing regeneration (`needsRegen`).
    - **Done‑when:**
        1. `processDirectories` function signature is updated for dependency injection.
        2. `processDirectories` returns the `needsRegen` map.
        3. Existing callers are updated to use the new signature (or temporarily adapted).
        4. Unit tests for `processDirectories` (if any) pass.
    - **Depends‑on:** none

- [x] **T005 · Bugfix · P2: remove flawed `glance_file_missing` check in logging**
    - **Context:** cr-04 Fix "glance_file_missing" Logic
    - **Action:**
        1. Remove the condition checking `filepath.Base(dir) == "glance.md"` within the regeneration logic.
        2. Ensure logging uses the actual regeneration reason derived from `ShouldRegenerate` or related boolean checks.
    - **Done‑when:**
        1. Incorrect log messages mentioning `glance_file_missing` for directories are no longer generated.
        2. Log messages accurately reflect the determined reason for regeneration.
    - **Depends‑on:** none

## Testing
- [x] **T003 · Test · P1: update integration tests to use production `processDirectories`**
    - **Context:** cr-02 Refactor Tests to Use Production Code (Steps 2-3)
    - **Action:**
        1. Modify integration tests to call the refactored production `processDirectories` function, providing necessary test doubles for dependencies.
        2. Remove the custom `ProcessDirectoriesWithTracking` test helper function.
        3. Update test assertions to validate behavior based on production code execution.
    - **Done‑when:**
        1. Integration tests directly invoke the production `processDirectories`.
        2. `ProcessDirectoriesWithTracking` helper function is deleted.
        3. All relevant integration tests pass using the production code path.
    - **Depends‑on:** [T002]

- [x] **T004 · Test · P2: refactor forced regeneration test to use actual force mechanism**
    - **Context:** cr-03 Fix Forced Regeneration Test
    - **Action:**
        1. Modify the forced regeneration test to utilize the application's actual force flag/mechanism when calling the refactored `processDirectories`.
        2. Remove manual simulation of forced regeneration behavior.
        3. Verify that the test correctly asserts parent directory regeneration based on the real force logic.
    - **Done‑when:**
        1. Test triggers regeneration via the application's intended force mechanism.
        2. Test successfully verifies the expected outcome of forced regeneration.
    - **Depends‑on:** [T002]

- [x] **T006 · Bugfix · P2: remove inverted condition blocking windows path test**
    - **Context:** cr-05 Fix Windows Path Test Logic
    - **Action:**
        1. Locate the test responsible for Windows path logic.
        2. Remove the `if filepath.Separator != '\\'` condition (or its equivalent) that prevents the test from running on Windows.
    - **Done‑when:**
        1. The conditional logic preventing the test execution on Windows is removed.
        2. The test is able to run (and ideally pass) on a Windows environment.
        3. The test still passes on non-Windows environments.
    - **Depends‑on:** none

- [x] **T008 · Chore · P3: remove or make conditional excessive `t.Logf` calls in integration tests**
    - **Context:** cr-07 Remove Excessive Test Logging
    - **Action:**
        1. Review integration tests for `t.Logf` calls used for debugging rather than essential test output.
        2. Remove unnecessary `t.Logf` statements.
        3. Optionally, wrap potentially useful debug logs with `if testing.Verbose()`.
    - **Done‑when:**
        1. Test output (`go test ./...`) is clean and concise when run without the `-v` flag.
        2. Essential test failure information remains clear.
    - **Depends‑on:** none

## Logging
- [x] **T007 · Refactor · P2: standardize structured log field names**
    - **Context:** cr-06 Standardize Log Field Names
    - **Action:**
        1. Define standard names (e.g., `error` vs `err`, `path` vs `file_path`).
        2. Update all structured logging calls across the codebase to use the defined standard field names.
    - **Done‑when:**
        1. All structured log output uses consistent field names (e.g., `error`, `path`).
        2. Code compiles and tests pass.
    - **Depends‑on:** none
