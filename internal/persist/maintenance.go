package persist

import (
	"errors"
	"fmt"
	"time"

	"example.com/temporary-share-gateway/internal/model"
	"go.etcd.io/bbolt"
)

type IntegrityReport struct {
	Shares         int
	Audits         int
	Configurations int
	Active         int
	Expired        int
	Revoked        int
}

func (s *Store) Integrity(now time.Time) (IntegrityReport, error) {
	if err := s.ValidateReady(); err != nil {
		return IntegrityReport{}, err
	}
	records, err := s.ListShares()
	if err != nil {
		return IntegrityReport{}, err
	}
	audits, err := s.ListAudits()
	if err != nil {
		return IntegrityReport{}, err
	}
	configs, err := s.listConfigs()
	if err != nil {
		return IntegrityReport{}, err
	}
	report := IntegrityReport{Shares: len(records), Audits: len(audits), Configurations: len(configs)}
	for _, record := range records {
		if err := record.Validate(); err != nil {
			return report, fmt.Errorf("share %s: %w", record.Token, err)
		}
		switch {
		case record.Status == model.StatusActive && !record.IsExpired(now) && !record.IsExhausted():
			report.Active++
		case record.Status == model.StatusExpired || record.IsExpired(now):
			report.Expired++
		case record.Status == model.StatusRevoked:
			report.Revoked++
		}
	}
	return report, nil
}

func (s *Store) PruneAuditsBefore(cutoff time.Time) (int, error) {
	if cutoff.IsZero() {
		return 0, errors.New("audit cutoff is required")
	}
	entries, err := s.ListAudits()
	if err != nil {
		return 0, err
	}
	removed := 0
	for _, entry := range entries {
		if entry.At.Before(cutoff) {
			if err := s.DeleteAudit(entry.ID); err != nil {
				return removed, err
			}
			removed++
		}
	}
	return removed, nil
}

func (s *Store) CountByStatus(status string) (int, error) {
	if !model.ValidStatus(status) {
		return 0, fmt.Errorf("invalid status %q", status)
	}
	records, err := s.ListShares()
	if err != nil {
		return 0, err
	}
	count := 0
	for _, record := range records {
		if record.Status == status {
			count++
		}
	}
	return count, nil
}

func (s *Store) DeleteAllShares() (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return 0, errors.New("store is closed")
	}
	removed := 0
	err := s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketNames.Shares)
		return bucket.ForEach(func(key, _ []byte) error {
			if err := bucket.Delete(key); err != nil {
				return err
			}
			removed++
			return nil
		})
	})
	return removed, err
}
