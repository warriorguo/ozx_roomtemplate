// Package fsstore provides a filesystem-backed implementation of store.Store
// that speaks the OZX TilemapData on-disk format directly.
//
// Layout (matches Assets/StreamingAssets/TilemapData in ozx_base):
//
//	<root>/
//	├── normal/
//	│   ├── all_boss_3_01.json
//	│   ├── all_boss_3_01.json.meta   (Unity-generated, we never touch it)
//	│   └── ...
//	├── basement/
//	├── cave/
//	└── test/
//
// Each .json file is a bare TemplatePayload (no Template envelope). The
// envelope fields the API still exposes (id, created_at, updated_at,
// computed counts) are synthesised on read:
//
//   - id        = "<category>__<basename>" (double-underscore separates the
//                 category folder from the OZX shape_stage_doors_seq stem)
//   - name      = the file's basename without .json
//   - created_at, updated_at = file mtime (best we have without an envelope)
//   - id as UUID = uuid.NewSHA1(namespace, "<category>/<basename>")
//   - counts    = model.ComputeTemplateStats over the payload
//
// Writes go to <file>.json.tmp then rename. We never write *.meta —
// Unity regenerates it on next import. On delete we remove both the
// .json and the .json.meta (if present) so the project tree stays tidy.
package fsstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"tile-backend/internal/model"
	"tile-backend/internal/store"

	"github.com/google/uuid"
)

const (
	fileSuffix     = ".json"
	metaSuffix     = ".meta"
	tempSuffix     = ".json.tmp"
	defaultPerm    = 0o644
	defaultDirPerm = 0o755
	idSeparator    = "__"
	maxSeq         = 99
	defaultCategory = "normal"
	defaultShape    = "all"
	defaultStage    = "none"
)

// uuidNamespace is the v5 namespace used to derive stable IDs from a file's
// relative path. A fixed namespace means a given category/basename always
// produces the same UUID across machines.
var uuidNamespace = uuid.MustParse("9b3b6f7e-2c5a-4f0c-9c2d-2d6a3e7b1f80")

// Store is a filesystem-backed implementation of store.Store, OZX flavour.
type Store struct {
	rootDir string
	mu      sync.RWMutex
}

// New returns a Store rooted at rootDir. The directory (and standard
// category subfolders) are created if missing.
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

// Create writes a new template. The destination filename is derived from the
// payload's roomCategory / roomShape / stageType / openDoors, with the next
// free numeric sequence appended.
//
// The caller-supplied ID, CreatedAt, UpdatedAt, and Thumbnail are ignored;
// the OZX file format has no slot for them and they'd just rot if we tried
// to preserve them.
func (s *Store) Create(ctx context.Context, t model.Template) (*model.Template, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cat := categoryOf(&t)
	if err := s.ensureCategory(cat); err != nil {
		return nil, err
	}

	basename, err := s.allocateBasename(cat, &t)
	if err != nil {
		return nil, err
	}
	relPath := filepath.Join(cat, basename+fileSuffix)
	fullPath := filepath.Join(s.rootDir, relPath)

	model.ComputeTemplateStats(&t)

	if err := writePayload(fullPath, &t.Payload); err != nil {
		return nil, err
	}
	return s.synthesise(relPath, fullPath, &t.Payload)
}

// Get loads a template by id (`<category>__<basename>` form).
func (s *Store) Get(ctx context.Context, id string) (*model.Template, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	relPath, err := s.pathForID(id)
	if err != nil {
		return nil, err
	}
	fullPath := filepath.Join(s.rootDir, relPath)
	payload, err := readPayload(fullPath)
	if err != nil {
		return nil, err
	}
	return s.synthesise(relPath, fullPath, payload)
}

