// Package web embeds the built SPA so the standalone binary
// (cmd/ozx-roomeditor) can serve the frontend from the same process that
// serves the API.
//
// The Vite build output is copied into ./dist by `make build-local`. A stub
// index.html is committed so the binary still compiles before the frontend
// has been built.
package web

import (
	"embed"
	"io/fs"
)

//go:embed dist
var raw embed.FS

// Assets returns the embedded SPA file system, rooted at the dist directory
// (so callers can serve "/" → "dist/index.html" without a prefix).
func Assets() (fs.FS, error) {
	return fs.Sub(raw, "dist")
}
