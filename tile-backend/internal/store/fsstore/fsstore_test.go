package fsstore

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"tile-backend/internal/model"
	"tile-backend/internal/store"
)

// fixture builds a 4x4 template with the metadata the OZX filename pattern
// needs (shape, stage, openDoors, category).
func fixture(name string) model.Template {
	shape := "all"
	stage := "boss"
	category := "normal"
	openDoors := 3
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
			RoomShape:    &shape,
			StageType:    &stage,
			RoomCategory: &category,
			OpenDoors:    &openDoors,
			Meta: model.TemplateMeta{Name: name, Version: 1, Width: 4, Height: 4},
		},
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

func TestNew_EmptyRoot(t *testing.T) {
	if _, err := New(""); err == nil {
		t.Fatal("expected error for empty rootDir")
	}
}

func TestCreate_AllocatesOZXFilename(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	saved, err := s.Create(ctx, fixture("alpha"))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	want := filepath.Join(s.RootDir(), "normal", "all_boss_3_01.json")
	if _, err := os.Stat(want); err != nil {
		t.Errorf("expected file at %s: %v", want, err)
	}
	if saved.RoomCategory == nil || *saved.RoomCategory != "normal" {
		t.Errorf("RoomCategory not propagated: %+v", saved.RoomCategory)
	}
}

func TestCreate_SequencesWithinCategory(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	for i := 1; i <= 3; i++ {
		if _, err := s.Create(ctx, fixture("alpha")); err != nil {
			t.Fatalf("Create %d: %v", i, err)
		}
	}
	for i := 1; i <= 3; i++ {
		want := filepath.Join(s.RootDir(), "normal", "all_boss_3_"+twoDigits(i)+".json")
		if _, err := os.Stat(want); err != nil {
			t.Errorf("seq %d missing at %s", i, want)
		}
	}
}

func TestCreate_FillsGapsInSequence(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		if _, err := s.Create(ctx, fixture("a")); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}
	// Delete the middle one; next Create should reuse the freed seq.
	mid := filepath.Join(s.RootDir(), "normal", "all_boss_3_02.json")
	if err := os.Remove(mid); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if _, err := s.Create(ctx, fixture("filler")); err != nil {
		t.Fatalf("Create after gap: %v", err)
	}
	if _, err := os.Stat(mid); err != nil {
		t.Errorf("expected seq 02 reused: %v", err)
	}
}

