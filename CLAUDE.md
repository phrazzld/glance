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

This project uses **mandatory vulnerability scanning** to ensure dependency security. All code changes must pass security gates before deployment.

### Vulnerability Scanning

* **Vulnerable dependencies BLOCK builds** - update dependencies to fix
* **Automatic scanning** runs on all commits and pull requests via `govulncheck`
* **Local testing:** `go install golang.org/x/vuln/cmd/govulncheck@latest && govulncheck ./...`

### Working with Security Failures

When vulnerability scanning fails:

1. **Update vulnerable dependencies:** `go get -u && go mod tidy`
2. **Verify fixes locally:** `govulncheck ./...`
3. **Commit and push** the dependency updates

### Security Resources

* **Guide:** See `docs/guides/security-scanning.md` for complete procedures
* **Local Scanning:** Install govulncheck: `go install golang.org/x/vuln/cmd/govulncheck@latest`
* **Quick Fix:** Update dependencies: `go get -u && go mod tidy`
