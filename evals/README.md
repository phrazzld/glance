# Glance Evals

Promptfoo-based eval pipeline for measuring and comparing prompt quality.

## What This Tests

Glance's LLM call takes a rendered prompt (directory path + file contents + sub-glance summaries) and produces a structured markdown summary. The evals verify:

1. **Format compliance** — output contains `## Purpose`, `## Key Roles`, `## Dependencies and Caveats`
2. **Constraint adherence** — no CLI flag descriptions, no recommendations, no speculation
3. **Quality** — LLM-as-judge factual accuracy score (DeepSeek V3 rubric)
4. **Sub-glance inheritance** — parent summaries reference child directory topics

## Structure

```
evals/
├── promptfooconfig.yaml   # Smoke suite (7 cases, runs in CI)
├── bakeoff.yaml           # v1 vs v2 prompt comparison
├── provider.js            # Custom provider: renders Go template + calls OpenRouter
├── assertions/
│   └── glance.assert.js   # Format + constraint validators
├── fixtures/              # Frozen directory snapshots (JSON)
│   ├── errors_pkg.json    # glance/errors — format test baseline
│   ├── config_pkg.json    # glance/config — CLI flag constraint test
│   ├── llm_pkg.json       # glance/llm — quality + constraint tests
│   ├── minimal.json       # Synthetic single-file — minimal format test
│   └── with_subglances.json  # Synthetic parent — sub-glance inheritance test
└── prompts/
    ├── v1.txt             # Legacy prompt (before fix/issue-52)
    └── v2.txt             # Current prompt (constrained format)
```

## Prerequisites

```bash
# OPENROUTER_API_KEY must be set (used for generation + llm-rubric judge)
export OPENROUTER_API_KEY=sk-or-...

cd evals
npm install
```

## Running Evals

**Smoke suite (7 cases):**
```bash
cd evals
npx promptfoo eval --config promptfooconfig.yaml
```

**Bakeoff (v1 vs v2, all fixtures):**
```bash
cd evals
npx promptfoo eval --config bakeoff.yaml
```

**View results in browser:**
```bash
npx promptfoo view
```

**Output to JSON for scripting:**
```bash
npx promptfoo eval --config promptfooconfig.yaml --output results.json
```

## CI Integration

- **`eval-smoke.yml`** — triggers on PRs touching `llm/prompt.go`, `llm/service.go`, or `evals/`. Non-blocking (`continue-on-error: true`). Posts summary comment to PR.
- **`eval-nightly.yml`** — runs full suite + bakeoff nightly. Files a GitHub issue if failure rate exceeds 20%.

Both workflows require `OPENROUTER_API_KEY` as a repository secret.

## Fixture Format

```json
{
  "directory": "glance/errors",
  "files": {
    "errors.go": "package errors\n..."
  },
  "subGlances": []
}
```

`subGlances` is an array of strings — the rendered content of child `.glance.md` files, passed to the prompt as subdirectory summaries.

## Adding Test Cases

1. Add a fixture to `fixtures/` (frozen snapshot from the actual codebase)
2. Add a test case to `promptfooconfig.yaml` referencing the fixture
3. Choose an assertion: `hasRequiredHeaders`, `noCliSpeculation`, `noRecommendations`, `noSpeculation`, or `llm-rubric`

## Assertion Reference

| Function | Tests |
|----------|-------|
| `hasRequiredHeaders` | Output has `## Purpose`, `## Key Roles`, `## Dependencies and Caveats` |
| `noCliSpeculation` | No `--flag` patterns, `[default: ...]`, or short flags in backticks |
| `noRecommendations` | No "I recommend", "you should", "consider using" |
| `noSpeculation` | No "likely", "probably", "it seems", "appears to" |

LLM-as-judge uses `openrouter:deepseek/deepseek-v3.2` with threshold 0.7.
