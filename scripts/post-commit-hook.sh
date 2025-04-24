#!/bin/bash
# Post-commit hook for Glance
# This hook runs Glance asynchronously after a commit to update documentation

# Get the root directory of the repository
REPO_ROOT=$(git rev-parse --show-toplevel)

# Log file for the async process
LOG_FILE="${REPO_ROOT}/.git/glance-post-commit.log"

# Function to run Glance
run_glance() {
    echo "Starting Glance documentation generation at $(date)" > "$LOG_FILE"

    # Change to the repository root directory
    cd "$REPO_ROOT" || exit 1

    # Run Glance (assuming the binary is available in the path or built locally)
    # Use the current directory as the target
    if [ -x "${REPO_ROOT}/glance" ]; then
        # If the binary exists in the repo root
        "${REPO_ROOT}/glance" . >> "$LOG_FILE" 2>&1
    elif command -v glance &> /dev/null; then
        # If the binary is available in PATH
        glance . >> "$LOG_FILE" 2>&1
    else
        # Try to build and run from source
        echo "Glance binary not found, building from source..." >> "$LOG_FILE"
        go build -o glance . && ./glance . >> "$LOG_FILE" 2>&1
    fi

    # Report completion status
    if [ $? -eq 0 ]; then
        echo "Glance documentation generation completed successfully at $(date)" >> "$LOG_FILE"
    else
        echo "Glance documentation generation failed at $(date)" >> "$LOG_FILE"
    fi
}

# Run Glance in the background to not block the commit process
(run_glance) &

# Inform the user
echo "Glance documentation generation started in the background"
echo "Check ${LOG_FILE} for progress and results"

exit 0
