package share

import (
	"context"
	"testing"
	"time"

	"example.com/temporary-share-gateway/internal/clock"
	"example.com/temporary-share-gateway/internal/model"
	"example.com/temporary-share-gateway/internal/persist"
	"example.com/temporary-share-gateway/internal/security"
)

func TestRegistryLifecycleAndRateLimit(t *testing.T) {
	store, err := persist.Open(t.TempDir() + "/registry.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	now := clock.NewFixed(clock.Unix(1700000000))
	service := NewService(store, now, AllowAllResources{})
	registry := NewRegistry(service, DefaultPolicy())
	_, err = registry.Issue(context.Background(), model.TokenGrant{Token: "registry-token", ResourceID: "asset", Label: "download", ExpiresAt: now.Now().Add(time.Hour), Uses: 3})
	if err != nil || registry.Label("registry-token") != "download" {
		t.Fatalf("issue failed: %v", err)
	}
	list, err := registry.List(context.Background())
	if err != nil || len(list) != 1 {
		t.Fatalf("list = %#v %v", list, err)
	}
	if _, err := registry.Rotate(context.Background(), "registry-token", now.Now().Add(2*time.Hour), 2); err != nil {
		t.Fatal(err)
	}
	limiter := NewLimiter(service, 1, time.Minute)
	allowed, _, err := limiter.Allow(context.Background(), "client")
	if err != nil || !allowed {
		t.Fatalf("first rate call: %v", err)
	}
	allowed, _, err = limiter.Allow(context.Background(), "client")
	if err != nil || allowed {
		t.Fatal("rate limit did not reject")
	}
	lifecycle := NewLifecycle(service, security.NewAuditor(store))
	if _, err := lifecycle.Sweep(context.Background()); err != nil {
		t.Fatal(err)
	}
}
