// Package fsstore provides a filesystem-backed implementation of store.Store.
//
// Each template is persisted as a single JSON file named <id>.json under a
// configured root directory. The store is intended for the local-client
// variant of the editor; concurrency is handled with a single RWMutex
// because in that mode there is exactly one human user at a time.
package fsstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"tile-backend/internal/model"
	"tile-backend/internal/store"

	"github.com/google/uuid"
)

const (
	fileSuffix     = ".json"
	tempSuffix     = ".json.tmp"
	defaultPerm    = 0o644
	defaultDirPerm = 0o755
)

// Store is a filesystem-backed implementation of store.Store.
type Store struct {
	rootDir string
	mu      sync.RWMutex
}

// New returns a Store whose data lives under rootDir. The directory is created
// if it does not already exist.
func New(rootDir string) (*Store, error) {
	if rootDir == "" {
		return nil, errors.New("fsstore: rootDir is required")
	}
	abs, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("fsstore: resolve rootDir: %w", err)
	}
	if err := os.MkdirAll(abs, defaultDirPerm); err != nil {
		return nil, fmt.Errorf("fsstore: create rootDir: %w", err)
	}
	return &Store{rootDir: abs}, nil
}

// RootDir returns the resolved absolute path the store reads and writes.
func (s *Store) RootDir() string { return s.rootDir }

// Create writes a new template. The id is generated if not supplied.
// It fails with an error if a template with the same id already exists.
func (s *Store) Create(ctx context.Context, t model.Template) (*model.Template, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	now := time.Now().UTC()
	t.CreatedAt = now
	t.UpdatedAt = now

	model.ComputeTemplateStats(&t)

	path := s.pathFor(t.ID.String())
	if _, err := os.Stat(path); err == nil {
		return nil, fmt.Errorf("fsstore: template %s already exists", t.ID)
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("fsstore: stat existing: %w", err)
	}

	if err := writeAtomic(path, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

// Get loads a template by id. Returns store.ErrNotFound if no such file exists.
func (s *Store) Get(ctx context.Context, id string) (*model.Template, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return readTemplate(s.pathFor(id))
}

// Update overwrites the template at id with the supplied value, preserving the
// original CreatedAt timestamp. Returns store.ErrNotFound if no such file exists.
func (s *Store) Update(ctx context.Context, id string, t model.Template) (*model.Template, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.pathFor(id)
	existing, err := readTemplate(path)
	if err != nil {
		return nil, err
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("fsstore: invalid id %q: %w", id, err)
	}
	t.ID = parsedID
	t.CreatedAt = existing.CreatedAt
	t.UpdatedAt = time.Now().UTC()
	t.ViewCount = existing.ViewCount

	model.ComputeTemplateStats(&t)

	if err := writeAtomic(path, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

// Delete removes the template file. Returns store.ErrNotFound if absent.
func (s *Store) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.pathFor(id)
	err := os.Remove(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return store.ErrNotFound
		}
		return fmt.Errorf("fsstore: delete %s: %w", id, err)
	}
	return nil
}

// List scans the root directory, applies the same filters as the SQL backend,
// sorts the result by UpdatedAt descending, and returns a paginated slice.
// total is the count after filtering, before pagination.
func (s *Store) List(ctx context.Context, params model.ListTemplatesQueryParams) ([]model.TemplateSummary, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(s.rootDir)
	if err != nil {
		return nil, 0, fmt.Errorf("fsstore: read dir: %w", err)
	}

	matches := make([]model.TemplateSummary, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, fileSuffix) || strings.HasSuffix(name, tempSuffix) {
			continue
		}
		t, err := readTemplate(filepath.Join(s.rootDir, name))
		if err != nil {
			// Skip unreadable / malformed files; a single bad file should not
			// take the whole list offline.
			continue
		}
		if !matchesFilters(t, params) {
			continue
		}
		matches = append(matches, summaryOf(t))
	}

	sort.SliceStable(matches, func(i, j int) bool {
		return matches[i].UpdatedAt.After(matches[j].UpdatedAt)
	})

	total := len(matches)
	start := params.Offset
	if start > total {
		start = total
	}
	end := start
	if params.Limit > 0 {
		end = start + params.Limit
	} else {
		end = total
	}
	if end > total {
		end = total
	}

	return matches[start:end], total, nil
}

// HealthCheck stats the root directory so callers can detect a deleted or
// unmounted folder.
func (s *Store) HealthCheck(ctx context.Context) error {
	info, err := os.Stat(s.rootDir)
	if err != nil {
		return fmt.Errorf("fsstore: rootDir unreachable: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("fsstore: rootDir is not a directory: %s", s.rootDir)
	}
	return nil
}

func (s *Store) pathFor(id string) string {
	return filepath.Join(s.rootDir, id+fileSuffix)
}

// readTemplate reads and unmarshals a single template file. Returns
// store.ErrNotFound if the file is absent.
func readTemplate(path string) (*model.Template, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, store.ErrNotFound
		}
		return nil, fmt.Errorf("fsstore: read %s: %w", path, err)
	}
	var t model.Template
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, fmt.Errorf("fsstore: parse %s: %w", path, err)
	}
	return &t, nil
}

