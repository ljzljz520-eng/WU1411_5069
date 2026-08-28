package persist

import (
	"errors"
	"fmt"
	"sort"

	"example.com/temporary-share-gateway/internal/model"
	"go.etcd.io/bbolt"
)

func (s *Store) SaveAuditEntry(entry model.AuditEntry) error {
	if entry.ID == "" {
		return errors.New("audit id is required")
	}
	data, err := encode(entry)
	if err != nil {
		return err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return errors.New("store is closed")
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketNames.Audits).Put([]byte(entry.ID), data)
	})
}

func (s *Store) ListAudits() ([]model.AuditEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return nil, errors.New("store is closed")
	}
	entries := make([]model.AuditEntry, 0)
	err := s.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketNames.Audits).ForEach(func(_, value []byte) error {
			var entry model.AuditEntry
			if err := decode(value, &entry); err != nil {
				return err
			}
			entries = append(entries, entry)
			return nil
		})
	})
	sort.Slice(entries, func(i, j int) bool { return entries[i].At.Before(entries[j].At) })
	return entries, err
}

func (s *Store) DeleteAudit(id string) error {
	if id == "" {
		return errors.New("audit id is required")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return errors.New("store is closed")
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketNames.Audits).Delete([]byte(id))
	})
}

func (s *Store) SaveConfig(config model.GatewayConfig) error {
	if err := config.Validate(); err != nil {
		return err
	}
	data, err := encode(config)
	if err != nil {
		return err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return errors.New("store is closed")
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketNames.Configs).Put([]byte(config.Name), data)
	})
}

func (s *Store) LoadConfig(name string) (model.GatewayConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return model.GatewayConfig{}, errors.New("store is closed")
	}
	var config model.GatewayConfig
	err := s.db.View(func(tx *bbolt.Tx) error {
		value := tx.Bucket(bucketNames.Configs).Get([]byte(name))
		if value == nil {
			return ErrNotFound
		}
		return decode(cloneBytes(value), &config)
	})
	return config, err
}

func (s *Store) SaveRateWindow(window model.RateWindow) error {
	if window.Key == "" {
		return fmt.Errorf("rate key is required")
	}
	data, err := encode(window)
	if err != nil {
		return err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return errors.New("store is closed")
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketNames.Windows).Put([]byte(window.Key), data)
	})
}

func (s *Store) LoadRateWindow(key string) (model.RateWindow, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return model.RateWindow{}, errors.New("store is closed")
	}
	var window model.RateWindow
	err := s.db.View(func(tx *bbolt.Tx) error {
		value := tx.Bucket(bucketNames.Windows).Get([]byte(key))
		if value == nil {
			return ErrNotFound
		}
		return decode(cloneBytes(value), &window)
	})
	return window, err
}
