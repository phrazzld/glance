# TODO

## Pre-commit & CI Tests
- [ ] **T001:** Update pre-commit test results documentation
    - **Action:** Run `./precommit-tests/run_tests.sh` with proper environment and update `hook_test_results.md` and `hook_testing_summary.md` with current, accurate results.
    - **Depends On:** None
    - **Type:** Chore
    - **Priority:** P1

- [ ] **T002:** Fix staging in pre-commit test script
    - **Action:** Add `git add --force <test-file>` and cleanup steps in `run_tests.sh` to make test hooks fail properly for merge conflicts, large files, and case conflicts.
    - **Depends On:** None
    - **Type:** Bugfix
    - **Priority:** P1

- [ ] **T003:** Remove `-short` flag from go-unit-tests pre-commit hook
    - **Action:** Edit `.pre-commit-config.yaml` to remove `-short` flag from the `go-unit-tests` hook for consistency with CI.
    - **Depends On:** None
    - **Type:** Bugfix
    - **Priority:** P1

- [ ] **T004:** Document removal of `-short` flag in PRECOMMIT.md
    - **Action:** Update docs to explain the change and rationale for consistent local/CI testing.
    - **Depends On:** [T003]
    - **Type:** Chore
    - **Priority:** P2

## Filesystem Refactoring
- [ ] **T005:** Refactor filesystem.ListDirsWithIgnores for consolidation
    - **Action:** Enhance `filesystem.ListDirsWithIgnores` to handle all BFS use cases currently covered by both implementations.
    - **Depends On:** None
    - **Type:** Refactor
    - **Priority:** P2

- [ ] **T006:** Update glance.go to use consolidated BFS logic
    - **Action:** Modify `glance.go` to call the refactored `filesystem.ListDirsWithIgnores` instead of its own implementation.
    - **Depends On:** [T005]
    - **Type:** Refactor
    - **Priority:** P2

- [ ] **T007:** Remove duplicate BFS code from glance.go
    - **Action:** Delete the redundant `listAllDirsWithIgnores` function from `glance.go` after migration is complete.
    - **Depends On:** [T006]
    - **Type:** Refactor
    - **Priority:** P2

- [ ] **T008:** Refactor GatherLocalFiles to use shared ignore functions
    - **Action:** Replace inline gitignore matching in `filesystem/reader.go:GatherLocalFiles` with calls to existing `ShouldIgnoreFile` and `ShouldIgnoreDir` functions.
    - **Depends On:** None
    - **Type:** Refactor
    - **Priority:** P2

## Testing & Mocks
- [ ] **T009:** Export shared llm.MockClient from llm package
    - **Action:** Rename and export the mock struct in `llm/client_test.go` to a dedicated `llm/testing.go` file.
    - **Depends On:** None
    - **Type:** Refactor
    - **Priority:** P2

- [ ] **T010:** Update integration_test.go to use shared llm.MockClient
    - **Action:** Remove local mock definition from `integration_test.go` and use the shared mock from the llm package.
    - **Depends On:** [T009]
    - **Type:** Refactor
    - **Priority:** P2

## CI/CD Configuration
- [x] **T011:** Remove insecure curl|sh install from precommit.yml
    - **Action:** Remove `curl | sh` golangci-lint installation in `.github/workflows/precommit.yml` and rely on pre-commit's framework for installation.
    - **Depends On:** None
    - **Type:** Security
    - **Priority:** P0

- [x] **T012:** Document secure golangci-lint installation in GITHUB_ACTIONS.md
    - **Action:** Update documentation to explain the secure installation method for golangci-lint.
    - **Depends On:** [T011]
    - **Type:** Chore
    - **Priority:** P2

- [ ] **T013:** Remove explicit shell: bash from build.yml workflow
    - **Action:** Remove the `shell: bash` directive from `.github/workflows/build.yml` to fix Windows CI builds.
    - **Depends On:** None
    - **Type:** Bugfix
    - **Priority:** P1

- [ ] **T014:** Decide and document standardized golangci-lint invocation method
    - **Action:** Choose and document whether to standardize on pre-commit or direct invocation for golangci-lint.
    - **Depends On:** None
    - **Type:** Chore
    - **Priority:** P2

- [ ] **T015:** Update lint.yml workflow based on standardized invocation
    - **Action:** Modify `.github/workflows/lint.yml` to use the chosen standardized invocation method.
    - **Depends On:** [T014]
    - **Type:** Chore
    - **Priority:** P2

- [ ] **T016:** Remove redundant golangci-lint configuration/versions
    - **Action:** Eliminate duplicated version specifications or configuration for golangci-lint across workflow files.
    - **Depends On:** [T014, T015]
    - **Type:** Chore
    - **Priority:** P2

### Clarifications & Assumptions
- [ ] **Issue:** Decision needed on standard golangci-lint invocation method
    - **Context:** Task T014 requires deciding whether to standardize on pre-commit hooks or direct golangci-lint invocation across all environments
    - **Blocking?:** Yes (for T015, T016)
