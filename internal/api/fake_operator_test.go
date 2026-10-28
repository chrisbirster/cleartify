package api

import (
 "context"
 "bytes"
 "crypto/sha256"
 "encoding/base64"
 "encoding/hex"
 "encoding/json"
 "github.com/bluesky-social/indigo/atproto/auth/oauth"
 "github.com/bluesky-social/indigo/atproto/syntax"
 "net/http"
 "time"
 "github.com/chrisbirster/cleartify/internal/atproto"
 "fmt"
 "net/http/httptest"
 "github.com/chrisbirster/cleartify/internal/database"
 "strings"
 "testing"
 "github.com/chrisbirster/cleartify/internal/marketplace"
)
func TestFakeOperatorRoutesFailClosed(t *testing.T){
 paths:=[]string{"/api/admin/fake-payments/e/resolve","/api/admin/fake-payments/e/refund"}
 for _,path:=range paths {
  for _,opts:=range []Options{{},{FakeCheckoutEnabled:true,FakeCheckout:&marketplace.Service{Enabled:true}}} {
   w:=httptest.NewRecorder();New(opts).ServeHTTP(w,httptest.NewRequest("POST",path,strings.NewReader("{}")))
   if w.Code!=404{t.Fatalf("ungated route %s returned %d",path,w.Code)}
  }
  opts:=Options{FakeCheckoutEnabled:true,FakeCheckout:&marketplace.Service{Enabled:true},FakeAccounting:&marketplace.FakeAccounting{Enabled:true},AdminToken:"test-admin",AppOrigin:"http://localhost:3000"}
  h:=New(opts)
  w:=httptest.NewRecorder();h.ServeHTTP(w,httptest.NewRequest("POST",path,strings.NewReader("{}")))
  if w.Code!=401{t.Fatalf("no token %s: %d",path,w.Code)}
  r:=httptest.NewRequest("POST",path,strings.NewReader("{}"));r.Header.Set("Authorization","Bearer test-admin")
  w=httptest.NewRecorder();h.ServeHTTP(w,r)
  if w.Code!=403{t.Fatalf("no origin %s: %d",path,w.Code)}
 }
}

func TestFakeBuyerDisputeRouteDisabled(t *testing.T){
 for _,opts:=range []Options{{},{FakeCheckoutEnabled:true,FakeCheckout:&marketplace.Service{Enabled:true},FakeAccounting:&marketplace.FakeAccounting{Enabled:true}}} {
  w:=httptest.NewRecorder();New(opts).ServeHTTP(w,httptest.NewRequest("POST","/api/me/fake-entitlements/e/dispute",strings.NewReader("{}")))
  if w.Code!=404{t.Fatalf("OAuth required; got %d",w.Code)}
 }
}

