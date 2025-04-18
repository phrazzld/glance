# Pre-commit Hook Test Results

This document contains the results of testing each pre-commit hook with specifically crafted test files.

Tests run on: Fri Apr 18 17:31:40 JST 2025

## Testing: go-fmt

**File:** test_go_fmt.go

**Description:** Tests go-fmt hook with improperly formatted Go code

**Results:**
```
go fmt...................................................................Failed
- hook id: go-fmt
- exit code: 1

Executable `run-go-fmt.sh` not found

```

## Testing: go-imports

**File:** test_go_imports.go

**Description:** Tests go-imports hook with unorganized imports

**Results:**
```
go imports...............................................................Failed
- hook id: go-imports
- exit code: 1

Executable `run-go-imports.sh` not found

```

## Testing: go-vet

**File:** test_go_vet.go

**Description:** Tests go-vet hook with suspicious code patterns

**Results:**
```
go vet...................................................................Failed
- hook id: go-vet
- duration: 0s
- exit code: 1

Executable `run-go-vet.sh` not found

```

## Testing: golangci-lint

**File:** test_golangci_lint.go

**Description:** Tests golangci-lint hook with linting issues

**Results:**
```
go-repo linter...........................................................Failed
- hook id: golangci-lint
- exit code: 3

Error: can't load config: unsupported version of the configuration: "" See https://golangci-lint.run/product/migration-guide for migration instructions
Failed executing command with error: can't load config: unsupported version of the configuration: "" See https://golangci-lint.run/product/migration-guide for migration instructions

golangci-lint............................................................Failed
- hook id: golangci-lint
- duration: 0.05s
- exit code: 3

Error: can't load config: unsupported version of the configuration: "" See https://golangci-lint.run/product/migration-guide for migration instructions
Failed executing command with error: can't load config: unsupported version of the configuration: "" See https://golangci-lint.run/product/migration-guide for migration instructions

```

## Testing: go-unit-tests

**File:** test_go_unit_test.go

**Description:** Tests go-unit-tests hook with a failing test

**Results:**
```
go test..................................................................Failed
- hook id: go-unit-tests
- exit code: 1

Executable `run-go-unit-tests.sh` not found

```

## Testing: go-build

**File:** test_go_build_error.go

**Description:** Tests go-build hook with build errors

**Results:**
```
go build.................................................................Failed
- hook id: go-build
- exit code: 1

Executable `run-go-build.sh` not found

```

## Testing: trailing-whitespace

**File:** test_trailing_whitespace.txt

**Description:** Tests trailing-whitespace hook with trailing spaces

**Results:**
```
Remove trailing whitespace...............................................Passed
```

## Testing: end-of-file-fixer

**File:** test_end_of_file.txt

**Description:** Tests end-of-file-fixer hook with missing final newline

**Results:**
```
Fix end of files.........................................................Failed
- hook id: end-of-file-fixer
- exit code: 1

Fixing precommit-tests/test_end_of_file.txt

```

## Testing: mixed-line-ending

**File:** test_mixed_line_ending.txt

**Description:** Tests mixed-line-ending hook with mixed line endings

**Results:**
```
Normalize line endings...................................................Failed
- hook id: mixed-line-ending
- exit code: 1

precommit-tests/test_mixed_line_ending.txt: fixed mixed line endings

```

## Testing: check-yaml

**File:** test_invalid_yaml.yaml

**Description:** Tests check-yaml hook with invalid YAML

**Results:**
```
Check YAML syntax........................................................Failed
- hook id: check-yaml
- exit code: 1

mapping values are not allowed here
  in "precommit-tests/test_invalid_yaml.yaml", line 3, column 14

```

## Testing: check-json

**File:** test_invalid_json.json

**Description:** Tests check-json hook with invalid JSON

**Results:**
```
Check JSON syntax........................................................Failed
- hook id: check-json
- exit code: 1

precommit-tests/test_invalid_json.json: Failed to json decode (Expecting ',' delimiter: line 5 column 3 (char 65))

```

## Testing: check-merge-conflict

**File:** test_merge_conflict.txt

**Description:** Tests check-merge-conflict hook with merge conflict markers

**Results:**
```
Check for merge conflicts................................................Passed
```

## Testing: detect-secrets

**File:** test_secrets.txt

**Description:** Tests detect-secrets hook with fake API keys

**Results:**
```
[INFO] Installing environment for https://github.com/Yelp/detect-secrets.
[INFO] Once installed this environment will be reused.
[INFO] This may take a few minutes...
Detect secrets...........................................................Failed
- hook id: detect-secrets
- exit code: 1

ERROR: Potential secrets about to be committed to git repo!

Secret Type: Base64 High Entropy String
Location:    precommit-tests/test_secrets.txt:3

Secret Type: Secret Keyword
Location:    precommit-tests/test_secrets.txt:3

Secret Type: AWS Access Key
Location:    precommit-tests/test_secrets.txt:4

Possible mitigations:
  - For information about putting your secrets in a safer place, please ask in
    #security
  - Mark false positives with an inline `pragma: allowlist secret` comment

If a secret has already been committed, visit
https://help.github.com/articles/removing-sensitive-data-from-a-repository

```

## Testing: detect-private-key

**File:** test_private_key.txt

**Description:** Tests detect-private-key hook with fake private key

**Results:**
```
Detect private keys......................................................Failed
- hook id: detect-private-key
- exit code: 1

Private key found: precommit-tests/test_private_key.txt

```

## Testing: check-added-large-files

**File:** test_large_file.txt

**Description:** Tests check-added-large-files hook with a 6MB file

**Results:**
```
Check for large files....................................................Passed
```

## Testing: check-case-conflict

**File:** CaseConflict.txt

**Description:** Tests check-case-conflict hook with case conflicts

**Results:**
```
Check for case conflicts.................................................Passed
```

