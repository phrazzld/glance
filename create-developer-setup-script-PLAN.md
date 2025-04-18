# T022: Create Developer Setup Script

## Task Description
Create a script to help developers set up all necessary tooling, including pre-commit hooks, to simplify onboarding.

## Approach

I'll create a comprehensive developer setup script that builds upon the existing `setup-precommit.sh` and extends it to cover all necessary tooling and environment setup for new developers. This script will serve as a one-stop solution for setting up the development environment.

### 1. Examine Current Setup

- Review the existing `setup-precommit.sh` script
- Identify what's already covered and what's missing
- Check the `DEVELOPMENT_PHILOSOPHY.md` for any specific tooling requirements

### 2. Script Expansion Plan

Create a new script `setup-dev-environment.sh` that will:

1. **Check Prerequisites**
   - Go version (1.23+ as per project requirements)
   - Git installation
   - Basic build tools

2. **Install Core Dependencies**
   - Maintain pre-commit installation
   - Maintain golangci-lint installation
   - Add installation of any other required tools

3. **Set Up Development Environment**
   - Configure Go environment variables if needed
   - Set up local environment file template (.env)
   - Check and configure git settings

4. **Install Pre-commit Hooks**
   - Leverage the existing setup-precommit.sh script
   - Ensure hooks are installed correctly

5. **Verify GitHub CLI Installation** (for GitHub Actions interaction)
   - Check if GitHub CLI is installed
   - Provide installation instructions if not

6. **Run Verification Checks**
   - Verify the Go installation and version
   - Run a simple build test
   - Run pre-commit hooks to test the setup

7. **Provide Next Steps**
   - Display information about what to do next
   - Point to relevant documentation

### 3. Implementation

- Write the new script with clear error handling
- Add detailed comments and help text
- Include OS detection for platform-specific instructions
- Make the script idempotent (safe to run multiple times)

### 4. Testing

- Test the script on a clean environment
- Verify all tooling is installed correctly
- Ensure the script handles error cases gracefully

### 5. Documentation

- Update README.md to reference the new developer setup script
- Add script usage instructions to relevant documentation
