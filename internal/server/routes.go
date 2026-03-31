package server

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/api/health", s.allowMethod("GET", s.handleHealth))
	s.mux.HandleFunc("/api/auth/session", s.allowMethod("GET", s.handleSession))
	s.mux.HandleFunc("/api/signals", s.allowMethod("GET", s.handleSignals))
	s.mux.HandleFunc("/api/clusters", s.allowMethod("GET", s.handleClusters))
	s.mux.HandleFunc("/api/clusters/", s.handleClusterActions)
	s.mux.HandleFunc("/api/work-objects", s.allowMethod("GET", s.handleWorkObjects))
	s.mux.HandleFunc("/api/work-objects/", s.handleWorkObjectActions)
	s.mux.HandleFunc("/app", s.handleApp)
	s.mux.HandleFunc("/app/", s.handleApp)
	s.mux.HandleFunc("/", s.handleRoot)
}
