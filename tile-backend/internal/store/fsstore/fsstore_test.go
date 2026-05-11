package fsstore

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"tile-backend/internal/model"
	"tile-backend/internal/store"

	"github.com/google/uuid"
)

// fixture returns a valid 4x4 template ready to write to the store.
func fixture(name string) model.Template {
	staticCount := 2
	roomType := "full"
	return model.Template{
		Name:    name,
		Version: 1,
		Width:   4,
		Height:  4,
		Payload: model.TemplatePayload{
			Ground: [][]int{{1, 1, 1, 1}, {1, 1, 1, 1}, {1, 1, 1, 1}, {1, 1, 1, 1}},
			Static: [][]int{{0, 1, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 1, 0, 0}},
			Chaser: [][]int{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}},
			Zoner:  [][]int{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}},
			DPS:    [][]int{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}},
			MobAir: [][]int{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}},
			Meta: model.TemplateMeta{
				Name:    name,
				Version: 1,
				Width:   4,
				Height:  4,
			},
		},
		StaticCount: &staticCount,
		RoomType:    &roomType,
	}
}

func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return s
}

func TestNew_CreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "templates")
	s, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("expected dir to exist: %v", err)
	}
	if s.RootDir() == "" {
		t.Fatal("RootDir empty")
	}
}

func TestNew_EmptyRoot(t *testing.T) {
	if _, err := New(""); err == nil {
		t.Fatal("expected error for empty rootDir")
	}
}

func TestCreate_RoundTrip(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	saved, err := s.Create(ctx, fixture("alpha"))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if saved.ID == uuid.Nil {
		t.Fatal("expected generated ID")
	}
	if saved.CreatedAt.IsZero() || saved.UpdatedAt.IsZero() {
		t.Fatal("expected timestamps populated")
	}

	got, err := s.Get(ctx, saved.ID.String())
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Name != "alpha" {
		t.Errorf("name: want alpha, got %q", got.Name)
	}
	if got.Payload.Meta.Width != 4 {
		t.Errorf("payload not round-tripped: %#v", got.Payload.Meta)
	}
}

func TestCreate_RejectsCollision(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	id := uuid.New()
	tmpl := fixture("alpha")
	tmpl.ID = id

	if _, err := s.Create(ctx, tmpl); err != nil {
		t.Fatalf("first Create: %v", err)
	}
	if _, err := s.Create(ctx, tmpl); err == nil {
		t.Fatal("expected second Create to fail with id collision")
	}
}

