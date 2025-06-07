#!/bin/bash

# govulncheck-with-retry.sh
# Wrapper script for govulncheck with retry logic and network failure handling
#
# Usage: govulncheck-with-retry.sh [govulncheck arguments]
#
# Environment Variables:
#   GOVULNCHECK_CONFIG_PATH - Path to .govulncheck.yaml config file [default: .govulncheck.yaml]
#   GOVULNCHECK_RETRY_ATTEMPTS - Number of retry attempts [default: from config]
#   GOVULNCHECK_TIMEOUT_SECONDS - Timeout in seconds [default: from config]
#   GOVULNCHECK_RETRY_DELAY - Delay between retries in seconds [default: 5]

set -euo pipefail

# Script configuration
SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]}")"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
CONFIG_PATH="${GOVULNCHECK_CONFIG_PATH:-${PROJECT_ROOT}/.govulncheck.yaml}"
RETRY_DELAY="${GOVULNCHECK_RETRY_DELAY:-5}"

# Load configuration parsing functions
if [ -f "${PROJECT_ROOT}/config/vulnerability.go" ]; then
    # Use Go-based config parsing if available
    CONFIG_LOADER="go"
else
    # Fallback to yq-based parsing
    CONFIG_LOADER="yq"
fi

# Logging function with correlation ID support
log_with_correlation() {
    local level="$1"
    local message="$2"
    local correlation_id="${CORRELATION_ID:-govulncheck-retry-$(date +%Y%m%d-%H%M%S)}"
    local timestamp=$(date -u '+%Y-%m-%dT%H:%M:%SZ')

    # Structured JSON logging
    local log_entry="{\"timestamp\": \"$timestamp\", \"level\": \"$level\", \"service_name\": \"govulncheck-wrapper\", \"correlation_id\": \"$correlation_id\", \"message\": \"$message\"}"
    echo "$log_entry" >&2
}

# Parse configuration file
parse_config() {
    local config_file="$1"

    if [ ! -f "$config_file" ]; then
        log_with_correlation "WARN" "Configuration file $config_file not found, using defaults"
        echo "retry_attempts=2"
        echo "timeout_seconds=300"
        return
    fi

    if command -v yq >/dev/null 2>&1; then
        local retry_attempts=$(yq eval '.retry_attempts // 2' "$config_file" 2>/dev/null || echo "2")
        local timeout_seconds=$(yq eval '.timeout_seconds // 300' "$config_file" 2>/dev/null || echo "300")
        echo "retry_attempts=$retry_attempts"
        echo "timeout_seconds=$timeout_seconds"
    else
        log_with_correlation "WARN" "yq not available, using default configuration values"
        echo "retry_attempts=2"
        echo "timeout_seconds=300"
    fi
}

# Check network connectivity to vulnerability database
check_network_connectivity() {
    local vuln_db_url="https://vuln.go.dev"
    local timeout=10

    log_with_correlation "INFO" "Checking connectivity to vulnerability database"

    if command -v curl >/dev/null 2>&1; then
        if curl -s --max-time "$timeout" --head "$vuln_db_url" >/dev/null 2>&1; then
            log_with_correlation "INFO" "Vulnerability database is reachable"
            return 0
        else
            log_with_correlation "WARN" "Unable to reach vulnerability database at $vuln_db_url"
            return 1
        fi
    elif command -v wget >/dev/null 2>&1; then
        if wget -q --timeout="$timeout" --spider "$vuln_db_url" >/dev/null 2>&1; then
            log_with_correlation "INFO" "Vulnerability database is reachable"
            return 0
        else
            log_with_correlation "WARN" "Unable to reach vulnerability database at $vuln_db_url"
            return 1
        fi
    else
        log_with_correlation "WARN" "Neither curl nor wget available, skipping connectivity check"
        return 0  # Assume connectivity is OK if we can't check
    fi
}

# Determine if error is network-related
is_network_error() {
    local exit_code="$1"
    local output="$2"

    # Common network-related indicators
    local network_indicators=(
        "timeout"
        "network"
        "connection"
        "dns"
        "resolve"
        "unreachable"
        "connection refused"
        "no route to host"
        "temporary failure in name resolution"
    )

    # Check exit code (124 = timeout)
    if [ "$exit_code" -eq 124 ]; then
        return 0
    fi

    # Check output for network-related errors
    local output_lower=$(echo "$output" | tr '[:upper:]' '[:lower:]')
    for indicator in "${network_indicators[@]}"; do
        if echo "$output_lower" | grep -q "$indicator"; then
            return 0
        fi
    done

    return 1
}

