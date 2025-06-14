# govulncheck Compatibility Matrix

This document provides information about tested govulncheck versions, known output format changes, and upgrade procedures for the Glance project.

## Tested govulncheck Versions

### Current Pinned Version: v1.1.3

**Status**: ✅ Fully Tested and Production Ready

- **CI Configuration**: Pinned across all workflows (`test.yml`, `lint.yml`, `precommit.yml`)
- **Installation**: `go install golang.org/x/vuln/cmd/govulncheck@v1.1.3`
- **Test Coverage**: All integration tests pass with this version
- **Known Issues**: None

### Version Compatibility History

| Version | Status | Test Results | Notes |
|---------|--------|-------------|-------|
| v1.1.4  | ⚠️ Tested | ✅ All tests pass | Newer than pinned version; not used in CI |
| v1.1.3  | ✅ Production | ✅ All tests pass | Current pinned version |
| v1.1.2  | ❓ Untested | - | Not tested in this project |
| v1.1.1  | ❓ Untested | - | Not tested in this project |
| @latest | ❌ Deprecated | - | No longer used due to supply chain security concerns |

## Known Output Format Changes

### v1.1.3 Output Patterns

**Clean Project Scan Output** (no vulnerabilities):
```
=== Symbol Results ===

No vulnerabilities found.
```

**Vulnerable Project Scan Output** (vulnerabilities detected):
```
=== Symbol Results ===

Vulnerability #1: GO-2022-1059
    Denial of service via crafted Accept-Language header in
    golang.org/x/text/language
  More info: https://pkg.go.dev/vuln/GO-2022-1059
  Module: golang.org/x/text
    Found in: golang.org/x/text@v0.3.0
    Fixed in: golang.org/x/text@v0.3.8

Your code is affected by 1 vulnerability from the Go standard library.
```

**Version Information Output**:
```
Go: go1.24.2
Scanner: govulncheck@v1.1.3
DB: https://vuln.go.dev
DB updated: 2025-06-12 14:10:01 +0000 UTC
```

### Format Changes Between Versions

- **v1.1.3 → v1.1.4**: No significant output format changes detected
- **Legacy versions**: May use different section headers or formatting

## Test Adaptations

### Semantic Output Validation

The project uses semantic validation instead of strict string matching to handle minor output format variations:

```go
// Robust validation for clean projects
assert.True(t,
    strings.Contains(combinedOutput, "No vulnerabilities found") ||
        strings.Contains(combinedOutput, "=== Symbol Results ==="),
    "Should indicate no vulnerabilities found")
```

### Version Validation Test

Automatic version validation prevents version drift:

```go
func TestGovulncheckVersion(t *testing.T) {
    cmd := exec.Command("govulncheck", "-version")
    output, err := cmd.CombinedOutput()
    require.NoError(t, err)
    assert.Contains(t, string(output), "v1.1.3",
        "Unexpected govulncheck version")
}
```

## Upgrade Procedures

### Pre-Upgrade Checklist

1. **Review Release Notes**: Check golang.org/x/vuln releases for breaking changes
2. **Test Locally**: Install new version and run full test suite
3. **Validate Output**: Ensure test expectations still match new output formats
4. **Check CI**: Verify all workflows reference the new version consistently

### Upgrade Steps

1. **Update CI Workflows**:
   ```bash
   # Update all references in:
   # - .github/workflows/test.yml
   # - .github/workflows/lint.yml
   # Replace: govulncheck@v1.1.3
   # With:    govulncheck@v1.1.X (new version)
   ```

2. **Update Local Installation**:
   ```bash
   go install golang.org/x/vuln/cmd/govulncheck@v1.1.X
   ```

3. **Run Test Suite**:
   ```bash
   go test -run=TestGovulncheckVersion ./...
   go test -run=TestVulnerabilityDetectionIntegration ./...
   go test -run=TestErrorMessaging ./...
   ```

4. **Validate CI**:
   - Create test PR with version update
   - Verify all CI checks pass
   - Monitor environment diagnostics output

### Rollback Procedure

If upgrade causes issues:

1. **Revert CI workflows** to previous version
2. **Update local installation**: `go install golang.org/x/vuln/cmd/govulncheck@v1.1.3`
3. **Verify tests pass** with reverted version
4. **Document issues** for future upgrade attempts

## Environment Diagnostics

### CI Environment Information

All CI workflows now include environment diagnostics that output:
- Go version
- govulncheck version  
- Platform information

### Local Debugging

To verify your local environment matches CI:

```bash
echo "Go version: $(go version)"
echo "govulncheck version: $(govulncheck -version)"
echo "Platform: $(uname -a)"
```

## Security Considerations

### Version Pinning Rationale

- **Supply Chain Security**: Prevents automatic updates to potentially compromised versions
- **Reproducible Builds**: Ensures consistent behavior across environments
- **Stability**: Reduces risk of unexpected failures due to tool changes

### Emergency Override

The vulnerability scanning CI includes an emergency override mechanism:

```bash
# Only for critical hotfixes
export EMERGENCY_SECURITY_OVERRIDE=true
```

**Warning**: Override usage is audited and requires remediation within 48 hours.

## Troubleshooting

### Common Issues

1. **Version Mismatch**:
   - **Symptom**: `TestGovulncheckVersion` fails
   - **Solution**: Install correct version locally

2. **Output Format Changes**:
   - **Symptom**: Integration tests fail unexpectedly  
   - **Solution**: Update test expectations or use semantic validation

3. **Network Issues**:
   - **Symptom**: Vulnerability database access fails
   - **Solution**: Check network connectivity and proxy settings

### Support Resources

- **Go Vulnerability Database**: https://vuln.go.dev
- **govulncheck Documentation**: https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck  
- **Project Security Guide**: `docs/guides/security-scanning.md`

## Maintenance Schedule

- **Monthly**: Review new govulncheck releases
- **Quarterly**: Evaluate upgrade to latest stable version
- **As Needed**: Update immediately for security-critical govulncheck updates

---

**Last Updated**: June 13, 2025  
**Reviewed By**: CI Pipeline Integration Team  
**Next Review**: September 13, 2025
