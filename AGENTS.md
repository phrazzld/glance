# AGENTS.md

Operational playbook for AI agents working in this repository.

## Commit Conventions

**Format:** Conventional Commits — `type(scope): description`

**Types:** `feat` (minor), `fix` (patch), `chore`, `docs`, `refactor`, `test`, `ci`

**Scopes:** `output`, `config`, `llm`, `deps`, `release` — optional but preferred when applicable.

**Rules:**
- Always write detailed multiline commit messages explaining the why
- Never add sign-off lines or co-author tags
- `[skip ci]` is reserved for automated release commits only
- Breaking changes use `feat!:` or `BREAKING CHANGE:` footer (triggers major release)

**Examples:**
```text
feat(llm): add Claude provider as fallback tier

Add a new LLM client implementation for Anthropic's Claude API,
integrated as tier 4 in the failover chain after OpenRouter/Grok.

This provides an additional safety net when both Gemini and
OpenRouter are experiencing outages.
```

```text
fix(filesystem): resolve symlink traversal in path validation

ValidatePathWithinBase now resolves symlinks before checking
containment, preventing symlinks inside base from pointing
to files outside the security boundary.

Closes #42
```

## Testing Guidelines

**Framework:** `github.com/stretchr/testify` — `assert` for soft checks, `require` for hard stops.

**Mandatory flags:** `-race` on every test run. CI enforces this.

**File naming:** `{name}_test.go` in the same package (white-box testing).

**Function naming:** `TestFunctionName` top-level, `t.Run("scenario description", ...)` for subtests.

**Mocking strategy:**
- Function variable injection for test seams (not constructor injection)
- Shared mocks in `internal/mocks/` package
- `llm.MockClientAdapter` bridges `mocks.LLMClient` → `llm.Client` (import cycle breaker)
- Use `testify/mock` for external boundaries; direct function replacement for internal seams

**Test helpers:** Use `t.Helper()`. Create temp dirs with `os.MkdirTemp`, clean up with `defer os.RemoveAll`.

**Coverage:** No hard threshold enforced, but coverage artifacts are uploaded to CI.

## PR Guidelines

**Branch naming:** `type/short-description` or `type/issue-{N}` (e.g., `fix/issue-50`, `feat/claude-provider`).

**Required sections** (no PR template exists — follow this manually):

1. **Summary** — What changed and why. Link issue if applicable.
2. **Changes** — Concise list of modified files/packages.
3. **Test coverage** — What tests were added/modified.
4. **Manual QA** — Steps to verify (build command, test command, expected output).

**Review:** Cerberus AI review runs automatically on all PRs. Address council feedback.

**Merge strategy:** Squash merge to master. Release automation triggers after tests pass.

## Coding Style

**Language:** Go 1.24. Standard library preferred over external dependencies.

**Naming:** Standard Go conventions — `CamelCase` exported, `camelCase` unexported.

**Error handling:** Use `glance/errors` package for typed errors with codes and suggestions. Wrap with context: `errors.WrapFileError(err, path, "reading config")`.

**Patterns in use:**
- Functional options (`ClientOption func(*ClientOptions)`)
- Immutable builders (config `With*` methods return copies)
- Function variable injection for test seams
- Composite/decorator for failover (`FallbackClient` wraps `[]Client`)

**File length:** 500 lines recommended, 1000 hard max (enforced by pre-commit).

**Linting:** golangci-lint v2.1.2 with `errcheck`, `govet`, `ineffassign`, `staticcheck`, `unused`, `misspell`, `gosec`. Tests excluded from linting.

## Version Pinning Invariants

These versions must stay in sync across all locations:

| Tool | Version | Locations |
|------|---------|-----------|
| golangci-lint | v2.1.2 | `.golangci.yml`, `.pre-commit-config.yaml`, `lint.yml` |
| govulncheck | v1.1.3 | `lint.yml`, `test.yml`, `precommit.yml` |
| Go | 1.24 | `go.mod`, all workflow files |

**GitHub Actions** are pinned to full 40-char commit SHAs, not version tags.

## Issue Workflow

**Templates available:** `bug_report.md`, `feature_request.md`, `refactoring.md`

**Labels:** `type:bug`, `type:feature`, `type:refactor`

**Picking work:** Check open issues, prioritize bugs. Create a branch matching `type/issue-{N}`.

## Definition of Done

- [ ] All tests pass: `go test -race ./...`
- [ ] Linting clean: `golangci-lint run --config=.golangci.yml --timeout=2m`
- [ ] No vulnerabilities: `govulncheck ./...`
- [ ] Pre-commit hooks pass: `pre-commit run --all-files`
- [ ] New code has tests covering behavior, not implementation
- [ ] Commit messages follow conventional commits spec
- [ ] No suppressed linter warnings — fix root causes

## Security Boundaries

**Never without human approval:**
- Modify `.github/workflows/` security gates (govulncheck, gosec)
- Set `EMERGENCY_SECURITY_OVERRIDE`
- Add `#nosec` annotations without preceding `ValidateFilePath` call
- Modify path validation logic in `filesystem/utils.go`
- Change file permissions from `0600`
- Add new external dependencies without vulnerability check

**Always verify:**
- Path validation before any `os.ReadFile`/`os.Open`
- Empty `baseDir` rejection in all validation functions
- Gitignore chain propagation when modifying scanner
