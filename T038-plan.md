# T038 Plan: Implement Path Validation for External Inputs

## Task Description
Implement path validation for external file path inputs to prevent path traversal vulnerabilities.

## Implementation Approach

### 1. Summary
We'll implement path validation at each identified entry point using the validation utilities we created in T037. The CONSULTANT-PLAN.md document identified 6 key locations:

1. **Target Directory Input (CLI Argument)** in `config/loadconfig.go`
2. **Custom Prompt File Input (CLI Flag)** in `config/loadconfig.go`
3. **Fallback Prompt File ("prompt.txt")** in both `config/loadconfig.go` and `llm/prompt.go`
4. **Reading Source Files During Directory Scan** in `filesystem/reader.go` and `glance.go`
5. **Writing Generated glance.md Files** in `glance.go`
6. **Reading Existing glance.md from Subdirectories** in `glance.go`

### 2. Implementation Steps

#### A. For the Target Directory (CLI Argument) in LoadConfig:
1. Already uses `filepath.Abs()` but missing `filepath.Clean()`
2. Will store this as the "trusted root" boundary for other path operations
3. Will use our new `ValidateDirPath()` function

#### B. For the Custom Prompt File (CLI Flag):
1. Add validation using our new `ValidateFilePath()` function
2. Define appropriate base directory (CWD or project dir)
3. Set proper validation flags (must exist, must be file)

#### C. For the Fallback Prompt File:
1. Add validation in both `config/loadconfig.go` and `llm/prompt.go`
2. Use relative path handling with appropriate context

#### D. For Reading Source Files:
1. Add validation in `filesystem/reader.go` or ensure caller validation
2. Add validation in relevant `glance.go` functions

#### E. For Writing glance.md Files:
1. Add validation in `glance.go` before file write operations

#### F. For Reading Existing glance.md Files:
1. Add validation in `gatherSubGlances` function

### 3. Important Considerations
1. All paths will be normalized with `filepath.Clean()`
2. All paths will be converted to absolute paths
3. All paths will be validated to ensure they don't escape the base directory
4. Proper error handling will be added throughout

### 4. Testing Strategy
1. Run unit tests after implementing validation at each point
2. Focus on verifying paths are properly validated

## Implementation Order
1. Update `config/loadconfig.go` first to establish trusted root directory
2. Update prompt file handling in both locations
3. Update file reading and writing operations in `glance.go`
4. Run all tests to ensure nothing is broken

This implementation follows the plan detailed in CONSULTANT-PLAN.md and will provide comprehensive protection against path traversal vulnerabilities.