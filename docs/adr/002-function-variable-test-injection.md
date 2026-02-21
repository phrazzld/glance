# 002. Function Variable Injection for Test Seams

Date: 2025-01-01 (reconstructed from code patterns)

## Status

Accepted

## Context

Go's strict type system and lack of monkey-patching makes testing code that calls external APIs challenging. Common approaches: constructor injection (interfaces in constructor params), test doubles via interfaces, or function variable replacement.

Constructor injection adds complexity to public APIs and creates interface-heavy designs. The codebase is a CLI tool with straightforward call chains — heavy DI is unnecessary.

## Decision

Use package-level function variables as test seams. Six swappable functions exist:

- `setupLLMServiceFunc` — replace entire LLM stack
- `dirChecker` — replace directory validation
- `loadPromptTemplate` — replace prompt file loading
- `validateFilePath` — replace path validation
- `createGeminiClient` — replace Gemini factory
- `createOpenRouterClient` — replace OpenRouter factory

Tests save the original, replace with a test double, and restore via `defer`.

## Consequences

**Good:**
- Zero-overhead in production (no interface indirection at runtime)
- Simple test setup — no factory builders or DI containers
- Public API stays clean (no injected dependencies in constructors)

**Bad:**
- Package-level mutable state — not safe for parallel test execution on the same variable
- Tests must remember to save/restore (forgetting leaks state across tests)
- Less discoverable than constructor injection — seams are implicit
