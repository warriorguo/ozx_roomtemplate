package store

import (
	"context"
	"sync"
	"tile-backend/internal/model"
)

// SwappableStore is a Store wrapper whose backing implementation can be
// replaced atomically at runtime. The local-client editor uses it so the
// "Switch project" flow can repoint at a new templates directory without
// restarting the binary.
//
// All method calls grab a read lock to look up the current backend, then
// delegate; Swap takes a write lock to install a new one. Since the load is
// O(1) (no copies), this is essentially free for callers.
type SwappableStore struct {
	mu      sync.RWMutex
	current Store
}

// NewSwappableStore wraps an initial Store.
func NewSwappableStore(initial Store) *SwappableStore {
	return &SwappableStore{current: initial}
}

// Swap atomically replaces the underlying store. Returns the previous one in
// case the caller wants to close or inspect it.
func (s *SwappableStore) Swap(next Store) Store {
	s.mu.Lock()
	defer s.mu.Unlock()
	prev := s.current
	s.current = next
	return prev
}

// Current returns the active backend. Useful in tests.
func (s *SwappableStore) Current() Store {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.current
}

func (s *SwappableStore) Create(ctx context.Context, t model.Template) (*model.Template, error) {
	return s.Current().Create(ctx, t)
}

func (s *SwappableStore) Get(ctx context.Context, id string) (*model.Template, error) {
	return s.Current().Get(ctx, id)
}

func (s *SwappableStore) Update(ctx context.Context, id string, t model.Template) (*model.Template, error) {
	return s.Current().Update(ctx, id, t)
}

func (s *SwappableStore) Delete(ctx context.Context, id string) error {
	return s.Current().Delete(ctx, id)
}

func (s *SwappableStore) List(ctx context.Context, params model.ListTemplatesQueryParams) ([]model.TemplateSummary, int, error) {
	return s.Current().List(ctx, params)
}

func (s *SwappableStore) HealthCheck(ctx context.Context) error {
	return s.Current().HealthCheck(ctx)
}
