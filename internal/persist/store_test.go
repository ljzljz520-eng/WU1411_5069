package persist

import (
	"testing"
	"time"

	"example.com/temporary-share-gateway/internal/model"
)

func TestPersistenceSurvivesReopen(t *testing.T) {
	path := t.TempDir() + "/persistent.db"
	store, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	record := model.ShareRecord{Token: "persisted", ResourceID: "asset-9", ExpiresAt: time.Unix(1700003600, 0).UTC(), Remaining: 4, Status: model.StatusActive}
	if err := store.SaveShare(record); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	if err := store.Reopen(); err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	got, err := store.LoadShare("persisted")
	if err != nil {
		t.Fatal(err)
	}
	if got.ResourceID != record.ResourceID || got.Remaining != record.Remaining {
		t.Fatalf("record = %#v", got)
	}
}

func TestShareListingAndExpiry(t *testing.T) {
	store, err := Open(t.TempDir() + "/listing.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	for _, token := range []string{"b", "a"} {
		if err := store.SaveShare(model.ShareRecord{Token: token, ResourceID: "asset", ExpiresAt: time.Unix(10, 0).UTC(), Remaining: 1, Status: model.StatusActive}); err != nil {
			t.Fatal(err)
		}
	}
	changed, err := store.MarkExpired(time.Unix(20, 0).UTC())
	if err != nil || changed != 2 {
		t.Fatalf("expired = %d %v", changed, err)
	}
	list, err := store.ListShares()
	if err != nil || len(list) != 2 || list[0].Token != "a" {
		t.Fatalf("list = %#v %v", list, err)
	}
}
