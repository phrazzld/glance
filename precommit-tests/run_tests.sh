#!/bin/bash
# Script to run pre-commit hooks on test files and document results

cd "$(dirname "$0")/.."
ROOT_DIR="$(pwd)"
TEST_DIR="$ROOT_DIR/precommit-tests"
RESULTS_FILE="$TEST_DIR/hook_test_results.md"

echo "# Pre-commit Hook Test Results" > "$RESULTS_FILE"
echo "" >> "$RESULTS_FILE"
echo "This document contains the results of testing each pre-commit hook with specifically crafted test files." >> "$RESULTS_FILE"
echo "" >> "$RESULTS_FILE"
echo "Tests run on: $(date)" >> "$RESULTS_FILE"
echo "" >> "$RESULTS_FILE"

# Function to test a specific file with pre-commit
test_file() {
    local file="$1"
    local hook="$2"
    local description="$3"

    echo "## Testing: $hook" >> "$RESULTS_FILE"
    echo "" >> "$RESULTS_FILE"
    echo "**File:** $(basename "$file")" >> "$RESULTS_FILE"
    echo "" >> "$RESULTS_FILE"
    echo "**Description:** $description" >> "$RESULTS_FILE"
    echo "" >> "$RESULTS_FILE"
    echo "**Results:**" >> "$RESULTS_FILE"
    echo '```' >> "$RESULTS_FILE"

    if [ -f "$file" ]; then
        # Create a backup of the original file
        cp "$file" "${file}.bak" 2>/dev/null || true

        # Stage the file with force to ensure it's in the git index
        # This is necessary for hooks that check staged content
        cd "$ROOT_DIR"
        git add --force "$file" 2>/dev/null

        # Run the pre-commit hook on the file - don't output git messages to results
        pre-commit run "$hook" --files "$file" >> "$RESULTS_FILE" 2>&1 || true

        # Unstage the file to clean up - don't output reset messages to results file
        git reset HEAD "$file" 2>/dev/null

        # Restore original file if modified
        if [ -f "${file}.bak" ]; then
            # Detect if hook modified the file
            if ! cmp -s "$file" "${file}.bak"; then
                echo "File was modified by hook. Showing original -> modified changes:" >> "$RESULTS_FILE"
                diff -u "${file}.bak" "$file" >> "$RESULTS_FILE" 2>&1 || true
            fi

            # Restore the original test file for future runs
            mv "${file}.bak" "$file" 2>/dev/null
        fi
    else
        echo "File not found: $file" >> "$RESULTS_FILE"
    fi

    echo '```' >> "$RESULTS_FILE"
    echo "" >> "$RESULTS_FILE"
}

# Test Go formatting hooks
test_file "$TEST_DIR/test_go_fmt.go" "go-fmt" "Tests go-fmt hook with improperly formatted Go code"
test_file "$TEST_DIR/test_go_imports.go" "go-imports" "Tests go-imports hook with unorganized imports"

# Test Go code analysis hooks
test_file "$TEST_DIR/test_go_vet.go" "go-vet" "Tests go-vet hook with suspicious code patterns"
test_file "$TEST_DIR/test_golangci_lint.go" "golangci-lint" "Tests golangci-lint hook with linting issues"

# Test Go test hooks
test_file "$TEST_DIR/test_go_unit_test.go" "go-unit-tests" "Tests go-unit-tests hook with a failing test"
test_file "$TEST_DIR/test_go_build_error.go" "go-build" "Tests go-build hook with build errors"

# Test file formatting hooks
test_file "$TEST_DIR/test_trailing_whitespace.txt" "trailing-whitespace" "Tests trailing-whitespace hook with trailing spaces"
test_file "$TEST_DIR/test_end_of_file.txt" "end-of-file-fixer" "Tests end-of-file-fixer hook with missing final newline"
test_file "$TEST_DIR/test_mixed_line_ending.txt" "mixed-line-ending" "Tests mixed-line-ending hook with mixed line endings"
test_file "$TEST_DIR/test_invalid_yaml.yaml" "check-yaml" "Tests check-yaml hook with invalid YAML"
test_file "$TEST_DIR/test_invalid_json.json" "check-json" "Tests check-json hook with invalid JSON"
test_file "$TEST_DIR/test_merge_conflict.txt" "check-merge-conflict" "Tests check-merge-conflict hook with merge conflict markers"

# Test security hooks
echo "## Testing: detect-secrets" >> "$RESULTS_FILE"
echo "" >> "$RESULTS_FILE"
echo "**File:** test_secrets.txt" >> "$RESULTS_FILE"
echo "" >> "$RESULTS_FILE"
echo "**Description:** Tests detect-secrets hook with fake API keys" >> "$RESULTS_FILE"
echo "" >> "$RESULTS_FILE"
echo "**Results:**" >> "$RESULTS_FILE"
echo '```' >> "$RESULTS_FILE"
cd "$ROOT_DIR"
# Create a backup of the original file
cp "$TEST_DIR/test_secrets.txt" "$TEST_DIR/test_secrets.txt.bak" 2>/dev/null || true

# Stage the file with force to ensure it's in the git index
git add --force "$TEST_DIR/test_secrets.txt" 2>/dev/null

# Run the hook
pre-commit run detect-secrets --files "$TEST_DIR/test_secrets.txt" >> "$RESULTS_FILE" 2>&1 || true

# Unstage the file to clean up
git reset HEAD "$TEST_DIR/test_secrets.txt" 2>/dev/null

# Restore original file if modified
if [ -f "$TEST_DIR/test_secrets.txt.bak" ]; then
    # Detect if hook modified the file
    if ! cmp -s "$TEST_DIR/test_secrets.txt" "$TEST_DIR/test_secrets.txt.bak"; then
        echo "File was modified by hook. Showing original -> modified changes:" >> "$RESULTS_FILE"
        diff -u "$TEST_DIR/test_secrets.txt.bak" "$TEST_DIR/test_secrets.txt" >> "$RESULTS_FILE" 2>&1 || true
    fi

    # Restore the original test file for future runs
    mv "$TEST_DIR/test_secrets.txt.bak" "$TEST_DIR/test_secrets.txt" 2>/dev/null
fi
echo '```' >> "$RESULTS_FILE"
echo "" >> "$RESULTS_FILE"
test_file "$TEST_DIR/test_private_key.txt" "detect-private-key" "Tests detect-private-key hook with fake private key"

# Generate and test large file
if [ -x "$TEST_DIR/generate_large_file.sh" ]; then
    cd "$TEST_DIR"
    ./generate_large_file.sh
    cd "$ROOT_DIR"
    test_file "$TEST_DIR/test_large_file.txt" "check-added-large-files" "Tests check-added-large-files hook with a 6MB file"
fi

# Test case conflict - create a temporary case-conflicting file
touch "$TEST_DIR/caseconflict.txt"
echo "This is a lowercase version to create a conflict" > "$TEST_DIR/caseconflict.txt"
test_file "$TEST_DIR/CaseConflict.txt" "check-case-conflict" "Tests check-case-conflict hook with case conflicts"
test_file "$TEST_DIR/caseconflict.txt" "check-case-conflict" "Tests check-case-conflict hook with lowercase variant"
# Clean up the temporary file
rm -f "$TEST_DIR/caseconflict.txt"

echo "Tests completed. Results saved to $RESULTS_FILE"
echo "To view results: cat $RESULTS_FILE"
