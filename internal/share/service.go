package share

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"example.com/temporary-share-gateway/internal/clock"
	"example.com/temporary-share-gateway/internal/model"
	"example.com/temporary-share-gateway/internal/persist"
)

type ResourceCatalog interface {
	Exists(resourceID string) bool
}

type AllowAllResources struct{}

func (AllowAllResources) Exists(resourceID string) bool { return resourceID != "" }

type ShareService struct {
	store     *persist.Store
	clock     clock.Clock
	resources ResourceCatalog
	gate      Gate
	sequence  uint64
}

func NewService(store *persist.Store, now clock.Clock, resources ResourceCatalog) *ShareService {
	if now == nil {
		now = clock.NewFixed(clock.Unix(1700000000))
	}
	if resources == nil {
		resources = AllowAllResources{}
	}
	return &ShareService{store: store, clock: now, resources: resources, gate: NoopGate{}}
}

func (s *ShareService) SetGate(gate Gate) {
	if gate == nil {
		s.gate = NoopGate{}
		return
	}
	s.gate = gate
}

func (s *ShareService) Create(ctx context.Context, grant model.TokenGrant) (model.ShareRecord, error) {
	if err := contextErr(ctx); err != nil {
		return model.ShareRecord{}, err
	}
	grant = grant.Normalize()
	if err := validateGrant(grant, s.clock.Now()); err != nil {
		return model.ShareRecord{}, err
	}
	s.sequence++
	record := model.ShareRecord{
		Token: grant.Token, ResourceID: grant.ResourceID, ExpiresAt: grant.ExpiresAt,
		Remaining: grant.Uses, CreatedAt: s.clock.Now(), Status: model.StatusActive,
		Version: s.sequence,
	}
	if err := record.Validate(); err != nil {
		return model.ShareRecord{}, err
	}
	if err := s.store.SaveShare(record); err != nil {
		return model.ShareRecord{}, err
	}
	return record, nil
}

func (s *ShareService) Lookup(ctx context.Context, token string) (model.ShareRecord, error) {
	if err := contextErr(ctx); err != nil {
		return model.ShareRecord{}, err
	}
	record, err := s.store.LoadShare(strings.TrimSpace(token))
	if err != nil {
		if errors.Is(err, persist.ErrNotFound) {
			return model.ShareRecord{}, denial(ErrTokenInvalid, 404)
		}
		return model.ShareRecord{}, err
	}
	return record, nil
}

func (s *ShareService) Authorize(ctx context.Context, token, requestID string) (model.ShareRecord, error) {
	if err := contextErr(ctx); err != nil {
		return model.ShareRecord{}, err
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return model.ShareRecord{}, denial(ErrTokenMissing, 401)
	}
	record, err := s.store.LoadShare(token)
	if err != nil {
		if errors.Is(err, persist.ErrNotFound) {
			return model.ShareRecord{}, denial(ErrTokenInvalid, 404)
		}
		return model.ShareRecord{}, err
	}
	now := s.clock.Now()
	if record.Status == model.StatusRevoked {
		return model.ShareRecord{}, denial(ErrRevoked, 410)
	}
	if record.IsExpired(now) {
		record.Status = model.StatusExpired
		_ = s.store.UpdateShare(record)
		return model.ShareRecord{}, denial(ErrExpired, 410)
	}
	if record.Remaining <= 0 {
		return model.ShareRecord{}, denial(ErrExhausted, 429)
	}
	if !s.resources.Exists(record.ResourceID) {
		return model.ShareRecord{}, denial(ErrResourceGone, 404)
	}
	s.gate.Wait()
	record.Remaining--
	record.LastUsedAt = now
	record.Version++
	if record.Remaining == 0 {
		record.Status = model.StatusExpired
	}
	if err := s.store.UpdateShare(record); err != nil {
		return model.ShareRecord{}, err
	}
	return record, nil
}

func (s *ShareService) Revoke(ctx context.Context, token string) error {
	if err := contextErr(ctx); err != nil {
		return err
	}
	return s.store.RevokeShare(strings.TrimSpace(token))
}

func (s *ShareService) Expire(ctx context.Context) (int, error) {
	if err := contextErr(ctx); err != nil {
		return 0, err
	}
	return s.store.MarkExpired(s.clock.Now())
}

func (s *ShareService) ActiveCount(ctx context.Context) (int, error) {
	if err := contextErr(ctx); err != nil {
		return 0, err
	}
	return s.store.CountActive(s.clock.Now())
}

func validateGrant(grant model.TokenGrant, now time.Time) error {
	if strings.TrimSpace(grant.Token) == "" {
		return fmt.Errorf("token is required")
	}
	if strings.TrimSpace(grant.ResourceID) == "" {
		return fmt.Errorf("resource id is required")
	}
	if grant.ExpiresAt.IsZero() || !grant.ExpiresAt.After(now) {
		return fmt.Errorf("expiration must be in the future")
	}
	if grant.Uses < 1 {
		return fmt.Errorf("uses must be positive")
	}
	return nil
}

func contextErr(ctx context.Context) error {
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
