package toolutils

import (
	"context"
	stdErrors "errors"
	"strings"
	"time"

	appErrors "github.com/algonius/algonius-wallet/native/pkg/errors"
)

// RetryPolicy controls timeout and retry behavior for tool RPC-style operations.
type RetryPolicy struct {
	MaxAttempts    int
	AttemptTimeout time.Duration
	BaseDelay      time.Duration
	MaxDelay       time.Duration
}

// DefaultRetryPolicy is tuned for lightweight MCP tool calls.
var DefaultRetryPolicy = RetryPolicy{
	MaxAttempts:    3,
	AttemptTimeout: 5 * time.Second,
	BaseDelay:      200 * time.Millisecond,
	MaxDelay:       2 * time.Second,
}

// NormalizeChainName converts chain aliases to a canonical name.
func NormalizeChainName(chain string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(chain))
	switch normalized {
	case "eth", "ethereum":
		return "ethereum", nil
	case "bsc", "binance", "binance smart chain":
		return "bsc", nil
	case "sol", "solana":
		return "solana", nil
	default:
		return "", appErrors.ValidationError("chain", "supported values: ethereum|eth, bsc|binance, solana|sol").WithSuggestion("Set chain to ethereum, bsc, or solana")
	}
}

// ExecuteWithRetry runs fn with per-attempt timeout and exponential backoff on retryable errors.
func ExecuteWithRetry[T any](ctx context.Context, policy RetryPolicy, fn func(context.Context) (T, error)) (T, error) {
	var zero T

	if policy.MaxAttempts <= 0 {
		policy.MaxAttempts = 1
	}
	if policy.BaseDelay <= 0 {
		policy.BaseDelay = 200 * time.Millisecond
	}
	if policy.MaxDelay <= 0 {
		policy.MaxDelay = 2 * time.Second
	}

	var lastErr error
	for attempt := 1; attempt <= policy.MaxAttempts; attempt++ {
		attemptCtx := ctx
		cancel := func() {}
		if policy.AttemptTimeout > 0 {
			attemptCtx, cancel = context.WithTimeout(ctx, policy.AttemptTimeout)
		}

		result, err := fn(attemptCtx)
		cancel()
		if err == nil {
			return result, nil
		}
		lastErr = err

		if !IsRetryableError(err) || attempt == policy.MaxAttempts {
			break
		}

		delay := policy.BaseDelay * time.Duration(1<<(attempt-1))
		if delay > policy.MaxDelay {
			delay = policy.MaxDelay
		}

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return zero, ctx.Err()
		case <-timer.C:
		}
	}

	if lastErr == nil {
		return zero, context.Canceled
	}
	return zero, lastErr
}

// IsRetryableError returns true for transient network/RPC style failures.
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	if stdErrors.Is(err, context.DeadlineExceeded) || stdErrors.Is(err, context.Canceled) {
		return true
	}

	msg := strings.ToLower(err.Error())
	retryableSignals := []string{
		"timeout",
		"temporarily unavailable",
		"connection reset",
		"connection refused",
		"broken pipe",
		"eof",
		"network",
		"rpc",
		"rate limit",
		"too many requests",
		"429",
	}
	for _, signal := range retryableSignals {
		if strings.Contains(msg, signal) {
			return true
		}
	}
	return false
}

// ClassifyError maps runtime errors to standardized application error codes.
func ClassifyError(operation string, err error) *appErrors.Error {
	if err == nil {
		return nil
	}
	if stdErrors.Is(err, context.DeadlineExceeded) || strings.Contains(strings.ToLower(err.Error()), "timeout") {
		return appErrors.TimeoutError(operation)
	}
	if IsRetryableError(err) {
		return appErrors.RPCError(operation, err)
	}
	return appErrors.InternalError(operation, err)
}
