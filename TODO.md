# todo

## ci/cd workflows
- [x] **T031 · chore · p0: replace insecure curl|sh install with github action**
    - **context:** CR-01: Secure CI: Replace `curl | sh` with GitHub Action
    - **action:**
        1. Remove `curl | sh` golangci-lint install step from `.github/workflows/lint.yml`
        2. Add `uses: golangci/golangci-lint-action@v4` step with appropriate arguments
    - **done‑when:**
        1. No `curl | sh` present in workflow
        2. Lint job passes with GitHub Action
    - **depends‑on:** none

- [x] **T032 · chore · p1: determine and standardize golangci-lint version**
    - **context:** CR-02: Standardize golangci-lint version & configuration
    - **action:**
        1. Decide on target golangci-lint version (e.g., v1.57.0)
        2. Update `rev:` in `.pre-commit-config.yaml` to this version
    - **done‑when:**
        1. Decision on standard version is made
        2. Pre-commit config is updated
    - **depends‑on:** none

- [x] **T033 · chore · p1: configure golangci-lint-action version in CI workflow**
    - **context:** CR-02: Standardize golangci-lint version & configuration
    - **action:**
        1. Configure GitHub Action with version from T032
        2. Ensure args point to `.golangci.yml`
    - **done‑when:**
        1. CI workflow uses same version as pre-commit config
        2. CI workflow uses standard `.golangci.yml` config
    - **depends‑on:** [T031, T032]

- [x] **T034 · chore · p1: update docs for standard golangci-lint version**
    - **context:** CR-02: Standardize golangci-lint version & configuration
    - **action:**
        1. Update docs (`README.md`, `docs/LINTING.md`, `docs/GITHUB_ACTIONS.md`) to reference standard version
    - **done‑when:**
        1. All documentation consistently references the same version
    - **depends‑on:** [T032]

- [x] **T035 · chore · p1: re-enable pre-commit golangci-lint hook**
    - **context:** CR-07: Re-enable pre-commit linter
    - **action:**
        1. Edit `.pre-commit-config.yaml` to uncomment the golangci-lint hook
        2. Ensure args point to `.golangci.yml`
        3. Test locally with `pre-commit run --all-files`
    - **done‑when:**
        1. Hook is active and runs during pre-commit
        2. Hook runs using correct config
    - **depends‑on:** [T032]

## security
- [x] **T036 · chore · p1: audit all #nosec annotations**
    - **context:** CR-03: Address excessive `#nosec` suppressions
    - **action:**
        1. List all `#nosec` usages in codebase (G304/G306/etc)
        2. For each, document justification or mark for removal
    - **done‑when:**
        1. Complete audit with status of each suppression
    - **depends‑on:** none

- [x] **T037 · feature · p1: identify locations requiring path validation**
    - **context:** CR-08: Implement file path validation
    - **action:**
        1. Identify all code locations accepting file paths from external sources (CLI, config, env)
        2. Document each location with type of validation needed
    - **done‑when:**
        1. Complete list of code locations requiring validation
    - **depends‑on:** none

- [x] **T038 · feature · p1: implement path validation for external inputs**
    - **context:** CR-08: Implement file path validation
    - **action:**
        1. Add validation: `filepath.Clean`, `filepath.Abs`, and prefix-checking
        2. Restrict paths to within allowed base directories
    - **done‑when:**
        1. All external paths are validated before use
        2. Implementation passes basic tests
    - **depends‑on:** [T037]

- [x] **T039 · test · p1: add path traversal security tests**
    - **context:** CR-08: Implement file path validation
    - **action:**
        1. Add tests for traversal attempts (e.g., `../etc/passwd`, absolute paths)
        2. Test symlink handling and other edge cases
    - **done‑when:**
        1. Tests confirm protection against traversal and path manipulation
    - **depends‑on:** [T038]

- [x] **T040 · refactor · p2: remove unnecessary #nosec suppressions**
    - **context:** CR-03: Address excessive `#nosec` suppressions
    - **action:**
        1. Remove G304 suppressions where path validation is now implemented
        2. Remove G306 suppressions where permissions are now documented
    - **done‑when:**
        1. Unnecessary suppressions removed, code passes static analysis
    - **depends‑on:** [T036, T038, T044]

- [ ] **T041 · chore · p2: create SECURITY_SUPPRESSIONS.md document**
    - **context:** CR-03: Address excessive `#nosec` suppressions
    - **action:**
        1. Create document explaining security decisions
        2. Include justification for any remaining suppressions
    - **done‑when:**
        1. Document exists with clear rationale for necessary suppressions
    - **depends‑on:** [T040]

