# Security Vulnerability Scanning Guide

This document provides guidance for the vulnerability scanning system in the Glance project. The system uses `govulncheck` to block builds with vulnerable dependencies.

## Overview

The vulnerability scanner automatically runs in CI and blocks builds when vulnerabilities are detected in Go dependencies.

### How It Works

- `govulncheck` scans all dependencies during CI
- Builds fail if any vulnerabilities are found
- Developers update dependencies to fix vulnerabilities
- Builds pass once dependencies are secure

## Developer Workflow

When vulnerabilities are detected:

1. Check the CI failure logs for vulnerability details
2. Update dependencies to fix vulnerabilities:
   ```bash
   go get -u && go mod tidy
   ```
3. Verify fixes locally:
   ```bash
   go install golang.org/x/vuln/cmd/govulncheck@latest
   govulncheck ./...
   ```
4. Commit and push the dependency updates

## Local Testing

Install and run govulncheck locally:

```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

## Emergency Override Protocol

In critical situations where a hotfix must be deployed despite vulnerabilities:

1. **Activate Override**: Set `EMERGENCY_SECURITY_OVERRIDE=true` in the CI environment
2. **Document Justification**: The override reason is automatically logged
3. **Create Tracking Issue**: Manually create a GitHub issue with `security-debt` label
4. **48-Hour Remediation**: Security issues must be resolved within 48 hours
5. **Audit Trail**: All override usage is logged for compliance review

### Example Override Usage:
```yaml
# In GitHub Actions workflow
env:
  EMERGENCY_SECURITY_OVERRIDE: true
```

**⚠️ Warning**: Emergency overrides should only be used for critical production hotfixes. All usage is audited.

## Resources

- [Go Vulnerability Database](https://vuln.go.dev)
- [Dependency Management Guide](https://go.dev/doc/modules/managing-dependencies)
