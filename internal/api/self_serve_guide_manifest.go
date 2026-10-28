package api

import (
 "net/http"
 "github.com/chrisbirster/cleartify/internal/protocols"
)

// Same publication visibility constraints as selfServeGuideInstall.
// A public manifest is discoverable, but installation still requires a private
// Clear cookie and server-side permission, digest and version rechecks.
func(s *Server) selfServeGuideManifest(w http.ResponseWriter,r *http.Request){
 w.Header().Set("Cache-Control","no-store")
 if s.installFlow==nil || s.installFlow.Store==nil {
  writeJSON(w,503,map[string]string{"error":"unavailable"});return
 }
 id:=r.PathValue("id")
 if id=="" || len(id)>128 {writeJSON(w,404,map[string]string{"error":"publication_unavailable"});return}
 var count int
 err:=s.installFlow.Store.DB.QueryRowContext(r.Context(),`SELECT COUNT(*) FROM published_protocols p WHERE p.id=? AND p.status='published'
 AND EXISTS(SELECT 1 FROM creator_publications c WHERE c.protocol_id=p.id AND c.visibility='public')
 AND NOT EXISTS(SELECT 1 FROM creator_takedown_requests t WHERE t.protocol_id=p.id AND t.status='hidden')`,id).Scan(&count)
 if err!=nil {writeJSON(w,503,map[string]string{"error":"unavailable"});return}
 if count!=1 {writeJSON(w,404,map[string]string{"error":"publication_unavailable"});return}
 manifest,err:=s.installFlow.Preview(r.Context(),id)
 if err!=nil {
  if err==protocols.ErrInstallUnavailable {writeJSON(w,404,map[string]string{"error":"publication_unavailable"});return}
  writeJSON(w,503,map[string]string{"error":"unavailable"});return
 }
 writeJSON(w,200,map[string]any{"manifestSha256":manifest.ManifestSHA256,"specSha256":manifest.SpecSHA256,"permissions":manifest.Permissions,"version":manifest.Version})
}
