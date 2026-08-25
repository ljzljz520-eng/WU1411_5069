package security

import (
	"context"
	"fmt"
	"sync"
	"time"

	"example.com/temporary-share-gateway/internal/model"
	"example.com/temporary-share-gateway/internal/persist"
)

type Auditor struct {
	store *persist.Store
	mu    sync.Mutex
	seq   uint64
}

func NewAuditor(store *persist.Store) *Auditor { return &Auditor{store: store} }

func (a *Auditor) Record(ctx context.Context, event model.AccessEvent, token string) error {
	if ctx != nil {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
	a.mu.Lock()
	a.seq++
	sequence := a.seq
	a.mu.Unlock()
	entry := model.AuditEntry{
		ID: fmt.Sprintf("audit-%06d", sequence), RequestID: event.RequestID,
		TokenHash: HashToken(token), ResourceID: event.ResourceID, Action: "authorize",
		Outcome: event.Outcome, Detail: SafeDetail(event.Reason, 160), At: event.At,
	}
	return a.store.SaveAuditEntry(entry)
}

func (a *Auditor) RecordSystem(ctx context.Context, action, detail string, now time.Time) error {
	return a.Record(ctx, model.AccessEvent{Outcome: model.OutcomeAllow, Reason: detail, At: now}, "system:"+action)
}

func (a *Auditor) Entries() ([]model.AuditEntry, error) { return a.store.ListAudits() }
