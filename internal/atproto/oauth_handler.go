package atproto

import (
 "context"
 "crypto/rand"
 "encoding/base64"
 "encoding/json"
 "errors"
 "fmt"
 "html"
 "net/http"
 "net/url"
 "strings"
 "time"

 "github.com/bluesky-social/indigo/atproto/auth/oauth"
 "github.com/bluesky-social/indigo/atproto/syntax"
 "github.com/chrisbirster/cleartify/internal/database"
)

const flowCookieName="cleartify_atproto_flow"
const sessionCookieName="cleartify_atproto_session"
const sessionTTL=24*time.Hour

// OAuthHandler delegates PAR, PKCE, DPoP, issuer and subject checks to Indigo.
// It is mounted ONLY when ATPROTO_OAUTH_ENABLED=true and config is valid.
type OAuthHandler struct {
 client *oauth.ClientApp
 store *OAuthStore
 publicOrigin string
 appOrigin string
 secure bool
 onVerifiedLogin func(context.Context,string)error
}

func NewOAuthHandler(db *database.Store, publicOrigin, appOrigin, secretKey string) (*OAuthHandler,error){
 urls,err:=ConfigureOAuthURLs(publicOrigin);if err!=nil{return nil,err}
 box,err:=NewSecretBox(secretKey);if err!=nil{return nil,err}
 if appOrigin=="" {return nil,fmt.Errorf("APP_ORIGIN required")}
 target,err:=url.Parse(appOrigin);if err!=nil||target.Host==""||target.Scheme!="https"||target.Path!=""||target.RawQuery!=""||target.Fragment!="" {return nil,fmt.Errorf("invalid fixed APP_ORIGIN")}
 config:=oauth.NewPublicConfig(urls.ClientID,urls.Callback,[]string{"atproto"})
 store:=&OAuthStore{DB:db,Box:box}
 return &OAuthHandler{client:oauth.NewClientApp(&config,store),store:store,publicOrigin:strings.TrimSuffix(publicOrigin,"/"),appOrigin:appOrigin,secure:strings.HasPrefix(publicOrigin,"https://")},nil
}
// SetOnVerifiedLogin registers a local, post-OAuth hook. It does not call a PDS.
func(h *OAuthHandler) SetOnVerifiedLogin(fn func(context.Context,string)error){h.onVerifiedLogin=fn}

// LocalAuthenticatedDID checks only locally stored OAuth browser and encrypted
// OAuth session rows; it never accepts a DID from request parameters.
func(h *OAuthHandler) LocalAuthenticatedDID(r *http.Request)(string,bool){
 c,err:=r.Cookie(sessionCookieName)
 if err!=nil||len(c.Value)<32{return "",false}
 var did,expires string
 err=h.store.DB.DB.QueryRowContext(r.Context(),`SELECT b.account_did,b.expires_at FROM atproto_browser_sessions b
 INNER JOIN atproto_oauth_sessions s ON s.account_did=b.account_did AND s.session_id=b.oauth_session_id
 WHERE b.token_digest=? LIMIT 1`,digest(c.Value)).Scan(&did,&expires)
 if err!=nil{return "",false}
 expiry,err:=time.Parse(time.RFC3339Nano,expires)
 if err!=nil||!time.Now().UTC().Before(expiry)||!strings.HasPrefix(did,"did:"){return "",false}
 return did,true
}

