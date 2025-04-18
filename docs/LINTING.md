# Golangci-lint Standardization

This document establishes the standardized approach for golangci-lint usage across the Glance project.

## Standardized Approach

After careful consideration of workflow efficiency, consistency, and best practices, we have established the following standardized approach for golangci-lint invocation:

### Primary Method: Pre-commit Hooks (Local Development)

**For local development and routine code checking, pre-commit hooks are the standard method for running golangci-lint.**

- Developers must install and use pre-commit hooks in their local environment
- golangci-lint is configured in `.pre-commit-config.yaml` as part of the pre-commit hooks
- This method ensures consistent code quality checks before code is committed

### Secondary Method: Official GitHub Action (CI Pipeline)

**For CI workflows, the golangci-lint-action is the standard method for running golangci-lint.**

- CI pipelines use the official golangci-lint GitHub Action
- This provides additional features like caching and detailed reporting optimized for CI environments
- Can be configured to focus on new or changed code for PR checks

Both methods use the same `.golangci.yml` configuration file to ensure consistent linting rules across environments.

## Rationale

This standardization approach was chosen for the following reasons:

1. **Dual Environment Optimization:**
   - Pre-commit hooks provide immediate feedback in the local development environment
   - GitHub Actions provide deeper analysis with better reporting features in CI environments

2. **Consistency with Security Requirements:**
   - Both methods follow security best practices for tool installation
   - Neither approach uses insecure installation methods like curl | sh

3. **Configuration Consistency:**
   - Both approaches use the same `.golangci.yml` configuration file
   - The same linting rules apply in both local and CI environments

4. **Performance Optimization:**
   - Pre-commit hooks are optimized for speed in local environments
   - GitHub Actions leverage caching and parallel execution in CI

## Implementation Details

### Local Development (Pre-commit Hooks)

golangci-lint is configured in `.pre-commit-config.yaml`:

```yaml
- repo: https://github.com/golangci/golangci-lint
  rev: v1.57.0
  hooks:
    - id: golangci-lint
      name: golangci-lint
      description: Fast Go linters runner that uses .golangci.yml config
      entry: golangci-lint run
      types: [go]
      language: system
      pass_filenames: false
      args: ["--config=.golangci.yml", "--timeout=2m"]
      exclude: '^vendor/|^precommit-tests/'
```

### CI Pipeline (GitHub Action)

golangci-lint is configured in `.github/workflows/lint.yml`:

```yaml
- name: Install golangci-lint
  uses: golangci/golangci-lint-action@v4
  with:
    version: v1.57.0
    args: --config=.golangci.yml
    only-new-issues: true
```

## Version Consistency

To ensure consistency across environments:

1. The golangci-lint version should be specified and kept synchronized between:
   - `.pre-commit-config.yaml` (in the `rev:` field)
   - `.github/workflows/lint.yml` (in the `version:` field)

2. When updating golangci-lint, both locations must be updated to the same version

## Migration Plan

All existing implementations have been reviewed and align with this standardization. Future changes to linting configuration should follow this established pattern.

## Related Documentation

- [GITHUB_ACTIONS.md](GITHUB_ACTIONS.md) - For details on GitHub Actions workflow configuration
- [PRECOMMIT.md](PRECOMMIT.md) - For details on pre-commit hooks configuration and usage
