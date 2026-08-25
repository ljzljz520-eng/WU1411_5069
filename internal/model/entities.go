package model

import "time"

type ShareRecord struct {
	Token      string    `json:"token"`
	ResourceID string    `json:"resource_id"`
	ExpiresAt  time.Time `json:"expires_at"`
	Remaining  int       `json:"remaining"`
	CreatedAt  time.Time `json:"created_at"`
	LastUsedAt time.Time `json:"last_used_at"`
	Status     string    `json:"status"`
	Version    uint64    `json:"version"`
}

type AccessEvent struct {
	ID         string    `json:"id"`
	Token      string    `json:"token"`
	ResourceID string    `json:"resource_id"`
	Outcome    string    `json:"outcome"`
	Reason     string    `json:"reason"`
	At         time.Time `json:"at"`
	RequestID  string    `json:"request_id"`
}

type GatewayConfig struct {
	Name             string        `json:"name"`
	MaxBodyBytes     int64         `json:"max_body_bytes"`
	RequestTimeout   time.Duration `json:"request_timeout"`
	AuditRetention   int           `json:"audit_retention"`
	RequireRequestID bool          `json:"require_request_id"`
}

type AuditEntry struct {
	ID         string    `json:"id"`
	RequestID  string    `json:"request_id"`
	TokenHash  string    `json:"token_hash"`
	ResourceID string    `json:"resource_id"`
	Action     string    `json:"action"`
	Outcome    string    `json:"outcome"`
	Detail     string    `json:"detail"`
	At         time.Time `json:"at"`
}

type RateWindow struct {
	Key       string    `json:"key"`
	StartedAt time.Time `json:"started_at"`
	Hits      int       `json:"hits"`
	Limit     int       `json:"limit"`
}

type TokenGrant struct {
	Token      string    `json:"token"`
	ResourceID string    `json:"resource_id"`
	ExpiresAt  time.Time `json:"expires_at"`
	Uses       int       `json:"uses"`
	Label      string    `json:"label"`
}

func (r ShareRecord) IsExpired(now time.Time) bool {
	return !now.Before(r.ExpiresAt)
}

func (r ShareRecord) IsExhausted() bool {
	return r.Remaining <= 0
}

func (r ShareRecord) Active(now time.Time) bool {
	return r.Status == "active" && !r.IsExpired(now) && !r.IsExhausted()
}

func (r ShareRecord) Clone() ShareRecord {
	return r
}

func (c GatewayConfig) Validate() error {
	if c.Name == "" {
		return ErrInvalidConfig("name is required")
	}
	if c.MaxBodyBytes <= 0 {
		return ErrInvalidConfig("max body must be positive")
	}
	if c.RequestTimeout <= 0 {
		return ErrInvalidConfig("timeout must be positive")
	}
	if c.AuditRetention < 1 {
		return ErrInvalidConfig("audit retention must be positive")
	}
	return nil
}

type invalidConfigError string

func (e invalidConfigError) Error() string { return string(e) }

func ErrInvalidConfig(reason string) error { return invalidConfigError(reason) }
