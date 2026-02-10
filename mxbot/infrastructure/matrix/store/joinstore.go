package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"maunium.net/go/mautrix/id"
)

type JoinStore struct {
	path string
	mu   sync.Mutex
	// room_id -> join_ts_millis
	JoinTS map[id.RoomID]int64 `json:"join_ts"`
}

func NewJoinStore(path string) (*JoinStore, error) {

	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	js := &JoinStore{
		path:   abs,
		JoinTS: map[id.RoomID]int64{},
	}

	b, err := os.ReadFile(js.path)
	if err != nil {
		if os.IsNotExist(err) {
			return js, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(b, &js.JoinTS); err != nil {
		return nil, err
	}
	return js, nil
}

func (s *JoinStore) Get(roomID id.RoomID) (int64, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.JoinTS[roomID]
	return v, ok
}

// Set without conditions overwrites the value.
func (s *JoinStore) Set(roomID id.RoomID, tsMillis int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.JoinTS[roomID] = tsMillis
	return s.saveLocked()
}

// SetIfLater saves tsMillis only if it's LATER than what's already saved (or if the value doesn't exist yet).
// This is what's needed for "from the last entry."
func (s *JoinStore) SetIfLater(roomID id.RoomID, tsMillis int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	old, ok := s.JoinTS[roomID]
	if ok && old != 0 && old >= tsMillis {
		return nil
	}
	s.JoinTS[roomID] = tsMillis
	return s.saveLocked()
}

func (s *JoinStore) SetIfEarlier(roomID id.RoomID, tsMillis int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	old, ok := s.JoinTS[roomID]
	if ok && old != 0 && old <= tsMillis {
		return nil
	}
	s.JoinTS[roomID] = tsMillis
	return s.saveLocked()
}

func (s *JoinStore) saveLocked() error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	b, err := json.MarshalIndent(s.JoinTS, "", "  ")
	if err != nil {
		return err
	}

	tmp := filepath.Join(dir, filepath.Base(s.path)+".tmp")
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}

	return os.Rename(tmp, s.path)
}

func MaxTime(a, b time.Time) time.Time {
	if a.IsZero() {
		return b
	}
	if b.IsZero() {
		return a
	}
	if a.After(b) {
		return a
	}
	return b
}
