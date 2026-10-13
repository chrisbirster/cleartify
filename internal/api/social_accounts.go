package api

import(
 "errors"
 "net/http"
 "github.com/chrisbirster/cleartify/internal/social"
)

// Social account creation never consumes a user-controlled DID, phone or
// email. It requires a valid, decryptable Indigo browser OAuth session.
func(s *Server) registerSocialAccount(w http.ResponseWriter,r *http.Request){
 if s.socialAuth==nil||s.socialAccounts==nil {
  writeJSON(w,http.StatusServiceUnavailable,map[string]string{"error":"social_login_not_configured"});return
 }
 if r.Header.Get("Origin")!=s.appOrigin {
  writeJSON(w,http.StatusForbidden,map[string]string{"error":"invalid_origin"});return
 }
 did,ok:=s.socialAuth.SocialSessionDID(r)
 if !ok{
  writeJSON(w,http.StatusUnauthorized,map[string]string{"error":"atproto_login_required"});return
 }
 acc,err:=s.socialAccounts.EnsureAuthenticated(r.Context(),did)
 if err!=nil{
  if errors.Is(err,social.ErrSocialIdentityRequired){writeJSON(w,http.StatusUnauthorized,map[string]string{"error":"atproto_login_required"});return}
  writeJSON(w,http.StatusInternalServerError,map[string]string{"error":"internal_error"});return
 }
 w.Header().Set("Cache-Control","no-store")
 writeJSON(w,http.StatusCreated,acc)
}
