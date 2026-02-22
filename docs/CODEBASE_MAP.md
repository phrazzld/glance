---
last_mapped: 2026-02-21T17:00:00Z
total_files: 45
total_lines: 12533
---

# Codebase Map

> Generated with Claude Code on 2026-02-21; maintained manually and not auto-updated.

## System Overview

Glance is a Go CLI tool that recursively scans a directory tree and generates a `.glance.md` file in each directory. Each summary is produced by an LLM pipeline with a multi-tier failover chain (Gemini → Gemini stable → Grok via OpenRouter), retries with backoff, and gitignore-aware file filtering.

The tool processes directories bottom-up (leaves first) so parent summaries can incorporate child summaries. It tracks modification times to avoid unnecessary regeneration on subsequent runs.

## Architecture

```text
┌─────────────────────────────────────────────────┐
│  main() / glance.go                             │
│  CLI entry → config → scan → process → debrief  │
└──────┬──────────┬──────────┬──────────┬─────────┘
       │          │          │          │
  ┌────▼───┐ ┌───▼────┐ ┌───▼──┐ ┌────▼───┐
  │ config │ │filesys │ │ llm  │ │  ui    │
  │        │ │tem     │ │      │ │        │
  │ CLI    │ │ BFS    │ │Client│ │Spinner │
  │ flags  │ │ scan   │ │(iface│ │        │
  │ env    │ │ ignore │ │)     │ │        │
  │ prompt │ │ read   │ │      │ │        │
  │ tmpl   │ │ valid  │ │      │ │        │
  └────────┘ └────────┘ └──┬───┘ └────────┘
                            │
              ┌─────────────┼─────────────┐
              │             │             │
        ┌─────▼────┐ ┌─────▼────┐ ┌──────▼─────┐
        │ Gemini   │ │ Gemini   │ │ OpenRouter │
        │ Client   │ │ Client   │ │ Client     │
        │ (flash)  │ │ (stable) │ │ (Grok)     │
        └──────────┘ └──────────┘ └────────────┘
```

### Failover Chain

```text
Tier 1: gemini-3-flash-preview (primary, Gemini SDK)
  ↓ all retries exhausted
Tier 2: gemini-2.5-flash (stable fallback, Gemini SDK)
  ↓ all retries exhausted
Tier 3: x-ai/grok-4.1-fast (cross-provider, OpenRouter REST)
```

Each tier gets `retriesPerTier` attempts with exponential backoff (200ms base, 30s cap, ±20% jitter) before advancing. `FallbackClient` is the sole retry owner — `GeminiClient.Generate` and `Service` each make a single attempt.

## Directory Structure

```text
glance/
├── glance.go              # Core: main(), scan, process loop, debrief
├── config/
│   ├── config.go          # Config struct + builder methods
│   ├── loadconfig.go      # CLI flag parsing, env loading
│   ├── template.go        # Prompt template file loading
│   └── vulnerability.go   # govulncheck config (CI only)
├── errors/
│   └── errors.go          # Typed error hierarchy (GlanceError interface)
├── filesystem/
│   ├── scanner.go         # BFS directory traversal + gitignore chains
│   ├── ignore.go          # File/dir ignore decisions
│   ├── reader.go          # File reading, UTF-8 sanitization, truncation
│   ├── utils.go           # Path validation, mod-time, regen logic
│   └── logger.go          # Package-level injectable logger
├── llm/
│   ├── client.go          # Client interface + GeminiClient impl
│   ├── client_adapter.go  # Mock adapter (breaks import cycle)
│   ├── backoff.go         # Shared ExponentialBackoff with jitter
│   ├── fallback_client.go # Multi-tier failover composite client (sole retry owner)
│   ├── openrouter_client.go # OpenRouter REST client
│   ├── prompt.go          # Template rendering + file formatting
│   └── service.go         # App-layer orchestration (single-attempt)
├── ui/
│   └── feedback.go        # Spinner + error reporting
├── internal/mocks/
│   └── llm_client.go      # Testify mock for llm.Client
├── scripts/               # Dev setup, pre-commit, govulncheck retry
├── docs/                  # Guides, design docs, performance tests
└── .github/workflows/     # CI: test, lint, build, release, perf
```

