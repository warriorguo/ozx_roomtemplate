package http

import (
	"io"
	"io/fs"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// MountFrontend registers a catch-all under "/" that serves the SPA bundle.
// API routes registered earlier on the router (e.g. /api/v1/*, /health) take
// precedence because chi routes the most specific match first.
//
// The handler implements the standard single-page-app fallback: requests for
// real files in the bundle (index.html, /assets/foo.js, ...) are served
// directly; everything else returns index.html so the client-side router
// can take over.
func MountFrontend(r chi.Router, files fs.FS, logger *zap.Logger) error {
	if _, err := fs.Stat(files, "index.html"); err != nil {
		return err
	}
	fileServer := http.FileServer(http.FS(files))

	r.Get("/*", func(w http.ResponseWriter, req *http.Request) {
		rel := strings.TrimPrefix(req.URL.Path, "/")
		if rel == "" {
			serveIndex(w, req, files, logger)
			return
		}
		if info, err := fs.Stat(files, rel); err == nil && !info.IsDir() {
			fileServer.ServeHTTP(w, req)
			return
		}
		serveIndex(w, req, files, logger)
	})
	return nil
}

func serveIndex(w http.ResponseWriter, req *http.Request, files fs.FS, logger *zap.Logger) {
	f, err := files.Open("index.html")
	if err != nil {
		logger.Error("Failed to open index.html", zap.Error(err))
		http.Error(w, "frontend bundle missing", http.StatusInternalServerError)
		return
	}
	defer f.Close()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// SPA shell should not be cached aggressively or users will get stale UI
	// after a redeploy; assets under /assets/ are hashed and benefit from
	// long caching, but the FileServer already does the right thing there.
	w.Header().Set("Cache-Control", "no-cache")
	if _, err := io.Copy(w, f); err != nil {
		logger.Debug("Copy index.html: client likely disconnected", zap.Error(err))
	}
}