# Run govulncheck with timeout and capture output
run_govulncheck_attempt() {
    local timeout_seconds="$1"
    local attempt_number="$2"
    shift 2
    local govulncheck_args=("$@")

    log_with_correlation "INFO" "Running govulncheck attempt $attempt_number with ${timeout_seconds}s timeout"

    local temp_output=$(mktemp)
    local start_time=$(date +%s)

    # Run govulncheck with timeout
    local exit_code=0
    timeout "${timeout_seconds}s" govulncheck "${govulncheck_args[@]}" > "$temp_output" 2>&1 || exit_code=$?

    local end_time=$(date +%s)
    local duration_ms=$(((end_time - start_time) * 1000))
    local output=$(cat "$temp_output")

    # Clean up temp file
    rm -f "$temp_output"

    log_with_correlation "INFO" "Govulncheck attempt $attempt_number completed in ${duration_ms}ms with exit code $exit_code"

    # Return results via global variables (bash limitation)
    ATTEMPT_EXIT_CODE=$exit_code
    ATTEMPT_OUTPUT="$output"
    ATTEMPT_DURATION_MS=$duration_ms

    return $exit_code
}

# Main retry logic
run_with_retry() {
    local config_values
    config_values=$(parse_config "$CONFIG_PATH")

    # Extract configuration values
    local retry_attempts=$(echo "$config_values" | grep retry_attempts | cut -d'=' -f2)
    local timeout_seconds=$(echo "$config_values" | grep timeout_seconds | cut -d'=' -f2)

    # Override with environment variables if provided
    retry_attempts="${GOVULNCHECK_RETRY_ATTEMPTS:-$retry_attempts}"
    timeout_seconds="${GOVULNCHECK_TIMEOUT_SECONDS:-$timeout_seconds}"

    log_with_correlation "INFO" "Starting govulncheck with retry configuration: attempts=$retry_attempts, timeout=${timeout_seconds}s"

    # Check network connectivity before starting
    if ! check_network_connectivity; then
        log_with_correlation "WARN" "Network connectivity check failed, but proceeding with scan"
    fi

    local max_attempts=$((retry_attempts + 1))  # retry_attempts + initial attempt
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        log_with_correlation "INFO" "Attempt $attempt of $max_attempts"

        # Run govulncheck attempt
        if run_govulncheck_attempt "$timeout_seconds" "$attempt" "$@"; then
            # Success
            log_with_correlation "INFO" "Govulncheck completed successfully on attempt $attempt"
            echo "$ATTEMPT_OUTPUT"
            return 0
        else
            local exit_code=$ATTEMPT_EXIT_CODE
            local output="$ATTEMPT_OUTPUT"
            local duration_ms=$ATTEMPT_DURATION_MS

            log_with_correlation "ERROR" "Govulncheck attempt $attempt failed with exit code $exit_code"

            # Check if this is the last attempt
            if [ $attempt -eq $max_attempts ]; then
                log_with_correlation "ERROR" "All retry attempts exhausted, govulncheck failed"

                # Provide detailed error analysis
                if [ $exit_code -eq 124 ]; then
                    log_with_correlation "ERROR" "Scan timed out after ${timeout_seconds} seconds"
                    echo "❌ FAILURE: Vulnerability scan timed out after ${timeout_seconds} seconds" >&2
                    echo "This may indicate network issues or an unusually large codebase." >&2
                    echo "Consider increasing timeout_seconds in .govulncheck.yaml or checking network connectivity." >&2
                elif is_network_error "$exit_code" "$output"; then
                    log_with_correlation "ERROR" "Network-related error detected"
                    echo "❌ FAILURE: Network error during vulnerability scan" >&2
                    echo "Please check your network connectivity and try again." >&2
                    echo "If the problem persists, the vulnerability database may be temporarily unavailable." >&2
                else
                    log_with_correlation "ERROR" "Non-network error detected"
                    echo "❌ FAILURE: Govulncheck failed with exit code $exit_code" >&2
                fi

                # Output the captured output for debugging
                echo "$output"
                return $exit_code
            fi

            # Check if we should retry (only for network-related errors)
            if is_network_error "$exit_code" "$output"; then
                log_with_correlation "WARN" "Network error detected, will retry after ${RETRY_DELAY}s delay"
                echo "⚠️  Network error detected (attempt $attempt/$max_attempts), retrying in ${RETRY_DELAY}s..." >&2
                sleep "$RETRY_DELAY"
            else
                log_with_correlation "ERROR" "Non-network error detected, not retrying"
                echo "❌ FAILURE: Non-network error, not retrying" >&2
                echo "$output"
                return $exit_code
            fi
        fi

        attempt=$((attempt + 1))
    done

    # Should not reach here
    log_with_correlation "ERROR" "Unexpected end of retry loop"
    return 1
}

