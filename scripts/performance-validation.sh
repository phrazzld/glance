#!/bin/bash

# performance-validation.sh
# Runs vulnerability scanning performance validation tests and generates reports
#
# Usage: performance-validation.sh [options]
#
# Environment Variables:
#   PERFORMANCE_TEST_MODE - Type of tests to run (quick|full|benchmark) [default: quick]
#   PERFORMANCE_THRESHOLD_SECONDS - Maximum acceptable scan duration [default: 60]
#   PERFORMANCE_WARNING_SECONDS - Warning threshold for scan duration [default: 45]
#   PERFORMANCE_REPORTS_DIR - Directory for performance reports [default: performance-reports]
#   CI_ENVIRONMENT - Running in CI environment [default: false]

set -euo pipefail

# Configuration with defaults
PERFORMANCE_TEST_MODE="${PERFORMANCE_TEST_MODE:-quick}"
PERFORMANCE_THRESHOLD_SECONDS="${PERFORMANCE_THRESHOLD_SECONDS:-60}"
PERFORMANCE_WARNING_SECONDS="${PERFORMANCE_WARNING_SECONDS:-45}"
PERFORMANCE_REPORTS_DIR="${PERFORMANCE_REPORTS_DIR:-performance-reports}"
CI_ENVIRONMENT="${CI_ENVIRONMENT:-false}"

# Script configuration
SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]}")"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)

# Logging function
log() {
    local level="$1"
    local message="$2"
    echo "[$(date -u '+%Y-%m-%dT%H:%M:%SZ')] [$level] [performance-validation] $message" >&2
}

# Check prerequisites
check_prerequisites() {
    log "INFO" "Checking performance validation prerequisites..."

    # Check if govulncheck is available
    if ! command -v govulncheck >/dev/null 2>&1; then
        log "ERROR" "govulncheck is required but not installed"
        log "INFO" "Install with: go install golang.org/x/vuln/cmd/govulncheck@latest"
        exit 1
    fi

    # Check if we're in the right directory
    if [ ! -f "$PROJECT_ROOT/go.mod" ]; then
        log "ERROR" "Must be run from project root directory"
        exit 1
    fi

    # Check test data exists
    if [ ! -d "$PROJECT_ROOT/testdata/clean-project" ] || [ ! -d "$PROJECT_ROOT/testdata/vulnerable-project" ]; then
        log "ERROR" "Test data directories not found"
        exit 1
    fi

    log "INFO" "Prerequisites check passed"
}

# Create reports directory
setup_reports_directory() {
    mkdir -p "$PERFORMANCE_REPORTS_DIR"
    log "INFO" "Performance reports will be saved to: $PERFORMANCE_REPORTS_DIR"
}

# Run performance tests based on mode
run_performance_tests() {
    local test_mode="$1"
    log "INFO" "Running performance tests in $test_mode mode..."

    cd "$PROJECT_ROOT"

    case "$test_mode" in
        "quick")
            log "INFO" "Running quick performance validation..."
            go test -v -run=TestVulnerabilityScanPerformance -timeout=300s .
            ;;
        "full")
            log "INFO" "Running full performance test suite..."
            go test -v -run=TestVulnerabilityScanPerformance -timeout=300s .
            go test -v -run=TestParallelScanPerformance -timeout=600s .
            ;;
        "benchmark")
            log "INFO" "Running performance benchmarks..."
            go test -bench=BenchmarkVulnerabilityScan -run=^$ -timeout=600s .
            ;;
        *)
            log "ERROR" "Unknown test mode: $test_mode"
            exit 1
            ;;
    esac
}

