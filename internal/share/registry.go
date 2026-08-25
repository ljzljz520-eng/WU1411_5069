package share

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"example.com/temporary-share-gateway/internal/model"
)

type Registry struct {
	service *ShareService
	policy  Policy
	mu      sync.RWMutex
	labels  map[string]string
}

func NewRegistry(service *ShareService, policy Policy) *Registry {
	return &Registry{service: service, policy: policy, labels: make(map[string]string)}
}

func (r *Registry) Issue(ctx context.Context, grant model.TokenGrant) (model.ShareRecord, error) {
	if r == nil || r.service == nil {
		return model.ShareRecord{}, fmt.Errorf("share registry is not configured")
	}
	grant = grant.Normalize()
	if err := r.policy.Validate(grant, r.service.clock.Now()); err != nil {
		return model.ShareRecord{}, err
	}
	record, err := r.service.Create(ctx, grant)
	if err != nil {
		return model.ShareRecord{}, err
	}
	r.mu.Lock()
	r.labels[record.Token] = grant.Label
	r.mu.Unlock()
	return record, nil
}

func (r *Registry) Label(token string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.labels[strings.TrimSpace(token)]
}

func (r *Registry) Revoke(ctx context.Context, token string) error {
	if r == nil || r.service == nil {
		return fmt.Errorf("share registry is not configured")
	}
	return r.service.Revoke(ctx, token)
}

func (r *Registry) List(ctx context.Context) ([]model.ShareSummary, error) {
	if r == nil || r.service == nil {
		return nil, fmt.Errorf("share registry is not configured")
	}
	if err := contextErr(ctx); err != nil {
		return nil, err
	}
	records, err := r.service.store.ListShares()
	if err != nil {
		return nil, err
	}
	summaries := make([]model.ShareSummary, 0, len(records))
	for _, record := range records {
		summaries = append(summaries, record.Summary())
	}
	model.SortSummaries(summaries, false)
	return summaries, nil
}

func (r *Registry) Rotate(ctx context.Context, token string, expiresAt time.Time, uses int) (model.ShareRecord, error) {
	if r == nil || r.service == nil {
		return model.ShareRecord{}, fmt.Errorf("share registry is not configured")
	}
	old, err := r.service.Lookup(ctx, token)
	if err != nil {
		return model.ShareRecord{}, err
	}
	if expiresAt.IsZero() || !expiresAt.After(r.service.clock.Now()) {
		return model.ShareRecord{}, fmt.Errorf("new expiration must be in the future")
	}
	if uses < 1 || uses > r.policy.MaxUses {
		return model.ShareRecord{}, fmt.Errorf("new use count is outside policy")
	}
	old.ExpiresAt = expiresAt.UTC()
	old.Remaining = uses
	old.Status = model.StatusActive
	old.Version++
	if err := r.service.store.UpdateShare(old); err != nil {
		return model.ShareRecord{}, err
	}
	return old, nil
}
