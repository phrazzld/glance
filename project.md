# Project: Glance

## Vision
CLI tool that recursively scans directories and generates high-quality `.glance.md` summaries using LLM failover chains.

**North Star:** The go-to developer tool for instant codebase comprehension — every directory self-documenting, always current, zero maintenance. Drop-in integration: one workflow file and your repo documents itself forever.

**Target User:** Developers who want quick, automated documentation of directory contents across codebases.

**Current Focus:** Two tracks in parallel:
1. **Reliability** — fix known bugs, clean up technical debt
2. **Zero-friction integration** — GitHub Actions workflow, post-merge hooks; make it trivial to adopt and keep fresh

**Key Differentiators:**
- Multi-provider LLM failover (Gemini, OpenRouter/Grok) for reliability
- Smart regeneration via modification time checking
- Respects .gitignore, handles large files gracefully
- Zero config for basic use, customizable prompts for advanced use

## Domain Glossary

| Term | Definition |
|------|-----------|
| `.glance.md` | Output file per directory containing LLM-generated summary |
| `glance.md` | Legacy read-only filename (read but not written) |
| `FallbackClient` | Composite LLM client that tries providers in order |
| Bottom-up processing | Leaf directories first; parents incorporate child summaries |
| Mod-time tracking | Skip regeneration if directory contents unchanged |
| Prompt template | Customizable instruction sent to LLM; default in `llm.DefaultTemplate()` |
| Tier | One LLM provider+model in the failover chain |

## Active Focus

- **Theme 1:** Reliability and correctness (fix sentinels, determinism, path issues, UTF-8 truncation)
- **Theme 2:** Zero-friction CI integration (composite action in `.github/actions/glance/`, auto-commit on by default)
- **Key Issues (Reliability / v1.1 / now):** #51, #53, #60, #66 — all `effort/s`
- **Key Issues (CI Integration / Backlog / next):** #67 (action), #68 (docs)

## Quality Bar

- `go test -race ./...` — race detector mandatory
- `golangci-lint run --config=.golangci.yml` — zero warnings, never suppress
- `govulncheck ./...` — no known vulnerabilities
- No coverage gate yet (issue #32 tracks adding one)
- Conventional commits with detailed multiline messages

## Patterns to Follow

### Functional Options Config
```go
// Each With* returns new immutable copy
cfg := config.NewConfig().
    WithForce(true).
    WithPromptFile("custom.txt")
```

### Test Seam Injection
```go
// Function vars swapped in tests (6 seams in glance.go)
var generateSummary = service.Generate
// Tests: generateSummary = func(...) { return fakeResult, nil }
```

### Typed Errors
```go
return glanceerrors.New(glanceerrors.CodeLLMError, "provider unavailable").
    WithSuggestion("check GEMINI_API_KEY").
    WithCause(err)
```

## Lessons Learned

| Decision | Outcome | Lesson |
|----------|---------|--------|
| Three retry layers | Confusing, redundant | Only FallbackClient retries (fixed #64) |
| Absolute paths in LLM prompt | LLM gets machine-specific noise | Use relative paths (#51) |
| Go map for file contents | Non-deterministic prompt ordering | Sort keys or use ordered slice (#53) |

---
*Last updated: 2026-02-23*
*Updated during: /groom session*
