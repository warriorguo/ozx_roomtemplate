package http

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tile-backend/internal/model"
	"tile-backend/internal/store"
	"tile-backend/internal/store/fsstore"

	"go.uber.org/zap"
)

// ORT-158: the editor offers the project's registry and a save refuses a tag
// the game would throw on.
func TestTilemapTags_RegistryEndpointAndSaveCheck(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, store.TilemapTagsFile), []byte(`["dark","wide"]`), 0o644); err != nil {
		t.Fatal(err)
	}
	fs, err := fsstore.New(dir)
	if err != nil {
		t.Fatalf("fsstore.New: %v", err)
	}
	srv := httptest.NewServer(SetupRouter(store.NewSwappableStore(fs), nil, zap.NewNop(), nil))
	t.Cleanup(srv.Close)

	var reg store.TagRegistry
	mustDecode(t, getJSON(t, srv.URL+"/api/v1/tilemap-tags", http.StatusOK), &reg)
	if !reg.Found || strings.Join(reg.Tags, ",") != "dark,wide" {
		t.Fatalf("registry = %+v, want found [dark wide]", reg)
	}

	z := [][]int{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}}
	body := func(tags ...string) model.CreateTemplateRequest {
		return model.CreateTemplateRequest{Name: "tagged", Payload: model.TemplatePayload{
			Ground: [][]int{{1, 1, 1, 1}, {1, 1, 1, 1}, {1, 1, 1, 1}, {1, 1, 1, 1}},
			Static: z, Chaser: z, Zoner: z, DPS: z, MobAir: z, Tags: tags,
			Meta: model.TemplateMeta{Name: "tagged", Version: 1, Width: 4, Height: 4},
		}}
	}

	resp := postJSON(t, srv.URL+"/api/v1/templates", body("dark", "nope"), http.StatusBadRequest)
	if !strings.Contains(string(resp), "nope") {
		t.Errorf("rejection does not name the unregistered tag: %s", resp)
	}

	var created model.CreateTemplateResponse
	mustDecode(t, postJSON(t, srv.URL+"/api/v1/templates", body("wide", "dark"), http.StatusCreated), &created)
	var got model.Template
	mustDecode(t, getJSON(t, srv.URL+"/api/v1/templates/"+created.ID.String(), http.StatusOK), &got)
	if strings.Join(got.Payload.Tags, ",") != "wide,dark" {
		t.Errorf("tags via the API = %v, want [wide dark] as authored", got.Payload.Tags)
	}
}
