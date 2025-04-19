# Path Validation Security Implementation - Consultant Plan

## Task Overview

**T037 · feature · p1: identify locations requiring path validation**

This plan identifies all code locations in the Glance codebase that accept file paths from external sources (CLI, configuration, environment variables) and specifies the exact validation requirements for each to prevent path traversal vulnerabilities. This is a critical security task that will unblock the implementation of proper path validation (T038) and the eventual removal of `#nosec` suppressions (T040).

## 1. Core Findings: Path Validation Entry Points

### A. Primary External Input Sources

1. **Target Directory Input (CLI Argument)**
   - **Location**: `config/loadconfig.go` → `LoadConfig`
   - **Source**: CLI positional argument (e.g., `./glance /path/to/target`)
   - **Operations**: Drives all directory scanning and file operations
   - **Validation Required**:
     - Normalize: `filepath.Clean(path)`
     - Absolutize: `filepath.Abs(path)` (already done)
     - Verify existence and is directory (already done via `dirChecker.CheckDirectory`)
     - **Critical**: Store this as the "trusted root" boundary for all subsequent operations

2. **Custom Prompt File Input (CLI Flag)**
   - **Location**: `config/loadconfig.go` → `loadPromptTemplate`
   - **Source**: CLI flag (`--prompt-file /path/to/prompt.txt`)
   - **Operations**: Reads template file for LLM prompt
   - **Validation Required**:
     - Normalize: `filepath.Clean(path)`
     - Absolutize: `filepath.Abs(path)`
     - Boundary check: Must be within allowed directory (CWD or project dir)
     - Verify existence and is file (not directory)
   - **`#nosec` Ref**: G304 (line 119)

3. **Fallback Prompt File ("prompt.txt")**
   - **Location**:
     - `config/loadconfig.go` → `loadPromptTemplate` (line 127)
     - `llm/prompt.go` → `LoadTemplate` (line 68)
   - **Source**: Hardcoded string `prompt.txt` (assumed in CWD)
   - **Operations**: Reads default LLM prompt template
   - **Validation Required**:
     - Normalize: `filepath.Clean("prompt.txt")`
     - Absolutize and check within CWD context
     - Optionally verify is file (not directory)
   - **`#nosec` Ref**: G304 (lines 127 in loadconfig.go, 68 in prompt.go)

### B. Derived Path Operations (Internal)

4. **Reading Source Files During Directory Scan**
   - **Location**:
     - `filesystem/reader.go` → `ReadTextFile` (line 30), `IsTextFile` (line 81)
     - `glance.go` → `gatherLocalFiles` (line 457)
   - **Source**: Paths generated during directory traversal
   - **Operations**: Reads files for LLM prompt context
   - **Validation Required**:
     - Normalize: `filepath.Clean(path)`
     - Absolutize: `filepath.Abs(path)`
     - **Critical**: Verify path is within trusted root (prefix check)
   - **`#nosec` Ref**: G304 (lines 30, 81 in reader.go; 457, 477 in glance.go)

5. **Writing Generated glance.md Files**
   - **Location**: `glance.go` → `processDirectory` (line 253)
   - **Source**: Path derived by joining directory with "glance.md"
   - **Operations**: Writes generated summary
   - **Validation Required**:
     - Normalize: `filepath.Clean(glancePath)`
     - Absolutize: `filepath.Abs(glancePath)`
     - **Critical**: Verify path is within trusted root (prefix check)
     - Use consistent file permissions (separate task T051/T052)
   - **`#nosec` Ref**: G306 (line 253 in glance.go)

6. **Reading Existing glance.md from Subdirectories**
   - **Location**: `glance.go` → `gatherSubGlances` (line 387)
   - **Source**: Paths derived from subdirectories
   - **Operations**: Reads existing summaries
   - **Validation Required**:
     - Normalize: `filepath.Clean(path)`
     - Absolutize: `filepath.Abs(path)`
     - **Critical**: Verify path is within trusted root (prefix check)
   - **`#nosec` Ref**: G304 (line 387 in glance.go)

## 2. Standard Path Validation Workflow

For all file path operations, implement this validation workflow:

1. **Normalize the path**
   ```go
   cleanPath := filepath.Clean(path)
   ```

