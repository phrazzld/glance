# CI Resolution Tasks

## CRITICAL - Fix CI Blocking Issues

### T001: Fix Pre-commit Requirements Path - CRITICAL [x]
- **Priority**: P0 (Blocks all commits)
- **Description**: Update pre-commit workflow to use correct requirements.txt path after reorganization
- **Files**: `.github/workflows/precommit.yml` or similar pre-commit workflow file
- **Action**: Change `pip install -r requirements.txt` to `pip install -r docs/project/requirements.txt`
- **Validation**: Pre-commit workflow step passes in CI
- **Estimate**: 15 minutes
- **Completed**: Updated `.github/workflows/precommit.yml` to reference `docs/project/requirements.txt`

### T002: Pin govulncheck Version Across All Workflows - CRITICAL [x]
- **Priority**: P0 (Security + Stability)
- **Description**: Replace `@latest` with pinned version to prevent supply chain attacks and version drift
- **Files**: `.github/workflows/test.yml`, `.github/workflows/lint.yml`, `.github/workflows/performance.yml`
- **Action**: Replace all instances of `govulncheck@latest` with `govulncheck@v1.1.3`
- **Validation**: All workflows install consistent govulncheck version
- **Estimate**: 20 minutes
- **Completed**: Updated test.yml and lint.yml to use govulncheck@v1.1.3 (performance.yml was already pinned correctly)

## HIGH - Fix Core Test Failures

### T003: Update Clean Project Test Expectations - HIGH [x]
- **Priority**: P1 (Core functionality broken)
- **Description**: Fix TestVulnerabilityDetectionIntegration/Clean_project_scan output format expectations
- **Files**: `vulnerability_integration_test.go:84`
- **Action**: Update assertion to accept both old and new govulncheck output formats
- **Code Change**:
  ```go
  // Replace strict format check with semantic check
  assert.True(t,
      strings.Contains(combinedOutput, "No vulnerabilities found") ||
      strings.Contains(combinedOutput, "=== Symbol Results ==="),
      "Should indicate no vulnerabilities found")
  ```
- **Validation**: Clean project test passes
- **Estimate**: 30 minutes
- **Completed**: Updated test to use semantic check accepting both "No vulnerabilities found" and "=== Symbol Results ===" output formats

### T004: Fix Timeout Test Duration - HIGH [x]
- **Priority**: P1 (Test reliability)
- **Description**: Fix TestTimeoutHandling/Very_short_timeout test that doesn't actually timeout
- **Files**: `network_failure_test.go:120-124`
- **Action**: Reduce timeout from 1 second to 100ms OR create test scenario that genuinely takes longer
- **Options**:
  - Option A: Change timeout to `100 * time.Millisecond`
  - Option B: Use larger test project that takes longer to scan
- **Validation**: Timeout test properly demonstrates timeout behavior
- **Estimate**: 45 minutes
- **Completed**: Changed timeout from 1 second to 100ms (Option A), test now reliably times out and completes in ~100ms

### T005: Research and Create Actual Vulnerable Test Project - HIGH
- **Priority**: P1 (Test validity)
- **Description**: Fix TestVulnerabilityDetectionIntegration/Vulnerable_project_scan by using real vulnerabilities
- **Files**: `testdata/vulnerable-project/go.mod`, `testdata/vulnerable-project/main.go`
- **Action**:
  1. Research Go packages with known vulnerabilities compatible with govulncheck v1.1.3
  2. Update test project dependencies to include vulnerable packages
  3. Verify govulncheck detects vulnerabilities with non-zero exit code
- **Research Areas**:
  - Recent Go security advisories (golang.org/security)
  - Common vulnerable dependencies (old crypto packages, etc.)
  - govulncheck database for v1.1.3 timeframe
- **Validation**: Vulnerable project test returns non-zero exit code and detects vulnerabilities
- **Estimate**: 2 hours

## MEDIUM - Improve Test Reliability

