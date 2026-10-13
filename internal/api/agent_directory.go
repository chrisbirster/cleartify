package api

import (
 "log/slog"
 "net/http"
 "github.com/chrisbirster/cleartify/internal/agents"
)
func(s *Server) listAgentDirectory(w http.ResponseWriter,r *http.Request){
 if s.agentDirectoryDB==nil {writeJSON(w,503,map[string]string{"error":"not_configured"});return}
 entries,err:=agents.Directory(r.Context(),s.agentDirectoryDB,20)
 if err!=nil {slog.Error("agent directory unavailable","error",err);writeJSON(w,503,map[string]string{"error":"directory_unavailable"});return}
 w.Header().Set("Cache-Control","no-store")
 writeJSON(w,http.StatusOK,map[string]any{"agents":entries})
}
