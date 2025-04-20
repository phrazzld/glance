# Security Suppressions Documentation

This document explains the security decisions and justifications for any remaining security suppressions in the Glance codebase. It is intended to provide transparency and documentation about security practices.

## Security Principles

Glance follows these key security principles:

1. **Path Validation**: All file paths are validated to prevent path traversal attacks.
2. **Least Privilege**: File permissions are set to the most restrictive settings when writing files (0600).
3. **Safe Defaults**: Default configurations are secure by default.
4. **Defense in Depth**: Multiple layers of validation are used for critical operations.

## Remaining Security Suppressions

The codebase contains a limited number of carefully considered security suppressions. Each is documented below with its rationale and mitigations.

### G304: File inclusion via variable

G304 flags cases where a file path from a variable is passed to functions like `os.ReadFile()` without validation, which could lead to path traversal vulnerabilities.

#### Justified Suppressions

| File | Line | Suppression | Justification | Mitigations |
|------|------|-------------|---------------|-------------|
| `llm/prompt.go` | 87 | `#nosec G304 -- The path has been cleaned, made absolute, and verified to be a file` | Loading a custom prompt template provided by the user. | Before reading, the path is cleaned via `filepath.Clean()`, converted to absolute path, and verified to be an existing file (not a directory). |
| `llm/prompt.go` | 101 | `#nosec G304 -- The path has been cleaned and is in the current working directory` | Loading the default prompt template from the current working directory. | The path is constructed using the current working directory and `filepath.Join()`, then cleaned to normalize it. |
| `config/loadconfig.go` | 155 | `#nosec G304 -- The path has been cleaned, made absolute, and verified to be a file` | Loading a custom prompt template provided by the user. | Same validation as in `llm/prompt.go`: cleaned, absolutized, and verified to be a file. |
| `config/loadconfig.go` | 169 | `#nosec G304 -- Reading from a standard prompt.txt file in the current directory` | Loading the default prompt template from the current working directory. | Path is built using `filepath.Join()` with the current directory and a fixed filename, then cleaned. |
| `filesystem/reader.go` | 44, 112 | `#nosec G304 -- When baseDir is not provided, caller is responsible for path validation` | Allowing legacy code paths to call the file reading functions without validation. | This is only used when `baseDir` is explicitly empty, which indicates the caller has opted out of validation and is taking responsibility. This is typically only used in tests. |
| `filesystem/reader.go` | 49, 117 | `#nosec G304 -- Path has been validated to prevent path traversal` | Reading a file after validation. | Before reading, the path is fully validated by `ValidateFilePath()` which ensures the path is within the allowed base directory. |

### G306: Write file with sensitive permissions

G306 flags cases where file permissions are not secure enough when writing files.

#### Justified Suppressions

| File | Line | Suppression | Justification | Mitigations |
|------|------|-------------|---------------|-------------|
| `glance.go` | 252 | `#nosec G306 -- Changed to 0600 for security & path validated` | Writing to a glance.md file with 0600 permissions for added security. | Permissions are set to 0o600 (user read/write only), which is more secure than the default. Additionally, the path is fully validated using `filesystem.ValidateFilePath()`. |

## Future Guidelines

When considering new suppressions, follow these guidelines:

1. **Document the Why**: Always add a comment explaining why the suppression is necessary.
2. **Add Mitigations**: Implement and document alternative security controls.
3. **Minimize Scope**: Keep the suppression as narrow as possible.
4. **Review Regularly**: Re-evaluate suppressions periodically to see if they can be removed.

All security suppressions must be reviewed and approved during code review before being committed.
