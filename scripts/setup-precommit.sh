#!/bin/bash
# Pre-commit setup script for Glance
# Note: This project standardizes on golangci-lint v2.1.2 across all environments

set -e

echo "Setting up pre-commit hooks for Glance..."

# Check if pre-commit is installed
if ! command -v pre-commit &> /dev/null; then
    echo "pre-commit not found. Attempting to install..."

    # Try pip installation
    if command -v pip &> /dev/null; then
        pip install pre-commit
    # Try pip3 installation
    elif command -v pip3 &> /dev/null; then
        pip3 install pre-commit
    # Try brew installation
    elif command -v brew &> /dev/null; then
        brew install pre-commit
    else
        echo "Error: Could not install pre-commit. Please install manually:"
        echo "pip install pre-commit or brew install pre-commit"
        exit 1
    fi
fi

# Check if golangci-lint is installed
if ! command -v golangci-lint &> /dev/null; then
    echo "golangci-lint not found. Attempting to install..."

    # Try brew installation with specific version
    if command -v brew &> /dev/null; then
        # Note: We try to install a specific version but fall back to using pre-commit's managed version
        # since brew might not support the exact version pinning we need
        brew install golangci-lint
        echo "Note: Brew might install a different version than v2.1.2 required by this project."
        echo "The pre-commit hook will use the correct version v2.1.2 regardless of the installed version."
    # Try go installation
    elif command -v go &> /dev/null; then
        go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.1.2
    else
        echo "Warning: Could not install golangci-lint. Please install manually:"
        echo "brew install golangci-lint or go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.1.2"
    fi
fi

# Install the pre-commit hooks
echo "Installing git hooks..."
pre-commit install

# Run the hooks on all files
echo "Running pre-commit hooks on all files..."
pre-commit run --all-files

echo "Pre-commit setup complete!"
echo "For more information, see docs/PRECOMMIT.md"
