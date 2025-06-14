# Network Resilience for Vulnerability Scanning

## Overview

This document describes the network resilience features implemented for vulnerability scanning, ensuring reliable operation even under degraded network conditions. The system implements comprehensive retry logic, timeout handling, and graceful degradation to maintain security scanning capabilities.

## Network Failure Scenarios

### 1. Complete Network Isolation

**Scenario**: No network connectivity to vulnerability database
- **Detection**: Connection timeout or immediate connection failure
- **Behavior**: Retry with exponential backoff, clear error messaging
- **Resolution**: Wait for connectivity restoration or use emergency override

### 2. DNS Resolution Failures

**Scenario**: Unable to resolve vulnerability database hostname
- **Detection**: DNS resolution errors for `vuln.go.dev`
- **Behavior**: Network error classification, retry attempts
- **Resolution**: Check DNS configuration or wait for DNS service restoration

### 3. Slow Network Connections

**Scenario**: High latency or limited bandwidth to vulnerability database
- **Detection**: Scan duration approaching timeout thresholds
- **Behavior**: Extended timeout handling, performance monitoring
- **Resolution**: Automatic completion with performance warnings

### 4. Intermittent Connectivity

**Scenario**: Sporadic network failures during scanning
- **Detection**: Intermittent connection failures
- **Behavior**: Intelligent retry with network connectivity validation
- **Resolution**: Retry until stable connection or maximum attempts reached

## Retry Logic Implementation

### Configuration

```yaml
# .govulncheck.yaml
retry_attempts: 2          # Number of retry attempts
timeout_seconds: 300       # Maximum scan duration
```

### Retry Wrapper Script

The `scripts/govulncheck-with-retry.sh` script provides comprehensive retry functionality:

```bash
# Basic usage with retries
./scripts/govulncheck-with-retry.sh ./...

# Custom retry configuration
./scripts/govulncheck-with-retry.sh -r 5 -t 600 ./...

# Environment variable configuration
GOVULNCHECK_RETRY_ATTEMPTS=3 ./scripts/govulncheck-with-retry.sh ./...
```

### Retry Decision Logic

1. **Network Error Classification**: Automatically identifies network-related errors
2. **Retry Eligibility**: Only network errors trigger retries, not scan failures
3. **Exponential Backoff**: 5-second delay between retry attempts (configurable)
4. **Maximum Attempts**: Configurable limit to prevent infinite loops

## Error Classification

### Network Errors (Retryable)
- Connection timeout (exit code 124)
- Connection refused
- DNS resolution failures
- Network unreachable
- Temporary name resolution failures

### Non-Network Errors (Non-Retryable)
- Invalid command arguments
- Scan configuration errors
- Vulnerability findings (when configured to fail)
- Permission errors

## Timeout Handling

### Timeout Configuration

```yaml
timeout_seconds: 300  # 5 minutes default
```

### Timeout Behavior

1. **Process Timeout**: Uses `timeout` command to enforce maximum duration
2. **Early Detection**: Monitors scan progress for early timeout prediction
3. **Graceful Termination**: Clean process termination with clear error messages
4. **Resource Cleanup**: Ensures temporary files are properly cleaned up

### Timeout Error Messages

```
❌ FAILURE: Vulnerability scan timed out after 300 seconds
This may indicate network issues or an unusually large codebase.
Consider increasing timeout_seconds in .govulncheck.yaml
```

## Graceful Degradation

### Performance Monitoring

- **Baseline Performance**: ~3 seconds for typical projects
- **Warning Threshold**: Scans approaching 45 seconds
- **Failure Threshold**: Scans exceeding 60 seconds
- **Adaptive Behavior**: Automatic timeout extension for large codebases

### Degradation Scenarios

1. **Slow Network Performance**
   - **Detection**: Scan duration >10x baseline
   - **Response**: Continue with extended timeout
   - **Logging**: Performance degradation warnings

2. **Partial Connectivity**
   - **Detection**: Sporadic connection failures
   - **Response**: Retry with connectivity validation
   - **Logging**: Network stability warnings

3. **Database Unavailability**
   - **Detection**: Persistent connection failures
   - **Response**: Emergency override notification
   - **Logging**: Service availability errors

## Connectivity Validation

### Pre-Scan Validation

```bash
# Check vulnerability database connectivity
curl -s --max-time 10 --head https://vuln.go.dev
```

### Validation Results

- **Success**: Proceed with normal scanning
- **Failure**: Log warning but attempt scan (may fail quickly)
- **Unavailable**: Network error classification for retry logic

## Structured Logging

### Log Format

```json
{
  "timestamp": "2025-06-07T19:35:43Z",
  "level": "INFO",
  "service_name": "govulncheck-wrapper",
  "correlation_id": "govulncheck-retry-20250607-123543-aa2f8f8",
  "message": "Network error detected, will retry after 5s delay"
}
```

### Log Levels

