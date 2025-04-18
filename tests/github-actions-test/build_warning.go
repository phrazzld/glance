package githubactionstest

import (
	"fmt"
)

// BuildWarning function contains code that would generate build warnings
func BuildWarning() {
	// This variable is never used - should generate a warning
	warningVar := "I will cause a warning"
	
	// This will generate "possible misuse of fmt.Println" warning
	fmt.Println("Warning", "test")
}