package llm

import (
	"crypto/rand"
	"encoding/binary"
	"math"
	"time"
)

const jitterRatio = 0.20

// ExponentialBackoff returns wait time for the given 1-based attempt number.
// It uses base*2^(attempt-1), caps at maxWait, and applies up to 20% jitter.
func ExponentialBackoff(attempt int, base, maxWait time.Duration) time.Duration {
	if base <= 0 || maxWait <= 0 {
		return 0
	}

	if attempt <= 0 {
		attempt = 1
	}

	wait := base
	for i := 1; i < attempt && wait < maxWait; i++ {
		if wait > maxWait/2 {
			wait = maxWait
			break
		}
		wait *= 2
	}
	if wait > maxWait {
		wait = maxWait
	}

	jittered := applyJitter(wait)
	if jittered > maxWait {
		return maxWait
	}
	if jittered < 0 {
		return 0
	}
	return jittered
}

func applyJitter(wait time.Duration) time.Duration {
	f, err := randomFraction()
	if err != nil {
		return wait
	}

	// Random multiplier in [0.8, 1.2].
	multiplier := 1 - jitterRatio + (2*jitterRatio)*f
	return time.Duration(float64(wait) * multiplier)
}

func randomFraction() (float64, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0, err
	}

	v := binary.LittleEndian.Uint64(b[:])
	return float64(v) / float64(math.MaxUint64), nil
}