// Update overwrites the payload of an existing template. The on-disk
// filename does not change even if the payload's category/shape/etc.
// changed — renaming would break the Unity .meta linkage. The user can
// delete and re-create if they really need a new filename.
func (s *Store) Update(ctx context.Context, id string, t model.Template) (*model.Template, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	relPath, err := s.pathForID(id)
	if err != nil {
		return nil, err
	}
	fullPath := filepath.Join(s.rootDir, relPath)
	if _, err := os.Stat(fullPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, store.ErrNotFound
		}
		return nil, fmt.Errorf("fsstore: stat existing: %w", err)
	}

	model.ComputeTemplateStats(&t)
	if err := writePayload(fullPath, &t.Payload); err != nil {
		return nil, err
	}
	return s.synthesise(relPath, fullPath, &t.Payload)
}

// Delete removes the template file plus its Unity .meta sibling (if any).
// Removing the .meta is the polite thing to do: Unity will regenerate it on
// next import and the user won't be left with an orphaned meta file.
func (s *Store) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	relPath, err := s.pathForID(id)
	if err != nil {
		return err
	}
	fullPath := filepath.Join(s.rootDir, relPath)
	if err := os.Remove(fullPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return store.ErrNotFound
		}
		return fmt.Errorf("fsstore: delete %s: %w", id, err)
	}
	_ = os.Remove(fullPath + metaSuffix) // best-effort, don't error if absent
	return nil
}

// List walks the immediate category subfolders, reads every .json (skipping
// .meta and .tmp), applies filters, sorts by mtime desc, and paginates.
// total is the count after filtering, before pagination.
func (s *Store) List(ctx context.Context, params model.ListTemplatesQueryParams) ([]model.TemplateSummary, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	matches := make([]model.TemplateSummary, 0, 64)

	walkErr := s.walk(func(relPath, fullPath string, payload *model.TemplatePayload) {
		t, err := s.synthesise(relPath, fullPath, payload)
		if err != nil {
			return
		}
		if !matchesFilters(t, params) {
			return
		}
		matches = append(matches, summaryOf(t))
	})
	if walkErr != nil {
		return nil, 0, walkErr
	}

	sort.SliceStable(matches, func(i, j int) bool {
		return matches[i].UpdatedAt.After(matches[j].UpdatedAt)
	})

	total := len(matches)
	start := params.Offset
	if start > total {
		start = total
	}
	end := total
	if params.Limit > 0 && start+params.Limit < total {
		end = start + params.Limit
	}
	return matches[start:end], total, nil
}

// HealthCheck stats the root directory.
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

// --- internals --------------------------------------------------------

// walk visits every payload file under any first-level category subfolder,
// skipping .meta sidecars and .tmp leftovers from interrupted writes.
func (s *Store) walk(visit func(relPath, fullPath string, payload *model.TemplatePayload)) error {
	cats, err := os.ReadDir(s.rootDir)
	if err != nil {
		return fmt.Errorf("fsstore: read rootDir: %w", err)
	}
	for _, cat := range cats {
		if !cat.IsDir() {
			continue
		}
		catName := cat.Name()
		catDir := filepath.Join(s.rootDir, catName)
		entries, err := os.ReadDir(catDir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if !strings.HasSuffix(name, fileSuffix) || strings.HasSuffix(name, tempSuffix) || strings.HasSuffix(name, metaSuffix) {
				continue
			}
			full := filepath.Join(catDir, name)
			payload, err := readPayload(full)
			if err != nil {
				continue // unreadable / corrupt — skip rather than fail the whole list
			}
			rel := filepath.Join(catName, name)
			visit(rel, full, payload)
		}
	}
	return nil
}

