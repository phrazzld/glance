package githubactionstest

import (
	"fmt"
)

// MainFunction is a simple function that doesn't do much
func MainFunction() {
	fmt.Println("This is a test file for GitHub Actions workflows")
	
	// This will never be executed, but is here to reference the BuildWarning function
	if false {
		BuildWarning()
	}
}