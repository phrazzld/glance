# Remediation Plan – Sprint 1

## Executive Summary
This sprint targets critical security vulnerabilities, code duplication, and CI inconsistencies identified in the code review. We prioritize securing CI tooling and standardizing development workflows first, followed by addressing security issues with path validation and suppressions. Completing filesystem refactoring and fixing template handling will ensure the codebase is both secure and maintainable going forward.

## Strike List
| Seq | CR-ID | Title | Effort | Owner |
|-----|-------|-------|--------|-------|
| 1 | CR-01 | Secure CI: Replace `curl \| sh` with GitHub Action | XS | DevOps |
| 2 | CR-02 | Standardize golangci-lint version & configuration | S | DevOps |
| 3 | CR-07 | Re-enable pre-commit linter | XS | DevOps |
| 4 | CR-03 | Address excessive `#nosec` suppressions | M | Security |
| 5 | CR-08 | Implement file path validation | M | Security |
| 6 | CR-04 | Eliminate duplicate filesystem logic | L | Backend |
| 7 | CR-05 | Complete filesystem abstraction refactoring | M | Backend |
| 8 | CR-06 | Fix LLM template handling | S | Backend |
| 9 | CR-09 | Document file permissions rationale | XS | Docs |
| 10 | CR-10 | Reconcile license mismatch | XS | Docs |

## Detailed Remedies

### CR-01: Secure CI: Replace `curl | sh` with GitHub Action
- **Problem:** CI workflow installs golangci-lint via `curl | sh`, directly piping a script from the internet into a shell.
- **Impact:** Critical security vulnerability; exposes the build process to supply chain attacks if the download source is compromised.
- **Chosen Fix:** Replace `curl | sh` with the official `golangci/golangci-lint-action@v4`.
- **Steps:**
  1. Modify `.github/workflows/lint.yml`.
  2. Remove the `curl | sh` step for golangci-lint installation.
  3. Add a step using `uses: golangci/golangci-lint-action@v4`.
  4. Configure with the desired version (matching CR-02) and arguments.
- **Done When:** CI lint job passes using the GitHub Action; `curl | sh` pattern removed from workflow.

### CR-02: Standardize golangci-lint version & configuration
- **Problem:** Multiple different versions and configurations of golangci-lint across CI, pre-commit, and docs.
- **Impact:** Inconsistent linting results; quality gates bypassed; increased maintenance burden.
- **Chosen Fix:** Define a single source of truth for the version and ensure consistent configuration.
- **Steps:**
  1. Determine target golangci-lint version (e.g., v1.57.0).
  2. Update `rev:` in `.pre-commit-config.yaml` to this version.
  3. Configure GitHub Action to use the same version.
  4. Update documentation to reference this version.
  5. Ensure `.golangci.yml` is used by both pre-commit and CI.
- **Done When:** All references to golangci-lint use the same version; configuration is consistent.

### CR-07: Re-enable pre-commit linter
- **Problem:** The golangci-lint hook is commented out in `.pre-commit-config.yaml`.
- **Impact:** Developers can commit code with lint errors, causing CI failures later.
- **Chosen Fix:** Uncomment and configure the golangci-lint hook.
- **Steps:**
  1. Edit `.pre-commit-config.yaml` to uncomment the hook.
  2. Ensure `rev:` matches the version chosen in CR-02.
  3. Ensure `args:` point to the correct `.golangci.yml` file.
  4. Test locally with `pre-commit run --all-files`.
- **Done When:** golangci-lint runs during pre-commit; prevents committing lint violations.

### CR-03: Address excessive `#nosec` suppressions
- **Problem:** Numerous `#nosec` annotations for G304 (file paths) and G306 (permissions) without justification.
- **Impact:** Masks potential security vulnerabilities; violates coding standards; accumulates technical debt.
- **Chosen Fix:** Review all suppressions, remove unnecessary ones, add proper validation, document remaining ones.
- **Steps:**
  1. Audit all `#nosec` annotations in the codebase.
  2. For path security (G304), implement validation (see CR-08) and remove suppressions.
  3. For permissions (G306), document rationale (see CR-09) and remove suppressions.
  4. Create a `SECURITY_SUPPRESSIONS.md` document for any necessary global suppressions.
- **Done When:** All remaining suppressions have clear justification; unnecessary ones removed.

