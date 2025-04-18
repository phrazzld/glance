package githubactionstest

import (
	"testing"
)

// TestCompilationError contains a deliberate syntax error
func TestCompilationError(t *testing.T) {
	// Missing closing parenthesis - this won't compile
	t.Log("This has a syntax error"
	
	// Undefined variable
	t.Log(undefinedVariable)
}