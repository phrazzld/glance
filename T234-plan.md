# T234 Plan: Refactor gatherSubGlances signature to include baseDir

## Context
The `gatherSubGlances` function in `glance.go` currently validates each subdirectory against its parent directory instead of a common security boundary. This can lead to potential path traversal vulnerabilities when an attacker provides an absolute path outside the intended directory structure.

## Approach
1. Locate the `gatherSubGlances` function in `glance.go`
2. Change its signature from `func gatherSubGlances(subdirs []string) (string, error)` to `func gatherSubGlances(baseDir string, subdirs []string) (string, error)`
3. Update function documentation to reflect the new parameter

## Implementation Details
- The `baseDir` parameter will be used as the security boundary for all path validations within the function
- This is the first step in a series of tasks (T234-T238) to fix the path validation vulnerability
- The function signature change will cause compilation failures, but these will be addressed in the subsequent tasks

## Success Criteria
- The function signature includes `baseDir string` as its first parameter 
- The function documentation explains the purpose of the `baseDir` parameter
- The code compiles successfully (though tests using the old signature will fail)