### CR-08: Implement file path validation
- **Problem:** File paths from CLI args, config files, etc. are used without proper validation.
- **Impact:** Potential path traversal vulnerabilities allowing access outside intended directories.
- **Chosen Fix:** Add robust path validation for all externally-sourced paths.
- **Steps:**
  1. Identify all locations reading paths from CLI, config, etc.
  2. Implement validation: `filepath.Clean`, `filepath.Abs`, prefix checking.
  3. Add tests for path traversal attempts (e.g., `../etc/passwd`).
  4. Remove G304 suppressions where validation is added.
- **Done When:** All external paths are validated; tests confirm protection against traversal attacks.

### CR-04: Eliminate duplicate filesystem logic
- **Problem:** Core filesystem functions duplicated between `glance.go` and `filesystem/` package.
- **Impact:** Violates DRY; increases maintenance burden; potential for inconsistent behavior.
- **Chosen Fix:** Refactor to use only the `filesystem/` package implementation.
- **Steps:**
  1. Identify all duplicate functions in `glance.go`.
  2. Refactor `glance.go` to use equivalent `filesystem/` package functions.
  3. Delete duplicate implementations from `glance.go`.
  4. Update tests to ensure functionality is preserved.
- **Done When:** All filesystem logic lives only in `filesystem/` package; `glance.go` uses these functions.

### CR-05: Complete filesystem abstraction refactoring
- **Problem:** Partial refactoring - `filesystem/` uses `IgnoreChain` but `glance.go` still uses old types.
- **Impact:** Leaky abstraction; technical debt; hinders future improvements.
- **Chosen Fix:** Fully migrate to use `IgnoreChain` throughout the codebase.
- **Steps:**
  1. Update functions in `glance.go` to accept and use `IgnoreChain` instead of old types.
  2. Remove compatibility helpers (`ExtractGitignoreMatchers`, `CreateIgnoreChain`).
  3. Ensure all logic properly uses `IgnoreChain` abstractions.
- **Done When:** Only `IgnoreChain` is used throughout; compatibility helpers removed.

### CR-06: Fix LLM template handling
- **Problem:** LLM service ignores custom prompt templates specified in configuration.
- **Impact:** Custom templates don't work; configuration option is broken.
- **Chosen Fix:** Pass and use the template from configuration in the LLM service.
- **Steps:**
  1. Modify `llm.ServiceOptions` to include `PromptTemplate`.
  2. Add `WithPromptTemplate` option function.
  3. Update service to store and use the provided template.
  4. Remove fallback template loading in `GenerateGlanceMarkdown`.
  5. Add tests to verify custom templates work.
- **Done When:** Custom templates via `--prompt-file` work correctly.

### CR-09: Document file permissions rationale
- **Problem:** File write permissions changed to 0600 without documenting rationale.
- **Impact:** Unclear security decisions; potential for inconsistency in other file operations.
- **Chosen Fix:** Document rationale for 0600 and standardize usage with a constant.
- **Steps:**
  1. Add a `const DefaultFileMode = 0o600` in appropriate package.
  2. Replace hardcoded 0600 with the constant.
  3. Add comments explaining the security rationale.
  4. Audit for any other file write operations to ensure consistency.
- **Done When:** Consistent file permissions with clear documentation.

### CR-10: Reconcile license mismatch
- **Problem:** README claims Apache-2.0 license, but LICENSE file is MIT.
- **Impact:** Legal ambiguity; documentation inaccuracy.
- **Chosen Fix:** Update README to match the actual MIT license in LICENSE file.
- **Steps:**
  1. Edit README.md license section.
  2. Change text to state MIT license.
  3. Ensure link points to correct LICENSE file.
- **Done When:** README correctly reflects MIT license to match LICENSE file.

## Standards Alignment
- **Simplicity:** Removing code duplication (CR-04) and completing abstractions (CR-05) directly improve simplicity.
- **Modularity:** Fixing duplicated filesystem logic (CR-04) and completing the `IgnoreChain` refactoring (CR-05) enforce clean package boundaries.
- **Testability:** Standardizing tools (CR-01, CR-02, CR-07) provides a stable base for testing. Adding path validation (CR-08) improves security test coverage.
- **Coding Standards:** Addressing suppressions (CR-03), fixing template handling (CR-06), and documenting security decisions (CR-09) bring code in line with standards.
- **Security:** All fixes contribute to improved security, particularly the CI changes (CR-01), path validation (CR-08), and proper handling of suppressions (CR-03).

## Validation Checklist
- Automated tests pass (`go test ./...`)
- Static analyzers pass (`golangci-lint run`, `go vet ./...`)
- Pre-commit hooks pass locally (`pre-commit run --all-files`)
- CI workflows complete successfully
- Path traversal tests confirm security
- Custom templates work correctly
- All permissions are properly justified
- License is consistent
