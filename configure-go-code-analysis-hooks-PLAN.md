# T004: Configure Go-specific code analysis hooks

## Task Description
Add go-vet hook to the pre-commit config to detect suspicious code patterns.

## Approach
1. Examine the current .pre-commit-config.yaml file
2. Add or configure the go-vet hook in the Go-specific hooks section
3. Configure the hook with appropriate settings:
   - Proper command arguments for go vet
   - File pattern matching to only run on Go files
   - Verbose output for better debugging
4. Add appropriate description and documentation in the config
5. Validate the configuration to ensure it's properly formed