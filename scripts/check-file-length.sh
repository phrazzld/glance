#!/bin/bash
# check-file-length.sh - Check if files exceed specified line length limits
#
# Usage: ./check-file-length.sh [--warning-threshold=N] [--error-threshold=M] file1 [file2 ...]
#
# This script checks files for exceeding line count thresholds:
# - Issues a warning if a file has more than WARNING_THRESHOLD lines (default: 500)
# - Exits with error if a file has more than ERROR_THRESHOLD lines (default: 1000)

set -o errexit
set -o nounset
set -o pipefail

WARNING_THRESHOLD=500
ERROR_THRESHOLD=1000
HAS_WARNINGS=0
HAS_ERRORS=0

# Process arguments
FILES=()
for arg in "$@"; do
  case $arg in
    --warning-threshold=*)
      WARNING_THRESHOLD="${arg#*=}"
      ;;
    --error-threshold=*)
      ERROR_THRESHOLD="${arg#*=}"
      ;;
    *)
      FILES+=("$arg")
      ;;
  esac
done

# Display header
echo "Checking file lengths (warning: $WARNING_THRESHOLD lines, error: $ERROR_THRESHOLD lines)"

# Process files
for file in "${FILES[@]}"; do
  # Skip if not a regular file (e.g., directories)
  if [ ! -f "$file" ]; then
    continue
  fi

  # Count lines in file
  linecount=$(wc -l < "$file")

  # Check against thresholds
  if [ "$linecount" -gt "$ERROR_THRESHOLD" ]; then
    printf "\033[1;31mERROR: %s has %d lines (exceeds limit of %d)\033[0m\n" "$file" "$linecount" "$ERROR_THRESHOLD"
    HAS_ERRORS=1
  elif [ "$linecount" -gt "$WARNING_THRESHOLD" ]; then
    printf "\033[1;33mWARNING: %s has %d lines (exceeds recommended limit of %d)\033[0m\n" "$file" "$linecount" "$WARNING_THRESHOLD"
    HAS_WARNINGS=1
  fi
done

# Report summary
if [ "$HAS_ERRORS" -eq 1 ]; then
  echo -e "\033[1;31mCheck failed: Files exceeding maximum line count found\033[0m"
  echo "Please refactor files with more than $ERROR_THRESHOLD lines"
  exit 1
elif [ "$HAS_WARNINGS" -eq 1 ]; then
  echo -e "\033[1;33mCheck passed with warnings: Files exceeding recommended line count found\033[0m"
  echo "Consider refactoring files with more than $WARNING_THRESHOLD lines"
  exit 0
else
  echo -e "\033[1;32mCheck passed: All files are within line count limits\033[0m"
  exit 0
fi
