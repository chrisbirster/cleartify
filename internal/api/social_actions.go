package api

import (
 "errors"
 "net/http"

 "github.com/chrisbirster/cleartify/internal/social"
)
func(s *Server) authenticatedSocialMutation(w http.ResponseWriter,r *http.Request)(string,bool){
 if s.socialAuth==nil||s.socialActions==nil {writeJSON(w,503,map[string]string{"error":"social_unavailable"});return "",false}
 if s.appOrigin==""||r.Header.Get("Origin")!=s.appOrigin {writeJSON(w,403,map[string]string{"error":"invalid_origin"});return "",false}
 did,ok:=s.socialAuth.SocialSessionDID(r)
 if !ok{writeJSON(w,401,map[string]string{"error":"atproto_login_required"});return "",false}
 if s.moderationBroader && !s.enforceSubjectHold(w,r,did){return "",false}
 return did,true
}
func(s *Server) claimOwnProfile(w http.ResponseWriter,r *http.Request){
 did,ok:=s.authenticatedSocialMutation(w,r);if !ok{return}
 r.Body=http.MaxBytesReader(w,r.Body,4096)
 var req struct{Handle string `json:"handle"`}
 if !decodeJSON(w,r,&req){return}
 err:=s.socialActions.ClaimExisting(r.Context(),did,req.Handle)
 if err!=nil {
  if errors.Is(err,social.ErrOwnership){writeJSON(w,403,map[string]string{"error":"profile_ownership_mismatch"});return}
  if errors.Is(err,social.ErrNotClaimable){writeJSON(w,409,map[string]string{"error":"handle_not_verified_for_did"});return}
  writeJSON(w,500,map[string]string{"error":"internal_error"});return
 }
 w.Header().Set("Cache-Control","no-store")
 writeJSON(w,200,map[string]any{"claimed":true})
}
func(s *Server) updateOwnFollow(w http.ResponseWriter,r *http.Request){
 did,ok:=s.authenticatedSocialMutation(w,r);if !ok{return}
 err:=s.socialActions.SetFollow(r.Context(),did,r.PathValue("handle"),r.Method==http.MethodPut)
 if err!=nil{
  switch{
  case errors.Is(err,social.ErrOwnership):writeJSON(w,403,map[string]string{"error":"social_account_required"})
  case errors.Is(err,social.ErrNotFound):writeJSON(w,404,map[string]string{"error":"profile_not_found"})
  case errors.Is(err,social.ErrInvalidFollow):writeJSON(w,400,map[string]string{"error":"invalid_follow"})
  default:writeJSON(w,500,map[string]string{"error":"internal_error"})
  }
  return
 }
 w.Header().Set("Cache-Control","no-store")
 w.WriteHeader(http.StatusNoContent)
}
