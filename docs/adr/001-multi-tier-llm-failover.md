# 001. Multi-Tier LLM Failover Chain

Date: 2025-12-01 (reconstructed from git history)

## Status

Accepted

## Context

Glance depends on LLM APIs for its core function. Single-provider dependency means any API outage, rate limit, or model deprecation renders the tool unusable. Users running glance on large codebases need reliable completion even during provider issues.

## Decision

Implement a composite `FallbackClient` that wraps N `Client` implementations in an ordered tier chain. Each tier gets independent retries with exponential backoff before advancing to the next tier. The current chain is:

1. `gemini-3-flash-preview` (primary, fastest)
2. `gemini-2.5-flash` (stable fallback, same provider)
3. `x-ai/grok-4.1-fast` via OpenRouter (cross-provider, requires separate API key)

The `Client` interface abstracts provider differences. `GeminiClient` uses the Google GenAI SDK; `OpenRouterClient` uses raw HTTP against the OpenAI-compatible API.

## Consequences

**Good:**
- Tool remains functional during single-provider outages
- Cross-provider fallback protects against Gemini-wide failures
- Interface abstraction makes adding new providers straightforward

**Bad:**
- Three retry layers (Service + FallbackClient + individual client) can amplify to 64 API calls worst case
- Backoff formulas are inconsistent (quadratic in clients, exponential in FallbackClient)
- `OpenRouterClient.GenerateStream` is a fake stream (single chunk) — streaming behavior differs by tier
- `CountTokens` is unsupported on OpenRouter — returns error, handled as optional
