package server

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/api/health", s.allowMethod("GET", s.handleHealth))
	s.mux.HandleFunc("/api/auth/session", s.allowMethod("GET", s.handleSession))
	s.mux.HandleFunc("/api/me", s.allowMethod("GET", s.handleMe))
	s.mux.HandleFunc("/api/me/onboarding/complete", s.allowMethod("POST", s.handleCompleteOnboarding))
	s.mux.HandleFunc("/api/team", s.allowMethod("GET", s.handleTeam))
	s.mux.HandleFunc("/api/signals", s.handleSignals)
	s.mux.HandleFunc("/api/signals/", s.handleSignalActions)
	s.mux.HandleFunc("/api/clusters", s.allowMethod("GET", s.handleClusters))
	s.mux.HandleFunc("/api/clusters/", s.handleClusterActions)
	s.mux.HandleFunc("/api/decisions", s.handleDecisions)
	s.mux.HandleFunc("/api/decisions/", s.handleDecisionActions)
	s.mux.HandleFunc("/api/artifacts", s.handleArtifacts)
	s.mux.HandleFunc("/api/artifacts/", s.handleArtifactActions)
	s.mux.HandleFunc("/api/work-objects", s.allowMethod("GET", s.handleWorkObjects))
	s.mux.HandleFunc("/api/work-objects/", s.handleWorkObjectActions)
	s.mux.HandleFunc("/login", s.allowMethod("GET", s.handleLogin))
	s.mux.HandleFunc("/auth/callback", s.allowMethod("GET", s.handleAuthCallback))
	s.mux.HandleFunc("/logout", s.allowMethod("GET", s.handleLogout))
	s.mux.HandleFunc("/", s.handleRoot)
}
