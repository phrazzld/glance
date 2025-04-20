# todo

## blocker-issues
- [ ] **T201 · bug · p0: Fix Path Traversal Vulnerability in Prompt Template Loading**
    - **context:** Critical security vulnerability allowing arbitrary file reads via default or custom prompt templates due to missing path validation in `config/loadconfig.go` and `llm/prompt.go`. This is the top priority fix.
    - **action:**
        1. Modify prompt loading logic in `config/loadconfig.go:120-159` to use `filesystem.ValidateFilePath`. Validate default prompt against CWD. Validate custom prompt path after cleaning and making absolute.
        2. Modify potentially duplicated logic in `llm/prompt.go:54-91` similarly, ensuring `filesystem.ValidateFilePath` is used with appropriate `baseDir`.
        3. Update `#nosec G304` comments to accurately reflect the validation performed.
        4. Add specific unit tests demonstrating attempted path traversal (e.g., `../`, `/etc/passwd`) fails, while valid paths succeed.
    - **done‑when:**
        1. Code in `config/loadconfig.go` and `llm/prompt.go` uses `filesystem.ValidateFilePath` for all prompt file loading.
        2. New unit tests demonstrating path traversal prevention pass.
        3. Existing functionality for loading default and custom prompts remains intact.
    - **depends‑on:** []

- [ ] **T202 · docs · p0: Restore or Replace Missing DEVELOPMENT_PHILOSOPHY.md**
    - **context:** A core documentation file, `DEVELOPMENT_PHILOSOPHY.md`, is missing, causing broken references throughout the codebase.
    - **action:**
        1. Search git history for `DEVELOPMENT_PHILOSOPHY.md` and attempt to restore it to the root directory.
        2. If restoration is not possible/desired, create a new file (even if initially a placeholder) to satisfy references.
        3. Use `grep` or search tools to find all references to "DEVELOPMENT_PHILOSOPHY.md" in the codebase and documentation.
        4. Update all found references to point to the correct, valid location of the file.
    - **done‑when:**
        1. A file named `DEVELOPMENT_PHILOSOPHY.md` exists at the location expected by references (likely the project root).
        2. All references to `DEVELOPMENT_PHILOSOPHY.md` in the codebase and documentation point to the existing file.
    - **depends‑on:** []

- [ ] **T203 · docs · p0: Fix References to Missing Standards Documents**
    - **context:** Code and documentation reference other standards documents that are missing or outdated.
    - **action:**
        1. Identify all referenced standards documents (e.g., search for keywords like "standards", common filenames).
        2. For each referenced document, determine if it still exists and locate its correct path or URL.
        3. If a document exists, update all references in the codebase and documentation to point to the correct location.
        4. If a document is confirmed to be obsolete or no longer relevant, remove the references to it.
    - **done‑when:**
        1. All references to standards documents within the project point to valid, existing files/URLs or have been removed.
        2. A review confirms no broken links related to standards documents remain.
    - **depends‑on:** [T202]

## high-severity-issues
- [ ] **T204 · bug · p1: Remove Empty baseDir Escape Hatch in filesystem/reader.go**
    - **context:** The `filesystem/reader.go` functions have a code path allowing an empty `baseDir`, effectively bypassing path validation. This weakens security suppressions relying on this validation.
    - **action:**
        1. Modify `ValidateFilePath` and `ValidateDirPath` in `filesystem/reader.go` (around lines 44, 112) to require a non-empty `baseDir` parameter.
        2. Add explicit error checking: `if baseDir == "" { return "", errors.New("baseDir cannot be empty for validation") }`.
        3. Update all callers to provide a non-empty, contextually correct `baseDir` parameter for validation.
        4. Review and update associated `#nosec` comments where validation is now strictly enforced.
    - **done‑when:**
        1. `filesystem.ValidateFilePath` and `filesystem.ValidateDirPath` no longer permit an empty `baseDir` for validation.
        2. All callers have been updated to provide a valid `baseDir`.
        3. Unit tests confirm the rejection of empty `baseDir`.
    - **depends‑on:** [T201]

