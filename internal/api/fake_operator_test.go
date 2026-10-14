package api

import (
 "net/http/httptest"
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
