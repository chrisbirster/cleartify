package api

import (
 "context"
 "errors"
 "net/http"
 "net/http/httptest"
 "testing"
)
type eligibilityStub struct{allow bool;err error;seen string}
func(e *eligibilityStub)Eligible(_ context.Context,id string)(bool,error){e.seen=id;return e.allow,e.err}
func TestPrivateAlphaGate(t *testing.T){
 for _,tc:=range []struct{name string;flag bool;principal string;authErr error;allow bool;gateErr error;status int}{
 {"flag off",false,"user1",nil,true,nil,404},
 {"missing session",true,"",errors.New("missing"),true,nil,401},
 {"revoked membership",true,"user1",nil,false,nil,403},
 {"backend error",true,"user1",nil,true,errors.New("unavailable"),503},
 {"allowed",true,"user1",nil,true,nil,200},
 }{
  t.Run(tc.name,func(t *testing.T){
   gate:=&eligibilityStub{allow:tc.allow,err:tc.gateErr}
   h:=New(Options{PrivateAlphaRoutesEnabled:tc.flag,AlphaEligibility:gate,ClearPrincipalResolver:func(*http.Request)(string,error){return tc.principal,tc.authErr}})
   w:=httptest.NewRecorder()
   h.ServeHTTP(w,httptest.NewRequest(http.MethodGet,"/api/me/alpha-eligibility?userId=someone-else",nil))
   if w.Code!=tc.status {t.Fatalf("got %d expected %d: %s",w.Code,tc.status,w.Body.String())}
   if w.Code==200 && gate.seen!="user1"{t.Fatalf("trusted request userId rather than principal: %q",gate.seen)}
  })
 }
}
func TestPrivateAlphaMissingDependencies(t *testing.T){
 h:=New(Options{PrivateAlphaRoutesEnabled:true})
 w:=httptest.NewRecorder();h.ServeHTTP(w,httptest.NewRequest(http.MethodGet,"/api/me/alpha-eligibility",nil))
 if w.Code!=503{t.Fatalf("fail open: %d",w.Code)}
}
