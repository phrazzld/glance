# TODO

## Pre-commit & CI Tests
- [x] **T001:** Update pre-commit test results documentation
    - **Action:** Run `./precommit-tests/run_tests.sh` with proper environment and update `hook_test_results.md` and `hook_testing_summary.md` with current, accurate results.
    - **Depends On:** None
    - **Type:** Chore
    - **Priority:** P1

- [x] **T002:** Fix staging in pre-commit test script
    - **Action:** Add `git add --force <test-file>` and cleanup steps in `run_tests.sh` to make test hooks fail properly for merge conflicts, large files, and case conflicts.
    - **Depends On:** None
    - **Type:** Bugfix
    - **Priority:** P1

- [x] **T003:** Remove `-short` flag from go-unit-tests pre-commit hook
    - **Action:** Edit `.pre-commit-config.yaml` to remove `-short` flag from the `go-unit-tests` hook for consistency with CI.
    - **Depends On:** None
    - **Type:** Bugfix
    - **Priority:** P1

- [x] **T004:** Document removal of `-short` flag in PRECOMMIT.md
    - **Action:** Update docs to explain the change and rationale for consistent local/CI testing.
    - **Depends On:** [T003]
    - **Type:** Chore
    - **Priority:** P2

## Filesystem Refactoring
- [x] **T005:** Refactor filesystem.ListDirsWithIgnores for consolidation
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

- [x] **T013:** Remove explicit shell: bash from build.yml workflow
    - **Action:** Remove the `shell: bash` directive from `.github/workflows/build.yml` to fix Windows CI builds.
    - **Depends On:** None
    - **Type:** Bugfix
    - **Priority:** P1

- [x] **T014:** Decide and document standardized golangci-lint invocation method
    - **Action:** Choose and document whether to standardize on pre-commit or direct invocation for golangci-lint.
    - **Depends On:** None
    - **Type:** Chore
    - **Priority:** P2

- [x] **T015:** Update lint.yml workflow based on standardized invocation
    - **Action:** Modify `.github/workflows/lint.yml` to use the chosen standardized invocation method.
    - **Depends On:** [T014]
    - **Type:** Chore
    - **Priority:** P2

- [ ] **T016:** Remove redundant golangci-lint configuration/versions
    - **Action:** Eliminate duplicated version specifications or configuration for golangci-lint across workflow files. (Restructured into T017-T022)
    - **Depends On:** [T014, T015]
    - **Type:** Chore
    - **Priority:** P2

- [x] **T017:** Update `.golangci.yml` Configuration Format
    - **Action:** Modify `.golangci.yml` to use the modern structure with top-level `run:`, `linters:`, and `issues:` sections. Set explicit `version: "2"` for modern format compatibility. Configure the initial set of enabled linters and verify `run.timeout` is set to `2m`.
    - **Depends On:** []
    - **Type:** Bugfix
    - **Priority:** P0

- [x] **T018:** Update `.pre-commit-config.yaml` for golangci-lint
    - **Action:** Modify the golangci-lint hook in `.pre-commit-config.yaml` to set `rev:` to `v1.57.0` as the single source of truth for the version and change `language:` from `system` to `golang` for better reproducibility.
    - **Depends On:** [T017]
    - **Type:** Bugfix
    - **Priority:** P0

- [x] **T019:** Align CI Workflow golangci-lint Version
    - **Action:** Update `.github/workflows/lint.yml` to use the same golangci-lint version specified in `.pre-commit-config.yaml` (v1.57.0) and ensure it uses the correct configuration file with matching timeout settings.
    - **Depends On:** [T018]
    - **Type:** Bugfix
    - **Priority:** P0

- [ ] **T020:** Update Linting Documentation
    - **Action:** Update `docs/LINTING.md` to explain that modern golangci-lint doesn't require a `version:` field, and update `docs/PRECOMMIT.md` to emphasize that pre-commit manages golangci-lint installation when using `language: golang`.
    - **Depends On:** [T019]
    - **Type:** Documentation
    - **Priority:** P1

- [ ] **T021:** Verify Pre-commit and CI Linting Pass
    - **Action:** Run `pre-commit run golangci-lint --all-files` locally and verify that the golangci-lint hooks pass with the updated configuration.
    - **Depends On:** [T020]
    - **Type:** Chore
    - **Priority:** P0

- [ ] **T022:** Complete T016
    - **Action:** Verify all redundant golangci-lint configurations/versions are removed, and mark T016 as completed.
    - **Depends On:** [T021]
    - **Type:** Chore
    - **Priority:** P2

### Clarifications & Assumptions
- [x] **Issue:** Decision needed on standard golangci-lint invocation method
    - **Context:** Task T014 requires deciding whether to standardize on pre-commit hooks or direct golangci-lint invocation across all environments
    - **Blocking?:** Yes (for T015, T016)
    - **Resolution:** Standardized approach documented in docs/LINTING.md - using pre-commit hooks for local development and official GitHub Action for CI
