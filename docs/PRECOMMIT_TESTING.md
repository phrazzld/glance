# Pre-commit Hook Testing Guide

## Overview

The Glance project uses pre-commit hooks to enforce code quality standards, formatting rules, and security best practices. This document explains how to test these hooks to ensure they are functioning correctly.

## Hook Testing Process

We've developed a comprehensive testing approach to ensure all hooks are functioning as expected:

1. **Test Files**: The `precommit-tests/` directory contains test files specifically designed to trigger each hook.
2. **Testing Script**: The `precommit-tests/run_tests.sh` script automates running each hook against its corresponding test file.
3. **Results Documentation**: Test results are recorded in `precommit-tests/hook_test_results.md`.

## Hooks Being Tested

The testing suite verifies the following hook categories:

- **Go Formatting Hooks**: Ensure proper formatting and import organization in Go code.
- **Go Code Analysis Hooks**: Detect suspicious code patterns and potential bugs.
- **Go Linting Hooks**: Enforce style guidelines and catch common mistakes.
- **Go Testing Hooks**: Verify tests pass, code builds, and dependencies are tidy.
- **File Formatting Hooks**: Enforce consistent file formatting across all files.
- **File Size Limitation Hooks**: Prevent excessively large files and case conflicts.
- **Security Hooks**: Detect secrets, credentials, and private keys.

## Running the Tests

To run all hook tests:

```bash
cd /path/to/glance
./precommit-tests/run_tests.sh
```

This will execute each test and save the results to `precommit-tests/hook_test_results.md`.

## Interpreting Results

For each hook, the test results will show:

- Whether the hook identified the intentional issue
- The hook's output and return code
- Whether the hook performed any automatic fixes

If a hook fails to detect an issue or behaves unexpectedly, check the hook configuration in `.pre-commit-config.yaml`.

## Extending the Tests

To add tests for new hooks:

1. Create a test file in `precommit-tests/` that specifically triggers the hook
2. Add the test to `run_tests.sh`
3. Document the test in `precommit-tests/README.md`

## Troubleshooting

If a hook is not working as expected:

1. Verify the hook is correctly configured in `.pre-commit-config.yaml`
2. Check that the hook is installed and available (`pre-commit hook-list`)
3. Try running the hook manually with `pre-commit run <hook-id> --files <file>`
4. Check if there are any specific requirements or dependencies for the hook