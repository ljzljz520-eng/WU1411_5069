package persist

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"sync"

	"example.com/temporary-share-gateway/internal/model"
	"go.etcd.io/bbolt"
)

var ErrNotFound = errors.New("record not found")

type Store struct {
	mu   sync.RWMutex
	db   *bbolt.DB
	path string
}

func Open(path string) (*Store, error) {
	if path == "" {
		return nil, fmt.Errorf("database path is required")
	}
	db, err := bbolt.Open(path, 0o600, nil)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	store := &Store{db: db, path: path}
	if err := store.initialize(); err != nil {
		db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) initialize() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return errors.New("store is closed")
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		for _, name := range [][]byte{bucketNames.Shares, bucketNames.Audits, bucketNames.Configs, bucketNames.Windows} {
			if _, err := tx.CreateBucketIfNotExists(name); err != nil {
				return fmt.Errorf("create bucket %s: %w", name, err)
			}
		}
		return nil
	})
}

func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
		return nil
	}
	err := s.db.Close()
	s.db = nil
	return err
}

func (s *Store) Reopen() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db != nil {
		return errors.New("store must be closed before reopen")
	}
	db, err := bbolt.Open(s.path, 0o600, nil)
	if err != nil {
		return fmt.Errorf("reopen database: %w", err)
	}
	s.db = db
	return nil
}

func (s *Store) Path() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.path
}

func (s *Store) SaveShare(record model.ShareRecord) error {
	if record.Token == "" {
		return errors.New("token is required")
	}
	data, err := encode(record)
	if err != nil {
		return err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return errors.New("store is closed")
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketNames.Shares)
		if bucket.Get([]byte(record.Token)) != nil {
			return fmt.Errorf("share %q already exists", record.Token)
		}
		return bucket.Put([]byte(record.Token), data)
	})
}

func (s *Store) LoadShare(token string) (model.ShareRecord, error) {
	if token == "" {
		return model.ShareRecord{}, ErrNotFound
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return model.ShareRecord{}, errors.New("store is closed")
	}
	var record model.ShareRecord
	err := s.db.View(func(tx *bbolt.Tx) error {
		value := tx.Bucket(bucketNames.Shares).Get([]byte(token))
		if value == nil {
			return ErrNotFound
		}
		return decode(cloneBytes(value), &record)
	})
	return record, err
}

func (s *Store) UpdateShare(record model.ShareRecord) error {
	if record.Token == "" {
		return errors.New("token is required")
	}
	data, err := encode(record)
	if err != nil {
		return err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return errors.New("store is closed")
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketNames.Shares)
		if bucket.Get([]byte(record.Token)) == nil {
			return ErrNotFound
		}
		return bucket.Put([]byte(record.Token), data)
	})
}

func (s *Store) DeleteShare(token string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return errors.New("store is closed")
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketNames.Shares).Delete([]byte(token))
	})
}

func (s *Store) ListShares() ([]model.ShareRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return nil, errors.New("store is closed")
	}
	var records []model.ShareRecord
	err := s.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketNames.Shares).ForEach(func(_, value []byte) error {
			var record model.ShareRecord
			if err := decode(value, &record); err != nil {
				return err
			}
			records = append(records, record)
			return nil
		})
	})
	sort.Slice(records, func(i, j int) bool { return records[i].Token < records[j].Token })
	return records, err
}

func (s *Store) Exists() bool {
	_, err := os.Stat(s.Path())
	return err == nil
}