- **INFO**: Normal operation, retry attempts, completion
- **WARN**: Performance degradation, connectivity warnings
- **ERROR**: Network failures, timeout events, retry exhaustion

### Correlation IDs

Format: `govulncheck-retry-YYYYMMDD-HHMMSS-{git-sha}`
- Enables tracking across retry attempts
- Links with CI/CD pipeline logs
- Facilitates debugging and incident analysis

## Integration with CI/CD

### GitHub Actions Integration

```yaml
- name: Run vulnerability scan with network resilience
  run: |
    ./scripts/govulncheck-with-retry.sh ./...
  env:
    GOVULNCHECK_RETRY_ATTEMPTS: 3
    GOVULNCHECK_TIMEOUT_SECONDS: 600
```

### Exit Codes

- **0**: Scan completed successfully
- **1**: Scan failed (non-network error)
- **124**: Scan timed out
- **125**: Network connectivity issues
- **126**: Configuration error
- **127**: Prerequisites not met

## Testing Network Resilience

### Network Simulation Tests

```bash
# Enable network tests
export RUN_NETWORK_TESTS=true

# Run network failure tests
go test -v -run TestNetworkFailureScenarios

# Run retry logic tests  
go test -v -run TestRetryLogic
```

### Test Scenarios

1. **Complete Network Isolation**: Proxy to non-existent address
2. **DNS Failures**: Invalid DNS configuration
3. **Slow Networks**: Simulated high latency
4. **Intermittent Connectivity**: Periodic network isolation

### Performance Benchmarks

```bash
# Benchmark network performance impact
go test -bench=BenchmarkNetworkPerformance
```

## Emergency Override

### When Network Fails Completely

If vulnerability scanning cannot complete due to persistent network issues:

1. **Document the Network Issue**: Include error messages and correlation IDs
2. **Security Team Approval**: Get approval for emergency override
3. **Set Override Variable**: `EMERGENCY_SECURITY_OVERRIDE=true`
4. **Track Security Debt**: Automatic issue creation for follow-up

### Override Audit Trail

```json
{
  "override_timestamp": "2025-06-07T19:35:43Z",
  "correlation_id": "govulncheck-retry-20250607-123543-aa2f8f8",
  "reason": "Persistent network connectivity issues",
  "approved_by": "security-team",
  "remediation_deadline": "2025-06-09T19:35:43Z"
}
```

## Troubleshooting Network Issues

### Common Issues

1. **Firewall Blocking**: Corporate firewalls blocking HTTPS to vuln.go.dev
2. **Proxy Configuration**: HTTP/HTTPS proxy misconfiguration
3. **DNS Issues**: Corporate DNS not resolving external hostnames
4. **Rate Limiting**: Vulnerability database rate limiting requests

### Diagnostic Commands

```bash
# Test basic connectivity
curl -v https://vuln.go.dev

# Test with current proxy settings
curl -v --proxy $HTTP_PROXY https://vuln.go.dev

# DNS resolution test
nslookup vuln.go.dev

# Network timing test
time curl -s --head https://vuln.go.dev
```

### Configuration Adjustments

```yaml
# For slow networks
timeout_seconds: 900  # 15 minutes
retry_attempts: 5

# For intermittent connectivity  
timeout_seconds: 600  # 10 minutes
retry_attempts: 3
```

## Performance Impact

### Overhead Analysis

- **Normal Conditions**: <100ms overhead
- **Network Errors**: 5-15 seconds per retry attempt
- **Maximum Impact**: ~45 seconds with 3 retries and 5s delays

### Optimization Strategies

1. **Early Detection**: Quick connectivity checks before full scans
2. **Adaptive Timeouts**: Shorter timeouts for known fast environments
3. **Caching**: Future enhancement for vulnerability database caching
4. **Parallel Validation**: Concurrent network validation during scan preparation

## Future Enhancements

### Planned Improvements

1. **Database Caching**: Local caching of vulnerability database
2. **Network Quality Detection**: Automatic timeout adjustment based on detected quality
3. **Alternative Endpoints**: Fallback to mirror vulnerability databases
4. **Circuit Breaker**: Automatic bypass for persistent network issues

### Monitoring Integration

1. **Network Performance Metrics**: Track scan duration and retry rates
2. **Alerting**: Automatic alerts for persistent network issues
3. **Dashboard**: Real-time network health for vulnerability scanning
4. **Trend Analysis**: Historical network performance tracking

## Validation Results

Network resilience testing demonstrates robust handling of various failure scenarios:

✅ **Complete Network Isolation**: Proper retry behavior with clear error messages  
✅ **DNS Failures**: Appropriate error classification and retry logic  
✅ **Timeout Handling**: Clean termination within configured timeouts  
✅ **Graceful Degradation**: Continued operation under degraded conditions  
✅ **Error Classification**: Accurate distinction between network and scan errors  
✅ **Structured Logging**: Complete audit trail with correlation IDs  

The network resilience system ensures that security scanning remains reliable even under adverse network conditions while providing clear guidance for resolution.
