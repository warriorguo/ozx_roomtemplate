package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeRegistry(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, TilemapTagsFile), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestLoadTagRegistry(t *testing.T) {
	if reg := LoadTagRegistry(t.TempDir()); reg.Found || reg.Err != "" || len(reg.Tags) != 0 {
		t.Errorf("no file: got %+v, want not found, no error, no tags", reg)
	}

	reg := LoadTagRegistry(writeRegistry(t, `["dark","wide","boss_ready"]`))
	if !reg.Found || reg.Err != "" || strings.Join(reg.Tags, ",") != "dark,wide,boss_ready" {
		t.Errorf("valid file: got %+v, want the names in file order", reg)
	}

	many := make([]string, MaxTilemapTags+1)
	for i := range many {
		many[i] = `"t` + string(rune('a'+i%26)) + string(rune('a'+i/26)) + `"`
	}
	bad := map[string]string{
		"not an array": `{"tags":["a"]}`,
		"empty name":   `["a",""]`,
		"duplicate":    `["a","b","a"]`,
		"over 32":      "[" + strings.Join(many, ",") + "]",
	}
	for name, body := range bad {
		reg := LoadTagRegistry(writeRegistry(t, body))
		if !reg.Found || reg.Err == "" || len(reg.Tags) != 0 {
			t.Errorf("%s: got %+v, want found with an error and no tags offered", name, reg)
		}
	}
}

func TestCheckTags(t *testing.T) {
	reg := TagRegistry{Tags: []string{"dark", "wide"}, Found: true}
	if p := CheckTags(nil, TagRegistry{Err: "broken"}); p != nil {
		t.Errorf("untagged template must pass whatever the registry state, got %v", p)
	}
	if p := CheckTags([]string{"wide", "dark"}, reg); p != nil {
		t.Errorf("registered tags in any order: got %v", p)
	}
	if p := CheckTags([]string{"dark", "nope"}, reg); len(p) != 1 || !strings.Contains(p[0], `"nope"`) {
		t.Errorf("unregistered tag: got %v", p)
	}
	if p := CheckTags([]string{"dark", "dark"}, reg); len(p) != 1 || !strings.Contains(p[0], "twice") {
		t.Errorf("duplicate tag: got %v", p)
	}
	if p := CheckTags([]string{"dark"}, TagRegistry{}); len(p) != 1 || !strings.Contains(p[0], "has no") {
		t.Errorf("no registry: got %v", p)
	}
}
