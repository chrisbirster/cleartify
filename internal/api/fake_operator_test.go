package api

import (
 "context"
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
