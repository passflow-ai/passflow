package llm

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"
)

// RetryableError wraps an error with retryability information.
type RetryableError struct {
	Err       error
	Retryable bool
}

// Error implements the error interface.
func (e *RetryableError) Error() string {
	return e.Err.Error()
}

// Unwrap returns the underlying error.
func (e *RetryableError) Unwrap() error {
	return e.Err
}

// ResilientClient wraps an LLM client with retry and fallback capabilities.
type ResilientClient struct {
	primary   Client
	fallbacks []Client

	maxRetries int
	baseDelay  time.Duration
	maxDelay   time.Duration
}

// ResilientConfig configures the resilient client behavior.
type ResilientConfig struct {
	MaxRetries int           // default: 3
	BaseDelay  time.Duration // default: 1s
	MaxDelay   time.Duration // default: 30s
}

// NewResilientClient creates a new resilient client wrapper.
func NewResilientClient(primary Client, fallbacks []Client, config ResilientConfig) *ResilientClient {
	maxRetries := config.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}

	baseDelay := config.BaseDelay
	if baseDelay <= 0 {
		baseDelay = time.Second
	}

	maxDelay := config.MaxDelay
	if maxDelay <= 0 {
		maxDelay = 30 * time.Second
	}

	return &ResilientClient{
		primary:    primary,
		fallbacks:  fallbacks,
		maxRetries: maxRetries,
		baseDelay:  baseDelay,
		maxDelay:   maxDelay,
	}
}

// Complete implements Client with retry and fallback.
func (c *ResilientClient) Complete(ctx context.Context, messages []Message, tools []ToolDefinition) (*Response, error) {
	// Build list of all clients to try (primary + fallbacks)
	clients := make([]Client, 0, 1+len(c.fallbacks))
	clients = append(clients, c.primary)
	clients = append(clients, c.fallbacks...)

	var lastErr error
	for i, client := range clients {
		providerName := fmt.Sprintf("provider-%d", i)
		resp, err := c.tryWithRetry(ctx, client, messages, tools, providerName)
		if err == nil {
			return resp, nil
		}
		lastErr = err

		// Check if context is done before trying next fallback
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Log fallback attempt
		if i < len(clients)-1 {
			log.Printf("resilient: %s failed after retries, trying fallback provider-%d: %v", providerName, i+1, err)
		}
	}

	return nil, fmt.Errorf("resilient: all providers failed: %w", lastErr)
}

// tryWithRetry attempts to complete with retries using exponential backoff.
func (c *ResilientClient) tryWithRetry(
	ctx context.Context,
	client Client,
	messages []Message,
	tools []ToolDefinition,
	providerName string,
) (*Response, error) {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		// Check context before each attempt
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		resp, err := client.Complete(ctx, messages, tools)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryableError(err) {
			log.Printf("resilient: %s non-retryable error: %v", providerName, err)
			return nil, err
		}

		// Don't wait after the last attempt
		if attempt == c.maxRetries {
			break
		}

		// Calculate backoff and wait
		backoff := calculateBackoff(attempt, c.baseDelay, c.maxDelay)
		log.Printf("resilient: %s attempt %d failed, retrying in %v: %v", providerName, attempt+1, backoff, err)

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
			// Continue to next attempt
		}
	}

	return nil, lastErr
}

// isRetryableError determines if an error should trigger a retry.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Context errors are not retryable
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	// Check for RetryableError type
	var retryableErr *RetryableError
	if errors.As(err, &retryableErr) {
		return retryableErr.Retryable
	}

	// Default: treat unknown errors as retryable (transient network issues, etc.)
	return true
}

// calculateBackoff computes the delay for a retry attempt using exponential backoff with jitter.
func calculateBackoff(attempt int, baseDelay, maxDelay time.Duration) time.Duration {
	// Exponential backoff: baseDelay * 2^attempt
	multiplier := math.Pow(2, float64(attempt))
	delay := time.Duration(float64(baseDelay) * multiplier)

	// Cap at maxDelay
	if delay > maxDelay {
		delay = maxDelay
	}

	// Add jitter (10-20% of delay)
	jitter := time.Duration(float64(delay) * (0.1 + rand.Float64()*0.1))
	return delay + jitter
}