# Validate prerequisites
validate_prerequisites() {
    if ! command -v govulncheck >/dev/null 2>&1; then
        log_with_correlation "ERROR" "govulncheck is not installed or not in PATH"
        echo "❌ ERROR: govulncheck is required but not found in PATH" >&2
        echo "Install with: go install golang.org/x/vuln/cmd/govulncheck@latest" >&2
        exit 1
    fi

    if ! command -v timeout >/dev/null 2>&1; then
        log_with_correlation "ERROR" "timeout command is required but not found"
        echo "❌ ERROR: timeout command is required but not found" >&2
        exit 1
    fi

    log_with_correlation "INFO" "Prerequisites validated successfully"
}

# Handle script arguments
show_help() {
    cat <<EOF
Usage: $0 [options] [-- govulncheck arguments]

Options:
  -h, --help              Show this help message
  -c, --config PATH       Path to .govulncheck.yaml config file
  -r, --retry-attempts N  Number of retry attempts (overrides config)
  -t, --timeout N         Timeout in seconds (overrides config)
  -d, --retry-delay N     Delay between retries in seconds (default: 5)
  -v, --verbose           Enable verbose logging

Environment Variables:
  GOVULNCHECK_CONFIG_PATH      Configuration file path
  GOVULNCHECK_RETRY_ATTEMPTS   Number of retry attempts
  GOVULNCHECK_TIMEOUT_SECONDS  Timeout in seconds
  GOVULNCHECK_RETRY_DELAY      Delay between retries

Examples:
  $0 ./...                           # Scan current module with default config
  $0 -r 5 -t 600 ./...              # Scan with 5 retries and 10min timeout
  $0 -c custom.yaml ./...           # Use custom configuration file
  $0 -- -format json ./...          # Pass -format json to govulncheck

Exit Codes:
  0   - Scan completed successfully
  1   - Scan failed (non-network error)
  124 - Scan timed out
  125 - Network connectivity issues
  126 - Configuration error
  127 - Prerequisites not met
EOF
}

# Parse command line arguments
parse_args() {
    PARSED_GOVULNCHECK_ARGS=()

    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -c|--config)
                CONFIG_PATH="$2"
                shift 2
                ;;
            -r|--retry-attempts)
                export GOVULNCHECK_RETRY_ATTEMPTS="$2"
                shift 2
                ;;
            -t|--timeout)
                export GOVULNCHECK_TIMEOUT_SECONDS="$2"
                shift 2
                ;;
            -d|--retry-delay)
                RETRY_DELAY="$2"
                shift 2
                ;;
            -v|--verbose)
                export GOVULNCHECK_VERBOSE="true"
                shift
                ;;
            --)
                shift
                PARSED_GOVULNCHECK_ARGS+=("$@")
                break
                ;;
            *)
                PARSED_GOVULNCHECK_ARGS+=("$1")
                shift
                ;;
        esac
    done

    # If no arguments provided, default to scanning current module
    if [ ${#PARSED_GOVULNCHECK_ARGS[@]} -eq 0 ]; then
        PARSED_GOVULNCHECK_ARGS=("./...")
    fi
}

# Main execution
main() {
    local correlation_id="govulncheck-retry-$(date +%Y%m%d-%H%M%S)-$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
    export CORRELATION_ID="$correlation_id"

    log_with_correlation "INFO" "Starting govulncheck wrapper with correlation ID: $correlation_id"

    # Validate prerequisites
    validate_prerequisites

    # Parse arguments
    parse_args "$@"
    local govulncheck_args=("${PARSED_GOVULNCHECK_ARGS[@]}")

    log_with_correlation "INFO" "Govulncheck arguments: ${govulncheck_args[*]}"

    # Run with retry logic
    run_with_retry "${govulncheck_args[@]}"
}

# Run main function
main "$@"
