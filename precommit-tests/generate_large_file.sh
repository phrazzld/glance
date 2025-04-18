#!/bin/bash
# Script to generate a large file for testing check-added-large-files hook

file="test_large_file.txt"
echo "Generating large test file..."
dd if=/dev/zero of="$file" bs=1M count=6 2>/dev/null
echo "Created $file"