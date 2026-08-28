package model

import (
	"testing"
	"time"
)

func TestShareFiltersAndLabels(t *testing.T) {
	now := time.Unix(100, 0).UTC()
	records := []ShareRecord{
		{Token: "b", ResourceID: "asset", Status: StatusActive, ExpiresAt: now.Add(time.Hour), Remaining: 2, CreatedAt: now},
		{Token: "a", ResourceID: "other", Status: StatusRevoked, ExpiresAt: now.Add(time.Hour), CreatedAt: now.Add(time.Second)},
	}
	filtered := FilterShares(records, ShareFilter{AvailableOnly: true}, now)
	if len(filtered) != 1 || RemainingLabel(filtered[0].Remaining) != "available" {
		t.Fatal("filter mismatch")
	}
	if StatusLabel(StatusExpired) != "expired" || !SameResource("/asset/", "asset") {
		t.Fatal("label mismatch")
	}
	SortShares(records, true)
	if records[0].Token != "a" {
		t.Fatal("sort mismatch")
	}
}
