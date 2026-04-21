package store

import (
	"errors"
	"sort"
	"strings"
	"sync"

	"netclaw/proxy-core/internal/session"
)

type ListOptions struct {
	Query          string
	Host           string
	Method         string
	HasError       *bool
	TLSIntercepted *bool
	Limit          int
}

type SessionStore interface {
	Save(item session.Session) error
	List(options ListOptions) ([]session.Summary, error)
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

func (s *MemoryStore) List(options ListOptions) ([]session.Summary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]session.Summary, 0, len(s.items))
	for _, item := range s.items {
		if !matchesListOptions(item, options) {
			continue
		}
		result = append(result, item.ToSummary())
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].StartTime.After(result[j].StartTime)
	})

	if options.Limit > 0 && len(result) > options.Limit {
		result = result[:options.Limit]
	}

	return result, nil
}

func matchesListOptions(item session.Session, options ListOptions) bool {
	if options.Host != "" && !strings.EqualFold(item.Host, strings.TrimSpace(options.Host)) {
		return false
	}
	if options.Method != "" && !strings.EqualFold(item.Method, strings.TrimSpace(options.Method)) {
		return false
	}
	if options.HasError != nil && (item.Error != "") != *options.HasError {
		return false
	}
	if options.TLSIntercepted != nil && item.TLSIntercepted != *options.TLSIntercepted {
		return false
	}
	if query := strings.ToLower(strings.TrimSpace(options.Query)); query != "" {
		searchable := []string{
			item.ID,
			item.Method,
			item.Host,
			item.URL,
			item.Path,
			item.ContentType,
			item.Error,
		}
		matched := false
		for _, value := range searchable {
			if strings.Contains(strings.ToLower(value), query) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
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
