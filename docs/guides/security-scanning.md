# Security Vulnerability Scanning Guide

This document provides comprehensive guidance for the mandatory vulnerability scanning system integrated into the Glance project. The system uses `govulncheck` to identify security vulnerabilities in Go dependencies and enforces security policies in the CI/CD pipeline.

## Table of Contents

1. [Overview](#overview)
2. [Security Policy](#security-policy)
3. [Configuration](#configuration)
4. [Scan Behavior](#scan-behavior)
5. [Developer Workflow](#developer-workflow)
6. [Emergency Override Procedures](#emergency-override-procedures)
7. [Vulnerability Response](#vulnerability-response)
8. [Troubleshooting](#troubleshooting)
9. [Common Scenarios](#common-scenarios)
10. [Integration Details](#integration-details)

## Overview

The Glance project implements mandatory vulnerability scanning to ensure that security vulnerabilities in dependencies are identified and addressed before code reaches production. This system provides:

- **Fail-fast security gates** that block builds containing HIGH/CRITICAL vulnerabilities
- **Comprehensive reporting** with structured JSON output and GitHub Actions summaries
- **Emergency override capabilities** for critical production hotfixes
- **Full audit trails** with correlation ID tracking for compliance
- **Automated issue creation** for security debt tracking

### Key Components

- **govulncheck**: Go's official vulnerability scanner
- **GitHub Actions integration**: Automated scanning on every PR and push
- **Structured logging**: JSON logs with correlation IDs for observability
- **Emergency override system**: Controlled bypass mechanism for critical situations

## Security Policy

### Blocking Vulnerabilities

The system blocks builds when dependencies contain vulnerabilities of the following severities:

- **🔴 CRITICAL**: Immediate build failure, no exceptions
- **🟠 HIGH**: Immediate build failure, emergency override available
- **🟡 MEDIUM**: Advisory only, build continues with warning
- **🟢 LOW**: Advisory only, build continues with warning

### Policy Enforcement

1. **Automatic Enforcement**: All commits and pull requests are scanned
2. **No Bypass by Default**: HIGH/CRITICAL vulnerabilities always block deployment
3. **Emergency Override**: Available for critical production hotfixes with approval
4. **Compliance Tracking**: All overrides logged and tracked with 48-hour remediation requirement

## Configuration

### Scan Configuration File

The vulnerability scanner is configured via `.govulncheck.yaml` in the project root:

```yaml
# Severity levels that cause build failure
fail_on_severity:
  - "HIGH"
  - "CRITICAL"

# Scan timeout in seconds
timeout_seconds: 300

# Scan level: module, package, or symbol
scan_level: "symbol"

# Output format for reports
output_format: "json"

# Retry attempts for network failures
retry_attempts: 2

# Reporting configuration
reporting:
  include_details: true
  include_remediation: true
  generate_correlation_id: true
```

### Configuration Parameters

| Parameter | Description | Default | Options |
|-----------|-------------|---------|---------|
| `fail_on_severity` | Vulnerability severities that block builds | `["HIGH", "CRITICAL"]` | `["CRITICAL", "HIGH", "MEDIUM", "LOW"]` |
| `timeout_seconds` | Maximum scan duration | `300` | Any positive integer |
| `scan_level` | Depth of vulnerability analysis | `"symbol"` | `"module"`, `"package"`, `"symbol"` |
| `output_format` | Report format | `"json"` | `"json"`, `"text"` |
| `retry_attempts` | Network failure retry count | `2` | Any non-negative integer |

## Scan Behavior

### When Scans Run

Vulnerability scans are automatically triggered on:

- **Push to master branch** (excluding documentation changes)
- **Pull requests** targeting the master branch
- **Monthly dependency checks** (1st of each month at 01:00 UTC)

### Scan Process

1. **Configuration Loading**: Read settings from `.govulncheck.yaml`
2. **Correlation ID Generation**: Create unique identifier for traceability
3. **Vulnerability Database Update**: Fetch latest vulnerability data
4. **Dependency Analysis**: Scan all Go modules and dependencies
5. **Severity Assessment**: Classify findings by severity level
6. **Policy Enforcement**: Apply configured failure thresholds
7. **Report Generation**: Create structured JSON reports and GitHub summaries
8. **Artifact Upload**: Store reports for 30 days with correlation ID

### Scan Timing

- **Target Duration**: Under 60 seconds for typical projects
- **Timeout**: 300 seconds (configurable)
- **Parallel Execution**: Runs alongside linting and testing workflows

## Developer Workflow

### Normal Development Flow

1. **Develop and commit** code changes as usual
2. **Push to GitHub** or create pull request
3. **Automatic scan** runs in CI pipeline
4. **Review results** in GitHub Actions summary
5. **Address vulnerabilities** if any are found
6. **Re-commit and push** fixes

### When Vulnerabilities Are Found

#### 🔴 HIGH/CRITICAL Vulnerabilities

**Build Status**: ❌ BLOCKED

**Required Actions**:
1. Review vulnerability details in the GitHub Actions summary
2. Update affected dependencies to secure versions
3. Run `govulncheck ./...` locally to verify fixes
4. Commit and push updated dependencies

**Quick Fix Commands**:
```bash
# Update all dependencies
go get -u && go mod tidy

# Verify fixes
govulncheck ./...

# Commit updates
git add go.mod go.sum
git commit -m "fix: update dependencies to resolve security vulnerabilities"
```

#### 🟡 MEDIUM/LOW Vulnerabilities

**Build Status**: ✅ CONTINUES (with advisory)

**Recommended Actions**:
- Consider updating affected dependencies when convenient
- Monitor for security advisories on these packages
- Plan updates during next maintenance window

### Local Development

#### Install govulncheck

```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
```

#### Run Local Scans

```bash
# Quick scan
govulncheck ./...

# Detailed JSON output
govulncheck -format json ./... | jq

# Scan specific module
govulncheck ./path/to/module
```

#### Pre-commit Verification

```bash
# Run full pre-commit checks (includes vulnerability scan)
pre-commit run --all-files

# Or just vulnerability check
govulncheck ./...
```

## Emergency Override Procedures

### When to Use Emergency Override

Emergency override should **ONLY** be used for:

- **Critical production hotfixes** that cannot wait for dependency updates
- **Security patches** that must be deployed immediately
- **System outages** requiring immediate code deployment

### Override Process

#### 1. Security Team Approval

- Contact security team via designated channels
- Provide justification for emergency override
- Document business impact and risk assessment
- Obtain written approval with 48-hour remediation commitment

#### 2. Activate Override

Set the emergency override environment variable in the GitHub Actions workflow:

```yaml
env:
  EMERGENCY_SECURITY_OVERRIDE: true
```

Or configure it in repository secrets for reusable access.

#### 3. Deploy with Override

- Push commit with override activated
- Monitor build logs for override confirmation
- Verify automatic security debt issue creation

#### 4. Post-Override Requirements

**Immediate Actions** (within 1 hour):
- [ ] Document override justification in the auto-created GitHub issue
- [ ] Create remediation plan with specific timeline
- [ ] Assign security team member for tracking

**Within 48 Hours**:
- [ ] Update vulnerable dependencies
- [ ] Run vulnerability scan to confirm resolution
- [ ] Deploy updated dependencies
- [ ] Close security debt issue
- [ ] Conduct post-incident review

### Override Audit Trail

Every override activation generates:

- **Structured logs** with correlation ID and user information
- **GitHub issue** with `security-debt` label for tracking
- **Audit timestamps** for compliance reporting
- **Remediation timeline** with automatic reminders

## Vulnerability Response

### Immediate Response (HIGH/CRITICAL)

1. **Assess Impact**: Review vulnerability details and affected components
2. **Check Exploitability**: Determine if vulnerability affects your usage
3. **Update Dependencies**: Use latest secure versions
4. **Test Thoroughly**: Verify functionality after updates
5. **Deploy Quickly**: Prioritize security fixes

### Dependency Update Strategies

#### Option 1: Targeted Updates

```bash
# Update specific vulnerable package
go get package/name@latest
go mod tidy

# Verify fix
govulncheck ./...
```

#### Option 2: Comprehensive Updates

```bash
# Update all dependencies
go get -u all
go mod tidy

# Run full test suite
go test -race ./...

# Verify security fixes
govulncheck ./...
```

#### Option 3: Pin to Secure Version

```bash
# Pin to specific secure version
go get package/name@v1.2.3
go mod tidy
```

### Communication During Response

- **Stakeholder Updates**: Inform relevant teams of security findings
- **Progress Tracking**: Use GitHub issue comments for status updates
- **Documentation**: Record decisions and rationale for future reference

## Troubleshooting

### Common Issues

#### 1. Scan Timeouts

**Symptoms**:
- Build fails with timeout error after 300 seconds
- "Vulnerability scan timed out" message in logs

**Solutions**:
- Increase `timeout_seconds` in `.govulncheck.yaml`
- Check network connectivity to vulnerability database
- Consider reducing `scan_level` from `symbol` to `package`

```yaml
# Increase timeout for large codebases
timeout_seconds: 600
```

#### 2. Network Connectivity Issues

**Symptoms**:
- "Failed to fetch vulnerability database" errors
- Intermittent scan failures

**Solutions**:
- Verify https://vuln.go.dev accessibility
- Increase `retry_attempts` in configuration
- Check GitHub Actions network policies

#### 3. False Positive Vulnerabilities

**Symptoms**:
- Vulnerabilities reported for unused code paths
- Theoretical vulnerabilities in dead code

**Solutions**:
- Use `symbol`-level scanning to reduce false positives
- Review actual code usage vs. vulnerability paths
- Consider dependency replacement if issue persists

#### 4. Configuration File Issues

**Symptoms**:
- "Configuration file not found" errors
- Invalid configuration format warnings

**Solutions**:
- Ensure `.govulncheck.yaml` exists in project root
- Validate YAML syntax with `yamllint`
- Check file permissions and encoding

#### 5. Emergency Override Not Working

**Symptoms**:
- Override variable set but build still fails
- Override not recognized in logs

**Solutions**:
- Verify exact variable name: `EMERGENCY_SECURITY_OVERRIDE`
- Ensure value is exactly `"true"` (case-sensitive)
- Check variable scope (repository vs. environment)

### Debug Mode

Enable detailed logging for troubleshooting:

```bash
# Local debugging
GOVULNCHECK_DEBUG=1 govulncheck ./...

# CI debugging - add to workflow
- name: Debug vulnerability scan
  env:
    GOVULNCHECK_DEBUG: 1
    ACTIONS_STEP_DEBUG: true
  run: govulncheck -v ./...
```

### Log Analysis

#### Correlation ID Tracking

Every scan generates a correlation ID in the format: `vuln-scan-YYYYMMDD-HHMMSS-{git-sha}`

Use this ID to:
- Track related log entries across systems
- Reference specific scan instances
- Correlate with GitHub artifacts and issues

#### Structured Log Fields

```json
{
  "timestamp": "2024-01-15T10:30:45.123Z",
  "level": "ERROR",
  "service_name": "vulnerability-scanner",
  "correlation_id": "vuln-scan-20240115-103045-a1b2c3d",
  "git_commit": "a1b2c3d",
  "scan_result": "vulnerabilities_found",
  "vulnerability_counts": {
    "critical_count": 0,
    "high_count": 2,
    "medium_count": 1,
    "low_count": 0,
    "total_vulnerabilities": 3
  },
  "scan_duration_ms": 45230,
  "message": "Vulnerabilities detected - security policy violation"
}
```

## Common Scenarios

### Scenario 1: New HIGH Severity Vulnerability

**Situation**: New vulnerability discovered in widely-used dependency

**Response**:
1. **Immediate**: All new builds blocked automatically
2. **Assessment**: Review impact on current deployment
3. **Update**: Upgrade to patched version
4. **Testing**: Verify compatibility and functionality
5. **Deployment**: Push updates through normal process

**Timeline**: Target resolution within 24 hours

### Scenario 2: Critical Production Hotfix

**Situation**: Production system down, immediate fix required, but dependency has vulnerability

**Response**:
1. **Emergency Protocol**: Activate override with security team approval
2. **Risk Assessment**: Document risk vs. business impact
3. **Deploy**: Push critical fix with override
4. **Immediate Remediation**: Start dependency update process
5. **Follow-up**: Complete security fix within 48 hours

### Scenario 3: False Positive in Unused Code

**Situation**: Vulnerability reported in code path not used by application

**Response**:
1. **Investigation**: Verify code path is truly unused
2. **Documentation**: Record analysis in security issue
3. **Code Cleanup**: Remove unused dependencies if possible
4. **Alternative**: Pin to current version with justification
5. **Monitoring**: Watch for actual exploitability

### Scenario 4: Dependency Update Breaks Functionality

**Situation**: Security update introduces breaking changes

**Response**:
1. **Immediate**: Identify specific breaking changes
2. **Options Assessment**:
   - Find alternative secure dependency
   - Implement compatibility layer
   - Temporarily pin with monitoring
3. **Implementation**: Apply chosen solution
4. **Testing**: Comprehensive verification
5. **Documentation**: Record decisions and monitoring plan

### Scenario 5: Vulnerability Database Outage

**Situation**: govulncheck cannot reach vulnerability database

**Response**:
1. **Verification**: Confirm database availability at https://vuln.go.dev
2. **Temporary Measure**: Proceed with builds if critical
3. **Monitoring**: Set up alerts for database restoration
4. **Catch-up**: Run comprehensive scan once database available
5. **Process Review**: Consider backup scanning strategies

## Integration Details

### GitHub Actions Workflow

The vulnerability scanning is integrated into `.github/workflows/lint.yml`:

```yaml
vulnerability-scan:
  name: Run vulnerability scanning
  runs-on: ubuntu-latest
  timeout-minutes: 10
  steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.24'

    - name: Run vulnerability scan
      run: |
        govulncheck -format json ./... > vuln-report.json
        # Process results and apply policy
```

### Artifacts and Reports

#### Structured JSON Reports

Generated for each scan with correlation ID:

```json
{
  "scan_metadata": {
    "correlation_id": "vuln-scan-20240115-103045-a1b2c3d",
    "timestamp": "2024-01-15T10:30:45Z",
    "git_commit": "a1b2c3d",
    "scanner_version": "govulncheck@v1.1.3"
  },
  "vulnerability_summary": {
    "total_vulnerabilities": 2,
    "critical_count": 0,
    "high_count": 2,
    "medium_count": 0,
    "low_count": 0,
    "scan_result": "vulnerabilities_found"
  },
  "findings": [...],
  "remediation": {
    "update_commands": ["go get -u && go mod tidy"],
    "resources": ["https://vuln.go.dev"]
  }
}
```

#### GitHub Actions Summary

Provides immediate visual feedback:

- **🔍 Vulnerability Scan Summary** with status indicators
- **Severity breakdown** with color-coded counts  
- **Actionable guidance** based on findings
- **Quick links** to resources and artifacts

### Observability Integration

#### Structured Logging

All scan events generate JSON logs:
- Configuration loading
- Scan start/completion
- Vulnerability detection
- Override activation
- Build termination

#### Metrics Extraction

Available metrics for monitoring dashboards:
- `vulnerability_scan_duration_ms`: Scan execution time
- `vulnerability_count_by_severity`: Breakdown by severity
- `scan_result`: clean|vulnerabilities_found|timeout|error
- `override_usage_count`: Emergency override activations

### Compliance Features

#### Audit Trail

Complete record of all security-related actions:
- Scan executions with correlation IDs
- Override activations with approver information
- Remediation tracking with timestamps
- Policy violations with detailed context

#### Reporting Capabilities

- **Security dashboard**: Vulnerability trends over time
- **Compliance reports**: Override usage and remediation status
- **Incident tracking**: Correlation ID based investigation
- **Performance metrics**: Scan duration and success rates

---

## Resources

- **Go Vulnerability Database**: https://vuln.go.dev
- **govulncheck Documentation**: https://go.dev/doc/security/vuln/
- **Dependency Management Guide**: https://go.dev/doc/modules/managing-dependencies
- **GitHub Actions Workflows**: [github-actions.md](github-actions.md)
- **Security Policy**: Contact security team for current policies

## Support

For questions or issues with vulnerability scanning:

1. **Technical Issues**: Create GitHub issue with `security-scanning` label
2. **Policy Questions**: Contact security team via designated channels  
3. **Emergency Override**: Follow emergency contact procedures
4. **Documentation Updates**: Submit pull request with improvements
