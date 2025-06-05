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
| `config/template.go` (LoadPromptTemplate) | - | `#nosec G304 -- The path has been validated using filesystem.ValidateFilePath` | Loading a custom prompt template provided by the user. | Before reading, the path is cleaned, made absolute, and fully validated by `filesystem.ValidateFilePath()` which ensures the path is valid. |
| `config/template.go` (LoadPromptTemplate) | - | `#nosec G304 -- The path has been validated using filesystem.ValidateFilePath` | Loading the default prompt template from the current working directory. | Path is validated using `filesystem.ValidateFilePath()` against the current working directory as the baseDir. |
| `filesystem/reader.go` (ReadTextFile) | - | `#nosec G304 -- When baseDir is not provided, caller is responsible for path validation` | Allowing legacy code paths to call the file reading functions without validation. | This is only used when `baseDir` is explicitly empty, which indicates the caller has opted out of validation and is taking responsibility. This is typically only used in tests. |
| `filesystem/reader.go` (ReadTextFile, IsTextFile) | - | `#nosec G304 -- Path has been validated to prevent path traversal` | Reading a file after validation. | Before reading, the path is fully validated by `ValidateFilePath()` which ensures the path is within the allowed base directory. |

### G306: Write file with sensitive permissions

G306 flags cases where file permissions are not secure enough when writing files.

#### Justified Suppressions

| File | Line | Suppression | Justification | Mitigations |
|------|------|-------------|---------------|-------------|
| `glance.go` (processDirectory) | - | `#nosec G306 -- Using filesystem.DefaultFileMode (0600) for security & path validated` | Writing to a glance.md file with restrictive permissions for added security. | Permissions are set using `filesystem.DefaultFileMode` (0o600, user read/write only), which is more secure than the default. Additionally, the path is fully validated using `filesystem.ValidateFilePath()`. |

## Future Guidelines

When considering new suppressions, follow these guidelines:

1. **Document the Why**: Always add a comment explaining why the suppression is necessary.
2. **Add Mitigations**: Implement and document alternative security controls.
3. **Minimize Scope**: Keep the suppression as narrow as possible.
4. **Review Regularly**: Re-evaluate suppressions periodically to see if they can be removed.

All security suppressions must be reviewed and approved during code review before being committed.