# Analyze performance reports
analyze_performance_reports() {
    log "INFO" "Analyzing performance reports..."

    # Find the most recent performance report
    local latest_report=$(find . -name "performance-report-*.json" -type f -exec ls -t {} + 2>/dev/null | head -1)

    if [ -z "$latest_report" ] || [ ! -f "$latest_report" ]; then
        log "WARN" "No performance report found"
        return 0
    fi

    log "INFO" "Analyzing report: $latest_report"

    # Extract key metrics using jq if available
    if command -v jq >/dev/null 2>&1; then
        local avg_duration=$(jq -r '.summary.avg_duration' "$latest_report" 2>/dev/null || echo "unknown")
        local max_duration=$(jq -r '.summary.max_duration' "$latest_report" 2>/dev/null || echo "unknown")
        local total_scans=$(jq -r '.summary.total_scans' "$latest_report" 2>/dev/null || echo "unknown")

        log "INFO" "Performance Summary:"
        log "INFO" "  Total scans: $total_scans"
        log "INFO" "  Average duration: $avg_duration"
        log "INFO" "  Maximum duration: $max_duration"
        log "INFO" "  Threshold: ${PERFORMANCE_THRESHOLD_SECONDS}s"

        # Check if performance meets requirements
        if [ "$max_duration" != "unknown" ]; then
            # Convert Go duration to seconds for comparison
            local max_seconds=$(echo "$max_duration" | awk '
            {
                # Handle Go duration format (e.g., "3.289834958s", "2m30.5s", "1h2m3s")
                if (/^[0-9]+(\.[0-9]+)?s$/) {
                    # Simple seconds format: extract number before 's'
                    gsub(/s$/, "")
                    print int($0)
                } else if (/^[0-9]+(\.[0-9]+)?ms$/) {
                    # Milliseconds format: convert to seconds
                    gsub(/ms$/, "")
                    print int($0 / 1000)
                } else if (/^[0-9]+(\.[0-9]+)?ns$/) {
                    # Nanoseconds format: likely the raw value, convert to seconds
                    gsub(/ns$/, "")
                    print int($0 / 1000000000)
                } else if (/^[0-9]+$/) {
                    # Raw number, assume nanoseconds
                    print int($0 / 1000000000)
                } else {
                    # Default: try to extract first number
                    match($0, /[0-9]+(\.[0-9]+)?/)
                    if (RSTART > 0) {
                        num = substr($0, RSTART, RLENGTH)
                        print int(num)
                    } else {
                        print 0
                    }
                }
            }')

            if [ -n "$max_seconds" ] && [ "$max_seconds" -gt "$PERFORMANCE_THRESHOLD_SECONDS" ]; then
                log "ERROR" "Performance threshold exceeded: ${max_seconds}s > ${PERFORMANCE_THRESHOLD_SECONDS}s"
                generate_failure_report "$max_seconds"
                exit 1
            elif [ -n "$max_seconds" ] && [ "$max_seconds" -gt "$PERFORMANCE_WARNING_SECONDS" ]; then
                log "WARN" "Performance approaching threshold: ${max_seconds}s (warning at ${PERFORMANCE_WARNING_SECONDS}s)"
            else
                log "INFO" "Performance validation passed: ${max_seconds}s < ${PERFORMANCE_THRESHOLD_SECONDS}s"
            fi
        fi
    else
        log "WARN" "jq not available, skipping detailed analysis"
    fi

    # Move report to reports directory
    local report_name="performance-report-${TIMESTAMP}.json"
    cp "$latest_report" "$PERFORMANCE_REPORTS_DIR/$report_name"
    rm -f "$latest_report"  # Clean up original

    log "INFO" "Performance report saved: $PERFORMANCE_REPORTS_DIR/$report_name"
}

# Generate failure report for CI
generate_failure_report() {
    local max_duration="$1"

    cat > "$PERFORMANCE_REPORTS_DIR/performance-failure-${TIMESTAMP}.json" <<EOF
{
  "timestamp": "$(date -u '+%Y-%m-%dT%H:%M:%SZ')",
  "status": "FAILED",
  "reason": "Performance threshold exceeded",
  "details": {
    "max_duration_seconds": $max_duration,
    "threshold_seconds": $PERFORMANCE_THRESHOLD_SECONDS,
    "excess_seconds": $((max_duration - PERFORMANCE_THRESHOLD_SECONDS))
  },
  "recommendations": [
    "Investigate scan performance bottlenecks",
    "Consider optimizing dependency scanning",
    "Review vulnerability database caching options",
    "Check network connectivity and latency"
  ]
}
EOF

    log "ERROR" "Performance failure report generated"
}

# Generate GitHub Actions summary if in CI
generate_github_summary() {
    if [ "$CI_ENVIRONMENT" != "true" ] || [ -z "${GITHUB_STEP_SUMMARY:-}" ]; then
        return 0
    fi

    log "INFO" "Generating GitHub Actions summary..."

    cat >> "$GITHUB_STEP_SUMMARY" <<EOF
## 🚀 Vulnerability Scan Performance Validation

**Test Mode:** \`$PERFORMANCE_TEST_MODE\`
**Timestamp:** $(date -u '+%Y-%m-%d %H:%M:%S UTC')

### Performance Thresholds
- **Maximum Duration:** ${PERFORMANCE_THRESHOLD_SECONDS}s
- **Warning Threshold:** ${PERFORMANCE_WARNING_SECONDS}s

EOF

    # Add performance results if available
    local latest_report=$(find "$PERFORMANCE_REPORTS_DIR" -name "performance-report-*.json" -type f -exec ls -t {} + 2>/dev/null | head -1)

    if [ -n "$latest_report" ] && [ -f "$latest_report" ] && command -v jq >/dev/null 2>&1; then
        local avg_duration=$(jq -r '.summary.avg_duration' "$latest_report" 2>/dev/null || echo "unknown")
        local max_duration=$(jq -r '.summary.max_duration' "$latest_report" 2>/dev/null || echo "unknown")
        local total_scans=$(jq -r '.summary.total_scans' "$latest_report" 2>/dev/null || echo "unknown")

        cat >> "$GITHUB_STEP_SUMMARY" <<EOF
### 📊 Performance Results
- **Total Scans:** $total_scans
- **Average Duration:** $avg_duration
- **Maximum Duration:** $max_duration
- **Status:** ✅ Within thresholds

### 📁 Artifacts
Performance reports are available in the workflow artifacts.

EOF
    fi

    cat >> "$GITHUB_STEP_SUMMARY" <<EOF
---
*Performance validation completed at $(date -u '+%Y-%m-%d %H:%M:%S UTC')*
EOF
}

# Clean up old reports
cleanup_old_reports() {
    if [ -d "$PERFORMANCE_REPORTS_DIR" ]; then
        log "INFO" "Cleaning up reports older than 30 days..."
        find "$PERFORMANCE_REPORTS_DIR" -name "performance-report-*.json" -type f -mtime +30 -delete 2>/dev/null || true
        find "$PERFORMANCE_REPORTS_DIR" -name "performance-failure-*.json" -type f -mtime +7 -delete 2>/dev/null || true
    fi
}

# Main execution
main() {
    log "INFO" "Starting vulnerability scan performance validation"
    log "INFO" "Mode: $PERFORMANCE_TEST_MODE"
    log "INFO" "Threshold: ${PERFORMANCE_THRESHOLD_SECONDS}s"
    log "INFO" "Warning: ${PERFORMANCE_WARNING_SECONDS}s"

    check_prerequisites
    setup_reports_directory
    cleanup_old_reports

    # Run performance tests
    run_performance_tests "$PERFORMANCE_TEST_MODE"

    # Analyze results
    analyze_performance_reports

    # Generate CI summary if applicable
    generate_github_summary

    log "INFO" "Performance validation completed successfully"
}

# Handle script arguments
case "${1:-}" in
    "quick"|"full"|"benchmark")
        PERFORMANCE_TEST_MODE="$1"
        ;;
    "help"|"-h"|"--help")
        echo "Usage: $0 [quick|full|benchmark|help]"
        echo ""
        echo "Modes:"
        echo "  quick     - Run basic performance validation (default)"
        echo "  full      - Run comprehensive performance tests including parallel execution"
        echo "  benchmark - Run performance benchmarks for detailed analysis"
        echo ""
        echo "Environment Variables:"
        echo "  PERFORMANCE_TEST_MODE - Test mode override"
        echo "  PERFORMANCE_THRESHOLD_SECONDS - Maximum scan duration (default: 60)"
        echo "  PERFORMANCE_WARNING_SECONDS - Warning threshold (default: 45)"
        echo "  PERFORMANCE_REPORTS_DIR - Reports directory (default: performance-reports)"
        echo "  CI_ENVIRONMENT - Set to 'true' in CI (default: false)"
        exit 0
        ;;
    "")
        # Use default mode
        ;;
    *)
        log "ERROR" "Unknown argument: $1"
        echo "Use '$0 help' for usage information"
        exit 1
        ;;
esac

# Run main function
main "$@"
