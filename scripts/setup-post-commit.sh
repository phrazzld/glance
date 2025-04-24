#!/bin/bash
# Setup script for the post-commit hook that runs Glance

set -e

echo "Setting up post-commit hook for Glance..."

# Get the root directory of the repository
REPO_ROOT=$(git rev-parse --show-toplevel)

# Path to the post-commit hook in the Git hooks directory
POST_COMMIT_HOOK="${REPO_ROOT}/.git/hooks/post-commit"

# Path to our post-commit hook script
HOOK_SCRIPT="${REPO_ROOT}/scripts/post-commit-hook.sh"

# Check if the hook script exists
if [ ! -f "$HOOK_SCRIPT" ]; then
    echo "Error: Post-commit hook script not found at $HOOK_SCRIPT"
    exit 1
fi

# Make sure the hook script is executable
chmod +x "$HOOK_SCRIPT"

# Create or update the post-commit hook
cat > "$POST_COMMIT_HOOK" << EOF
#!/bin/bash
# This hook was set up by setup-post-commit.sh

# Run the post-commit hook script
"$HOOK_SCRIPT"

# Exit with the script's exit code
exit \$?
EOF

# Make the post-commit hook executable
chmod +x "$POST_COMMIT_HOOK"

echo "Post-commit hook installed successfully at $POST_COMMIT_HOOK"
echo "Glance will now automatically run after each commit to update documentation."
