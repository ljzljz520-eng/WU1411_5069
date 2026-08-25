package share

import (
	"context"
	"sync"
	"testing"
	"time"

	"example.com/temporary-share-gateway/internal/clock"
	"example.com/temporary-share-gateway/internal/model"
	"example.com/temporary-share-gateway/internal/persist"
)

func newShareFixture(t *testing.T) (*ShareService, *persist.Store, *clock.FixedClock) {
	t.Helper()
	store, err := persist.Open(t.TempDir() + "/shares.db")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	now := clock.NewFixed(clock.Unix(1700000000))
	return NewService(store, now, AllowAllResources{}), store, now
}

func TestShareCountAllowsOne(t *testing.T) {
	service, store, now := newShareFixture(t)
	_, err := service.Create(context.Background(), model.TokenGrant{Token: "one-use", ResourceID: "resource-1", ExpiresAt: now.Now().Add(time.Hour), Uses: 1})
	if err != nil {
		t.Fatal(err)
	}
	gate := NewControlledGate(2)
	service.SetGate(gate)
	results := make(chan error, 2)
	var group sync.WaitGroup
	group.Add(2)
	for index := 0; index < 2; index++ {
		go func() {
			defer group.Done()
			_, callErr := service.Authorize(context.Background(), "one-use", "request")
			results <- callErr
		}()
	}
	gate.AwaitArrivals(2)
	group.Wait()
	close(results)
	allowed := 0
	for callErr := range results {
		if callErr == nil {
			allowed++
		}
	}
	if allowed != 1 {
		t.Fatalf("allowed calls = %d", allowed)
	}
	record, err := store.LoadShare("one-use")
	if err != nil {
		t.Fatal(err)
	}
	if record.Remaining != 0 {
		t.Fatalf("remaining = %d", record.Remaining)
	}
}

func TestCreateAndAuthorize(t *testing.T) {
	service, _, now := newShareFixture(t)
	record, err := service.Create(context.Background(), model.TokenGrant{Token: "multi", ResourceID: "resource-1", ExpiresAt: now.Now().Add(time.Hour), Uses: 2})
	if err != nil || record.Remaining != 2 {
		t.Fatalf("create: %#v %#v", record, err)
	}
	used, err := service.Authorize(context.Background(), "multi", "req-1")
	if err != nil || used.Remaining != 1 {
		t.Fatalf("authorize: %#v %#v", used, err)
	}
}

func TestExpiredAndRevokedTokensDenied(t *testing.T) {
	service, _, now := newShareFixture(t)
	_, err := service.Create(context.Background(), model.TokenGrant{Token: "expiring", ResourceID: "resource-1", ExpiresAt: now.Now().Add(time.Minute), Uses: 1})
	if err != nil {
		t.Fatal(err)
	}
	now.Advance(2 * time.Minute)
	if _, err := service.Authorize(context.Background(), "expiring", "req"); err == nil {
		t.Fatal("expired token was accepted")
	}
	_, err = service.Create(context.Background(), model.TokenGrant{Token: "revoked", ResourceID: "resource-1", ExpiresAt: now.Now().Add(time.Hour), Uses: 1})
	if err != nil {
		t.Fatal(err)
	}
	if err := service.Revoke(context.Background(), "revoked"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Authorize(context.Background(), "revoked", "req"); err == nil {
		t.Fatal("revoked token was accepted")
	}
}
