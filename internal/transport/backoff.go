package transport

import (
	"context"
	"fmt"
	"time"
)

type RetryPolicy struct {
	Attempts int
	Delay    time.Duration
	MaxDelay time.Duration
}

func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{Attempts: 3, Delay: 10 * time.Millisecond, MaxDelay: 250 * time.Millisecond}
}

func (p RetryPolicy) Validate() error {
	if p.Attempts < 1 || p.Delay < 0 || p.MaxDelay < p.Delay {
		return fmt.Errorf("retry policy is invalid")
	}
	return nil
}

func RunWithRetry(ctx context.Context, policy RetryPolicy, operation func() error) error {
	if err := policy.Validate(); err != nil {
		return err
	}
	if operation == nil {
		return fmt.Errorf("retry operation is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	var last error
	delay := policy.Delay
	for attempt := 0; attempt < policy.Attempts; attempt++ {
		if err := contextError(ctx); err != nil {
			return err
		}
		if err := operation(); err == nil {
			return nil
		} else {
			last = err
		}
		if attempt+1 < policy.Attempts && delay > 0 {
			timer := time.NewTimer(delay)
			select {
			case <-timer.C:
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			}
			if delay < policy.MaxDelay {
				delay *= 2
				if delay > policy.MaxDelay {
					delay = policy.MaxDelay
				}
			}
		}
	}
	return fmt.Errorf("operation failed after %d attempts: %w", policy.Attempts, last)
}

func contextError(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func BackoffSequence(policy RetryPolicy) []time.Duration {
	if policy.Validate() != nil {
		return nil
	}
	delays := make([]time.Duration, 0, policy.Attempts)
	delay := policy.Delay
	for index := 0; index < policy.Attempts; index++ {
		delays = append(delays, delay)
		if delay < policy.MaxDelay {
			delay *= 2
			if delay > policy.MaxDelay {
				delay = policy.MaxDelay
			}
		}
	}
	return delays
}
