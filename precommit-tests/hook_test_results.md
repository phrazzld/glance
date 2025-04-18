# Pre-commit Hook Test Results

This document contains the results of testing each pre-commit hook with specifically crafted test files.

Tests run on: Fri Apr 18 22:08:18 JST 2025

## Testing: go-fmt

**File:** test_go_fmt.go

**Description:** Tests go-fmt hook with improperly formatted Go code

**Results:**
```
go fmt...............................................(no files to check)Skipped
```

## Testing: go-imports

**File:** test_go_imports.go

**Description:** Tests go-imports hook with unorganized imports

**Results:**
```
go imports...........................................(no files to check)Skipped
```

## Testing: go-vet

**File:** test_go_vet.go

**Description:** Tests go-vet hook with suspicious code patterns

**Results:**
```
go vet...............................................(no files to check)Skipped
```

## Testing: golangci-lint

**File:** test_golangci_lint.go

**Description:** Tests golangci-lint hook with linting issues

**Results:**
```
golangci-lint........................................(no files to check)Skipped
```

## Testing: go-unit-tests

**File:** test_go_unit_test.go

**Description:** Tests go-unit-tests hook with a failing test

**Results:**
```
go test..............................................(no files to check)Skipped
```

## Testing: go-build

**File:** test_go_build_error.go

**Description:** Tests go-build hook with build errors

**Results:**
```
go build.............................................(no files to check)Skipped
```

## Testing: trailing-whitespace

**File:** test_trailing_whitespace.txt

**Description:** Tests trailing-whitespace hook with trailing spaces

**Results:**
```
Remove trailing whitespace...............................................Failed
- hook id: trailing-whitespace
- files were modified by this hook
```

## Testing: end-of-file-fixer

**File:** test_end_of_file.txt

**Description:** Tests end-of-file-fixer hook with missing final newline

**Results:**
```
Fix end of files.........................................................Failed
- hook id: end-of-file-fixer
- files were modified by this hook
```

## Testing: mixed-line-ending

**File:** test_mixed_line_ending.txt

**Description:** Tests mixed-line-ending hook with mixed line endings

**Results:**
```
Normalize line endings...................................................Failed
- hook id: mixed-line-ending
- files were modified by this hook
```

## Testing: check-yaml

**File:** test_invalid_yaml.yaml

**Description:** Tests check-yaml hook with invalid YAML

**Results:**
```
Check YAML syntax....................................(no files to check)Skipped
```

## Testing: check-json

**File:** test_invalid_json.json

**Description:** Tests check-json hook with invalid JSON

**Results:**
```
Check JSON syntax....................................(no files to check)Skipped
```

## Testing: check-merge-conflict

**File:** test_merge_conflict.txt

**Description:** Tests check-merge-conflict hook with merge conflict markers

**Results:**
```
Check for merge conflicts............................(no files to check)Skipped
```

## Testing: detect-secrets

**File:** test_secrets.txt

**Description:** Tests detect-secrets hook with fake API keys

**Results:**
```
Detect secrets.......................................(no files to check)Skipped
```

## Testing: detect-private-key

**File:** test_private_key.txt

**Description:** Tests detect-private-key hook with fake private key

**Results:**
```
Detect private keys..................................(no files to check)Skipped
```

## Testing: check-added-large-files

**File:** test_large_file.txt

**Description:** Tests check-added-large-files hook with a 6MB file

**Results:**
```
Check for large files................................(no files to check)Skipped
```

## Testing: check-case-conflict

**File:** CaseConflict.txt

**Description:** Tests check-case-conflict hook with case conflicts

**Results:**
```
Check for case conflicts.............................(no files to check)Skipped
```

## Testing: check-case-conflict

**File:** caseconflict.txt

**Description:** Tests check-case-conflict hook with lowercase variant

**Results:**
```
Check for case conflicts.................................................Failed
- hook id: check-case-conflict
- exit code: 1
- files were modified by this hook

Case-insensitivity conflict found: precommit-tests/CaseConflict.txt
Case-insensitivity conflict found: precommit-tests/caseconflict.txt

```
