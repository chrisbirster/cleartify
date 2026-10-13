package api

import (
 "net/http"
 "github.com/chrisbirster/cleartify/internal/clearsession"
)

func (s *Server) clearPrincipal(w http.ResponseWriter,r *http.Request)(string,bool){
 c,err:=r.Cookie("clear_session")
 if err!=nil {writeJSON(w,http.StatusUnauthorized,map[string]string{"error":"unauthenticated"});return "",false}
 id,err:=s.clearSessions.Authenticate(r.Context(),c.Value)
 if err!=nil {
  if err!=clearsession.ErrUnauthenticated {writeJSON(w,http.StatusServiceUnavailable,map[string]string{"error":"session_unavailable"});return "",false}
  writeJSON(w,http.StatusUnauthorized,map[string]string{"error":"unauthenticated"});return "",false
 }
 return id,true
}
func (s *Server) meClearSession(w http.ResponseWriter,r *http.Request){
 id,ok:=s.clearPrincipal(w,r);if !ok{return}
 w.Header().Set("Cache-Control","no-store")
 writeJSON(w,http.StatusOK,map[string]any{"authenticated":true,"userId":id})
}
func (s *Server) logoutClearSession(w http.ResponseWriter,r *http.Request){
 c,err:=r.Cookie("clear_session")
 if err==nil {
  if err=s.clearSessions.Revoke(r.Context(),c.Value);err!=nil {writeJSON(w,http.StatusServiceUnavailable,map[string]string{"error":"session_unavailable"});return}
 }
 http.SetCookie(w,&http.Cookie{Name:"clear_session",Path:"/api/me/",HttpOnly:true,Secure:r.TLS!=nil,SameSite:http.SameSiteStrictMode,MaxAge:-1})
 writeJSON(w,http.StatusOK,map[string]bool{"loggedOut":true})
}