func (h *OAuthHandler) Register(mux *http.ServeMux){
 mux.HandleFunc("GET /oauth/client-metadata.json",h.Metadata)
 mux.HandleFunc("GET /oauth/login",h.Login)
 mux.HandleFunc("POST /oauth/login",h.Login)
 mux.HandleFunc("GET /oauth/callback",h.Callback)
 mux.HandleFunc("POST /oauth/logout",h.Logout)
}
func(h *OAuthHandler) Metadata(w http.ResponseWriter,r *http.Request){
 meta:=h.client.Config.ClientMetadata()
 name:="Cleartify"
 meta.ClientName=&name
 meta.ClientURI=&h.publicOrigin
 if err:=meta.Validate(h.client.Config.ClientID);err!=nil {http.Error(w,"client metadata unavailable",500);return}
 w.Header().Set("Content-Type","application/json")
 _=json.NewEncoder(w).Encode(meta)
}
func(h *OAuthHandler) checkOrigin(r *http.Request) bool{return r.Header.Get("Origin")==h.publicOrigin}
func(h *OAuthHandler) cookie(w http.ResponseWriter,name,value string,maxAge int){
 http.SetCookie(w,&http.Cookie{Name:name,Value:value,Path:"/",HttpOnly:true,Secure:h.secure,SameSite:http.SameSiteLaxMode,MaxAge:maxAge})
}
func randomSecret()(string,error){
 b:=make([]byte,32)
 if _,err:=rand.Read(b);err!=nil{return "",err}
 return base64.RawURLEncoding.EncodeToString(b),nil
}
func (h *OAuthHandler) Login(w http.ResponseWriter,r *http.Request){
 if r.Method==http.MethodGet {
  w.Header().Set("Content-Type","text/html; charset=utf-8")
  _,_=w.Write([]byte(`<!doctype html><html><body><h1>Connect AT Protocol</h1><form method="post" action="/oauth/login"><label>Handle or DID <input name="identifier" required maxlength="253"></label><button type="submit">Sign in</button></form></body></html>`))
  return
 }
 if !h.checkOrigin(r){http.Error(w,"invalid origin",http.StatusForbidden);return}
 r.Body=http.MaxBytesReader(w,r.Body,4096)
 if err:=r.ParseForm();err!=nil{http.Error(w,"invalid form",400);return}
 identifier:=strings.TrimSpace(r.PostFormValue("identifier"))
 if identifier==""||len(identifier)>253||strings.ContainsAny(identifier," \n\r\t") {http.Error(w,"invalid identifier",400);return}
 nonce,err:=randomSecret();if err!=nil{http.Error(w,"unavailable",500);return}
 dest,err:=h.client.StartAuthFlow(bindNonce(r.Context(),nonce),identifier)
 if err!=nil{http.Error(w,"authorization could not start",400);return}
 parsed,err:=url.Parse(dest)
 if err!=nil||parsed.Scheme!="https"||parsed.Host==""||parsed.User!=nil{http.Error(w,"authorization redirect invalid",502);return}
 h.cookie(w,flowCookieName,nonce,600)
 w.Header().Set("Cache-Control","no-store")
 http.Redirect(w,r,dest,http.StatusSeeOther)
}
func (h *OAuthHandler) Callback(w http.ResponseWriter,r *http.Request){
 flow,err:=r.Cookie(flowCookieName)
 if err!=nil||len(flow.Value)<32 {http.Error(w,"authorization session missing",403);return}
 // State verification, issuer, PKCE verifier, DPoP-bound code exchange and
 // account DID validation are implemented by Indigo. The store atomically
 // consumes state only if this exact browser's nonce matches.
 session,err:=h.client.ProcessCallback(bindNonce(r.Context(),flow.Value),r.URL.Query())
 h.cookie(w,flowCookieName,"",-1)
 if err!=nil{http.Error(w,"authorization invalid or expired",400);return}
 token,err:=randomSecret();if err!=nil{http.Error(w,"unavailable",500);return}
 now:=time.Now().UTC()
 _,err=h.store.DB.DB.ExecContext(r.Context(),`INSERT INTO atproto_browser_sessions(token_digest,account_did,oauth_session_id,expires_at,created_at) VALUES(?,?,?,?,?)`,
  digest(token),session.AccountDID.String(),session.SessionID,now.Add(sessionTTL).Format(time.RFC3339Nano),now.Format(time.RFC3339Nano))
 if err!=nil{http.Error(w,"unable to create session",500);return}
 if err:=h.store.DB.Push(r.Context());err!=nil{http.Error(w,"unable to persist session",500);return}
 if h.onVerifiedLogin!=nil{
  if err:=h.onVerifiedLogin(r.Context(),session.AccountDID.String());err!=nil{
   http.Error(w,"social onboarding unavailable",http.StatusServiceUnavailable);return
  }
 }
 h.cookie(w,sessionCookieName,token,int(sessionTTL.Seconds()))
 w.Header().Set("Cache-Control","no-store")
 http.Redirect(w,r,h.appOrigin+"/settings/profile?atproto=connected",http.StatusSeeOther)
}
func(h *OAuthHandler) Logout(w http.ResponseWriter,r *http.Request){
 if !h.checkOrigin(r){http.Error(w,"invalid origin",http.StatusForbidden);return}
 defer h.cookie(w,sessionCookieName,"",-1)
 c,err:=r.Cookie(sessionCookieName);if err!=nil{w.WriteHeader(http.StatusNoContent);return}
 var did,id,expires string
 err=h.store.DB.DB.QueryRowContext(r.Context(),`SELECT account_did,oauth_session_id,expires_at FROM atproto_browser_sessions WHERE token_digest=?`,digest(c.Value)).Scan(&did,&id,&expires)
 if err!=nil{w.WriteHeader(http.StatusNoContent);return}
 _,_=h.store.DB.DB.ExecContext(r.Context(),`DELETE FROM atproto_browser_sessions WHERE token_digest=?`,digest(c.Value))
 _=h.store.DB.Push(r.Context())
 account,err:=syntax.ParseDID(did)
 if err==nil{
  // Indigo attempts remote revocation where the provider supports it,
  // and removes the encrypted local OAuth session.
  _=h.client.Logout(context.WithoutCancel(r.Context()),account,id)
 }
 w.WriteHeader(http.StatusNoContent)
}
var _=errors.Is
var _=html.EscapeString