- [ ] **T205 · bug · p1: Fix Path Validation in gatherSubGlances**
    - **context:** Path validation in `gatherSubGlances` (`glance.go:308`) needs strengthening using proper `baseDir` parameters to prevent potential traversal.
    - **action:**
        1. Validate the subdirectory path (`sd`) using `filesystem.ValidateDirPath` with an appropriate `baseDir`.
        2. Update the call to `filesystem.ReadTextFile` to use the validated directory path (`validDir`) as its `baseDir` parameter.
        3. Add unit tests involving nested structures and paths designed to test containment logic.
    - **done‑when:**
        1. `gatherSubGlances` uses `ValidateDirPath` for subdirectories and passes the validated path as `baseDir` to `ReadTextFile`.
        2. New unit tests pass, verifying path containment within the expected base directories.
    - **depends‑on:** [T204]

- [ ] **T206 · bug · p1: Fix Path Validation in readSubdirectories**
    - **context:** Path validation in `readSubdirectories` (`glance.go:339`) needs strengthening to ensure subdirectories cannot escape the parent directory.
    - **action:**
        1. Validate the constructed `fullPath` using `filesystem.ValidateDirPath`, ensuring the `baseDir` parameter is the parent directory (`validDir`).
        2. Add unit tests for nested directories and paths designed to test containment logic.
    - **done‑when:**
        1. `readSubdirectories` uses `ValidateDirPath` with the parent directory as `baseDir` to validate subdirectory entries.
        2. New unit tests pass, verifying path containment.
    - **depends‑on:** [T204]

- [ ] **T207 · refactor · p1: Centralize Prompt Template Loading Logic**
    - **context:** Duplicate logic for loading prompt templates exists in `config/loadconfig.go`, `llm/prompt.go`, and `llm/service.go`. This should be consolidated into a single, secure implementation.
    - **action:**
        1. Designate the `config` package as the sole owner of prompt template loading logic.
        2. Refactor the loading logic into a function within the `config` package, incorporating secure path validation.
        3. Remove the duplicate prompt loading code from `llm/prompt.go` and `llm/service.go`.
        4. Update the LLM service initialization to accept the loaded template via configuration options.
        5. Remove any duplicate security suppression justifications related to prompt loading.
    - **done‑when:**
        1. Prompt template loading logic resides only within the `config` package.
        2. The `llm` package receives prompt information via configuration and does not perform file loading itself.
        3. Path validation is applied correctly in the centralized logic.
    - **depends‑on:** [T201]

- [x] **T208 · chore · p1: Align golangci-lint Version in Scripts and Configs**
    - **context:** The `scripts/setup-precommit.sh` script installs a version of `golangci-lint` that differs from the version used in CI, leading to inconsistent linting results.
    - **action:**
        1. Identify the canonical `golangci-lint` version used by the project (check `.golangci.yml`, `pre-commit-config.yaml`, CI configuration).
        2. Update the installation command in `scripts/setup-precommit.sh:40` to use this exact version.
        3. Consider if the direct installation in the script is necessary or if `pre-commit`'s environment management is sufficient.
    - **done‑when:**
        1. The `golangci-lint` version in `setup-precommit.sh` matches the version in project configurations.
        2. Running linters locally yields results consistent with the CI pipeline.
    - **depends‑on:** []

- [x] **T209 · refactor · p1: Standardize gitignore Handling with IgnoreChain**
    - **context:** Gitignore pattern handling is inconsistent in `glance.go`, with potential direct use of raw `gitignore` types instead of the `filesystem.IgnoreChain` abstraction.
    - **action:**
        1. Review all functions in `glance.go` that handle file ignoring logic.
        2. Ensure these functions exclusively use the `filesystem.IgnoreChain` type.
        3. Remove any direct usage of lower-level `gitignore.GitIgnore` types.
        4. Verify that the `IgnoreChain` instance is correctly passed through the call stack.
    - **done‑when:**
        1. File ignoring logic in `glance.go` consistently uses `filesystem.IgnoreChain`.
        2. Direct usage of raw `gitignore` types is removed from high-level application logic.
        3. Ignore patterns function as expected across different parts of the application.
    - **depends‑on:** []