func TestGet_AcceptsBothIDForms(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	saved, err := s.Create(ctx, fixture("alpha"))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Form 1: the synthesised UUID round-trips (so the existing
	// /api/v1/templates/{uuid} surface keeps working unchanged).
	got, err := s.Get(ctx, saved.ID.String())
	if err != nil {
		t.Fatalf("Get by UUID: %v", err)
	}
	if got.Payload.Meta.Width != 4 {
		t.Errorf("payload not retrieved: %+v", got.Payload.Meta)
	}

	// Form 2: the path-derived canonical id works too.
	id := "normal" + idSeparator + "all_boss_3_01"
	got, err = s.Get(ctx, id)
	if err != nil {
		t.Fatalf("Get by path id: %v", err)
	}
	if got.StaticCount == nil || *got.StaticCount != 2 {
		t.Errorf("StaticCount: want 2, got %v", got.StaticCount)
	}
}

func TestGet_OnDiskFileIsBarePayload(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	if _, err := s.Create(ctx, fixture("alpha")); err != nil {
		t.Fatalf("Create: %v", err)
	}
	path := filepath.Join(s.RootDir(), "normal", "all_boss_3_01.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	// The on-disk format is a bare TemplatePayload — no top-level "id"
	// or "created_at" envelope keys.
	var top map[string]json.RawMessage
	if err := json.Unmarshal(raw, &top); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if _, hasEnvelope := top["id"]; hasEnvelope {
		t.Errorf("file should not contain envelope key 'id'")
	}
	if _, hasGround := top["ground"]; !hasGround {
		t.Errorf("file should contain payload key 'ground'")
	}
	if _, hasMeta := top["meta"]; !hasMeta {
		t.Errorf("file should contain payload key 'meta'")
	}
}

func TestGet_NotFound(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Get(context.Background(), "normal__missing")
	if !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUpdate_OverwritesButKeepsFilename(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	saved, _ := s.Create(ctx, fixture("alpha"))
	id := "normal" + idSeparator + "all_boss_3_01"

	revised := fixture("alpha-v2")
	revised.Width = 8
	revised.Height = 8
	revised.Payload.Meta.Width = 8
	revised.Payload.Meta.Height = 8

	time.Sleep(2 * time.Millisecond)
	updated, err := s.Update(ctx, id, revised)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Width != 8 {
		t.Errorf("payload not updated: %+v", updated.Payload.Meta)
	}
	// Filename should be unchanged.
	if _, err := os.Stat(filepath.Join(s.RootDir(), "normal", "all_boss_3_01.json")); err != nil {
		t.Errorf("filename should be stable across updates: %v", err)
	}
	// The synthesised ID is path-derived, so it should match the original.
	if updated.ID != saved.ID {
		t.Errorf("ID drift across update: saved=%v updated=%v", saved.ID, updated.ID)
	}
}

func TestUpdate_MissingFails(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Update(context.Background(), "normal__nope", fixture("ghost"))
	if !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDelete_RemovesJSONAndMeta(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	if _, err := s.Create(ctx, fixture("alpha")); err != nil {
		t.Fatalf("Create: %v", err)
	}
	jsonPath := filepath.Join(s.RootDir(), "normal", "all_boss_3_01.json")
	metaPath := jsonPath + ".meta"
	if err := os.WriteFile(metaPath, []byte("fileFormatVersion: 2"), 0o644); err != nil {
		t.Fatalf("write meta: %v", err)
	}

	if err := s.Delete(ctx, "normal"+idSeparator+"all_boss_3_01"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := os.Stat(jsonPath); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("json should be gone")
	}
	if _, err := os.Stat(metaPath); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("meta should also be removed")
	}
}

func TestDelete_NotFound(t *testing.T) {
	s := newTestStore(t)
	err := s.Delete(context.Background(), "normal__nope")
	if !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestList_SkipsMetaAndCorrupted(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	if _, err := s.Create(ctx, fixture("ok")); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Pretend a Unity meta file is sitting next to our template.
	jsonPath := filepath.Join(s.RootDir(), "normal", "all_boss_3_01.json")
	if err := os.WriteFile(jsonPath+".meta", []byte("yaml"), 0o644); err != nil {
		t.Fatalf("write meta: %v", err)
	}
	// Corrupt JSON should not derail the listing.
	bad := filepath.Join(s.RootDir(), "normal", "all_boss_3_02.json")
	if err := os.WriteFile(bad, []byte("not json"), 0o644); err != nil {
		t.Fatalf("write bad: %v", err)
	}
	// Stray .tmp from an interrupted prior write.
	tmp := filepath.Join(s.RootDir(), "normal", "all_boss_3_99.json.tmp")
	if err := os.WriteFile(tmp, []byte("{}"), 0o644); err != nil {
		t.Fatalf("write tmp: %v", err)
	}

	items, total, err := s.List(ctx, model.ListTemplatesQueryParams{Limit: 10})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 1 || len(items) != 1 || items[0].Name != "ok" {
		t.Errorf("expected only 'ok', got total=%d items=%+v", total, items)
	}
}

func TestList_SortsByMTimeDesc(t *testing.T) {
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
		t.Errorf("wrong order: got %v %v %v", items[0].ID, items[1].ID, items[2].ID)
	}
}

func TestList_WalksAllCategorySubfolders(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	makeIn := func(cat, name string) {
		tpl := fixture(name)
		c := cat
		tpl.Payload.RoomCategory = &c
		if _, err := s.Create(ctx, tpl); err != nil {
			t.Fatalf("Create %s/%s: %v", cat, name, err)
		}
	}
	for _, cat := range []string{"normal", "basement", "cave", "test"} {
		makeIn(cat, cat+"-room")
	}

	_, total, err := s.List(ctx, model.ListTemplatesQueryParams{Limit: 10})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 4 {
		t.Errorf("expected one per category (4), got %d", total)
	}
}

func TestList_PaginationAndNameFilter(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	for _, n := range []string{"alpha-1", "alpha-2", "beta-1", "alpha-3"} {
		tpl := fixture(n)
		if _, err := s.Create(ctx, tpl); err != nil {
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

func TestPathForID_RejectsBadInput(t *testing.T) {
	s := newTestStore(t)
	for _, bad := range []string{"", "no-separator", "__nope", "cat__", "cat/with/slash__base", "cat__base/with/slash"} {
		if _, err := s.pathForID(bad); err == nil {
			t.Errorf("expected error for id %q", bad)
		}
	}
}

func TestWriteAtomic_NoTempLeftOnSuccess(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	if _, err := s.Create(ctx, fixture("alpha")); err != nil {
		t.Fatalf("Create: %v", err)
	}
	cat := filepath.Join(s.RootDir(), "normal")
	entries, _ := os.ReadDir(cat)
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("found stale temp file: %s", e.Name())
		}
	}
}

func twoDigits(n int) string {
	if n < 10 {
		return "0" + string(rune('0'+n))
	}
	return string(rune('0'+n/10)) + string(rune('0'+n%10))
}
