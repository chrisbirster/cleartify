package server

import (
	"io/fs"
	"net/http"
	"path"
	"strings"

	"cleartify/internal/httpx"
)

func (s *Server) handleApp(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/app" || r.URL.Path == "/app/":
		s.serveIndex(w)
	case strings.HasPrefix(r.URL.Path, "/app/assets/"):
		s.staticAssets.ServeHTTP(w, r)
	case path.Ext(strings.TrimPrefix(r.URL.Path, "/app/")) != "":
		s.staticAssets.ServeHTTP(w, r)
	default:
		s.serveIndex(w)
	}
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.HasPrefix(r.URL.Path, "/api/"):
		httpx.WriteError(w, http.StatusNotFound, "not_found", "API route not found.")
	case r.Method == http.MethodGet:
		http.Redirect(w, r, s.config.AppMarketingURL, http.StatusTemporaryRedirect)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) serveIndex(w http.ResponseWriter) {
	indexHTML, err := fs.ReadFile(s.frontend, "index.html")
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "frontend_missing", "Embedded frontend index is missing.")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(indexHTML)
}
