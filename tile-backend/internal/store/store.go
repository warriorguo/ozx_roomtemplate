package store

import (
	"context"
	"errors"
	"tile-backend/internal/model"
)

// ErrNotFound is returned when a template does not exist in the store.
var ErrNotFound = errors.New("template not found")

// ErrNotImplemented is returned by stub stores that have no real backing.
var ErrNotImplemented = errors.New("store not implemented")

// Store is the storage abstraction for room templates.
//
// The local-client variant of the editor uses a filesystem-backed implementation;
// the interface deliberately omits cross-template aggregation operations
// (e.g. project listings) that only make sense against a relational database.
type Store interface {
	Create(ctx context.Context, template model.Template) (*model.Template, error)
	Get(ctx context.Context, id string) (*model.Template, error)
	Update(ctx context.Context, id string, template model.Template) (*model.Template, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, params model.ListTemplatesQueryParams) ([]model.TemplateSummary, int, error)
	HealthCheck(ctx context.Context) error
}
