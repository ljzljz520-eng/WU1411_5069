package share

import (
	"context"
	"fmt"
	"strings"
	"time"

	"example.com/temporary-share-gateway/internal/model"
)

type Limiter struct {
	service *ShareService
	limit   int
	window  time.Duration
}

func NewLimiter(service *ShareService, limit int, window time.Duration) *Limiter {
	if limit < 1 {
		limit = 1
	}
	if window <= 0 {
		window = time.Minute
	}
	return &Limiter{service: service, limit: limit, window: window}
}

func (l *Limiter) Allow(ctx context.Context, key string) (bool, model.RateWindow, error) {
	if l == nil || l.service == nil {
		return false, model.RateWindow{}, fmt.Errorf("rate limiter is not configured")
	}
	if err := contextErr(ctx); err != nil {
		return false, model.RateWindow{}, err
	}
	key = normalizeRateKey(key)
	if key == "" {
		return false, model.RateWindow{}, fmt.Errorf("rate key is required")
	}
	now := l.service.clock.Now()
	window, err := l.service.store.LoadRateWindow(key)
	if err != nil {
		if !isMissing(err) {
			return false, model.RateWindow{}, err
		}
		window = model.RateWindow{Key: key, StartedAt: now, Limit: l.limit}
	}
	if window.StartedAt.IsZero() || now.Sub(window.StartedAt) >= l.window {
		window.StartedAt = now
		window.Hits = 0
		window.Limit = l.limit
	}
	if window.Limit != l.limit {
		window.Limit = l.limit
	}
	if window.Hits >= window.Limit {
		return false, window, nil
	}
	window.Hits++
	if err := l.service.store.SaveRateWindow(window); err != nil {
		return false, model.RateWindow{}, err
	}
	return true, window, nil
}

func (l *Limiter) Reset(key string) error {
	if l == nil || l.service == nil {
		return fmt.Errorf("rate limiter is not configured")
	}
	key = normalizeRateKey(key)
	if key == "" {
		return fmt.Errorf("rate key is required")
	}
	return l.service.store.SaveRateWindow(model.RateWindow{Key: key, StartedAt: l.service.clock.Now(), Limit: l.limit})
}

func (l *Limiter) Remaining(window model.RateWindow) int {
	if window.Limit <= 0 {
		return 0
	}
	remaining := window.Limit - window.Hits
	if remaining < 0 {
		return 0
	}
	return remaining
}

func normalizeRateKey(key string) string {
	return strings.TrimSpace(key)
}

func isMissing(err error) bool {
	return err != nil && strings.Contains(err.Error(), "record not found")
}
