package api

import (
 "errors"
 "net/http"
 "strings"
 "github.com/chrisbirster/cleartify/internal/clearsession"
)
func(s *Server) startReturningLogin(w http.ResponseWriter,r *http.Request){
 w.Header().Set("Cache-Control","no-store")
 if s.returningLogin==nil {writeJSON(w,503,map[string]string{"error":"not_configured"});return}
 var req struct {Phone string `json:"phone"`}
 if !decodeJSON(w,r,&req){return}
 if len(req.Phone)<8||len(req.Phone)>16||!strings.HasPrefix(req.Phone,"+"){writeJSON(w,400,map[string]string{"error":"invalid_phone"});return}
 if err:=s.returningLogin.Start(r.Context(),req.Phone);err!=nil {writeJSON(w,503,map[string]string{"error":"login_unavailable"});return}
 writeJSON(w,202,map[string]bool{"accepted":true})
}
func(s *Server) verifyReturningLogin(w http.ResponseWriter,r *http.Request){
 w.Header().Set("Cache-Control","no-store")
 if s.returningLogin==nil {writeJSON(w,503,map[string]string{"error":"not_configured"});return}
 var req struct {Phone string `json:"phone"`;Code string `json:"code"`}
 if !decodeJSON(w,r,&req){return}
 token,err:=s.returningLogin.Verify(r.Context(),req.Phone,req.Code)
 if err!=nil {
  if errors.Is(err,clearsession.ErrChallenge){writeJSON(w,401,map[string]string{"error":"invalid_challenge"});return}
  writeJSON(w,503,map[string]string{"error":"login_unavailable"});return
 }
 http.SetCookie(w,&http.Cookie{Name:"clear_session",Value:token,Path:"/api/me/",HttpOnly:true,Secure:strings.HasPrefix(s.appOrigin,"https://"),SameSite:http.SameSiteStrictMode,MaxAge:int(clearsession.IdleLimit.Seconds())})
 writeJSON(w,200,map[string]bool{"authenticated":true})
}