// pathForID maps an id to a relative path under the root. Two id forms are
// accepted:
//
//  1. `<category>__<basename>` — the cheap canonical form.
//  2. A UUID string — matched against the deterministic uuid.NewSHA1 derived
//     from every payload file's relative path. This lets the frontend keep
//     using the UUIDs returned in TemplateSummary without us teaching the
//     React side about the OZX naming convention.
func (s *Store) pathForID(id string) (string, error) {
	if strings.Contains(id, idSeparator) {
		cat, base, ok := strings.Cut(id, idSeparator)
		if !ok || cat == "" || base == "" {
			return "", fmt.Errorf("fsstore: invalid id %q: expected '<category>%s<basename>'", id, idSeparator)
		}
		if strings.ContainsAny(cat, "/\\") || strings.ContainsAny(base, "/\\") {
			return "", fmt.Errorf("fsstore: invalid id %q: contains path separator", id)
		}
		return filepath.Join(cat, base+fileSuffix), nil
	}
	if target, err := uuid.Parse(id); err == nil {
		if rel, ok := s.findRelPathByUUID(target); ok {
			return rel, nil
		}
		return "", store.ErrNotFound
	}
	return "", fmt.Errorf("fsstore: invalid id %q: expected '<category>%s<basename>' or a UUID", id, idSeparator)
}

// findRelPathByUUID scans every payload filename under each category subfolder
// and returns the first whose synthesised v5 UUID matches target. Reads no
// file contents — comparison is purely on relative path strings.
func (s *Store) findRelPathByUUID(target uuid.UUID) (string, bool) {
	cats, err := os.ReadDir(s.rootDir)
	if err != nil {
		return "", false
	}
	for _, cat := range cats {
		if !cat.IsDir() {
			continue
		}
		entries, err := os.ReadDir(filepath.Join(s.rootDir, cat.Name()))
		if err != nil {
			continue
		}
		for _, e := range entries {
			name := e.Name()
			if e.IsDir() || !strings.HasSuffix(name, fileSuffix) || strings.HasSuffix(name, tempSuffix) || strings.HasSuffix(name, metaSuffix) {
				continue
			}
			rel := filepath.Join(cat.Name(), name)
			if uuid.NewSHA1(uuidNamespace, []byte(rel)) == target {
				return rel, true
			}
		}
	}
	return "", false
}

// idForPath inverts pathForID.
func idForPath(relPath string) string {
	dir, file := filepath.Split(relPath)
	dir = strings.TrimRight(dir, "/")
	base := strings.TrimSuffix(file, fileSuffix)
	if dir == "" {
		return base
	}
	return dir + idSeparator + base
}

// ensureCategory creates <root>/<cat> if it does not already exist.
func (s *Store) ensureCategory(cat string) error {
	if err := os.MkdirAll(filepath.Join(s.rootDir, cat), defaultDirPerm); err != nil {
		return fmt.Errorf("fsstore: create category %s: %w", cat, err)
	}
	return nil
}

// allocateBasename picks the next free `<shape>_<stage>_<doors>_<NN>` in a
// category folder. Returns the basename WITHOUT the .json suffix.
func (s *Store) allocateBasename(cat string, t *model.Template) (string, error) {
	shape := shapeOf(t)
	stage := stageOf(t)
	doors := doorsOf(t)
	prefix := fmt.Sprintf("%s_%s_%d_", shape, stage, doors)

	catDir := filepath.Join(s.rootDir, cat)
	entries, err := os.ReadDir(catDir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("fsstore: read category dir: %w", err)
	}

	used := make(map[int]bool, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, fileSuffix) {
			continue
		}
		seqStr := strings.TrimSuffix(strings.TrimPrefix(name, prefix), fileSuffix)
		if n, err := strconv.Atoi(seqStr); err == nil {
			used[n] = true
		}
	}
	for n := 1; n <= maxSeq; n++ {
		if !used[n] {
			return fmt.Sprintf("%s%02d", prefix, n), nil
		}
	}
	return "", fmt.Errorf("fsstore: no free sequence number under %d for %q", maxSeq, prefix)
}

