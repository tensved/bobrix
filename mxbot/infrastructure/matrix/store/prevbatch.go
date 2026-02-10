package store

import (
	"sync"

	"maunium.net/go/mautrix/id"
)

type PrevBatchStore struct {
	mu sync.Mutex
	m  map[id.RoomID]string
}

func NewPrevBatchStore() *PrevBatchStore {
	return &PrevBatchStore{m: map[id.RoomID]string{}}
}

func (s *PrevBatchStore) Set(roomID id.RoomID, token string) {
	if token == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	// We need the most recent prev_batch, we can just overwrite it
	s.m[roomID] = token
}

func (s *PrevBatchStore) Get(roomID id.RoomID) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.m[roomID]
	return v, ok
}