// writeAtomic writes the template to <path>.tmp and renames over <path> to
// avoid leaving a partial file behind on crash.
func writeAtomic(path string, t *model.Template) error {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return fmt.Errorf("fsstore: marshal: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, defaultPerm); err != nil {
		return fmt.Errorf("fsstore: write tmp: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("fsstore: rename: %w", err)
	}
	return nil
}

// summaryOf projects a Template into a TemplateSummary (drops the payload
// blob, keeps the indexed/displayed fields).
func summaryOf(t *model.Template) model.TemplateSummary {
	return model.TemplateSummary{
		ID:             t.ID,
		Name:           t.Name,
		Version:        t.Version,
		Width:          t.Width,
		Height:         t.Height,
		Thumbnail:      t.Thumbnail,
		WalkableRatio:  t.WalkableRatio,
		RoomType:       t.RoomType,
		RoomCategory:   t.RoomCategory,
		RoomAttributes: t.RoomAttributes,
		DoorsConnected: t.DoorsConnected,
		OpenDoors:      t.OpenDoors,
		StaticCount:    t.StaticCount,
		ChaserCount:    t.ChaserCount,
		ZonerCount:     t.ZonerCount,
		DPSCount:       t.DPSCount,
		MobAirCount:    t.MobAirCount,
		StageType:      t.StageType,
		ViewCount:      t.ViewCount,
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
	}
}

// matchesFilters returns true if t satisfies every filter set on params.
// Unset filters are ignored.
func matchesFilters(t *model.Template, p model.ListTemplatesQueryParams) bool {
	if p.NameLike != "" && !strings.Contains(strings.ToLower(t.Name), strings.ToLower(p.NameLike)) {
		return false
	}
	if p.RoomType != "" {
		if t.RoomType == nil || *t.RoomType != p.RoomType {
			return false
		}
	}
	if p.StageType != "" {
		if t.StageType == nil || *t.StageType != p.StageType {
			return false
		}
	}
	if p.MinWalkableRatio != nil && (t.WalkableRatio == nil || *t.WalkableRatio < *p.MinWalkableRatio) {
		return false
	}
	if p.MaxWalkableRatio != nil && (t.WalkableRatio == nil || *t.WalkableRatio > *p.MaxWalkableRatio) {
		return false
	}
	if !inRange(t.StaticCount, p.MinStaticCount, p.MaxStaticCount) {
		return false
	}
	if !inRange(t.ChaserCount, p.MinChaserCount, p.MaxChaserCount) {
		return false
	}
	if !inRange(t.ZonerCount, p.MinZonerCount, p.MaxZonerCount) {
		return false
	}
	if !inRange(t.DPSCount, p.MinDPSCount, p.MaxDPSCount) {
		return false
	}
	if !inRange(t.MobAirCount, p.MinMobAirCount, p.MaxMobAirCount) {
		return false
	}
	if !doorMatches(t.DoorsConnected, p) {
		return false
	}
	return true
}

func inRange(v *int, lo, hi *int) bool {
	if lo != nil && (v == nil || *v < *lo) {
		return false
	}
	if hi != nil && (v == nil || *v > *hi) {
		return false
	}
	return true
}

func doorMatches(dc *model.DoorsConnected, p model.ListTemplatesQueryParams) bool {
	if p.TopDoorConnected != nil && (dc == nil || dc.Top != *p.TopDoorConnected) {
		return false
	}
	if p.RightDoorConnected != nil && (dc == nil || dc.Right != *p.RightDoorConnected) {
		return false
	}
	if p.BottomDoorConnected != nil && (dc == nil || dc.Bottom != *p.BottomDoorConnected) {
		return false
	}
	if p.LeftDoorConnected != nil && (dc == nil || dc.Left != *p.LeftDoorConnected) {
		return false
	}
	return true
}
