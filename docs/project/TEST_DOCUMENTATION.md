# Test Documentation

This document describes what tests verify in the simplified vulnerability scanning system.

## Test Categories

### 1. Integration Tests (`vulnerability_integration_test.go`)

**Purpose**: Verify end-to-end functionality of the simplified vulnerability scanning system.

**Tests**:
- `TestVulnerabilityDetectionIntegration`: Core vulnerability detection capability
  - Verifies govulncheck execution on clean and vulnerable projects
  - Validates scan duration stays under 30 seconds
  - Confirms expected output formats and content

- `TestConfigurationFileSupport`: Configuration file handling
  - Verifies `.govulncheck.yaml` files exist and are readable
  - Validates configuration contains required severity settings

- `TestScanOutputFormats`: Output format support
  - Tests default text format output
  - Tests JSON format output structure

- `TestScanPerformanceAndReliability`: Performance validation
  - Verifies scans complete within 60-second target (from requirements)
  - Tests consistency across multiple iterations
  - Measures average scan duration

- `TestIsolationAndConcurrency`: Concurrency and isolation
  - Verifies multiple scans can run simultaneously without interference
  - Tests thread safety of govulncheck execution

### 2. Network Failure Tests (`network_failure_test.go`)

**Purpose**: Verify behavior under adverse network conditions.

**Tests**:
- `TestNetworkFailureScenarios`: Network isolation scenarios
- `TestTimeoutHandling`: Timeout behavior with context cancellation
- `TestErrorMessaging`: Error message validation for network issues

### 3. Core Functionality Tests

**Purpose**: Verify main application functionality remains intact.

**Coverage**:
- File system operations and path validation
- LLM integration and mocking
- Configuration loading and template processing
- Git ignore pattern handling
- Documentation generation workflows

## Test Data

### Test Projects
- `testdata/clean-project/`: Go project with up-to-date dependencies
- `testdata/vulnerable-project/`: Go project with known vulnerable dependencies

### Configuration Files
- Each test project contains `.govulncheck.yaml` with:
  - `fail_on_severity: ["HIGH", "CRITICAL"]`
  - Standard timeout and scanning configurations

## Performance Targets

- **Scan Duration**: < 60 seconds (requirement)
- **Test Execution**: Current baseline ~4-5 seconds for integration tests
- **Concurrency**: Multiple simultaneous scans supported

## Test Isolation

- Each test changes to appropriate directory before running govulncheck
- Context timeouts prevent hanging tests
- Network tests are skipped unless `RUN_NETWORK_TESTS=true`
- Performance tests are skipped in short mode (`testing.Short()`)

## Coverage Verification

The test suite validates:
✅ Core vulnerability detection works
✅ Configuration file support
✅ Output format handling
✅ Performance within targets
✅ Network failure resilience
✅ Concurrent execution safety
✅ Timeout handling
✅ Error messaging clarity

## Running Tests

```bash
# All tests
go test ./...

# Integration tests only
go test -run TestVulnerability

# With race detection
go test -race ./...

# With network tests
RUN_NETWORK_TESTS=true go test ./...

# Performance tests
go test -run TestScanPerformance
```