## Module Guide

### Root Package (glance.go)

**Entry point:** `main()` → `config.LoadConfig` → `setupLLMService` → `scanDirectories` → `processDirectories` → `printDebrief`

**Key functions:**
- `processDirectories` — iterates leaf-first, calls LLM, writes `.glance.md`
- `gatherSubGlances` — reads child `.glance.md` files (with legacy `glance.md` fallback)
- `readSubdirectories` — lists non-hidden, non-ignored subdirs
- `setupLLMServiceFunc` — swappable function variable (test seam)

**Processing order:** BFS scan collects all dirs, then reversed for bottom-up processing. Parent regeneration bubbles up via `filesystem.BubbleUpParents` when a child is regenerated.

### config

Handles CLI flags (`--force`, `--prompt-file`), `.env` loading via godotenv, `GEMINI_API_KEY` validation, and prompt template resolution.

**Config builder:** Immutable functional-style — each `With*` method returns a new copy.

**Prompt fallback chain:** `--prompt-file` arg → `prompt.txt` in CWD → `llm.DefaultTemplate()`.

### errors

Typed error hierarchy with builder pattern: `New("msg").WithCode("X").WithSeverity(Critical).WithSuggestion("fix")`.

Types: `FileSystemError`, `APIError`, `ConfigError`, `ValidationError`. All implement `GlanceError` interface with `errors.Is`/`As` support.

**Warning:** Sentinel errors (e.g., `ErrFileNotFound`) are mutable — `WithCause()` modifies them in place. Concurrent use with different causes is unsafe.

### filesystem

Core file operations with security-first design.

- **scanner.go** — BFS with per-directory gitignore chain accumulation
- **ignore.go** — Centralized ignore logic; checks `.glance.md`, hidden files, `node_modules`, gitignore patterns
- **reader.go** — `ReadTextFile` with path validation, UTF-8 sanitization, binary detection via `http.DetectContentType`
- **utils.go** — Path validation (`ValidatePathWithinBase`, `ValidateFilePath`, `ValidateDirPath`), mod-time comparison, regen logic

**Security:** All file reads go through `ValidateFilePath` before `os.ReadFile`. Empty `baseDir` is rejected. Symlinks are NOT resolved (documented known gap).

### llm

LLM abstraction layer with interface-based design and composite failover.

- **Client interface** — `Generate`, `GenerateStream`, `CountTokens`, `Close`
- **GeminiClient** — Google GenAI SDK, functional options, single-attempt Generate
- **OpenRouterClient** — HTTP REST, fake streaming (single chunk), no token counting
- **FallbackClient** — Composite pattern wrapping N clients; sole retry owner with `ExponentialBackoff` (200ms base, 30s cap, ±20% jitter)
- **Service** — Builds prompts, calls client once, logs metadata
- **ExponentialBackoff** (`backoff.go`) — Shared utility: `base*2^(attempt-1)`, capped at maxWait, with cryptographic ±20% jitter

**Token management:** `CountTokens` is called for logging only. No automatic truncation — oversized prompts fail at the API and retry.

### ui

Terminal feedback via spinner (briandowns/spinner). Progress bar is used directly from `glance.go` via schollz/progressbar.

## Data Flow

```text
CLI args + env vars
    │
    ▼
LoadConfig() → Config{APIKey, TargetDir, Force, PromptTemplate}
    │
    ▼
setupLLMService() → FallbackClient{Gemini, Gemini-stable, Grok} → Service
    │
    ▼
scanDirectories() → BFS with gitignore chains → sorted dir list
    │
    ▼
reverse(dirs) → leaf-first order
    │
    ▼
for each dir:
    ├─ ShouldRegenerate(modtime comparison)
    ├─ readSubdirectories() → child dir names
    ├─ gatherSubGlances() → child .glance.md content
    ├─ gatherLocalFiles() → map[filename]content
    ├─ BuildPromptData() + GeneratePrompt() → rendered prompt
    ├─ Service.GenerateGlanceMarkdown() → single LLM call (FallbackClient retries)
    ├─ ValidateFilePath() → security check
    └─ os.WriteFile(.glance.md, 0600)
```

