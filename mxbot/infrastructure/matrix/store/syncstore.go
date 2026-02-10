package store

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
)

// FileSyncStore stores sync token and filter ID on disk.
// Good enough for tests/single instance. For prod лучше БД.
type FileSyncStore struct {
	path string
	mu   sync.Mutex
	data fileSyncStoreData
}

type fileSyncStoreData struct {
	NextBatch string `json:"next_batch"`
	FilterID  string `json:"filter_id"`
}

var _ mautrix.SyncStore = (*FileSyncStore)(nil)

func NewFileSyncStore(path string) (*FileSyncStore, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	s := &FileSyncStore{path: abs}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *FileSyncStore) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			s.data = fileSyncStoreData{}
			return nil
		}
		return err
	}
	return json.Unmarshal(b, &s.data)
}

func (s *FileSyncStore) saveLocked() error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	b, err := json.MarshalIndent(&s.data, "", "  ")
	if err != nil {
		return err
	}

	tmp := filepath.Join(dir, filepath.Base(s.path)+".tmp")
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}

	return os.Rename(tmp, s.path)
}

func (s *FileSyncStore) SaveNextBatch(_ context.Context, _ id.UserID, nextBatch string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.NextBatch = nextBatch
	return s.saveLocked()
}

func (s *FileSyncStore) LoadNextBatch(_ context.Context, _ id.UserID) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data.NextBatch, nil
}

func (s *FileSyncStore) SaveFilterID(_ context.Context, _ id.UserID, filterID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.FilterID = filterID
	return s.saveLocked()
}

func (s *FileSyncStore) LoadFilterID(_ context.Context, _ id.UserID) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data.FilterID, nil
}
