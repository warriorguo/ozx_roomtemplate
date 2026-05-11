package store

import (
	"context"
	"tile-backend/internal/model"
)

// StubStore is a placeholder implementation of Store. Every method returns
// ErrNotImplemented. It exists so the binary builds and serves /health while
// a real backend (e.g. the filesystem store in ORT-66) is being developed.
type StubStore struct{}

// NewStubStore returns a Store that always reports ErrNotImplemented.
func NewStubStore() *StubStore { return &StubStore{} }

func (s *StubStore) Create(ctx context.Context, template model.Template) (*model.Template, error) {
	return nil, ErrNotImplemented
}

func (s *StubStore) Get(ctx context.Context, id string) (*model.Template, error) {
	return nil, ErrNotImplemented
}

func (s *StubStore) Update(ctx context.Context, id string, template model.Template) (*model.Template, error) {
	return nil, ErrNotImplemented
}

func (s *StubStore) Delete(ctx context.Context, id string) error {
	return ErrNotImplemented
}

func (s *StubStore) List(ctx context.Context, params model.ListTemplatesQueryParams) ([]model.TemplateSummary, int, error) {
	return nil, 0, ErrNotImplemented
}

// HealthCheck is the one method that succeeds — the stub backend is always "up"
// in the sense that it is reachable; users get clear errors from the data
// operations instead.
func (s *StubStore) HealthCheck(ctx context.Context) error { return nil }