## Test Seams

| Variable | Location | Purpose |
|---|---|---|
| `setupLLMServiceFunc` | `glance.go` | Replace LLM client/service |
| `dirChecker` | `config/loadconfig.go` | Replace directory validation |
| `loadPromptTemplate` | `config/loadconfig.go` | Replace prompt file loader |
| `validateFilePath` | `config/template.go` | Replace path validator |
| `createGeminiClient` | `llm/client.go` | Replace Gemini factory |
| `createOpenRouterClient` | `llm/openrouter_client.go` | Replace OpenRouter factory |

All are package-level function variables enabling test injection without constructor changes.

## CI Pipeline

| Workflow | Trigger | Purpose |
|---|---|---|
| test.yml | push/PR to master | `go test -race` + coverage |
| lint.yml | push/PR to master | golangci-lint, go vet, govulncheck |
| build.yml | push/PR to master | Cross-platform build (Ubuntu + macOS) |
| precommit.yml | push/PR + weekly | Pre-commit hooks in CI |
| release.yml | after tests pass on master | Semantic release via Landfall |
| cerberus.yml | PR events | AI code review council |
| performance.yml | weekly + manual | govulncheck performance benchmarks |

## Conventions

- **Output filename:** `.glance.md` (dot-prefixed, hidden). Legacy `glance.md` is read but not written.
- **File permissions:** `0600` (owner read/write only) for all generated files.
- **Error codes:** `FS-001` through `FS-005`, `API-001` through `API-005`, `CFG-001` through `CFG-004`, `VAL-001` through `VAL-003`.
- **Logging:** logrus with structured fields (`directory`, `operation`, `token_count`, `model`).
- **Testing:** testify assert/require, function variable injection, `-race` mandatory.
- **Linting:** golangci-lint v2.1.2 with gosec. Tests excluded from linting.

## Gotchas

1. **Map iteration in prompts** — `FormatFileContents` iterates `map[string]string` non-deterministically. Same input produces different prompt orderings across runs.
2. **Single retry owner** — Only `FallbackClient` retries. `GeminiClient.Generate` and `Service.GenerateGlanceMarkdown` are single-attempt. Worst case: `(retriesPerTier+1) × len(tiers)` API calls per directory.
3. **Sentinel error mutation** — `errors.ErrFileNotFound.WithCause(err)` permanently mutates the global sentinel. Unsafe for concurrent use.
4. **Symlinks not resolved** — Path validation checks string prefixes, not resolved targets. A symlink inside base pointing outside base passes validation.
5. **Service.promptTemplate defaults to ""** — Callers must explicitly pass `WithPromptTemplate(llm.DefaultTemplate())` or get empty prompts.
6. **golangci-lint version sync** — Must match across `.golangci.yml`, `lint.yml`, and `.pre-commit-config.yaml`. Currently v2.1.2.
7. **govulncheck pinned at v1.1.3** — Must match across `lint.yml`, `test.yml`, `precommit.yml`.
8. **TruncateContent splits UTF-8** — Truncates at byte boundary, can produce broken codepoints.

## Navigation Guide

**To add a new LLM provider:** Create a new `Client` implementation in `llm/`, add it as a `FallbackTier` in `glance.go:createLLMService()`.

**To change the prompt template:** Edit `llm/prompt.go:DefaultTemplate()` or pass `--prompt-file`.

**To add a new ignore rule:** Update `filesystem/ignore.go:ShouldIgnoreFile` or `ShouldIgnoreDir`.

**To modify CLI flags:** Edit `config/loadconfig.go:LoadConfig`.

**To add a new error type:** Add to `errors/errors.go` following the `baseError` embedding pattern.

**To update CI quality gates:** Edit `.github/workflows/lint.yml` and `.pre-commit-config.yaml` (keep versions in sync).
