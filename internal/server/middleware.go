package server

import (
	"net/http"

	"cleartify/internal/httpx"
)

func (s *Server) allowMethod(method string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			httpx.MethodNotAllowed(w)
			return
		}
		next(w, r)
	}
}
