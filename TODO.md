# Todo

## Glance Parent Directory Regeneration

- [x] **T001 · Feature · P1: implement parent directory regeneration tracking**
    - **Context:** Solution Approach, Implementation Steps #1
    - **Action:**
        1. Modify `processDirectories` in `glance.go` to initialize and use a `needsRegen map[string]bool`.
        2. Update the `forceDir` check to incorporate `needsRegen[d]`.
        3. Add logic to call `filesystem.BubbleUpParents` after successful regeneration.
    - **Done‑when:**
        1. `processDirectories` function implements the parent propagation logic.
        2. Code compiles successfully.
    - **Depends‑on:** none

- [x] **T002 · Refactor · P2: enhance logging for regeneration reasons**
    - **Context:** Implementation Steps #2
    - **Action:**
        1. Update logging messages within `processDirectory` in `glance.go` to clearly indicate regeneration reason.
        2. Add debugging log for when parent directories are marked for regeneration.
    - **Done‑when:**
        1. Log messages clearly indicate regeneration reasons (e.g., forced, file change, parent marked).
    - **Depends‑on:** [T001]

- [x] **T003 · Test · P2: verify `filesystem.BubbleUpParents` unit tests**
    - **Context:** Testing Strategy #1
    - **Action:**
        1. Review existing tests for `filesystem.BubbleUpParents` in `filesystem/utils_test.go`.
        2. Add unit tests for missing scenarios (if any): deep nesting, root directory, multiple parents.
        3. Ensure tests verify the `needsRegen` map is correctly updated.
    - **Done‑when:**
        1. Unit tests for `filesystem.BubbleUpParents` are comprehensive and pass.
        2. Test coverage is adequate (determined by coverage report).
    - **Depends‑on:** none

- [x] **T004 · Test · P1: add integration test for parent regeneration on child change**
    - **Context:** Testing Strategy #2, #3
    - **Action:**
        1. Create a test with multi-level directory structure in `main_test.go` or `integration_test.go`.
        2. Run glance initially, modify a file in the deepest directory, run glance again.
        3. Verify glance.md files regenerated in the changed directory and all parents.
    - **Done‑when:**
        1. Integration test `TestParentRegenerationPropagation` passes.
        2. Test verifies parent regeneration by checking file modification times.
    - **Depends‑on:** [T001]

- [x] **T005 · Test · P2: add integration test for parent regeneration on forced child**
    - **Context:** Testing Strategy #3
    - **Action:**
        1. Create a test with multi-level directory structure.
        2. Run glance with force flag targeting only a child directory.
        3. Verify glance.md files regenerated in the forced child and all parents.
    - **Done‑when:**
        1. Integration test `TestForcedChildRegenerationBubblesUp` passes.
    - **Depends‑on:** [T001]

- [ ] **T006 · Test · P2: add integration test for no-change optimization**
    - **Context:** Testing Strategy #3
    - **Action:**
        1. Create a test setup and run glance initially, recording file modification times.
        2. Run glance again without changing any files.
        3. Verify no glance.md files were regenerated (modification times unchanged).
    - **Done‑when:**
        1. Integration test `TestNoChangesMeansNoRegeneration` passes.
    - **Depends‑on:** [T001]

- [ ] **T007 · Test · P2: add integration test for sibling directory isolation**
    - **Context:** Testing Strategy #2
    - **Action:**
        1. Create a test with branching directory structure (e.g., root/a/b and root/c).
        2. Run glance initially, modify a file in a deep subdirectory (root/a/b).
        3. Verify glance.md regenerated in modified path and parents, but NOT in sibling paths.
    - **Done‑when:**
        1. Test verifies that changes in one branch don't trigger regeneration in unrelated branches.
    - **Depends‑on:** [T001]
