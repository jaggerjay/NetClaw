package store

import (
	"errors"
	"sort"
	"sync"

	"netclaw/proxy-core/internal/session"
)

type SessionStore interface {
	Save(item session.Session) error
	List() ([]session.Summary, error)
	Get(id string) (*session.Session, error)
	Clear() error
}

type MemoryStore struct {
	mu    sync.RWMutex
	items map[string]session.Session
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{items: make(map[string]session.Session)}
}

func (s *MemoryStore) Save(item session.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[item.ID] = item
	return nil
}

func (s *MemoryStore) List() ([]session.Summary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]session.Summary, 0, len(s.items))
	for _, item := range s.items {
		result = append(result, item.ToSummary())
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].StartTime.After(result[j].StartTime)
	})

	return result, nil
}

func (s *MemoryStore) Get(id string) (*session.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.items[id]
	if !ok {
		return nil, errors.New("session not found")
	}
	copy := item
	return &copy, nil
}

func (s *MemoryStore) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = make(map[string]session.Session)
	return nil
}