- [ ] **T210 · bug · p1: Update Security Suppression for Prompt Loading**
    - **context:** Security suppression comments (`#nosec`) in `llm/prompt.go:101` and elsewhere need updating after path traversal fixes.
    - **action:**
        1. Locate all `#nosec G304` suppression comments related to prompt file reading.
        2. Ensure each comment accurately reflects the validation performed (i.e., "Path validated by filesystem.ValidateFilePath").
        3. If code was centralized, ensure no outdated suppressions remain.
    - **done‑when:**
        1. All `#nosec G304` comments associated with file reads accurately describe the validation performed.
        2. No incorrect or obsolete suppression comments related to prompt loading exist.
    - **depends‑on:** [T201, T207]

- [ ] **T211 · refactor · p1: Simplify Overly Complex Builder Patterns**
    - **context:** Builder patterns in `llm/client.go` and `llm/service.go` may be overly complex for the configuration required.
    - **action:**
        1. Replace builder patterns with direct struct initialization where appropriate.
        2. Keep functional options where they provide significant benefit.
        3. Remove unnecessary `With*` methods that become redundant.
        4. Verify through tests that all configuration pathways still function correctly.
    - **done‑when:**
        1. Initialization code for LLM clients/services is simplified.
        2. Unnecessary builder methods are removed.
        3. All existing configuration options still work correctly.
        4. Tests covering client/service initialization pass.
    - **depends‑on:** []

## medium-severity-issues
- [ ] **T212 · refactor · p2: Create Shared Test Utilities for Mock Client**
    - **context:** A mock LLM client implementation is duplicated across different test files.
    - **action:**
        1. Create a new internal package dedicated to shared testing utilities (e.g., `internal/testutil`).
        2. Move a canonical version of the `MockClient` into this package.
        3. Update all test files to import and use the shared version.
    - **done‑when:**
        1. A shared `internal/testutil` package exists with a single `MockClient` implementation.
        2. All relevant tests use the shared mock client.
        3. Tests continue to pass.
    - **depends‑on:** []

- [ ] **T213 · refactor · p2: Standardize Mocking Approach**
    - **context:** Mocking approaches are inconsistent, using a mix of interface injection and function variables.
    - **action:**
        1. Review code for different mocking techniques currently employed.
        2. Establish interface injection as the preferred standard mocking approach.
        3. Refactor areas using function variables to use interfaces and injected mocks.
        4. Update tests to use the standardized approach.
    - **done‑when:**
        1. Mocking primarily uses interface injection.
        2. Use of function variables for mocking is minimized or eliminated where interfaces are suitable.
        3. Tests pass using the standardized mocking approach.
    - **depends‑on:** [T212]

- [ ] **T214 · bug · p2: Include Error Context in Non-Verbose ReportError**
    - **context:** The `ui.ReportError` function hides useful error information when not in verbose mode.
    - **action:**
        1. Modify the `ReportError` function in `ui/feedback.go` to include the error message even in non-verbose mode.
        2. Implement a format similar to: `logrus.Errorf("❌ %s: %v", context, err)`.
        3. Keep stack trace output exclusive to verbose mode.
    - **done‑when:**
        1. Calling `ui.ReportError` with `verbose=false` logs both the context string and the error message.
        2. Calling `ui.ReportError` with `verbose=true` still includes the stack trace.
    - **depends‑on:** []

- [ ] **T215 · refactor · p2: Consolidate Regeneration Check Logic**
    - **context:** Redundant logic to determine if a glance file needs regeneration exists in multiple places.
    - **action:**
        1. Identify all places where regeneration checks occur.
        2. Ensure that `filesystem.ShouldRegenerate` is the single source of truth.
        3. Remove any duplicate logic in `processDirectory` and elsewhere.
    - **done‑when:**
        1. `filesystem.ShouldRegenerate` is the only function determining if regeneration is needed.
        2. Redundant checks in `processDirectory` are removed.
    - **depends‑on:** []

- [ ] **T216 · refactor · p2: Improve Path Handling in directoryChecker**
    - **context:** The current implementation forces absolute path conversion when relative paths would suffice.
    - **action:**
        1. Improve the `directoryChecker` interface to handle both relative and absolute paths correctly.
        2. Remove redundant path conversions in callers where possible.
    - **done‑when:**
        1. Path handling is simplified with fewer unnecessary conversions.
        2. Tests verify correct behavior with both relative and absolute paths.
    - **depends‑on:** []

