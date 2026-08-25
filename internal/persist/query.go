package persist

import (
	"errors"
	"time"

	"example.com/temporary-share-gateway/internal/model"
	"go.etcd.io/bbolt"
)

func (s *Store) MarkExpired(now time.Time) (int, error) {
	records, err := s.ListShares()
	if err != nil {
		return 0, err
	}
	changed := 0
	for _, record := range records {
		if record.Status == model.StatusActive && record.IsExpired(now) {
			record.Status = model.StatusExpired
			record.Version++
			if err := s.UpdateShare(record); err != nil {
				return changed, err
			}
			changed++
		}
	}
	return changed, nil
}

func (s *Store) RevokeShare(token string) error {
	record, err := s.LoadShare(token)
	if err != nil {
		return err
	}
	if record.Status == model.StatusRevoked {
		return nil
	}
	record.Status = model.StatusRevoked
	record.Version++
	return s.UpdateShare(record)
}

func (s *Store) CountActive(now time.Time) (int, error) {
	records, err := s.ListShares()
	if err != nil {
		return 0, err
	}
	count := 0
	for _, record := range records {
		if record.Active(now) {
			count++
		}
	}
	return count, nil
}

func (s *Store) ValidateReady() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return errors.New("store is closed")
	}
	return s.db.View(func(tx *bbolt.Tx) error {
		for _, name := range [][]byte{bucketNames.Shares, bucketNames.Audits, bucketNames.Configs, bucketNames.Windows} {
			if tx.Bucket(name) == nil {
				return errors.New("required bucket missing")
			}
		}
		return nil
	})
}
