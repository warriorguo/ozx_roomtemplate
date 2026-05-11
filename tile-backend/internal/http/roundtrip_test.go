package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"tile-backend/internal/model"
	"tile-backend/internal/store/fsstore"

	"go.uber.org/zap"
)

// TestFsStoreRoundTrip exercises POST → GET list → GET one → DELETE through the
// full handler stack with a real filesystem store under t.TempDir(). This is
// the acceptance test called out in ORT-66.
func TestFsStoreRoundTrip(t *testing.T) {
	fs, err := fsstore.New(t.TempDir())
	if err != nil {
		t.Fatalf("fsstore.New: %v", err)
	}
	logger := zap.NewNop()
	router := SetupRouter(fs, logger, nil)
	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)

	// 1. Create.
	createBody := model.CreateTemplateRequest{
		Name: "roundtrip",
		Payload: model.TemplatePayload{
			Ground: [][]int{{1, 1, 1, 1}, {1, 1, 1, 1}, {1, 1, 1, 1}, {1, 1, 1, 1}},
			Static: [][]int{{0, 1, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 1, 0, 0}},
			Chaser: [][]int{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}},
			Zoner:  [][]int{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}},
			DPS:    [][]int{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}},
			MobAir: [][]int{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}},
			Meta:   model.TemplateMeta{Name: "roundtrip", Version: 1, Width: 4, Height: 4},
		},
	}

	created := postJSON(t, srv.URL+"/api/v1/templates", createBody, http.StatusCreated)
	var createdResp model.CreateTemplateResponse
	mustDecode(t, created, &createdResp)
	if createdResp.Name != "roundtrip" {
		t.Fatalf("created name: %q", createdResp.Name)
	}
	id := createdResp.ID.String()

	// 2. List — should contain the new template.
	listBody := getJSON(t, srv.URL+"/api/v1/templates", http.StatusOK)
	var list model.ListTemplatesResponse
	mustDecode(t, listBody, &list)
	if list.Total != 1 || len(list.Items) != 1 || list.Items[0].ID.String() != id {
		t.Fatalf("list: total=%d items=%d", list.Total, len(list.Items))
	}

	// 3. Get by id.
	getBody := getJSON(t, srv.URL+"/api/v1/templates/"+id, http.StatusOK)
	var got model.Template
	mustDecode(t, getBody, &got)
	if got.Payload.Meta.Width != 4 {
		t.Errorf("payload not retrieved: %+v", got.Payload.Meta)
	}

	// 4. Delete.
	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/api/v1/templates/"+id, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("DELETE status: want 204, got %d", resp.StatusCode)
	}

	// 5. Confirm it's gone (404).
	resp2, err := http.Get(srv.URL + "/api/v1/templates/" + id)
	if err != nil {
		t.Fatalf("GET after delete: %v", err)
	}
	_ = resp2.Body.Close()
	if resp2.StatusCode != http.StatusNotFound {
		t.Errorf("after delete: want 404, got %d", resp2.StatusCode)
	}
}

func postJSON(t *testing.T, url string, body interface{}, wantStatus int) []byte {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != wantStatus {
		t.Fatalf("POST %s: want %d, got %d: %s", url, wantStatus, resp.StatusCode, data)
	}
	return data
}

func getJSON(t *testing.T, url string, wantStatus int) []byte {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != wantStatus {
		t.Fatalf("GET %s: want %d, got %d: %s", url, wantStatus, resp.StatusCode, data)
	}
	return data
}

func mustDecode(t *testing.T, data []byte, v interface{}) {
	t.Helper()
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("decode: %v: %s", err, data)
	}
}
