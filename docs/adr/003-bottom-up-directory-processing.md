# 003. Bottom-Up Directory Processing with Parent Regeneration

Date: 2025-01-01 (reconstructed from code patterns)

## Status

Accepted

## Context

Glance generates summaries for every directory in a tree. Parent summaries should incorporate information from child summaries. Processing order matters: if a parent is processed before its children, it won't have child summaries to reference.

Additionally, when a child changes, its parent's summary becomes stale because it referenced old child content.

## Decision

1. BFS scan collects all directories
2. Reverse the list for leaf-first (bottom-up) processing
3. Each directory's `.glance.md` includes child `.glance.md` content via `gatherSubGlances`
4. When a child is regenerated, `BubbleUpParents` marks all ancestors for regeneration
5. On subsequent runs, `ShouldRegenerate` compares `.glance.md` mod-time against the latest file mod-time in the directory

## Consequences

**Good:**
- Parent summaries always incorporate current child summaries
- Incremental runs only regenerate what changed (mod-time tracking)
- Bubble-up ensures ancestor staleness is handled automatically

**Bad:**
- A single file change at depth N triggers regeneration of N directories up to root
- The reverse-BFS approach requires holding all directory paths in memory (fine for typical trees)
- Sibling isolation is maintained (changing branch A doesn't touch branch B) but root is always affected