## filesystem
- [x] **T042 · refactor · p1: identify duplicate filesystem functions**
    - **context:** CR-04: Eliminate duplicate filesystem logic
    - **action:**
        1. List all functions in `glance.go` duplicated in `filesystem/`
        2. Map each to its canonical implementation
    - **done‑when:**
        1. Complete mapping document of duplicated functions
    - **depends‑on:** none

- [x] **T043 · refactor · p1: replace duplicated logic with filesystem package calls**
    - **context:** CR-04: Eliminate duplicate filesystem logic
    - **action:**
        1. Update `glance.go` to use `filesystem` package functions
        2. Ensure correct parameter passing
    - **done‑when:**
        1. `glance.go` calls the `filesystem` package for filesystem operations
    - **depends‑on:** [T042]

- [x] **T044 · refactor · p1: remove old filesystem code from glance.go**
    - **context:** CR-04: Eliminate duplicate filesystem logic
    - **action:**
        1. Delete obsolete functions from `glance.go`
    - **done‑when:**
        1. No duplicate implementations remain in `glance.go`
    - **depends‑on:** [T043]

- [ ] **T045 · test · p2: update tests after filesystem refactoring**
    - **context:** CR-04: Eliminate duplicate filesystem logic
    - **action:**
        1. Update tests to target canonical implementations
        2. Verify all functionality is preserved
    - **done‑when:**
        1. All tests pass with refactored code
    - **depends‑on:** [T044]

- [ ] **T046 · refactor · p2: migrate to IgnoreChain abstraction**
    - **context:** CR-05: Complete filesystem abstraction refactoring
    - **action:**
        1. Update functions to use `filesystem.IgnoreChain` instead of old types
        2. Remove compatibility helpers
    - **done‑when:**
        1. `filesystem.IgnoreChain` is used consistently
        2. `ExtractGitignoreMatchers` and `CreateIgnoreChain` are removed
    - **depends‑on:** [T045]

## llm
- [ ] **T047 · feature · p2: add PromptTemplate to ServiceOptions**
    - **context:** CR-06: Fix LLM template handling
    - **action:**
        1. Add `PromptTemplate` field to `llm.ServiceOptions` struct
    - **done‑when:**
        1. Field exists in struct
    - **depends‑on:** none

- [ ] **T048 · feature · p2: add WithPromptTemplate option function**
    - **context:** CR-06: Fix LLM template handling
    - **action:**
        1. Create `WithPromptTemplate(template string) ServiceOption` function
    - **done‑when:**
        1. Function exists and works
    - **depends‑on:** [T047]

- [ ] **T049 · refactor · p2: update service to use stored template**
    - **context:** CR-06: Fix LLM template handling
    - **action:**
        1. Modify service to store template from options
        2. Remove fallback template loading in `GenerateGlanceMarkdown`
    - **done‑when:**
        1. Service uses template from options
        2. Fallback loading is removed
    - **depends‑on:** [T048]

- [ ] **T050 · test · p2: add tests for custom template handling**
    - **context:** CR-06: Fix LLM template handling
    - **action:**
        1. Test service with custom template
        2. Verify generated content reflects template
    - **done‑when:**
        1. Tests pass for custom templates via `--prompt-file`
    - **depends‑on:** [T049]

## misc
- [ ] **T051 · feature · p3: add DefaultFileMode constant**
    - **context:** CR-09: Document file permissions rationale
    - **action:**
        1. Add `const DefaultFileMode = 0o600` in appropriate package
        2. Add comment explaining security rationale
    - **done‑when:**
        1. Constant is defined with documentation
    - **depends‑on:** none

- [ ] **T052 · refactor · p3: use DefaultFileMode in file writes**
    - **context:** CR-09: Document file permissions rationale
    - **action:**
        1. Replace hardcoded `0600` with `DefaultFileMode`
        2. Audit other file operations for consistency
    - **done‑when:**
        1. No direct `0600` literals in file operations
        2. All file writes use consistent permissions
    - **depends‑on:** [T051]

- [ ] **T053 · chore · p3: update README license reference**
    - **context:** CR-10: Reconcile license mismatch
    - **action:**
        1. Edit README.md license section to reference MIT
        2. Ensure link points to LICENSE file
    - **done‑when:**
        1. README correctly states MIT license
    - **depends‑on:** none

## clarifications & assumptions
- [ ] **issue:** Determine all locations accepting external file paths for validation
    - **context:** CR-08, step 1
    - **blocking?:** yes

- [ ] **issue:** Confirm all file write operations for permission standardization
    - **context:** CR-09, step 4
    - **blocking?:** no