2. **Convert to absolute path**
   ```go
   absPath, err := filepath.Abs(cleanPath)
   if err != nil {
       return fmt.Errorf("invalid path: %w", err)
   }
   ```

3. **Check path is within allowed boundary**
   ```go
   // Option 1: Simple prefix check
   if !strings.HasPrefix(absPath, baseDir+string(os.PathSeparator)) && absPath != baseDir {
       return fmt.Errorf("path %q is outside of allowed directory %q", path, baseDir)
   }

   // Option 2: Relative path check
   rel, err := filepath.Rel(baseDir, absPath)
   if err != nil || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
       return fmt.Errorf("path %q escapes the allowed base directory %q", path, baseDir)
   }
   ```

4. **Handle symlinks (if necessary)**
   ```go
   evalPath, err := filepath.EvalSymlinks(absPath)
   if err != nil {
       return fmt.Errorf("invalid path: %w", err)
   }
   // Then re-check boundary on evalPath
   ```

5. **Verify file/directory existence and type**
   ```go
   info, err := os.Stat(absPath)
   if err != nil {
       return fmt.Errorf("cannot access path %q: %w", path, err)
   }

   // Check if directory or file as appropriate
   if expectDir && !info.IsDir() {
       return fmt.Errorf("path %q is not a directory", path)
   }
   if !expectDir && info.IsDir() {
       return fmt.Errorf("path %q is a directory, expected a file", path)
   }
   ```

## 3. Recommended Implementation Approach

1. **Create a utility package for path validation**
   - Add shared validation functions to the `filesystem` package:
     ```go
     // ValidatePathUnderBase checks if a path is strictly contained within a base directory
     func ValidatePathUnderBase(path, baseDir string) error { ... }

     // ValidateFilePath checks if a path exists, is a file, and is under the base directory
     func ValidateFilePath(path, baseDir string) error { ... }

     // ValidateDirPath checks if a path exists, is a directory, and is under the base directory
     func ValidateDirPath(path, baseDir string) error { ... }
     ```

2. **Apply validation at system boundaries first**
   - Validate input paths immediately in `LoadConfig` and `loadPromptTemplate`
   - Store valid, absolute paths in the configuration

3. **Apply validation before all file operations**
   - Add validation before every `os.ReadFile`, `os.Open`, and `os.WriteFile` call
   - Focus especially on code with existing `#nosec` annotations

## 4. Specific Recommendations for Each Location

### LoadConfig (config/loadconfig.go)
- Already validates target directory existence but should also clean path
- Store the absolute, clean path as the trusted root for boundary checks

### loadPromptTemplate (config/loadconfig.go)
- Add validation before `os.ReadFile(path)` on line 119
- Determine if prompt file should be restricted to specific directories or allow arbitrary locations

### LoadTemplate (llm/prompt.go)
- Add validation before `os.ReadFile(path)` on lines 59 and 68
- Consider restricting to CWD or project dir for security

### Reader Functions (filesystem/reader.go)
- Add boundary checks in `ReadTextFile` and `IsTextFile`
- Or document that these are internal functions requiring caller validation

### File Operations in glance.go
- Add boundary checks before all read/write operations
- Refactor duplicate functionality to use filesystem package (future tasks T042-T044)

## 5. Testing Recommendations for T039

1. Create tests for path traversal attempts:
   - `../` sequences that try to escape the base directory
   - Absolute paths outside the base directory
   - Symlinks pointing outside the base directory

2. Test both positive and negative cases:
   - Valid paths should be accepted
   - Invalid paths should be rejected with appropriate errors

## 6. Next Steps

This plan directly unblocks:

1. **T038** - Implement path validation for external inputs
   - Use this document to guide implementation of validation at each location

2. **T040** - Remove unnecessary #nosec suppressions
   - Once validation is implemented, suppressions can be safely removed

Additionally, this relates to:

3. **T042-T044** - Filesystem refactoring
   - Path validation should be integrated into the refactored filesystem package

## Conclusion

This comprehensive path validation plan identifies all code locations requiring validation and provides specific recommendations for proper implementation. By following this plan, Glance will have a robust defense against path traversal attacks, meeting the security requirements of the development philosophy.