- [x] **T217 · security · p2: Remove Insecure Allowlist Pragma**
    - **context:** An "allowlist" pragma in parameter documentation may suppress legitimate security warnings.
    - **action:**
        1. Remove the pragma from parameter documentation in `llm/client.go:84`.
        2. If needed, move the pragma to specific variable assignments where suppression is actually required.
    - **done‑when:**
        1. No allowlist pragmas remain in parameter documentation.
        2. Any required suppressions are moved to specific assignment operations.
    - **depends‑on:** []

- [ ] **T218 · chore · p2: Rename CI Workflow Schedules**
    - **context:** CI workflow schedule descriptions are misleading and do not reflect their actual purpose.
    - **action:**
        1. Update workflow schedule names in `.github/workflows/lint.yml` and `.github/workflows/precommit.yml`.
        2. Ensure the descriptions accurately reflect what the workflows do.
    - **done‑when:**
        1. Workflow schedule names match their actual functionality.
        2. No misleading references to "dependency checks" or "security scans" remain unless those are actually performed.
    - **depends‑on:** []

- [ ] **T219 · chore · p2: Standardize Pre-commit Hook Language Configs**
    - **context:** Pre-commit hooks use inconsistent language configuration (some `system`, some `golang`).
    - **action:**
        1. Update `.pre-commit-config.yaml` to use `language: golang` for all Go-related hooks.
        2. Ensure version fields are consistent across hooks.
    - **done‑when:**
        1. All Go hooks use `language: golang` consistently.
        2. Pre-commit checks pass without errors.
    - **depends‑on:** []

- [ ] **T220 · docs · p2: Clean Up .gitignore Entries**
    - **context:** `.gitignore` contains entries for files or directories that no longer exist.
    - **action:**
        1. Review `.gitignore` to identify obsolete entries.
        2. Remove entries for deleted files or directories.
        3. Document any special patterns that should be retained.
    - **done‑when:**
        1. `.gitignore` only contains patterns for files/directories that exist or may be created.
        2. No references to deleted documentation or symlinks remain.
    - **depends‑on:** [T202, T203]

- [ ] **T221 · docs · p2: Document File Permission Settings**
    - **context:** The README lacks documentation about the secure file permissions used by the application.
    - **action:**
        1. Add a section to `README.md` explaining the use of `filesystem.DefaultFileMode`.
        2. Update `SECURITY_SUPPRESSIONS.md` to document the rationale for 0600 permissions.
    - **done‑when:**
        1. Documentation clearly explains the file permission model and its security implications.
    - **depends‑on:** []

- [ ] **T222 · docs · p2: Document LLM Model Name Change**
    - **context:** The default LLM model name changed without user-facing documentation, affecting outputs and costs.
    - **action:**
        1. Add a section to `README.md` noting the model name change from the old version to the new one.
        2. Document any potential impacts on output quality, performance, or costs.
    - **done‑when:**
        1. Documentation clearly explains the model change and its implications.
    - **depends‑on:** []

## low-severity-issues
- [x] **T223 · chore · p3: Use filesystem.DefaultFileMode in glance.go**
    - **context:** The `glance.go` file still uses a magic number (`0o600`) instead of the `filesystem.DefaultFileMode` constant.
    - **action:**
        1. Replace the hardcoded `0o600` in `glance.go:252` with `filesystem.DefaultFileMode`.
        2. Ensure the code compiles and behaves identically.
    - **done‑when:**
        1. No direct `0o600` literals remain in file operations.
        2. The application uses `filesystem.DefaultFileMode` consistently.
    - **depends‑on:** []

- [ ] **T224 · refactor · p3: Remove Unnecessary ShouldIgnorePath Function**
    - **context:** The `ShouldIgnorePath` function adds unnecessary indirection and complexity.
    - **action:**
        1. Identify usage of `ShouldIgnorePath` in the codebase.
        2. Replace calls with direct use of `ShouldIgnoreDir` or `ShouldIgnoreFile` as appropriate.
        3. Remove the redundant wrapper function.
    - **done‑when:**
        1. `ShouldIgnorePath` is removed from the codebase.
        2. All callers have been updated to use the appropriate direct function.
    - **depends‑on:** []

