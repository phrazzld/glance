package githubactionstest

import (
	"testing"
)

// TestFailing is a deliberately failing test
func TestFailing(t *testing.T) {
	expected := 10
	actual := 5
	
	// This assertion will fail
	if actual != expected {
		t.Errorf("Expected %d, got %d", expected, actual)
	}
}