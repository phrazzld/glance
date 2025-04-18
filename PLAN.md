# remediation plan – sprint 1

## executive summary
This sprint targets critical issues that undermine confidence in our CI/CD pipeline and increase technical debt. We prioritize fixing pre-commit test reliability, eliminating duplicated code paths, and addressing security concerns. The selected order addresses the most impactful issues first (testing correctness and code duplication), followed by security improvements and configuration consistency, delivering quick wins that unlock future refactoring.

## strike list
| seq | cr‑id | title                                        | effort | owner? |
|-----|-------|----------------------------------------------|--------|--------|
| 1   | cr‑09 | Correct pre-commit test results documentation | xs     | dev tools |
| 2   | cr‑10 | Fix pre-commit test script staging          | s      | dev tools |
| 3   | cr‑11 | Remove `-short` flag from test hook         | xs     | dev tools |
| 4   | cr‑12 | Consolidate duplicate BFS scanning logic     | m      | core dev |
| 5   | cr‑13 | Use shared gitignore matching logic          | s      | core dev |
| 6   | cr‑14 | Consolidate MockClient definitions           | xs     | core dev |
| 7   | cr‑03 | Secure golangci-lint installation in CI      | xs     | devops |
| 8   | cr‑04 | Fix bash shell usage on Windows CI          | xs     | devops |
| 9   | cr‑01 | Unify golangci-lint invocation patterns      | s      | devops |

## detailed remedies

### cr‑09 Correct pre-commit test results documentation
- **problem:** Pre-commit hook test results in `hook_test_results.md` are inaccurate and misleading.
- **impact:** Erodes trust in pre-commit checks, masks real issues, and creates confusion about what actually works.
- **chosen fix:** Re-run tests in a properly configured environment and update documentation.
- **steps:**
  1. Ensure local Go version and tooling match CI configuration.
  2. Run `./precommit-tests/run_tests.sh` with proper environment variables.
  3. Replace content in `hook_test_results.md` with current results.
  4. Update `hook_testing_summary.md` to match.
- **done‑when:** Documentation accurately reflects actual hook behavior with no false positives/negatives.

### cr‑10 Fix pre-commit test script staging
- **problem:** Tests for failing conditions (merge conflicts, large files, case conflicts) incorrectly report success.
- **impact:** Critical issues could slip through pre-commit checks, providing false sense of security.
- **chosen fix:** Modify test script to properly stage files before running relevant hooks.
- **steps:**
  1. Edit `precommit-tests/run_tests.sh` to add `git add --force <test-file>` before relevant hook tests.
  2. Add cleanup steps (`git reset HEAD <test-file>`) after each test.
  3. Rerun the test script and verify hooks now fail as expected.
  4. Update documentation with corrected results.
- **done‑when:** Hooks correctly fail on test files designed to trigger failures.

### cr‑11 Remove `-short` flag from test hook
- **problem:** Pre-commit's `go-unit-tests` hook uses `-short` flag, skipping tests that would run in CI.
- **impact:** Tests pass locally but fail in CI, creating frustration and reducing pre-commit effectiveness.
- **chosen fix:** Remove the `-short` flag for consistent behavior between environments.
- **steps:**
  1. Edit `.pre-commit-config.yaml` to remove `-short` flag from `go-unit-tests` hook entry.
  2. Test locally to ensure full test suite runs correctly.
  3. Document this change in `docs/PRECOMMIT.md`.
- **done‑when:** Both local pre-commit and CI run identical test suites.

### cr‑12 Consolidate duplicate BFS scanning logic
- **problem:** Two nearly identical BFS implementations exist in `filesystem/scanner.go` and `glance.go`.
- **impact:** Violates DRY principle, creates maintenance burden, and risks divergent behavior over time.
- **options:**
  - **Option 1: Move all BFS logic to filesystem package**
    * Refactor `filesystem.ListDirsWithIgnores` to handle all cases.
    * Update `glance.go` to use the filesystem package function.
    * Remove duplicated code from `glance.go`.
  - **Option 2: Delegate to one implementation**
    * Keep one version as the source of truth.
    * Modify the other to call it, passing through parameters.
- **standards check:**
  | philosophy | passes? | note |
  |------------|---------|------|
  | simplicity | ✔ | Reduces cognitive load with single implementation |
  | modularity | ✔ | Places directory traversal in filesystem package where it belongs |
  | testability | ✔ | One code path to test thoroughly |
  | coding std | ✔ | Follows Go conventions for package organization |
  | security | ✔ | No new attack vectors introduced |
- **recommendation:** Option 1 - Move all BFS logic to filesystem package
  - Centralizes responsibility for directory traversal where it belongs
  - Cleaner separation of concerns
  - More maintainable in the long run
- **effort:** m (≤ 3 days)

