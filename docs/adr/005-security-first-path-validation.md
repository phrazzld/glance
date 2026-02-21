# 005. Security-First Path Validation

Date: 2025-01-01 (reconstructed from code patterns)

## Status

Accepted

## Context

Glance reads arbitrary files from user-specified directories and sends their content to external LLM APIs. Path traversal vulnerabilities could expose sensitive files outside the intended scan directory. The tool runs with the user's full filesystem permissions.

## Decision

Enforce path containment at every file access point:

1. `ValidatePathWithinBase` — core check: `filepath.Clean` + `filepath.Abs` + `strings.HasPrefix(absPath, absBaseDir+sep)`
2. `ValidateFilePath` / `ValidateDirPath` — type-checked wrappers
3. Empty `baseDir` rejected at all entry points (prevents accidental wide-open validation)
4. All `os.ReadFile`/`os.Open` calls preceded by validation (`#nosec G304` only after validation)
5. Generated files use `0600` permissions (owner-only)

## Consequences

**Good:**
- Path traversal via `../` and absolute paths is blocked
- `#nosec` annotations are narrow and justified
- Defense-in-depth: multiple validation layers

**Bad:**
- Symlinks are NOT resolved — a symlink inside base pointing outside passes validation (documented known gap)
- URL-encoded paths (`%2e`) are not decoded — application layer must decode before calling validation
- String-prefix containment can have edge cases with path separators (mitigated by appending `os.PathSeparator`)
