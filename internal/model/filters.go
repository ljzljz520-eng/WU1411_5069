package model

import (
	"sort"
	"strings"
	"time"
)

type ShareFilter struct {
	Status        string
	ResourceID    string
	TokenPrefix   string
	ExpiresBefore time.Time
	AvailableOnly bool
}

func (f ShareFilter) Match(record ShareRecord, now time.Time) bool {
	if f.Status != "" && record.Status != f.Status {
		return false
	}
	if f.ResourceID != "" && !SameResource(record.ResourceID, f.ResourceID) {
		return false
	}
	if f.TokenPrefix != "" && !strings.HasPrefix(record.Token, f.TokenPrefix) {
		return false
	}
	if !f.ExpiresBefore.IsZero() && !record.ExpiresAt.Before(f.ExpiresBefore) {
		return false
	}
	if f.AvailableOnly && !record.Active(now) {
		return false
	}
	return true
}

func FilterShares(records []ShareRecord, filter ShareFilter, now time.Time) []ShareRecord {
	result := make([]ShareRecord, 0, len(records))
	for _, record := range records {
		if filter.Match(record, now) {
			result = append(result, record)
		}
	}
	return result
}

func SortShares(records []ShareRecord, newestFirst bool) {
	sort.SliceStable(records, func(i, j int) bool {
		if newestFirst {
			return records[i].CreatedAt.After(records[j].CreatedAt)
		}
		return records[i].CreatedAt.Before(records[j].CreatedAt)
	})
}

func SortSummaries(summaries []ShareSummary, byRemaining bool) {
	sort.SliceStable(summaries, func(i, j int) bool {
		if byRemaining && summaries[i].Remaining != summaries[j].Remaining {
			return summaries[i].Remaining > summaries[j].Remaining
		}
		return summaries[i].Token < summaries[j].Token
	})
}

func ParseStatus(value string) (string, bool) {
	value = strings.ToLower(strings.TrimSpace(value))
	if !ValidStatus(value) {
		return "", false
	}
	return value, true
}

func FilterSummaries(summaries []ShareSummary, status string) []ShareSummary {
	if status == "" {
		return append([]ShareSummary(nil), summaries...)
	}
	filtered := make([]ShareSummary, 0, len(summaries))
	for _, summary := range summaries {
		if summary.Status == status {
			filtered = append(filtered, summary)
		}
	}
	return filtered
}

func CountAvailable(records []ShareRecord, now time.Time) int {
	count := 0
	for _, record := range records {
		if record.Active(now) {
			count++
		}
	}
	return count
}
