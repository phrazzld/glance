# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build/Test/Lint Commands

* **Run all tests:** `go test ./...`
* **Run specific test:** `go test -run=TestName ./package` (e.g., `go test -run=TestLoadPromptTemplate .`)
* **Run tests with race detection:** `go test -race ./...`
* **Run vulnerability scan:** `govulncheck ./...`
* **Run golangci-lint:** `golangci-lint run --config=.golangci.yml --timeout=2m`
* **Format code:** `go fmt ./...`
* **Run pre-commit hooks:** `pre-commit run --all-files`

## Code Style Guidelines

* **Simplicity First:** Seek the simplest correct solution. Eliminate unnecessary complexity.
* **Modularity:** Build small, focused components with clear interfaces following package-by-feature structure.
* **Design for Testability:** Structure code for easy automated testing without mocking internal collaborators.
* **Error Handling:** Use the project's error package for consistent, structured error handling with context.
* **Naming:** Use descriptive names with standard Go conventions (CamelCase for exported, camelCase for private).
* **Documentation:** Code should be self-documenting. Comments explain rationale (why), not how.
* **NEVER suppress linter warnings/errors** - fix the root cause instead.
* **Conventional Commits:** All commit messages must follow the spec for automated versioning.
* **Always write detailed multiline conventional commit messages**
* **NEVER sign your commit messages -- your commit messages should be strictly detailed multiline conventional commit messages about the work done**

## Security Requirements

This project implements **mandatory vulnerability scanning** to ensure dependency security. All code changes must pass security gates before deployment.

### Vulnerability Scanning Policy

* **HIGH/CRITICAL vulnerabilities BLOCK builds** - no exceptions without emergency override
* **MEDIUM/LOW vulnerabilities** are logged but non-blocking  
* **Automatic scanning** runs on all commits and pull requests
* **Fail-fast enforcement** - builds terminate immediately on security policy violations

### Working with Security Failures

When vulnerability scanning detects HIGH/CRITICAL vulnerabilities:

1. **Review the GitHub Actions summary** for immediate details
2. **Update vulnerable dependencies:** `go get -u && go mod tidy`
3. **Verify fixes locally:** `govulncheck ./...`
4. **Commit and push** the dependency updates

### Emergency Override Protocol

For **critical production hotfixes only**:

1. **Get security team approval** with documented justification
2. **Set emergency override:** `EMERGENCY_SECURITY_OVERRIDE=true` in CI environment
3. **Create security debt issue** automatically generated with 48-hour remediation timeline
4. **Resolve vulnerabilities within 48 hours** as required by policy

**WARNING:** All emergency overrides are logged, audited, and tracked for compliance.

### Security Resources

* **Detailed Guide:** See `docs/guides/security-scanning.md` for complete procedures
* **Configuration:** Security policies defined in `.govulncheck.yaml`
* **Local Scanning:** Install govulncheck: `go install golang.org/x/vuln/cmd/govulncheck@latest`
* **Quick Fix:** Update dependencies: `go get -u && go mod tidy`

### Security Gate Failures

If your build fails due to vulnerability scanning:
- **DO NOT** bypass security gates with force pushes or overrides
- **DO** update vulnerable dependencies to secure versions
- **DO** verify fixes with local vulnerability scanning before pushing
- **DO** seek security team guidance for complex dependency issues

Remember to adhere to all principles outlined in the Development Philosophy. Quality gates require passing all pre-commit hooks, CI checks, **and security scanning**. Do not bypass hooks with `--no-verify`.
