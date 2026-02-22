package llm

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExponentialBackoff(t *testing.T) {
	base := 100 * time.Millisecond
	maxWait := 10 * time.Second

	t.Run("attempt 1 uses base with jitter", func(t *testing.T) {
		wait := ExponentialBackoff(1, base, maxWait)
		assert.GreaterOrEqual(t, wait, 80*time.Millisecond)
		assert.LessOrEqual(t, wait, 120*time.Millisecond)
	})

	t.Run("attempt 2 doubles base with jitter", func(t *testing.T) {
		wait := ExponentialBackoff(2, base, maxWait)
		assert.GreaterOrEqual(t, wait, 160*time.Millisecond)
		assert.LessOrEqual(t, wait, 240*time.Millisecond)
	})

	t.Run("caps at max wait", func(t *testing.T) {
		wait := ExponentialBackoff(16, time.Second, 3*time.Second)
		assert.GreaterOrEqual(t, wait, 2400*time.Millisecond)
		assert.LessOrEqual(t, wait, 3*time.Second)
	})

	t.Run("handles zero and negative attempts", func(t *testing.T) {
		testCases := []int{0, -1, -10}
		for _, attempt := range testCases {
			wait := ExponentialBackoff(attempt, base, maxWait)
			assert.GreaterOrEqual(t, wait, 80*time.Millisecond)
			assert.LessOrEqual(t, wait, 120*time.Millisecond)
		}
	})
}
