package api

import (
 "errors"
 "net/http"
 "github.com/chrisbirster/cleartify/internal/social"
)
func(s *Server) socialModerationPrincipal(w http.ResponseWriter,r *http.Request)(string,bool){
 if s.socialAuth==nil||s.moderation==nil {writeJSON(w,503,map[string]string{"error":"social_unavailable"});return "",false}
 if s.appOrigin==""||r.Header.Get("Origin")!=s.appOrigin {writeJSON(w,403,map[string]string{"error":"invalid_origin"});return "",false}
 did,ok:=s.socialAuth.SocialSessionDID(r)
 if !ok {writeJSON(w,401,map[string]string{"error":"atproto_login_required"});return "",false}
 return did,true
}
func(s *Server) setBlock(w http.ResponseWriter,r *http.Request){
 did,ok:=s.socialModerationPrincipal(w,r);if !ok{return}
 err:=s.moderation.Block(r.Context(),did,r.PathValue("did"),r.Method==http.MethodPut)
 if err!=nil{
  switch{
  case errors.Is(err,social.ErrInvalidModeration):writeJSON(w,400,map[string]string{"error":"invalid_block"})
  case errors.Is(err,social.ErrUnlistedSubject):writeJSON(w,404,map[string]string{"error":"profile_not_found"})
  default:writeJSON(w,500,map[string]string{"error":"internal_error"})
  }
  return
 }
 w.Header().Set("Cache-Control","no-store");w.WriteHeader(http.StatusNoContent)
}
func(s *Server) createSocialReport(w http.ResponseWriter,r *http.Request){
 did,ok:=s.socialModerationPrincipal(w,r);if !ok{return}
 var req struct{SubjectDID string `json:"subjectDid"`;Reason string `json:"reason"`}
 if !decodeJSON(w,r,&req){return}
 id,err:=s.moderation.Report(r.Context(),did,req.SubjectDID,req.Reason)
 if err!=nil{
  switch{
  case errors.Is(err,social.ErrInvalidModeration):writeJSON(w,400,map[string]string{"error":"invalid_report"})
  case errors.Is(err,social.ErrUnlistedSubject):writeJSON(w,404,map[string]string{"error":"profile_not_found"})
  default:writeJSON(w,500,map[string]string{"error":"internal_error"})
  }
  return
 }
 w.Header().Set("Cache-Control","no-store")
 writeJSON(w,http.StatusAccepted,map[string]any{"reportId":id,"status":"pending"})
}
