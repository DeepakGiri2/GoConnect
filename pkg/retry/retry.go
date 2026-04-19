package retry

import (
	"database/sql"
	"errors"
	"time"

	"github.com/goconnect/pkg/config"
)

type RetryableFunc func() error

func WithRetry(fn RetryableFunc, cfg config.RetryConfig) error {
	var lastErr error
	backoff := cfg.InitialBackoff

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		if !isRetryable(err) {
			return err
		}

		if attempt < cfg.MaxRetries {
			time.Sleep(backoff)
			backoff *= 2
			if backoff > cfg.MaxBackoff {
				backoff = cfg.MaxBackoff
			}
		}
	}

	return lastErr
}

func isRetryable(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, sql.ErrConnDone) {
		return true
	}

	errMsg := err.Error()
	retryableErrors := []string{
		"connection refused",
		"connection reset",
		"broken pipe",
		"timeout",
		"temporary",
		"deadlock",
	}

	for _, retryable := range retryableErrors {
		if contains(errMsg, retryable) {
			return true
		}
	}

	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
