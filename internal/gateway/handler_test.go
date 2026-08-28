package gateway

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"example.com/temporary-share-gateway/internal/clock"
	"example.com/temporary-share-gateway/internal/metrics"
	"example.com/temporary-share-gateway/internal/model"
	"example.com/temporary-share-gateway/internal/persist"
	"example.com/temporary-share-gateway/internal/security"
	"example.com/temporary-share-gateway/internal/share"
)

type handlerFixture struct {
	service *share.ShareService
	h       *Handler
	store   *persist.Store
	now     *clock.FixedClock
}

func newHandlerFixture(t *testing.T) handlerFixture {
	t.Helper()
	store, err := persist.Open(t.TempDir() + "/gateway.db")
	if err != nil {
		t.Fatal(err)
	}
	now := clock.NewFixed(clock.Unix(1700000000))
	resources := StaticResources{"asset-1": "payload"}
	service := share.NewService(store, now, resources)
	handler := NewHandler(service, security.NewAuditor(store), resources, metrics.New(), false)
	t.Cleanup(func() { _ = store.Close() })
	return handlerFixture{service: service, h: handler, store: store, now: now}
}

func TestPrimaryWorkflow(t *testing.T) {
	fixture := newHandlerFixture(t)
	_, err := fixture.service.Create(context.Background(), model.TokenGrant{Token: "http-token", ResourceID: "asset-1", ExpiresAt: fixture.now.Now().Add(time.Hour), Uses: 2})
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/share/asset-1", strings.NewReader(""))
	request.Header.Set("X-Share-Token", "http-token")
	request.Header.Set("X-Request-ID", "r-1")
	recorder := httptest.NewRecorder()
	fixture.h.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || recorder.Body.String() != "payload" {
		t.Fatalf("response = %d %q", recorder.Code, recorder.Body.String())
	}
}

func TestSecondaryWorkflow(t *testing.T) {
	fixture := newHandlerFixture(t)
	_, err := fixture.service.Create(context.Background(), model.TokenGrant{Token: "head-token", ResourceID: "asset-1", ExpiresAt: fixture.now.Now().Add(time.Hour), Uses: 1})
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodHead, "/share/asset-1", nil)
	request.Header.Set("X-Share-Token", "head-token")
	recorder := httptest.NewRecorder()
	fixture.h.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || recorder.Body.Len() != 0 {
		t.Fatalf("head response = %d %q", recorder.Code, recorder.Body.String())
	}
	if fixture.store.ValidateReady() != nil {
		t.Fatal("store not ready")
	}
}

func TestTertiaryWorkflow(t *testing.T) {
	fixture := newHandlerFixture(t)
	_, err := fixture.service.Create(context.Background(), model.TokenGrant{Token: "deny-token", ResourceID: "asset-1", ExpiresAt: fixture.now.Now().Add(time.Hour), Uses: 1})
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/share/asset-1", nil)
	recorder := httptest.NewRecorder()
	fixture.h.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("denial status = %d", recorder.Code)
	}
	entries, err := security.NewAuditor(fixture.store).Entries()
	if err != nil || len(entries) != 1 || entries[0].Outcome != model.OutcomeDeny {
		t.Fatalf("audit = %#v %v", entries, err)
	}
}

func TestInvalidResourceRequest(t *testing.T) {
	fixture := newHandlerFixture(t)
	request := httptest.NewRequest(http.MethodPost, "/share/asset-1", nil)
	recorder := httptest.NewRecorder()
	fixture.h.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", recorder.Code)
	}
}
