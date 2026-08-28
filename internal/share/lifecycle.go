package share

import (
	"context"
	"fmt"
	"time"

	"example.com/temporary-share-gateway/internal/model"
	"example.com/temporary-share-gateway/internal/security"
)

type Lifecycle struct {
	service *ShareService
	auditor *security.Auditor
}

func NewLifecycle(service *ShareService, auditor *security.Auditor) *Lifecycle {
	return &Lifecycle{service: service, auditor: auditor}
}

func (l *Lifecycle) Sweep(ctx context.Context) (int, error) {
	if l == nil || l.service == nil {
		return 0, fmt.Errorf("lifecycle is not configured")
	}
	changed, err := l.service.Expire(ctx)
	if err != nil {
		return changed, err
	}
	if changed > 0 && l.auditor != nil {
		_ = l.auditor.RecordSystem(ctx, "sweep", fmt.Sprintf("expired=%d", changed), l.service.clock.Now())
	}
	return changed, nil
}

func (l *Lifecycle) RevokeExpired(ctx context.Context) (int, error) {
	if l == nil || l.service == nil {
		return 0, fmt.Errorf("lifecycle is not configured")
	}
	records, err := l.service.store.ListShares()
	if err != nil {
		return 0, err
	}
	changed := 0
	for _, record := range records {
		if record.Status == model.StatusExpired {
			if err := l.service.store.RevokeShare(record.Token); err != nil {
				return changed, err
			}
			changed++
		}
	}
	return changed, nil
}

func (l *Lifecycle) RetainAudits(ctx context.Context, limit int) (int, error) {
	if l == nil || l.service == nil {
		return 0, fmt.Errorf("lifecycle is not configured")
	}
	if err := contextErr(ctx); err != nil {
		return 0, err
	}
	if limit < 1 {
		return 0, fmt.Errorf("retention limit must be positive")
	}
	entries, err := l.service.store.ListAudits()
	if err != nil {
		return 0, err
	}
	if len(entries) <= limit {
		return 0, nil
	}
	removed := 0
	for _, entry := range entries[:len(entries)-limit] {
		if err := l.deleteAudit(entry.ID); err != nil {
			return removed, err
		}
		removed++
	}
	return removed, nil
}

func (l *Lifecycle) deleteAudit(id string) error {
	return l.service.store.DeleteAudit(id)
}

func AuditAge(entry model.AuditEntry, now time.Time) time.Duration {
	if now.Before(entry.At) {
		return 0
	}
	return now.Sub(entry.At)
}

func IsOperational(record model.ShareRecord, now time.Time) bool {
	return record.Active(now) && record.Status != model.StatusExpired
}
