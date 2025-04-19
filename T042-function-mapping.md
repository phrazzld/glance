# Duplicate Filesystem Functions Mapping

This document maps functions in `glance.go` that duplicate functionality available in the `filesystem/` package, as part of T042: identify duplicate filesystem functions.

## Function Mapping

| `glance.go` Function | `filesystem/` Function | Description | Parameter Differences | Return Value Differences | Notes |
|----------------------|------------------------|-------------|------------------------|--------------------------|-------|
| `latestModTime` (L560-586) | `filesystem.LatestModTime` (utils.go L28-62) | Finds the most recent modification time in a directory | `ignoreChain` parameter type: <br>- `glance.go`: `[]*gitignore.GitIgnore`<br>- `filesystem/`: `IgnoreChain` | None | The `filesystem.LatestModTime` function uses `ShouldIgnoreDir` while the `glance.go` version implements similar logic directly with `isIgnored`. Both do the same thing but with different ignore chain types. |
| `shouldRegenerate` (L538-557) | `filesystem.ShouldRegenerate` (utils.go L64-112) | Determines if a glance.md file needs to be regenerated | `ignoreChain` parameter type: <br>- `glance.go`: `[]*gitignore.GitIgnore`<br>- `filesystem/`: `IgnoreChain` | None | Very similar behavior. The `filesystem.ShouldRegenerate` function uses more verbose logging. |
| `bubbleUpParents` (L588-597) | `filesystem.BubbleUpParents` (utils.go L114-139) | Marks parent directories for regeneration | None | None | The `filesystem.BubbleUpParents` function has more detailed logic to handle the edge case where a parent is the root. |
| `isIgnored` (L364-376) | `filesystem.MatchesGitignore` (ignore.go L134-175) | Checks if a path matches gitignore patterns | - `isIgnored` takes `rel` and `chain` parameters<br>- `MatchesGitignore` takes `path`, `baseDir`, `ignoreChain`, and `isDir` parameters | None | The `filesystem.MatchesGitignore` function has more comprehensive logic that handles relative paths properly. |
| `listAllDirsWithIgnores` (L276-348) | `filesystem.ListDirsWithIgnores` (scanner.go L43-135) | Performs BFS to collect directories with gitignore info | Return value uses different ignore chain types | None | The `filesystem.ListDirsWithIgnores` function uses the `IgnoreChain` type while `glance.go` uses raw `[]*gitignore.GitIgnore` slices. |
| `loadGitignore` (L350-361) | `filesystem.LoadGitignore` (scanner.go L137-156) | Loads a .gitignore file from a directory | None | None | Identical behavior. |
| `gatherLocalFiles` (L462-529) | `filesystem.GatherLocalFiles` (reader.go L143-280) | Collects text files from a directory | `ignoreChain` parameter type: <br>- `glance.go`: `[]*gitignore.GitIgnore`<br>- `filesystem/`: `[]IgnoreRule` | None | The `filesystem.GatherLocalFiles` version has more comprehensive path validation logic. |
| `readSubdirectories` (L419-460) | Not a direct equivalent, but functionality covered by `ListDirsWithIgnores` with filtering | Lists immediate subdirectories | N/A | N/A | Would need to be replaced by a combination of `ListDirsWithIgnores` and filtering for immediate children only. |
| `gatherSubGlances` (L389-417) | Not a direct equivalent, but similar file reading functions exist | Merges subdirectory glance.md files | N/A | N/A | Similar functionality can be achieved using `filesystem.ReadTextFile` and directory walking. |

## Notes on Non-Trivial Replacements

### `readSubdirectories`

The `readSubdirectories` function in `glance.go` (L419-460) doesn't have a direct equivalent in the `filesystem/` package, but its functionality can be achieved by:

1. Using `ListDirsWithIgnores` to get all directories
2. Filtering to include only immediate children of the given directory
3. Using the `ShouldIgnoreDir` function to apply the same filtering logic

### `gatherSubGlances`

The `gatherSubGlances` function in `glance.go` (L389-417) doesn't have a direct equivalent, but it can be implemented using:

1. `filesystem.ValidateDirPath` and `filesystem.ValidateFilePath` for path validation (already being used in the current implementation)
2. `filesystem.ReadTextFile` to read each glance.md file

## Compatibility Considerations

1. **IgnoreChain vs []*gitignore.GitIgnore**: The most significant difference is the use of different types for gitignore chains. The `filesystem/` package uses the `IgnoreChain` type (a slice of `IgnoreRule` structs), while `glance.go` uses raw `[]*gitignore.GitIgnore` slices. Conversion utilities `ExtractGitignoreMatchers` and `CreateIgnoreChain` are available in the `filesystem/` package.

2. **Path Validation**: The `filesystem/` functions generally have more comprehensive path validation to prevent path traversal attacks. Any replacements should maintain or improve this security aspect.

3. **Logging**: The `filesystem/` functions often include more detailed logging. When replacing functions, we should ensure the same level of logging is maintained.

## Recommendation for `glance.go` Functions to Replace

Based on this analysis, the following functions in `glance.go` should be replaced with their `filesystem/` equivalents in task T043:

1. `latestModTime` → `filesystem.LatestModTime`
2. `shouldRegenerate` → `filesystem.ShouldRegenerate`
3. `bubbleUpParents` → `filesystem.BubbleUpParents`
4. `isIgnored` → `filesystem.MatchesGitignore`
5. `listAllDirsWithIgnores` → `filesystem.ListDirsWithIgnores` (with type conversion)
6. `loadGitignore` → `filesystem.LoadGitignore`
7. `gatherLocalFiles` → `filesystem.GatherLocalFiles` (with type conversion)

For the functions without direct equivalents, we'll need to create custom implementations in task T043:

1. `readSubdirectories`: Create a custom implementation that uses `filesystem` package functions
2. `gatherSubGlances`: Create a custom implementation that uses `filesystem` package functions