func TestFakeOperatorHTTPWithLocalDB(t *testing.T) {
 ctx:=context.Background()
 store,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)};defer store.Close()
 if err=store.Migrate(ctx);err!=nil{t.Fatal(err)}
 for _,q:=range []string{
  `INSERT INTO social_profiles(did,handle,display_name,kind,public_visible,claimed,updated_at) VALUES('did:plc:seller','seller-audit.test','Seller','person',1,1,'now')`,
  `INSERT INTO published_protocols(id,creator_did,slug,version,title,summary,license,spec_json,status,created_at) VALUES('p','did:plc:seller','slug',1,'Title','','MIT','{}','published','now')`,
  `INSERT INTO marketplace_listings(id,protocol_id,creator_did,title,pricing_kind,price_minor,currency,status,created_at) VALUES('l','p','did:plc:seller','Title','one_time',101,'USD','listed','now')`,
 } {if _,err=store.DB.ExecContext(ctx,q);err!=nil{t.Fatal(err)}}
 svc:=&marketplace.Service{DB:store.DB,Provider:marketplace.FakeProvider{},Enabled:true}
 acct:=&marketplace.FakeAccounting{DB:store.DB,Enabled:true,FeeBPS:1000}
 opts:=Options{FakeCheckoutEnabled:true,FakeCheckout:svc,FakeAccounting:acct,AdminToken:"local-test-admin",AppOrigin:"https://example.test"}
 handler:=New(opts)
 call:=func(path,body,token,origin string)int{
  t.Helper()
  r:=httptest.NewRequest("POST",path,strings.NewReader(body))
  if token!="" {r.Header.Set("Authorization","Bearer "+token)}
  if origin!="" {r.Header.Set("Origin",origin)}
  w:=httptest.NewRecorder();handler.ServeHTTP(w,r);return w.Code
 }
 for _,route:=range []string{"/resolve","/refund"} {
  path:="/api/admin/fake-payments/missing"+route
  if got:=call(path,`{"refund":true}`,"bad","https://example.test");got!=401{t.Fatalf("wrong token %s: %d",route,got)}
  if got:=call(path,`{"refund":true}`,"local-test-admin","https://evil.test");got!=403{t.Fatalf("wrong origin %s: %d",route,got)}
 }
 buy:=func(buyer string)string{
  t.Helper()
  id,err:=svc.Purchase(ctx,buyer,"l");if err!=nil{t.Fatal(err)}
  return id
 }
 id:=buy("did:plc:buyer")
 other:=buy("did:plc:other")
 if err=acct.OpenDispute(ctx,"did:plc:other",id,"other");err==nil{t.Fatal("cross-buyer dispute accepted")}
 if err=acct.OpenDispute(ctx,"did:plc:buyer",id,"other");err!=nil{t.Fatal(err)}
 if err=acct.OpenDispute(ctx,"did:plc:other",other,"not_received");err!=nil{t.Fatal(err)}
 // Real HTTP routing and authorization: refund and no-refund adjudication.
 refundPath:=fmt.Sprintf("/api/admin/fake-payments/%s/refund",id)
 if got:=call(refundPath,"{}","local-test-admin","https://example.test");got!=200{t.Fatalf("refund %d",got)}
 if got:=call(refundPath,"{}","local-test-admin","https://example.test");got!=409{t.Fatalf("duplicate refund %d",got)}
 resolvePath:=fmt.Sprintf("/api/admin/fake-payments/%s/resolve",other)
 if got:=call(resolvePath,`{"refund":false}`,"local-test-admin","https://example.test");got!=200{t.Fatalf("resolve %d",got)}
 if got:=call(resolvePath,`{"refund":false}`,"local-test-admin","https://example.test");got!=409{t.Fatalf("duplicate resolution %d",got)}
 var refunds int
 if err=store.DB.QueryRowContext(ctx,`SELECT count(*) FROM marketplace_payment_audit WHERE entitlement_id=? AND event='fake_refund'`,id).Scan(&refunds);err!=nil||refunds!=1{t.Fatalf("refund audit %d %v",refunds,err)}
}

