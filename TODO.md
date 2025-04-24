# Todo

## Radical Progress Bar Simplification

- [x] **T001 · Refactor · P1: remove all progress bar code from ui/feedback.go**
    - **Context:** Radical simplification - eliminate entire abstraction layer
    - **Action:**
        1. Remove the `ProgressBar` interface definition.
        2. Remove the `ProgressTrackerFactory` interface definition and implementation.
        3. Remove the `ConcreteProgressBar` struct and all its methods.
        4. Remove all progress bar constructor functions.
        5. Remove all progress bar option functions.
        6. Keep `ProgressBarTheme` if needed for spinners, otherwise remove it too.
    - **Done‑when:**
        1. All progress bar related code is removed from `ui/feedback.go`.
        2. Spinner functionality (if needed) remains intact.
        3. Code won't compile but that's expected at this stage.
    - **Depends‑on:** none

- [x] **T002 · Feature · P1: implement direct progress bar usage in glance.go**
    - **Context:** Radical simplification - use library directly
    - **Action:**
        1. Import the progress bar library directly in `glance.go`:
           ```go
           import progressbar "github.com/schollz/progressbar/v3"
           ```
        2. Remove the `progressFactory` parameter from `processDirectories` function.
        3. Create the progress bar directly in `processDirectories`:
           ```go
           bar := progressbar.NewOptions(len(dirsList),
               progressbar.OptionSetDescription("Creating glance files"),
               progressbar.OptionShowCount(),
               progressbar.OptionSetWidth(40),
               progressbar.OptionSetPredictTime(false),
           )
           ```
        4. Update the progress bar directly:
           ```go
           // Ignore error for non-critical UI
           _ = bar.Add(1)
           ```
    - **Done‑when:**
        1. Progress bar is created directly in `glance.go` using the library.
        2. All factory/abstraction references are removed.
        3. Code compiles.
    - **Depends‑on:** none

- [x] **T003 · Feature · P1: add test output suppression mechanism**
    - **Context:** Radical simplification - test output handling
    - **Action:**
        1. Add a simple mechanism to suppress progress bar output in tests.
        2. Options:
           - Add a boolean parameter `testing bool` to `processDirectories`
           - Add a field to config: `cfg.SuppressProgressOutput`
           - Use build tags: `// +build !test`
        3. If testing/output suppression is enabled, redirect progress bar output:
           ```go
           if testing {
               bar.RenderTo(io.Discard)
           }
           ```
    - **Done‑when:**
        1. Progress bar output can be suppressed in tests.
        2. Code compiles.
    - **Depends‑on:** T002

- [x] **T004 · Test · P1: update integration tests**
    - **Context:** Radical simplification - adapt tests
    - **Action:**
        1. Identify and update all tests that involve progress bar.
        2. Use the new mechanism for suppressing output in tests.
        3. Remove any mock progress bar implementations.
        4. Ensure tests focus on functional behavior, not UI.
    - **Done‑when:**
        1. All tests pass.
        2. No UI output during tests.
        3. No reference to old progress bar abstractions.
    - **Depends‑on:** T001, T002, T003

- [x] **T005 · Test · P2: manually verify CLI progress bar appearance**
    - **Context:** Verify user experience
    - **Action:**
        1. Run the `glance` command locally on a sample set of directories.
        2. Visually verify the progress bar appears and functions correctly.
    - **Done‑when:**
        1. Progress bar appears and updates correctly in terminal.
        2. User experience is maintained.
    - **Depends‑on:** T002

- [x] **T006 · Chore · P2: run linters/formatters and fix issues**
    - **Context:** Code cleanup
    - **Action:**
        1. Run `go fmt ./...`
        2. Run `golangci-lint run --config=.golangci.yml`
        3. Fix any issues reported by linters
    - **Done‑when:**
        1. Code passes all linting and formatting checks
    - **Depends‑on:** T001, T002, T003, T004

- [x] **T007 · Chore · P3: update BACKLOG.md**
    - **Context:** Documentation update
    - **Action:**
        1. Locate all progress bar related tasks in `BACKLOG.md`.
        2. Mark them as complete or remove them.
        3. Add a note about the radical simplification if appropriate.
    - **Done‑when:**
        1. `BACKLOG.md` reflects the completed progress bar simplification.
    - **Depends‑on:** T001, T002, T003, T004, T005, T006

- [x] **T008 · Refactor · P3: clean up imports and dependencies**
    - **Context:** Code cleanup
    - **Action:**
        1. Remove any unused imports left after removing progress bar code.
        2. Run `go mod tidy` to clean up dependencies.
        3. Check if any dependencies have become unused and can be removed.
    - **Done‑when:**
        1. No unused imports in the codebase.
        2. Dependencies are clean.
    - **Depends‑on:** T001, T002, T003
