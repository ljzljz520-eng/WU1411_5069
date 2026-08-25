package persist

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"example.com/temporary-share-gateway/internal/model"
	"go.etcd.io/bbolt"
)

type Snapshot struct {
	ExportedAt time.Time             `json:"exported_at"`
	Shares     []model.ShareRecord   `json:"shares"`
	Audits     []model.AuditEntry    `json:"audits"`
	Configs    []model.GatewayConfig `json:"configs"`
}

func (s *Store) Snapshot(now time.Time) (Snapshot, error) {
	shares, err := s.ListShares()
	if err != nil {
		return Snapshot{}, err
	}
	audits, err := s.ListAudits()
	if err != nil {
		return Snapshot{}, err
	}
	configs, err := s.listConfigs()
	if err != nil {
		return Snapshot{}, err
	}
	if now.IsZero() {
		now = time.Unix(0, 0).UTC()
	}
	return Snapshot{ExportedAt: now.UTC(), Shares: shares, Audits: audits, Configs: configs}, nil
}

func (s *Store) SnapshotJSON(now time.Time) ([]byte, error) {
	snapshot, err := s.Snapshot(now)
	if err != nil {
		return nil, err
	}
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal snapshot: %w", err)
	}
	return data, nil
}

func (s *Store) Restore(snapshot Snapshot) error {
	if len(snapshot.Shares) == 0 && len(snapshot.Configs) == 0 && len(snapshot.Audits) == 0 {
		return errors.New("snapshot is empty")
	}
	for _, record := range snapshot.Shares {
		if err := record.Validate(); err != nil {
			return err
		}
		if _, err := s.LoadShare(record.Token); errors.Is(err, ErrNotFound) {
			if err := s.SaveShare(record); err != nil {
				return err
			}
		} else if err == nil {
			if err := s.UpdateShare(record); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	for _, config := range snapshot.Configs {
		if err := s.SaveConfig(config); err != nil {
			return err
		}
	}
	for _, entry := range snapshot.Audits {
		if err := s.SaveAuditEntry(entry); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) listConfigs() ([]model.GatewayConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return nil, errors.New("store is closed")
	}
	configs := make([]model.GatewayConfig, 0)
	err := s.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketNames.Configs).ForEach(func(_, value []byte) error {
			var config model.GatewayConfig
			if err := decode(value, &config); err != nil {
				return err
			}
			configs = append(configs, config)
			return nil
		})
	})
	sort.Slice(configs, func(i, j int) bool { return configs[i].Name < configs[j].Name })
	return configs, err
}

func ParseSnapshot(data []byte) (Snapshot, error) {
	if len(data) == 0 {
		return Snapshot{}, errors.New("snapshot payload is empty")
	}
	var snapshot Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return Snapshot{}, fmt.Errorf("parse snapshot: %w", err)
	}
	return snapshot, nil
}
