package api

import (
 "errors"
 "net/http"
 "github.com/chrisbirster/cleartify/internal/protocols"
)
func(s *Server) createCreatorDraft(w http.ResponseWriter,r *http.Request){
 if s.socialAuth==nil||s.creatorDrafts==nil {writeJSON(w,503,map[string]string{"error":"creator_unavailable"});return}
 if s.appOrigin==""||r.Header.Get("Origin")!=s.appOrigin {writeJSON(w,403,map[string]string{"error":"invalid_origin"});return}
 did,ok:=s.socialAuth.SocialSessionDID(r)
 if !ok {writeJSON(w,401,map[string]string{"error":"atproto_login_required"});return}
 var in protocols.DraftInput
 if !decodeJSON(w,r,&in){return}
 id,err:=s.creatorDrafts.Create(r.Context(),did,in)
 if err!=nil {
  switch{
  case errors.Is(err,protocols.ErrInvalidDraft):writeJSON(w,400,map[string]string{"error":"invalid_draft"})
  case errors.Is(err,protocols.ErrDraftConflict):writeJSON(w,409,map[string]string{"error":"version_conflict"})
  default:writeJSON(w,500,map[string]string{"error":"internal_error"})
  }
  return
 }
 w.Header().Set("Cache-Control","no-store")
 writeJSON(w,http.StatusCreated,map[string]any{"id":id,"status":"draft","published":false})
}
