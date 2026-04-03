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
	case strings.HasPrefix(r.URL.Path, "/api/"):
		httpx.WriteError(w, http.StatusNotFound, "not_found", "API route not found.")
	case r.URL.Path == "/" || r.URL.Path == "":
		s.serveIndex(w)
	case r.URL.Path == "/app" || r.URL.Path == "/app/":
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	case strings.HasPrefix(r.URL.Path, "/app/"):
		http.Redirect(w, r, strings.TrimPrefix(r.URL.Path, "/app"), http.StatusTemporaryRedirect)
	case strings.HasPrefix(r.URL.Path, "/assets/"):
		s.staticAssets.ServeHTTP(w, r)
	case path.Ext(strings.TrimPrefix(r.URL.Path, "/")) != "":
		s.staticAssets.ServeHTTP(w, r)
	case r.Method == http.MethodGet || r.Method == http.MethodHead:
		s.serveIndex(w)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	s.handleApp(w, r)
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