// synthesise wraps a bare TemplatePayload into a Template envelope. The
// envelope fields not present in the OZX format are derived deterministically:
//   - ID = uuid.NewSHA1(uuidNamespace, relPath)
//   - Name = the filename basename without .json
//   - CreatedAt = UpdatedAt = file mtime (best approximation)
//   - StaticCount/ChaserCount/... = ComputeTemplateStats over the payload
func (s *Store) synthesise(relPath, fullPath string, payload *model.TemplatePayload) (*model.Template, error) {
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("fsstore: stat %s: %w", fullPath, err)
	}
	mtime := info.ModTime().UTC()

	dir, file := filepath.Split(relPath)
	base := strings.TrimSuffix(file, fileSuffix)
	name := payload.Meta.Name
	if name == "" {
		name = base
	}
	cat := strings.TrimRight(dir, string(filepath.Separator))

	t := &model.Template{
		ID:        uuid.NewSHA1(uuidNamespace, []byte(relPath)),
		Name:      name,
		Version:   payload.Meta.Version,
		Width:     payload.Meta.Width,
		Height:    payload.Meta.Height,
		Payload:   *payload,
		CreatedAt: mtime,
		UpdatedAt: mtime,
	}
	// Mirror the payload-level fields into the envelope columns so the
	// frontend filters that read from TemplateSummary keep working.
	if payload.RoomShape != nil {
		shape := *payload.RoomShape
		if shape == "all" {
			full := "full"
			t.RoomType = &full
		} else {
			t.RoomType = &shape
		}
	}
	if payload.RoomCategory != nil {
		t.RoomCategory = payload.RoomCategory
	} else if cat != "" {
		c := cat
		t.RoomCategory = &c
	}
	if payload.StageType != nil {
		t.StageType = payload.StageType
	}
	if payload.OpenDoors != nil {
		t.OpenDoors = payload.OpenDoors
	}
	model.ComputeTemplateStats(t)
	return t, nil
}

// categoryOf returns the category subfolder to write a template into.
// Falls back to `defaultCategory` if neither the envelope nor the payload
// supplied one.
func categoryOf(t *model.Template) string {
	if t.RoomCategory != nil && *t.RoomCategory != "" {
		return *t.RoomCategory
	}
	if t.Payload.RoomCategory != nil && *t.Payload.RoomCategory != "" {
		return *t.Payload.RoomCategory
	}
	return defaultCategory
}

// shapeOf returns the file-naming shape segment. The OZX convention uses
// "all" / "bridge" / "platform" / "none"; "full" (DB-style) is mapped to
// "all" to match the existing files on disk.
func shapeOf(t *model.Template) string {
	if t.Payload.RoomShape != nil && *t.Payload.RoomShape != "" {
		return *t.Payload.RoomShape
	}
	if t.RoomType != nil && *t.RoomType != "" {
		s := *t.RoomType
		if s == "full" {
			return "all"
		}
		return s
	}
	return defaultShape
}

// stageOf returns the file-naming stage segment.
func stageOf(t *model.Template) string {
	if t.Payload.StageType != nil && *t.Payload.StageType != "" {
		return *t.Payload.StageType
	}
	if t.StageType != nil && *t.StageType != "" {
		return *t.StageType
	}
	return defaultStage
}

// doorsOf returns the openDoors bitmask used in the file name.
func doorsOf(t *model.Template) int {
	if t.Payload.OpenDoors != nil {
		return *t.Payload.OpenDoors
	}
	if t.OpenDoors != nil {
		return *t.OpenDoors
	}
	return 0
}

// summaryOf projects a Template into a TemplateSummary (drops the heavy
// payload blob, keeps the indexed/displayed fields).
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

// readPayload reads and unmarshals a single OZX payload file. Returns
// store.ErrNotFound if the file is absent.
func readPayload(path string) (*model.TemplatePayload, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, store.ErrNotFound
		}
		return nil, fmt.Errorf("fsstore: read %s: %w", path, err)
	}
	var p model.TemplatePayload
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("fsstore: parse %s: %w", path, err)
	}
	return &p, nil
}

// writePayload writes the bare payload via tmp + rename.
func writePayload(path string, payload *model.TemplatePayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("fsstore: marshal: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), defaultDirPerm); err != nil {
		return fmt.Errorf("fsstore: mkdir: %w", err)
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