func TestFakeBuyerDisputeUsesOAuthBrowserCookieNotClaimedDID(t *testing.T) {
 ctx:=context.Background()
 store,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)}
 defer store.Close()
 if err=store.Migrate(ctx);err!=nil{t.Fatal(err)}
 oauthHandler,err:=atproto.NewOAuthHandler(store,"https://api.example.test","https://app.example.test",base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{6},32)))
 if err!=nil{t.Fatal(err)}
 for _,q:=range []string{
  `INSERT INTO social_profiles(did,handle,display_name,kind,public_visible,claimed,updated_at) VALUES('did:plc:seller','seller-audit.test','Seller','person',1,1,'now')`,
  `INSERT INTO published_protocols(id,creator_did,slug,version,title,summary,license,spec_json,status,created_at) VALUES('p','did:plc:seller','slug',1,'Title','','MIT','{}','published','now')`,
  `INSERT INTO marketplace_listings(id,protocol_id,creator_did,title,pricing_kind,price_minor,currency,status,created_at) VALUES('l','p','did:plc:seller','Title','one_time',101,'USD','listed','now')`,
 } {if _,err=store.DB.ExecContext(ctx,q);err!=nil{t.Fatal(err)}}
 svc:=&marketplace.Service{DB:store.DB,Provider:marketplace.FakeProvider{},Enabled:true}
 accounting:=&marketplace.FakeAccounting{DB:store.DB,Enabled:true,FeeBPS:1000}
 aID,err:=svc.Purchase(ctx,"did:web:buyer-a.example","l");if err!=nil{t.Fatal(err)}
 bID,err:=svc.Purchase(ctx,"did:web:buyer-b.example","l");if err!=nil{t.Fatal(err)}
 // Mirrors the local OAuth browser/session fixtures: a digest-keyed browser
 // cookie joined to an existing OAuth session. No PDS or callback is invoked.
 seed:=func(did,token string,expires time.Time) {
  t.Helper()
  sessionID:="session-"+did
  parsed,parseErr:=syntax.ParseDID(did);if parseErr!=nil{t.Fatal(parseErr)}
  raw,marshalErr:=json.Marshal(oauth.ClientSessionData{AccountDID:parsed,SessionID:sessionID});if marshalErr!=nil{t.Fatal(marshalErr)}
  box,boxErr:=atproto.NewSecretBox(base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{6},32)));if boxErr!=nil{t.Fatal(boxErr)}
  ciphertext,sealErr:=box.Seal(raw,"session:"+did+":"+sessionID);if sealErr!=nil{t.Fatal(sealErr)}
  _,err:=store.DB.ExecContext(ctx,`INSERT INTO atproto_oauth_sessions(account_did,session_id,payload_ciphertext,created_at,updated_at) VALUES(?,?,?,?,?)`,did,sessionID,ciphertext,"now","now")
  if err!=nil{t.Fatal(err)}
  hash:=sha256.Sum256([]byte(token))
  _,err=store.DB.ExecContext(ctx,`INSERT INTO atproto_browser_sessions(token_digest,account_did,oauth_session_id,expires_at,created_at) VALUES(?,?,?,?,?)`,hex.EncodeToString(hash[:]),did,sessionID,expires.UTC().Format(time.RFC3339Nano),"now")
  if err!=nil{t.Fatal(err)}
 }
 tokenA:=strings.Repeat("a",40)
 tokenB:=strings.Repeat("b",40)
 tokenExpired:=strings.Repeat("e",40)
 tokenRevoked:=strings.Repeat("r",40)
 seed("did:web:buyer-a.example",tokenA,time.Now().Add(time.Hour))
 seed("did:web:buyer-b.example",tokenB,time.Now().Add(time.Hour))
 seed("did:web:expired.example",tokenExpired,time.Now().Add(-time.Hour))
 seed("did:web:revoked.example",tokenRevoked,time.Now().Add(time.Hour))
 revokedDigest:=sha256.Sum256([]byte(tokenRevoked))
 if _,err=store.DB.ExecContext(ctx,`DELETE FROM atproto_browser_sessions WHERE token_digest=?`,hex.EncodeToString(revokedDigest[:]));err!=nil{t.Fatal(err)}
 handler:=New(Options{FakeCheckoutEnabled:true,FakeCheckout:svc,FakeAccounting:accounting,OAuth:oauthHandler,AppOrigin:"https://app.example.test"})
 request:=func(entitlement,token,origin,body,claimed string)int {
  t.Helper()
  uri:="/api/me/fake-entitlements/"+entitlement+"/dispute?buyerDid="+claimed
  r:=httptest.NewRequest(http.MethodPost,uri,strings.NewReader(body))
  if token!=""{r.AddCookie(&http.Cookie{Name:"cleartify_atproto_session",Value:token})}
  if origin!=""{r.Header.Set("Origin",origin)}
  r.Header.Set("X-Buyer-DID",claimed)
  r.Header.Set("Content-Type","application/json")
  w:=httptest.NewRecorder();handler.ServeHTTP(w,r)
  return w.Code
 }
 goodBody:=`{"reason":"not_received"}`
 // OAuth cookie A authorizes only A's purchase.
 // B must not be able to open A's still-unclaimed dispute.
 if got:=request(aID,tokenB,"https://app.example.test",goodBody,"did:web:buyer-a.example");got!=409{t.Fatalf("cross-buyer DID spoof: %d",got)}
 if got:=request(aID,tokenA,"https://app.example.test",goodBody,"did:web:buyer-b.example");got!=201{t.Fatalf("buyer A request: %d",got)}
 // JSON-supplied buyer identity is forbidden by DisallowUnknownFields.
 if got:=request(bID,tokenB,"https://app.example.test",`{"reason":"other","buyerDid":"did:web:buyer-a.example"}`,"did:web:buyer-a.example");got!=400{t.Fatalf("body DID spoof: %d",got)}
 if got:=request(bID,tokenB,"https://evil.example",goodBody,"did:web:buyer-a.example");got!=403{t.Fatalf("wrong origin: %d",got)}
 for name,token:=range map[string]string{"missing":"","expired":tokenExpired,"revoked":tokenRevoked} {
  if got:=request(bID,token,"https://app.example.test",goodBody,"did:web:buyer-a.example");got!=401{t.Fatalf("%s session: %d",name,got)}
 }
 if got:=request(bID,tokenB,"https://app.example.test",goodBody,"did:web:buyer-a.example");got!=201{t.Fatalf("buyer B own dispute: %d",got)}
 if got:=request(bID,tokenB,"https://app.example.test",goodBody,"did:web:buyer-a.example");got!=409{t.Fatalf("duplicate dispute: %d",got)}
 var count int
 if err=store.DB.QueryRowContext(ctx,`SELECT count(*) FROM fake_payment_disputes WHERE entitlement_id IN (?,?)`,aID,bID).Scan(&count);err!=nil||count!=2{t.Fatalf("unexpected dispute count %d %v",count,err)}
}
