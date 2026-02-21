// Package llm provides abstractions and implementations for interacting with
// Large Language Model APIs in the glance application.
package llm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	customerrors "glance/errors"
)

const (
	defaultFallbackBackoff    = 250 * time.Millisecond
	defaultFallbackMaxBackoff = 4 * time.Second
)

// FallbackTier defines a model/provider tier in a failover chain.
type FallbackTier struct {
	Name   string
	Client Client
}

// FallbackClient tries generation with retries on each tier, then falls back
// to the next tier when a tier is exhausted.
type FallbackClient struct {
	tiers          []FallbackTier
	retriesPerTier int
	baseBackoff    time.Duration
	maxBackoff     time.Duration
}

// NewFallbackClient creates a fallback client with sensible backoff defaults.
func NewFallbackClient(tiers []FallbackTier, retriesPerTier int) (Client, error) {
	return NewFallbackClientWithBackoff(
		tiers,
		retriesPerTier,
		defaultFallbackBackoff,
		defaultFallbackMaxBackoff,
	)
}

// NewFallbackClientWithBackoff creates a fallback client with explicit backoff settings.
func NewFallbackClientWithBackoff(
	tiers []FallbackTier,
	retriesPerTier int,
	baseBackoff time.Duration,
	maxBackoff time.Duration,
) (Client, error) {
	if len(tiers) == 0 {
		return nil, customerrors.NewValidationError("at least one fallback tier is required", nil).
			WithCode("LLM-001")
	}
	if retriesPerTier < 0 {
		return nil, customerrors.NewValidationError("retries per tier cannot be negative", nil).
			WithCode("LLM-002")
	}
	if baseBackoff <= 0 {
		return nil, customerrors.NewValidationError("base backoff must be greater than zero", nil).
			WithCode("LLM-003")
	}
	if maxBackoff <= 0 {
		return nil, customerrors.NewValidationError("max backoff must be greater than zero", nil).
			WithCode("LLM-004")
	}

	cleanTiers := make([]FallbackTier, 0, len(tiers))
	for i, tier := range tiers {
		if tier.Client == nil {
			return nil, customerrors.NewValidationError(
				fmt.Sprintf("fallback tier %d has nil client", i),
				nil,
			).WithCode("LLM-005")
		}

		name := strings.TrimSpace(tier.Name)
		if name == "" {
			name = fmt.Sprintf("tier-%d", i+1)
		}

		cleanTiers = append(cleanTiers, FallbackTier{
			Name:   name,
			Client: tier.Client,
		})
	}

	return &FallbackClient{
		tiers:          cleanTiers,
		retriesPerTier: retriesPerTier,
		baseBackoff:    baseBackoff,
		maxBackoff:     maxBackoff,
	}, nil
}

// Generate tries each fallback tier with exponential backoff retries.
func (c *FallbackClient) Generate(ctx context.Context, prompt string) (string, error) {
	var lastErr error
	maxAttempts := c.retriesPerTier + 1

	for tierIdx, tier := range c.tiers {
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			if ctx.Err() != nil {
				return "", ctx.Err()
			}

			result, err := tier.Client.Generate(ctx, prompt)
			if err == nil {
				if tierIdx > 0 || attempt > 1 {
					logrus.WithFields(logrus.Fields{
						"tier_name":       tier.Name,
						"tier_index":      tierIdx + 1,
						"tier_count":      len(c.tiers),
						"attempt":         attempt,
						"attempts_tier":   maxAttempts,
						"retries_tier":    c.retriesPerTier,
						"failover_used":   tierIdx > 0,
						"tier_retry_used": attempt > 1,
					}).Info("LLM generation succeeded after retry/failover")
				}
				return result, nil
			}

			lastErr = err

			logFields := logrus.Fields{
				"tier_name":       tier.Name,
				"tier_index":      tierIdx + 1,
				"tier_count":      len(c.tiers),
				"attempt":         attempt,
				"attempts_tier":   maxAttempts,
				"retries_tier":    c.retriesPerTier,
				"error":           err,
				"will_failover":   attempt == maxAttempts && tierIdx < len(c.tiers)-1,
				"will_retry_tier": attempt < maxAttempts,
			}

			if attempt < maxAttempts {
				wait := c.retryBackoff(attempt)
				logFields["backoff_ms"] = wait.Milliseconds()
				logrus.WithFields(logFields).Warn("LLM tier attempt failed, retrying tier")

				if sleepErr := sleepWithContext(ctx, wait); sleepErr != nil {
					return "", sleepErr
				}
				continue
			}

			logrus.WithFields(logFields).Warn("LLM tier exhausted, trying fallback tier")
		}
	}

	return "", customerrors.WrapAPIError(lastErr, "all LLM fallback tiers failed").
		WithCode("LLM-006").
		WithSuggestion("Check provider connectivity, API keys, or reduce prompt size")
}

// CountTokens attempts token counting across tiers until one succeeds.
func (c *FallbackClient) CountTokens(ctx context.Context, prompt string) (int, error) {
	var lastErr error
	for _, tier := range c.tiers {
		tokens, err := tier.Client.CountTokens(ctx, prompt)
		if err == nil {
			return tokens, nil
		}
		lastErr = err
	}

	return 0, customerrors.WrapAPIError(lastErr, "failed to count tokens across fallback tiers").
		WithCode("LLM-007")
}

// GenerateStream attempts streaming from each tier until one starts successfully.
func (c *FallbackClient) GenerateStream(ctx context.Context, prompt string) (<-chan StreamChunk, error) {
	var lastErr error
	for _, tier := range c.tiers {
		stream, err := tier.Client.GenerateStream(ctx, prompt)
		if err == nil {
			return stream, nil
		}
		lastErr = err
	}

	return nil, customerrors.WrapAPIError(lastErr, "failed to start streaming across fallback tiers").
		WithCode("LLM-008")
}

// Close closes all underlying clients.
func (c *FallbackClient) Close() {
	for _, tier := range c.tiers {
		tier.Client.Close()
	}
}

func (c *FallbackClient) retryBackoff(attempt int) time.Duration {
	if attempt <= 0 {
		return c.baseBackoff
	}

	backoff := c.baseBackoff * time.Duration(1<<(attempt-1))
	if backoff > c.maxBackoff {
		return c.maxBackoff
	}
	return backoff
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
