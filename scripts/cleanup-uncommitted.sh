#!/bin/bash
# Script to clean up uncommitted changes after PR creation
set -e

echo "Cleaning up uncommitted changes after PR creation..."

# 1. Commit helpful documentation and code style fixes
echo "Committing documentation and code style improvements..."
git add docs/*.md docs/prompts/*.md errors/README.md glance.go integration_test.go ui/feedback_test.go
git commit -m "Fix whitespace and end-of-file formatting issues in docs and core code

This commit addresses code style issues by:
- Adding newlines at end of files where missing
- Removing unnecessary whitespace in core files
- Ensuring consistent formatting across documentation

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>"

# 2. Commit useful planning documents
echo "Committing planning documents..."
git add PLAN.md TASK-PROMPT.md ticket-task.md
git commit -m "Add task planning documents

These documents provide context for the implementation:
- PLAN.md: Overall implementation approach
- TASK-PROMPT.md: Task planning instructions
- ticket-task.md: Task decomposition guidelines

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>"

# 3. Restore test files with intentional issues
echo "Restoring test files with intentional issues..."
git restore tests/github-actions-test/* precommit-tests/*

# 4. Check for any remaining changes
echo "Remaining uncommitted changes:"
git status

echo "Cleanup complete!"
echo "IMPORTANT: Make sure to manually review any remaining uncommitted changes"
