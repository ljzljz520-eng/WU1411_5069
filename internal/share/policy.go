package share

import (
	"fmt"
	"time"

	"example.com/temporary-share-gateway/internal/model"
)

type Policy struct {
	MaxUses      int
	MaxLifetime  time.Duration
	RequireLabel bool
}

func DefaultPolicy() Policy {
	return Policy{MaxUses: 100, MaxLifetime: 24 * time.Hour, RequireLabel: false}
}

func (p Policy) Validate(grant model.TokenGrant, now time.Time) error {
	if p.MaxUses < 1 || p.MaxLifetime <= 0 {
		return fmt.Errorf("invalid share policy")
	}
	if grant.Uses > p.MaxUses {
		return fmt.Errorf("uses exceed policy")
	}
	if grant.ExpiresAt.Sub(now) > p.MaxLifetime {
		return fmt.Errorf("lifetime exceeds policy")
	}
	if p.RequireLabel && grant.Label == "" {
		return fmt.Errorf("label is required")
	}
	return nil
}

func (p Policy) RemainingPercent(record model.ShareRecord) int {
	if p.MaxUses <= 0 || record.Remaining <= 0 {
		return 0
	}
	percent := record.Remaining * 100 / p.MaxUses
	if percent > 100 {
		return 100
	}
	return percent
}