### T006: Add Better Test Error Context - MEDIUM
- **Priority**: P2 (Debugging aid)
- **Description**: Improve test failure messages with full context for easier debugging
- **Files**: `vulnerability_integration_test.go`, `network_failure_test.go`
- **Action**: Add detailed error messages showing actual vs expected output
- **Code Pattern**:
  ```go
  assert.Contains(t, output, expected,
      "Expected pattern not found.\nFull output:\n%s\nExpected pattern: %s\nActual length: %d",
      output, expected, len(output))
  ```
- **Validation**: Test failures provide actionable debugging information
- **Estimate**: 45 minutes

### T007: Add govulncheck Version Validation Test - MEDIUM
- **Priority**: P2 (Prevent future issues)
- **Description**: Add test to validate govulncheck version matches expectations
- **Files**: New test in appropriate test file
- **Action**: Create test that checks govulncheck version and warns if unexpected
- **Code**:
  ```go
  func TestGovulncheckVersion(t *testing.T) {
      cmd := exec.Command("govulncheck", "version")
      output, err := cmd.CombinedOutput()
      require.NoError(t, err)
      // Check for expected version pattern
      assert.Contains(t, string(output), "v1.1.3",
          "Unexpected govulncheck version: %s", output)
  }
  ```
- **Validation**: CI fails early if wrong govulncheck version detected
- **Estimate**: 30 minutes

### T008: Improve Network Test Error Pattern Matching - MEDIUM
- **Priority**: P2 (Test reliability, from synthesis recommendations)
- **Description**: Fix network failure test assertions to check broader error patterns
- **Files**: `network_failure_test.go` (multiple locations with network error checks)
- **Action**: Implement the synthesis recommendation for comprehensive error pattern matching
- **Code**:
  ```go
  combinedOutput := strings.ToLower(result.StdOut + result.StdErr + result.ErrorMessage)
  assert.True(t,
      strings.Contains(combinedOutput, "network") ||
      strings.Contains(combinedOutput, "connection") ||
      strings.Contains(combinedOutput, "refused") ||
      strings.Contains(combinedOutput, "timeout") ||
      strings.Contains(combinedOutput, "dial"),
      "Error output should contain network-related error indicators")
  ```
- **Validation**: Network tests properly detect various error types
- **Estimate**: 30 minutes

## LOW - Documentation and Prevention

### T009: Document govulncheck Compatibility Matrix - LOW
- **Priority**: P3 (Prevention)
- **Description**: Create documentation for tested govulncheck versions and behaviors
- **Files**: New `docs/project/govulncheck-compatibility.md`
- **Content**:
  - Tested govulncheck versions
  - Known output format changes
  - Upgrade procedures and testing checklist
- **Validation**: Clear documentation exists for future maintainers
- **Estimate**: 1 hour

### T010: Add CI Environment Diagnostic Information - LOW
- **Priority**: P3 (Debugging aid)
- **Description**: Add CI step to output environment details for troubleshooting
- **Files**: CI workflow files
- **Action**: Add step to output govulncheck version, Go version, and other relevant info
- **Code**:
  ```yaml
  - name: Environment Diagnostics
    run: |
      echo "Go version: $(go version)"
      echo "govulncheck version: $(govulncheck version)"
      echo "Platform: $(uname -a)"
  ```
- **Validation**: CI logs contain useful diagnostic information
- **Estimate**: 20 minutes

---

## Task Dependencies

```
T001 (Fix pre-commit) → Independent (blocks all other work)
T002 (Pin versions) → Independent (required for test stability)
T003 (Clean test) → Depends on T002
T004 (Timeout test) → Depends on T002  
T005 (Vulnerable test) → Depends on T002
T006-T010 → Can be done in parallel after T002
```

## Success Criteria

- [ ] All CI checks pass (pre-commit + Go tests + existing passing checks)
- [ ] govulncheck functionality works reliably with pinned version
- [ ] Test suite provides meaningful, reliable validation
- [ ] No security vulnerabilities introduced (version pinning addressed)
- [ ] Future maintainability improved through documentation

## Estimated Total Time: 6-8 hours

**Critical Path**: T001 → T002 → T003/T004/T005 (parallel) → Validation
**Can be done in 2-3 focused work sessions**
