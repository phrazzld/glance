# 004. Dot-Prefixed Output Filename (.glance.md)

Date: 2025-06-01 (from commit f956d59)

## Status

Accepted (migrated from `glance.md` in v1.1.1)

## Context

The original output filename `glance.md` caused conflicts with build systems (e.g., documentation generators, static site builders) that process all `.md` files in a directory. Users reported their build pipelines picking up glance summaries as documentation pages.

## Decision

Rename output from `glance.md` to `.glance.md` (dot-prefixed, hidden on Unix). Maintain backward compatibility:

- `gatherSubGlances` checks `.glance.md` first, falls back to `glance.md`
- `ShouldRegenerate` forces regen when only legacy `glance.md` exists (triggers migration)
- `ShouldIgnoreFile` skips both filenames

## Consequences

**Good:**
- Build systems skip hidden files by default — no more conflicts
- Smooth migration path — old summaries are read, new ones are written with new name
- Files are less visible in directory listings (appropriate for generated content)

**Bad:**
- Two filenames to check everywhere (constants `GlanceFilename` and `LegacyGlanceFilename`)
- Legacy `glance.md` files are never cleaned up automatically — they persist alongside new `.glance.md`
- Users must use `ls -a` or equivalent to see output files
