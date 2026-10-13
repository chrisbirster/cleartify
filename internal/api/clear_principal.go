package api

import (
 "net/http"
 "github.com/chrisbirster/cleartify/internal/clearsession"
)

// ClearSessionPrincipal binds a private route to a live server-side Clear session.
// A nil service leaves route authorization unavailable; no token may come
// from a query parameter or body.
func ClearSessionPrincipal(sessions *clearsession.Service) ClearPrincipal {
 if sessions==nil{return nil}
 return func(r *http.Request)(string,error){
  cookie,err:=r.Cookie("clear_session")
  if err!=nil{return "",clearsession.ErrUnauthenticated}
  return sessions.Authenticate(r.Context(),cookie.Value)
 }
}