### cr‑13 Use shared gitignore matching logic
- **problem:** `filesystem/reader.go:GatherLocalFiles` reimplements gitignore matching instead of using existing methods.
- **impact:** Duplicate code increases maintenance burden and risks inconsistent gitignore handling.
- **chosen fix:** Refactor to use existing `ShouldIgnoreFile` and `ShouldIgnoreDir` functions.
- **steps:**
  1. Modify the `filepath.WalkDir` callback in `GatherLocalFiles`.
  2. Replace inline ignore matching logic with calls to `ShouldIgnoreFile` and `ShouldIgnoreDir`.
  3. Remove redundant logic while preserving special case handling if needed.
  4. Verify all tests pass after refactoring.
- **done‑when:** Duplicated ignore logic is removed and shared functions are used consistently.

### cr‑14 Consolidate MockClient definitions
- **problem:** Mock implementations of `llm.Client` are duplicated in `integration_test.go` and `llm/client_test.go`.
- **impact:** Violates DRY principle and creates maintenance burden when mock behavior needs changes.
- **chosen fix:** Create a shared, exported mock in the `llm` package for reuse.
- **steps:**
  1. Export the mock from `llm/client_test.go` by renaming it (e.g., to `MockClient` or `TestClient`).
  2. Move it to a dedicated testing file if needed (e.g., `llm/testing.go`).
  3. Update `integration_test.go` to import and use the shared mock.
  4. Verify tests still pass with the shared implementation.
- **done‑when:** Only one mock definition exists and all tests use it successfully.

### cr‑03 Secure golangci-lint installation in CI
- **problem:** CI uses `curl | sh` to install golangci-lint, creating security risk.
- **impact:** Potential vector for supply chain attacks through compromised install scripts.
- **chosen fix:** Use the official GitHub Action or let pre-commit handle installation.
- **steps:**
  1. Remove the `curl | sh` installation in `.github/workflows/precommit.yml`.
  2. Let pre-commit handle tool installation based on `.pre-commit-config.yaml`.
  3. If needed, add GitHub Action `golangci/golangci-lint-action` instead.
  4. Document security considerations in `docs/GITHUB_ACTIONS.md`.
- **done‑when:** No insecure installation methods remain in CI workflows.

### cr‑04 Fix bash shell usage on Windows CI
- **problem:** `build.yml` explicitly specifies `shell: bash`, which may fail on Windows runners.
- **impact:** CI build failures on Windows platforms without Git Bash installed.
- **chosen fix:** Remove the explicit shell override.
- **steps:**
  1. Edit `.github/workflows/build.yml` to remove the `shell: bash` directive.
  2. Allow GitHub Actions to select the appropriate shell per OS.
  3. If needed, add OS-specific commands via conditional logic.
- **done‑when:** CI builds succeed on all platforms consistently.

### cr‑01 Unify golangci-lint invocation patterns
- **problem:** Different methods for invoking golangci-lint between pre-commit and CI workflows.
- **impact:** Potential for inconsistent behavior and version drift between environments.
- **chosen fix:** Standardize on a single approach for both environments.
- **steps:**
  1. Decide whether to standardize on pre-commit or direct invocation.
  2. If using pre-commit in CI, update `.github/workflows/lint.yml` to use `pre-commit run golangci-lint`.
  3. If standardizing on direct invocation, ensure versions match exactly.
  4. Remove any redundant configuration or version specifications.
- **done‑when:** Single source of truth for linting configuration with consistent behavior between environments.

## standards alignment

- These fixes directly address core principles from DEVELOPMENT_PHILOSOPHY.md:
  - **Simplicity First**: Eliminate duplication in BFS and gitignore logic (cr-12, cr-13, cr-14)
  - **Modularity**: Proper responsibility assignment in filesystem package (cr-12, cr-13)
  - **Design for Testability**: Correct pre-commit test behavior and documentation (cr-09, cr-10, cr-11)
  - **Maintainability**: Unified linting approaches and centralized mocks (cr-01, cr-14)
  - **Automation**: Consistent behavior between local and CI environments (cr-01, cr-11)
  - **Security**: Removal of shell-pipe installs in CI (cr-03)

- The remediation prioritizes first fixing test reliability (build trust), then eliminating duplicated code (simplify), and finally addressing security/configuration issues (secure/stabilize).

## validation checklist

- [ ] All unit and integration tests pass locally and in CI
- [ ] Pre-commit hook test suite reports accurate results
- [ ] Hook test failures occur as expected for invalid inputs
- [ ] No duplicate BFS or gitignore code remains
- [ ] Only one MockClient implementation exists
- [ ] CI workflows use secure installation methods
- [ ] Windows CI builds succeed without shell override
- [ ] GolangCI-lint invocation is consistent in all environments
- [ ] No new lint warnings or static analysis issues
- [ ] Documentation accurately reflects the current state
