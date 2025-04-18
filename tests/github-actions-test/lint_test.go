package githubactionstest

import (
	"testing"
)

// TestLintIssues contains deliberate linting issues
func TestLintIssues(t *testing.T) {
	// Unused variable (golangci-lint should catch this)
	unusedVar := "I'm not used anywhere"
	
	// Formatting issue - extra spacing (go fmt should catch this)
	result :=  "badly formatted"
	
	// Inefficient string concatenation (golangci-lint should catch this)
	s := ""
	for i := 0; i < 10; i++ {
		s = s + "a"
	}
	
	t.Log("This test doesn't actually test anything")
}