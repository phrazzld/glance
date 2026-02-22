# CLAUDE.md

## What This Is

Glance is a Go CLI that recursively scans a directory tree and generates `.glance.md` summaries using LLMs. It uses a multi-tier failover chain (Gemini → Gemini stable → Grok/OpenRouter), processes directories bottom-up so parents incorporate child summaries, and tracks mod-times to skip unnecessary regeneration.

## Essential Commands

* **Build:** `go build -o glance`
* **Run:** `./glance [--force] [--prompt-file path] [directory]`
* **Test:** `go test -race ./...`
* **Test specific:** `go test -run=TestName ./package`
* **Lint:** `golangci-lint run --config=.golangci.yml --timeout=2m`
* **Format:** `go fmt ./...`
* **Vuln scan:** `govulncheck ./...`
* **Pre-commit:** `pre-commit run --all-files`

## Architecture

```text
main → config.LoadConfig → setupLLMService → scanDirectories (BFS)
     → processDirectories (leaf-first) → LLM generate → write .glance.md
```

Packages: `config` (CLI/env), `filesystem` (scan/read/validate), `llm` (client interface + failover), `errors` (typed hierarchy), `ui` (spinner).

See [docs/CODEBASE_MAP.md](docs/CODEBASE_MAP.md) for full architecture.

## Tech Stack

* **Go 1.24** — `google.golang.org/genai` (Gemini SDK), logrus, testify, progressbar
* **LLM providers:** Gemini API (primary), OpenRouter REST (fallback)
* **CI:** GitHub Actions — test, lint, build, govulncheck, semantic release via Landfall

## Quality Gates

CI blocks merge on any failure. Local pre-commit hooks mirror CI.

| Gate | Command | Notes |
|------|---------|-------|
| Tests | `go test -race ./...` | Race detector mandatory |
| Lint | golangci-lint v2.1.2 | errcheck, govet, staticcheck, gosec, etc. |
| Vuln scan | govulncheck v1.1.3 | Blocks on any vulnerability |
| File length | pre-commit hook | 500 warn, 1000 error |
| Secrets | detect-secrets | Baseline at `.secrets.baseline` |

**NEVER suppress linter warnings** — fix the root cause.

**NEVER lower quality gates** — write code to meet them.

## Code Style

* **Simplicity first.** Seek the simplest correct solution.
* **Package-by-feature** with clear interfaces.
* **Error handling:** Use `glance/errors` package for typed errors with codes and suggestions.
* **Naming:** Standard Go conventions. CamelCase exported, camelCase private.
* **Comments:** Explain why, not how. Code should be self-documenting.
* **Conventional Commits:** Always detailed multiline messages. Never sign commits.

## Gotchas

* **Version sync:** golangci-lint (v2.1.2) and govulncheck (v1.1.3) versions must match across `.pre-commit-config.yaml`, `.golangci.yml`, and all workflow files.
* **GitHub Actions pinned to full SHA** — never use version tags.
* **Output file is `.glance.md`** (dot-prefix) — legacy `glance.md` is read but not written.
* **File permissions:** All output uses `0600`. Security boundary enforced by `ValidateFilePath` before every read.
* **Prompt map ordering:** `FormatFileContents` iterates a Go map — non-deterministic order across runs.
* **Three retry layers:** Service + FallbackClient + individual client. Can amplify to 64 API calls worst case.
* **Sentinel errors are mutable** — known bug tracked in issue #60; `WithCause()` modifies globals and should return a new error value instead.
* **Symlinks not resolved** in path validation — documented known gap.

## Environment

| Variable | Required | Purpose |
|----------|----------|---------|
| `GEMINI_API_KEY` | Yes | Google Gemini API access |
| `OPENROUTER_API_KEY` | No | Cross-provider fallback (Grok) |
| `GLANCE_LOG_LEVEL` | No | `debug`, `info` (default), `warn`, `error` |

Supports `.env` file in CWD (godotenv, system env takes precedence).

## Deployment

CLI tool — no server deployment. Distributed as compiled binary.

* **Release:** Automated via Landfall after tests pass on master. Conventional commits drive semver.
* **Install:** `go install` or `go build -o glance`
* **Security override:** Break-glass override exists for CI emergencies; procedure is human-only. AI agents must never set or suggest security overrides.
