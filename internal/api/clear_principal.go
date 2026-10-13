package api

import(
 "errors"
 "net/http"
 "github.com/chrisbirster/cleartify/internal/clearsession"
)
// ClearSessionPrincipal derives the principal solely from an authenticated,
// non-revoked server-side Clear cookie; request user IDs are ignored.
func ClearSessionPrincipal(sessions *clearsession.Service) ClearPrincipal {
 return func(r *http.Request)(string,error){
  if sessions==nil{return "",errors.New("Clear sessions unavailable")}
  cookie,err:=r.Cookie("clear_session");if err!=nil{return "",clearsession.ErrUnauthenticated}
  return sessions.Authenticate(r.Context(),cookie.Value)
 }
}
