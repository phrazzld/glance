# Pre-commit Hook Testing

This directory contains test files and scripts for verifying the functionality of pre-commit hooks configured in the Glance project.

## Purpose

The purpose of these tests is to ensure that all pre-commit hooks correctly identify and address issues in the codebase. Each test file is designed to trigger a specific hook by intentionally violating the rule that the hook enforces.

## Test Files

| File | Purpose | Hook Tested |
|------|---------|-------------|
| `test_go_fmt.go` | Contains improperly formatted Go code | go-fmt |
| `test_go_imports.go` | Contains unorganized imports | go-imports |
| `test_go_vet.go` | Contains suspicious code patterns | go-vet |
| `test_golangci_lint.go` | Contains linting issues | golangci-lint |
| `test_go_unit_test.go` | Contains a failing test | go-unit-tests |
| `test_go_build_error.go` | Contains build errors | go-build |
| `test_trailing_whitespace.txt` | Contains trailing whitespace | trailing-whitespace |
| `test_end_of_file.txt` | Missing final newline | end-of-file-fixer |
| `test_mixed_line_ending.txt` | Contains mixed line endings | mixed-line-ending |
| `test_invalid_yaml.yaml` | Contains invalid YAML | check-yaml |
| `test_invalid_json.json` | Contains invalid JSON | check-json |
| `test_merge_conflict.txt` | Contains merge conflict markers | check-merge-conflict |
| `test_secrets.txt` | Contains fake API keys | detect-secrets |
| `test_private_key.txt` | Contains a fake private key | detect-private-key |
| `CaseConflict.txt` / `caseconflict.txt` | Files with conflicting names | check-case-conflict |
| `generate_large_file.sh` | Script to generate a large file | check-added-large-files |

## Running the Tests

To run all tests:

```bash
./precommit-tests/run_tests.sh
```

This script will:
1. Run each test file against its corresponding hook
2. Document the results in `hook_test_results.md`
3. Report any hooks that don't function as expected

## Expected Results

Each hook should detect the issue in its corresponding test file and either:
1. Fix the issue automatically (for fixable issues like formatting)
2. Report the issue and prevent the commit (for non-fixable issues)

## Notes

- These test files are deliberately problematic and should not be committed to the repository.
- The test scripts use temporary files or isolate the tests to avoid affecting the actual codebase.
- Test files are prefixed with `test_` to make their purpose clear.