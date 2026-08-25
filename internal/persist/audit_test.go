package persist

import (
	"testing"
	"time"

	"example.com/temporary-share-gateway/internal/model"
)

func TestAuditAndConfigRoundTrip(t *testing.T) {
	store, err := Open(t.TempDir() + "/audit.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	config := model.GatewayConfig{Name: "gateway", MaxBodyBytes: 100, RequestTimeout: time.Second, AuditRetention: 10}
	if err := store.SaveConfig(config); err != nil {
		t.Fatal(err)
	}
	gotConfig, err := store.LoadConfig("gateway")
	if err != nil || gotConfig.Name != config.Name {
		t.Fatalf("config = %#v %v", gotConfig, err)
	}
	entry := model.AuditEntry{ID: "audit-1", Action: "authorize", Outcome: model.OutcomeAllow, At: time.Unix(10, 0).UTC()}
	if err := store.SaveAuditEntry(entry); err != nil {
		t.Fatal(err)
	}
	entries, err := store.ListAudits()
	if err != nil || len(entries) != 1 || entries[0].ID != entry.ID {
		t.Fatalf("entries = %#v %v", entries, err)
	}
}
