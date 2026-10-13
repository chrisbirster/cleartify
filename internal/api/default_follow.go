package api

import(
 "errors"
 "net/http"
 "github.com/chrisbirster/cleartify/internal/social"
)

// The caller's follower DID always comes from the locally verified OAuth
// browser session. The request body/path cannot impersonate other users.
func(s *Server) removeDefaultFollow(w http.ResponseWriter,r *http.Request){
 if s.oauth==nil||s.defaultFollow==nil {writeJSON(w,http.StatusServiceUnavailable,map[string]string{"error":"not_configured"});return}
 if r.Header.Get("Origin")!=s.appOrigin {writeJSON(w,http.StatusForbidden,map[string]string{"error":"invalid_origin"});return}
 did,ok:=s.oauth.LocalAuthenticatedDID(r)
 if !ok {writeJSON(w,http.StatusUnauthorized,map[string]string{"error":"atproto_login_required"});return}
 if err:=s.defaultFollow.RemoveChrisFollow(r.Context(),did);err!=nil {
  if errors.Is(err,social.ErrDefaultTargetUnavailable) {writeJSON(w,http.StatusNotFound,map[string]string{"error":"chris_profile_unavailable"});return}
  writeJSON(w,http.StatusInternalServerError,map[string]string{"error":"internal_error"});return
 }
 w.Header().Set("Cache-Control","no-store")
 w.WriteHeader(http.StatusNoContent)
}
