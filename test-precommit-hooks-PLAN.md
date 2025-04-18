# T010: Test Pre-commit Hooks with Sample Changes

## Overview

This plan outlines the approach for testing all pre-commit hooks configured in the Glance project to ensure they function as expected. We'll create a suite of test files designed to trigger specific hooks, execute the hooks, and document the results.

## Implementation Approach

We'll create a structured testing approach using temporary files that intentionally violate the rules enforced by each hook. By committing these files, we can verify that the hooks correctly identify and address issues.

## Testing Matrix

| Hook Type | Hook ID | Test Case Description |
|-----------|---------|------------------------|
| Go Formatting | go-fmt | Create Go file with improper formatting |
| Go Formatting | go-imports | Create Go file with unorganized imports |
| Go Analysis | go-vet | Create Go file with suspicious code patterns |
| Go Linting | golangci-lint | Create Go file with linting issues |
| Go Testing | go-unit-tests | Create a failing test file |
| Go Testing | go-mod-tidy | Modify go.mod file with inconsistencies |
| Go Testing | go-build | Create Go file with build errors |
| File Formatting | trailing-whitespace | Create file with trailing whitespace |
| File Formatting | end-of-file-fixer | Create file without final newline |
| File Formatting | mixed-line-ending | Create file with mixed line endings |
| File Formatting | check-yaml | Create invalid YAML file |
| File Formatting | check-json | Create invalid JSON file |
| File Size | check-added-large-files | Create large file exceeding limits |
| File Size | check-case-conflict | Create files with case conflicts |
| Security | detect-secrets | Create file with embedded API key |
| Security | detect-private-key | Create file with private key content |
| Security | no-commit-to-branch | Attempt commit to protected branch |

## Detailed Steps

### 1. Preparation
1. Create a dedicated testing directory `precommit-tests/` to contain all test files
2. Document each test case with a clear description and expected outcome

### 2. Create Test Files
1. For each hook, create a specific test file that should trigger the hook
2. Ensure the violations are clear and minimal to isolate the hook's behavior
3. Add comments in each file explaining the intentional issue

### 3. Test Execution
1. Run the pre-commit hooks against all test files using `pre-commit run --files <test_files>`
2. Document each hook's response to the test cases
3. Verify that hooks catch the expected issues and produce the expected output

### 4. Documentation
1. Create documentation with test results for each hook
2. Include examples of the violations and the hook's response
3. Document any issues or unexpected behaviors found during testing

### 5. Refinement
1. Fix any hooks that don't behave as expected
2. Adjust hook configurations if needed based on test results
3. Re-run tests to verify fixes and document updated results

## Success Criteria

The implementation will be considered successful when:
1. All pre-commit hooks detect the intentional violations in the test files
2. The hook responses are documented and match the expected behavior
3. Any issues discovered during testing are addressed

## Implementation Notes

- We'll use a script to automate running the hooks on all test files
- All test files will be contained in a separate directory to avoid accidental commits
- Test files will be prefixed with `test_` to make their purpose clear
- Each test will focus on a single hook to ensure clear results