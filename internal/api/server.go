package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/chrisbirster/cleartify/internal/waitlist"
)

type Server struct {
	waitlist  *waitlist.Service
	appOrigin string
}

func New(waitlistService *waitlist.Service, appOrigin string) http.Handler {
	s := &Server{
		waitlist:  waitlistService,
		appOrigin: strings.TrimRight(appOrigin, "/"),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("POST /api/waitlist", s.joinWaitlist)

	return s.middleware(mux)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"service": "cleartify-api",
	})
}

func (s *Server) joinWaitlist(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 32<<10)

	var req struct {
		Name         string `json:"name"`
		Email        string `json:"email"`
		Phone        string `json:"phone"`
		EmailUpdates bool   `json:"emailUpdates"`
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "invalid_request",
		})
		return
	}

	_, err := s.waitlist.Join(r.Context(), waitlist.Signup{
		Name:         req.Name,
		Email:        req.Email,
		Phone:        req.Phone,
		EmailUpdates: req.EmailUpdates,
		Source:       "landing",
	})
	if err != nil {
		var validation *waitlist.ValidationError
		switch {
		case errors.As(err, &validation):
			writeJSON(w, http.StatusBadRequest, map[string]any{
				"error":   "validation_error",
				"field":   validation.Field,
				"message": validation.Message,
			})
		case errors.Is(err, waitlist.ErrPhoneInUse):
			writeJSON(w, http.StatusConflict, map[string]any{
				"error":   "phone_in_use",
				"message": "That mobile number is already on the waitlist with another email.",
			})
		default:
			slog.Error("waitlist signup failed", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]any{
				"error": "internal_error",
			})
		}
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"joined": true})
}

func (s *Server) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		if s.appOrigin != "" {
			origin := r.Header.Get("Origin")
			if origin == s.appOrigin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
				w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
			}
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		started := time.Now()
		next.ServeHTTP(w, r)
		slog.Debug("request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(started))
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Error("write response", "error", err)
	}
}
