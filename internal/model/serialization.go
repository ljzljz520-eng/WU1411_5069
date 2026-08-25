package model

import (
	"fmt"
	"strings"
	"time"
)

type ShareSummary struct {
	Token      string    `json:"token"`
	ResourceID string    `json:"resource_id"`
	Status     string    `json:"status"`
	ExpiresAt  time.Time `json:"expires_at"`
	Remaining  int       `json:"remaining"`
	Version    uint64    `json:"version"`
}

func (r ShareRecord) Summary() ShareSummary {
	return ShareSummary{Token: r.Token, ResourceID: r.ResourceID, Status: r.Status, ExpiresAt: r.ExpiresAt, Remaining: r.Remaining, Version: r.Version}
}

func (r ShareRecord) Validate() error {
	if strings.TrimSpace(r.Token) == "" {
		return fmt.Errorf("record token is required")
	}
	if strings.TrimSpace(r.ResourceID) == "" {
		return fmt.Errorf("record resource is required")
	}
	if !ValidStatus(r.Status) {
		return fmt.Errorf("record status %q is invalid", r.Status)
	}
	if r.ExpiresAt.IsZero() {
		return fmt.Errorf("record expiration is required")
	}
	if r.Remaining < 0 {
		return fmt.Errorf("record remaining cannot be negative")
	}
	return nil
}

func (g TokenGrant) Normalize() TokenGrant {
	g.Token = strings.TrimSpace(g.Token)
	g.ResourceID = strings.Trim(strings.TrimSpace(g.ResourceID), "/")
	g.Label = strings.TrimSpace(g.Label)
	g.ExpiresAt = g.ExpiresAt.UTC()
	return g
}

func (e AccessEvent) Validate() error {
	if e.ID == "" || e.Token == "" {
		return fmt.Errorf("event identity is required")
	}
	if !ValidOutcome(e.Outcome) {
		return fmt.Errorf("event outcome is invalid")
	}
	if e.At.IsZero() {
		return fmt.Errorf("event timestamp is required")
	}
	return nil
}

func StatusLabel(status string) string {
	switch status {
	case StatusActive:
		return "available"
	case StatusExpired:
		return "expired"
	case StatusRevoked:
		return "revoked"
	default:
		return "unknown"
	}
}

func RemainingLabel(remaining int) string {
	if remaining <= 0 {
		return "depleted"
	}
	if remaining == 1 {
		return "last-use"
	}
	return "available"
}

func TruncateID(value string, limit int) string {
	if limit < 1 {
		return ""
	}
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}

func SameResource(left, right string) bool {
	return strings.Trim(strings.TrimSpace(left), "/") == strings.Trim(strings.TrimSpace(right), "/")
}

func Lifetime(record ShareRecord) time.Duration {
	if record.CreatedAt.IsZero() || record.ExpiresAt.Before(record.CreatedAt) {
		return 0
	}
	return record.ExpiresAt.Sub(record.CreatedAt)
}
