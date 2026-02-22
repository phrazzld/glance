package llm

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExponentialBackoffJitter(t *testing.T) {
	const samples = 100
	base := 100 * time.Millisecond
	maxWait := 10 * time.Second

	results := make([]time.Duration, samples)
	for i := range results {
		results[i] = ExponentialBackoff(1, base, maxWait)
	}

	// All samples must be within ±20% of base.
	lo, hi := 80*time.Millisecond, 120*time.Millisecond
	for _, d := range results {
		assert.GreaterOrEqual(t, d, lo)
		assert.LessOrEqual(t, d, hi)
	}

	// Jitter must produce variance: at least two distinct values across 100 draws.
	first := results[0]
	allSame := true
	for _, d := range results[1:] {
		if d != first {
			allSame = false
			break
		}
	}
	assert.False(t, allSame, "expected jitter to produce distinct values across %d samples", samples)
}

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

	t.Run("returns zero for non-positive base", func(t *testing.T) {
		assert.Equal(t, time.Duration(0), ExponentialBackoff(1, 0, maxWait))
		assert.Equal(t, time.Duration(0), ExponentialBackoff(1, -time.Second, maxWait))
	})

	t.Run("returns zero for non-positive maxWait", func(t *testing.T) {
		assert.Equal(t, time.Duration(0), ExponentialBackoff(1, base, 0))
		assert.Equal(t, time.Duration(0), ExponentialBackoff(1, base, -time.Second))
	})

	t.Run("caps when base exceeds maxWait", func(t *testing.T) {
		wait := ExponentialBackoff(1, 5*time.Second, time.Second)
		// base > maxWait: doubling loop doesn't run, wait caps to maxWait
		assert.LessOrEqual(t, wait, time.Second)
	})

	t.Run("large attempt does not overflow", func(t *testing.T) {
		wait := ExponentialBackoff(100, base, maxWait)
		// Must cap at maxWait regardless of attempt count
		assert.GreaterOrEqual(t, wait, 8*time.Second) // maxWait * 0.8
		assert.LessOrEqual(t, wait, maxWait)
	})
}