func TestGet_NotFound(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Get(context.Background(), uuid.NewString())
	if !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGet_CorruptedFile(t *testing.T) {
	s := newTestStore(t)
	bad := filepath.Join(s.RootDir(), uuid.NewString()+".json")
	if err := os.WriteFile(bad, []byte("{not json"), 0o644); err != nil {
		t.Fatalf("write bad: %v", err)
	}
	id := strings.TrimSuffix(filepath.Base(bad), ".json")
	if _, err := s.Get(context.Background(), id); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestUpdate_OverwritesPayloadButPreservesCreatedAt(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	saved, err := s.Create(ctx, fixture("alpha"))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	createdAt := saved.CreatedAt
	time.Sleep(2 * time.Millisecond) // ensure UpdatedAt advances measurably

	revised := fixture("alpha-v2")
	revised.Width = 8
	revised.Height = 8
	revised.Payload.Meta.Width = 8
	revised.Payload.Meta.Height = 8

	updated, err := s.Update(ctx, saved.ID.String(), revised)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if !updated.CreatedAt.Equal(createdAt) {
		t.Errorf("CreatedAt should be preserved: want %v got %v", createdAt, updated.CreatedAt)
	}
	if !updated.UpdatedAt.After(createdAt) {
		t.Errorf("UpdatedAt should advance: %v vs %v", updated.UpdatedAt, createdAt)
	}
	if updated.Name != "alpha-v2" {
		t.Errorf("name not updated: %q", updated.Name)
	}
	if updated.ID != saved.ID {
		t.Errorf("ID should remain %v, got %v", saved.ID, updated.ID)
	}
}

func TestUpdate_MissingFails(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Update(context.Background(), uuid.NewString(), fixture("ghost"))
	if !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDelete_MissingReturnsNotFound(t *testing.T) {
	s := newTestStore(t)
	err := s.Delete(context.Background(), uuid.NewString())
	if !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDelete_RemovesFile(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	saved, err := s.Create(ctx, fixture("alpha"))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := s.Delete(ctx, saved.ID.String()); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := s.Get(ctx, saved.ID.String()); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestList_Empty(t *testing.T) {
	s := newTestStore(t)
	items, total, err := s.List(context.Background(), model.ListTemplatesQueryParams{Limit: 20})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 0 || len(items) != 0 {
		t.Errorf("expected empty, got total=%d items=%d", total, len(items))
	}
}

func TestList_SortsByUpdatedAtDesc(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	first, _ := s.Create(ctx, fixture("first"))
	time.Sleep(2 * time.Millisecond)
	second, _ := s.Create(ctx, fixture("second"))
	time.Sleep(2 * time.Millisecond)
	third, _ := s.Create(ctx, fixture("third"))

	items, total, err := s.List(ctx, model.ListTemplatesQueryParams{Limit: 10})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 3 {
		t.Errorf("total: want 3, got %d", total)
	}
	if items[0].ID != third.ID || items[1].ID != second.ID || items[2].ID != first.ID {
		t.Errorf("wrong order: %v", []uuid.UUID{items[0].ID, items[1].ID, items[2].ID})
	}
}

func TestList_PaginationAndNameFilter(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	for _, n := range []string{"alpha-1", "alpha-2", "beta-1", "alpha-3"} {
		if _, err := s.Create(ctx, fixture(n)); err != nil {
			t.Fatalf("Create %s: %v", n, err)
		}
		time.Sleep(time.Millisecond)
	}

	items, total, err := s.List(ctx, model.ListTemplatesQueryParams{Limit: 2, Offset: 0, NameLike: "alpha"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 3 {
		t.Errorf("alpha total: want 3, got %d", total)
	}
	if len(items) != 2 {
		t.Errorf("page size: want 2, got %d", len(items))
	}

	page2, _, _ := s.List(ctx, model.ListTemplatesQueryParams{Limit: 2, Offset: 2, NameLike: "alpha"})
	if len(page2) != 1 {
		t.Errorf("page2 size: want 1, got %d", len(page2))
	}
}

func TestList_FiltersByRoomTypeAndCounts(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	// ComputeTemplateStats recomputes counts from the payload, so we vary the
	// Static layer rather than setting StaticCount directly.
	mk := func(name, room string, staticCells [][]int) model.Template {
		tpl := fixture(name)
		tpl.RoomType = &room
		tpl.Payload.Static = staticCells
		return tpl
	}

	// "a" has 2 static cells (in fixture default), "b" has 0, "c" has 6.
	zero := [][]int{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}}
	many := [][]int{{1, 1, 0, 0}, {1, 1, 0, 0}, {0, 0, 1, 1}, {0, 0, 0, 0}}

	for _, tpl := range []model.Template{
		mk("a", "full", fixture("").Payload.Static), // 2 statics
		mk("b", "bridge", zero),                     // 0 statics
		mk("c", "full", many),                       // 6 statics
	} {
		if _, err := s.Create(ctx, tpl); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	min := 1
	max := 5
	items, total, err := s.List(ctx, model.ListTemplatesQueryParams{
		Limit:          10,
		RoomType:       "full",
		MinStaticCount: &min,
		MaxStaticCount: &max,
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 1 {
		t.Errorf("filtered total: want 1, got %d", total)
	}
	if len(items) != 1 || items[0].Name != "a" {
		t.Errorf("expected only 'a', got %+v", items)
	}
}

func TestList_IgnoresTempFilesAndUnreadable(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	if _, err := s.Create(ctx, fixture("ok")); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Leftover atomic-write temp file from a hypothetical crash.
	tmp := filepath.Join(s.RootDir(), uuid.NewString()+".json.tmp")
	if err := os.WriteFile(tmp, []byte("{}"), 0o644); err != nil {
		t.Fatalf("write tmp: %v", err)
	}
	// Corrupted entry.
	bad := filepath.Join(s.RootDir(), uuid.NewString()+".json")
	if err := os.WriteFile(bad, []byte("not json"), 0o644); err != nil {
		t.Fatalf("write bad: %v", err)
	}

	items, total, err := s.List(ctx, model.ListTemplatesQueryParams{Limit: 10})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 1 || len(items) != 1 || items[0].Name != "ok" {
		t.Errorf("expected only 'ok', got %+v", items)
	}
}

func TestList_DoorConnectivityFilter(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	yes, no := true, false

	mk := func(name string, top bool) model.Template {
		t := fixture(name)
		t.DoorsConnected = &model.DoorsConnected{Top: top}
		return t
	}
	for _, tpl := range []model.Template{mk("topOpen", true), mk("topClosed", false)} {
		if _, err := s.Create(ctx, tpl); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	open, total, err := s.List(ctx, model.ListTemplatesQueryParams{Limit: 10, TopDoorConnected: &yes})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 1 || open[0].Name != "topOpen" {
		t.Errorf("topOpen filter: %+v total=%d", open, total)
	}

	closed, total2, _ := s.List(ctx, model.ListTemplatesQueryParams{Limit: 10, TopDoorConnected: &no})
	if total2 != 1 || closed[0].Name != "topClosed" {
		t.Errorf("topClosed filter: %+v total=%d", closed, total2)
	}
}

func TestWriteAtomic_NoTempLeftOnSuccess(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	if _, err := s.Create(ctx, fixture("alpha")); err != nil {
		t.Fatalf("Create: %v", err)
	}
	entries, _ := os.ReadDir(s.RootDir())
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("found stale temp file: %s", e.Name())
		}
	}
}

func TestHealthCheck(t *testing.T) {
	s := newTestStore(t)
	if err := s.HealthCheck(context.Background()); err != nil {
		t.Errorf("HealthCheck: %v", err)
	}
}

func TestHealthCheck_FailsWhenRootMissing(t *testing.T) {
	s := newTestStore(t)
	if err := os.RemoveAll(s.RootDir()); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if err := s.HealthCheck(context.Background()); err == nil {
		t.Fatal("expected HealthCheck to fail with missing dir")
	}
}
