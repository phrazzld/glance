# T020: Create Test Pull Request

## Task Description
Create a test pull request with mixed good/bad code to verify GitHub Actions workflows run correctly and identify issues appropriately.

## Approach

1. **Create a new branch** for the test PR
   - Branch off from the current branch (add-precommit-and-github-actions)
   - Name the branch something descriptive like "test-github-actions"

2. **Prepare test changes** that will trigger different GitHub Actions workflows:
   - **For testing the lint workflow:**
     - Introduce a Go formatting issue (missing or extra spacing)
     - Add a linting error (unused variable, inefficient code)
     - Create a file with trailing whitespace

   - **For testing the test workflow:**
     - Add a failing test case
     - Add a test with a compilation error

   - **For testing the build workflow:**
     - Add code that builds successfully but with warnings
     - Add a README update (to test path-ignore functionality)

3. **Create the Pull Request**
   - Push the branch to GitHub
   - Create a PR targeting the master branch
   - Include a clear PR description explaining this is a test PR to verify GitHub Actions

4. **Document Results**
   - Verify all workflows trigger appropriately
   - Confirm status checks appear on the PR
   - Document any issues encountered during the process

5. **Clean Up**
   - Close the PR without merging
   - Document the PR number for future reference