- [ ] **T225 · test · p3: Standardize Test Model Names and File Permissions**
    - **context:** Tests use outdated model names and inconsistent file permissions.
    - **action:**
        1. Update all mock models to use the current model name.
        2. Replace hardcoded permission constants with `filesystem.DefaultFileMode` where appropriate.
    - **done‑when:**
        1. All tests use the current model name and consistent permissions.
        2. Tests pass without modification to behavior.
    - **depends‑on:** []

- [ ] **T226 · test · p3: Fix Potential Race Condition in logCapture**
    - **context:** The `logCapture` test utility may not be thread-safe if tests run in parallel.
    - **action:**
        1. Update `logCapture` in test files to use a thread-safe buffer.
        2. Or add appropriate test flags to prevent parallel execution.
    - **done‑when:**
        1. Log capturing in tests works reliably without race conditions.
    - **depends‑on:** []

- [ ] **T227 · test · p3: Add t.Helper() to Test Helper Functions**
    - **context:** Test helper functions are missing `t.Helper()` calls, making error reporting less precise.
    - **action:**
        1. Identify all test helper functions that take a `t *testing.T` parameter.
        2. Add `t.Helper()` as the first line in each helper function.
    - **done‑when:**
        1. All test helper functions include `t.Helper()`.
        2. Test failures report the correct line number in the test file.
    - **depends‑on:** []

- [ ] **T228 · chore · p3: Remove Redundant goimports Installation in CI**
    - **context:** The CI workflow manually installs `goimports` when pre-commit would manage it.
    - **action:**
        1. Remove redundant `go install` commands for tools managed by pre-commit.
        2. Verify CI workflows still run correctly after removal.
    - **done‑when:**
        1. CI workflow is simplified without redundant tool installation.
        2. Pre-commit hooks continue to work in CI.
    - **depends‑on:** []

- [ ] **T229 · refactor · p3: Simplify OS-Specific Build Logic**
    - **context:** The build workflow uses complex OS-specific conditionals that could be simplified.
    - **action:**
        1. Refactor `.github/workflows/build.yml` to use standard GOOS-based logic.
        2. Remove redundant verification steps where possible.
    - **done‑when:**
        1. Build workflow is simplified while maintaining functionality.
        2. Builds succeed for all target platforms.
    - **depends‑on:** []

- [ ] **T230 · refactor · p3: Remove Dead Code**
    - **context:** Unused code (e.g., `queueItem` struct) and redundant comments remain in the codebase.
    - **action:**
        1. Remove the `queueItem` struct in `glance.go`.
        2. Clean up redundant or outdated comments.
    - **done‑when:**
        1. No unused structs or functions remain in the codebase.
        2. Comments are accurate and up-to-date.
    - **depends‑on:** []

- [ ] **T231 · refactor · p3: Centralize Logging Logic**
    - **context:** Logging code is duplicated across multiple files, leading to inconsistent formatting.
    - **action:**
        1. Create shared logging utility functions if needed.
        2. Standardize log level usage and formatting across the codebase.
    - **done‑when:**
        1. Logging is consistent and non-duplicative across the codebase.
    - **depends‑on:** []

- [ ] **T232 · chore · p3: Refine Pre-commit Ignore Patterns**
    - **context:** Pre-commit ignore patterns may be too broad, potentially skipping important files.
    - **action:**
        1. Review `.pre-commit-config.yaml` exclude patterns.
        2. Narrow exclude patterns to only necessary files/directories.
    - **done‑when:**
        1. Pre-commit hooks run on all relevant files without unnecessary exclusions.
    - **depends‑on:** []

- [ ] **T233 · test · p3: Add Test Coverage Enforcement to CI**
    - **context:** The CI pipeline doesn't enforce minimum test coverage requirements.
    - **action:**
        1. Add a coverage gate to `.github/workflows/test.yml` that fails if coverage drops below a threshold.
        2. Set appropriate thresholds for each package or component.
    - **done‑when:**
        1. CI fails if test coverage falls below the defined threshold.
        2. Current code passes the coverage requirements.
    - **depends‑on:** []
