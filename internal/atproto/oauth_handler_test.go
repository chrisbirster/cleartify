package atproto

import (
 "bytes"
 "context"
 "encoding/base64"
 "errors"
 "net/http"
 "net/http/httptest"
 "testing"

 "github.com/bluesky-social/indigo/atproto/auth/oauth"
 "github.com/chrisbirster/cleartify/internal/database"
)

func oauthTestStore(t *testing.T)*OAuthStore{
 t.Helper()
 db,err:=database.Open(context.Background(),database.Config{Path:":memory:"})
 if err!=nil{t.Fatal(err)}
 t.Cleanup(func(){_ = db.Close()})
 if err:=db.Migrate(context.Background());err!=nil{t.Fatal(err)}
 box,err:=NewSecretBox(base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{3},32)))
 if err!=nil{t.Fatal(err)}
 return &OAuthStore{DB:db,Box:box}
}
func TestOAuthRequestBrowserBoundConsumed(t *testing.T){
 store:=oauthTestStore(t)
 first:=bindNonce(context.Background(),"browser-nonce-aaaaaaaaaaaaaaaaaaaaaaaaaaaa")
 second:=bindNonce(context.Background(),"browser-nonce-bbbbbbbbbbbbbbbbbbbbbbbbbbbb")
 if err:=store.SaveAuthRequestInfo(first,oauth.AuthRequestData{State:"test-state",PKCEVerifier:"private-verifier"});err!=nil{t.Fatal(err)}
 var raw []byte
 if err:=store.DB.DB.QueryRow("SELECT payload_ciphertext FROM atproto_oauth_requests WHERE state=?", "test-state").Scan(&raw);err!=nil{t.Fatal(err)}
 if bytes.Contains(raw,[]byte("private-verifier")){t.Fatal("PKCE verifier stored in cleartext")}
 if _,err:=store.GetAuthRequestInfo(second,"test-state");!errors.Is(err,ErrInvalidState){t.Fatalf("wrong browser state: %v",err)}
 info,err:=store.GetAuthRequestInfo(first,"test-state")
 if err!=nil||info.PKCEVerifier!="private-verifier"{t.Fatalf("valid callback rejected: %+v %v",info,err)}
 if _,err:=store.GetAuthRequestInfo(first,"test-state");!errors.Is(err,ErrInvalidState){t.Fatalf("callback replay accepted: %v",err)}
}
func TestOAuthLoginDisabledAndOriginCheck(t *testing.T){
 store:=oauthTestStore(t)
 h,err:=NewOAuthHandler(store.DB,"https://api.cleartify.com","https://cleartify.com",base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{6},32)))
 if err!=nil{t.Fatal(err)}
 mux:=http.NewServeMux()
 // No handler registration simulates the default-off API configuration.
 req:=httptest.NewRequest(http.MethodGet,"/oauth/login",nil)
 w:=httptest.NewRecorder()
 mux.ServeHTTP(w,req)
 if w.Code!=http.StatusNotFound{t.Fatalf("disabled OAuth exposed: %d",w.Code)}
 h.Register(mux)
 req=httptest.NewRequest(http.MethodPost,"/oauth/login",nil)
 req.Header.Set("Origin","https://attacker.example")
 w=httptest.NewRecorder()
 mux.ServeHTTP(w,req)
 if w.Code!=http.StatusForbidden{t.Fatalf("cross-origin login accepted: %d",w.Code)}
 req=httptest.NewRequest(http.MethodGet,"/oauth/callback?code=fake&state=fake",nil)
 w=httptest.NewRecorder()
 mux.ServeHTTP(w,req)
 if w.Code!=http.StatusForbidden{t.Fatalf("callback without browser nonce accepted: %d",w.Code)}
}
func TestOAuthMetadataExactRedirect(t *testing.T){
 store:=oauthTestStore(t)
 h,err:=NewOAuthHandler(store.DB,"https://api.cleartify.com","https://cleartify.com",base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{6},32)))
 if err!=nil{t.Fatal(err)}
 mux:=http.NewServeMux();h.Register(mux)
 req:=httptest.NewRequest(http.MethodGet,"/oauth/client-metadata.json",nil)
 w:=httptest.NewRecorder();mux.ServeHTTP(w,req)
 if w.Code!=http.StatusOK{t.Fatalf("metadata: %d %s",w.Code,w.Body.String())}
 if !bytes.Contains(w.Body.Bytes(),[]byte("https://api.cleartify.com/oauth/callback")){t.Fatal("metadata missing exact callback")}
}
