package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TilemapTagsFile is OZX's tilemap tag registry (ORT-158 / OZX-2152). It sits at
// the root of the TilemapData folder, next to the category folders, and holds an
// ordered array of names: a name's index is its bit in RoomTilemapData.TagMask.
const TilemapTagsFile = "tilemap_tags.json"

// MaxTilemapTags is the registry's ceiling: the game keeps the mask in a uint.
const MaxTilemapTags = 32

// TagRegistry is the registry as the editor sees it. Found is false when the
// project has no registry file, which is not an error: such a project simply
// has no tags to offer. Err is set when the file exists but the game would
// refuse it (bad JSON, over 32 names, empty or duplicate names).
type TagRegistry struct {
	Tags  []string `json:"tags"`
	Found bool     `json:"found"`
	Path  string   `json:"path,omitempty"`
	Err   string   `json:"error,omitempty"`
}

// TagRegistryProvider is implemented by stores that live inside an OZX
// project and can therefore read its registry.
type TagRegistryProvider interface {
	TilemapTags() TagRegistry
}

// TilemapTagsOf returns s's registry, or an empty "not found" one when the
// store has none (the stub store).
func TilemapTagsOf(s Store) TagRegistry {
	if p, ok := s.(TagRegistryProvider); ok {
		return p.TilemapTags()
	}
	return TagRegistry{Tags: []string{}}
}

// LoadTagRegistry reads <dir>/tilemap_tags.json and applies the same rules the
// game's loader does, so the editor never offers a registry OZX would reject.
func LoadTagRegistry(dir string) TagRegistry {
	path := filepath.Join(dir, TilemapTagsFile)
	reg := TagRegistry{Tags: []string{}, Path: path}
	data, err := os.ReadFile(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			reg.Found = true
			reg.Err = err.Error()
		}
		return reg
	}
	reg.Found = true
	var names []string
	if err := json.Unmarshal(data, &names); err != nil {
		reg.Err = fmt.Sprintf("%s is not a JSON array of names: %v", TilemapTagsFile, err)
		return reg
	}
	if len(names) > MaxTilemapTags {
		reg.Err = fmt.Sprintf("%s lists %d names; the mask holds at most %d", TilemapTagsFile, len(names), MaxTilemapTags)
		return reg
	}
	seen := make(map[string]bool, len(names))
	for i, n := range names {
		if strings.TrimSpace(n) == "" {
			reg.Err = fmt.Sprintf("%s entry %d is empty", TilemapTagsFile, i)
			return reg
		}
		if seen[n] {
			reg.Err = fmt.Sprintf("%s lists %q twice", TilemapTagsFile, n)
			return reg
		}
		seen[n] = true
	}
	reg.Tags = names
	return reg
}

// CheckTags reports every problem with a template's tags against reg: a name
// the registry does not list (the game throws on load), or a name given twice.
// An untagged template is always fine, whatever state the registry is in.
func CheckTags(tags []string, reg TagRegistry) []string {
	if len(tags) == 0 {
		return nil
	}
	var problems []string
	if reg.Err != "" {
		return []string{"the tag registry is invalid: " + reg.Err}
	}
	known := make(map[string]bool, len(reg.Tags))
	for _, n := range reg.Tags {
		known[n] = true
	}
	seen := make(map[string]bool, len(tags))
	for _, n := range tags {
		switch {
		case seen[n]:
			problems = append(problems, fmt.Sprintf("tag %q is listed twice", n))
		case !known[n] && !reg.Found:
			problems = append(problems, fmt.Sprintf("tag %q: this project has no %s", n, TilemapTagsFile))
		case !known[n]:
			problems = append(problems, fmt.Sprintf("tag %q is not in %s", n, TilemapTagsFile))
		}
		seen[n] = true
	}
	return problems
}
