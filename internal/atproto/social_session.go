package atproto

import(
 "net/http"
 "strings"
 "time"
)

// SocialSessionDID derives identity strictly from the opaque browser cookie,
// a non-expired Turso browser-session record and its associated OAuth session.
// A request body, handle, email or phone number cannot substitute for OAuth.
func(h *OAuthHandler) SocialSessionDID(r *http.Request)(string,bool){
 c,err:=r.Cookie(sessionCookieName);if err!=nil||len(c.Value)<32{return "",false}
 var did,expiryRaw string
 err=h.store.DB.DB.QueryRowContext(r.Context(),`SELECT b.account_did,b.expires_at FROM atproto_browser_sessions b
 JOIN atproto_oauth_sessions s ON s.account_did=b.account_did AND s.session_id=b.oauth_session_id
 WHERE b.token_digest=? LIMIT 1`,digest(c.Value)).Scan(&did,&expiryRaw)
 if err!=nil{return "",false}
 expiry,err:=time.Parse(time.RFC3339Nano,expiryRaw)
 if err!=nil||!time.Now().UTC().Before(expiry)||!strings.HasPrefix(did,"did:"){return "",false}
 return did,true
}
