package api
import(
 "errors"
 "net/http"
 "net/http/httptest"
 "testing"
)
func TestAlphaPrivateRouteAuthorization(t *testing.T){
 for _,tc:=range []struct{name string;flag bool;principal string;authErr error;allow bool;gateErr error;want int}{
  {"off",false,"",errors.New("no session"),false,nil,204},
  {"spoofed user",true,"",errors.New("no session"),true,nil,401},
  {"revoked membership",true,"trusted",nil,false,nil,403},
  {"membership backend failure",true,"trusted",nil,true,errors.New("DB down"),503},
  {"authorized",true,"trusted",nil,true,nil,204},
 }{
  t.Run(tc.name,func(t *testing.T){
   gate:=&eligibilityStub{allow:tc.allow,err:tc.gateErr}
   s:=&Server{privateAlphaEnabled:tc.flag,alphaEligibility:gate,clearPrincipalResolver:func(*http.Request)(string,error){return tc.principal,tc.authErr}}
   next:=s.alphaPrivateRoute(func(w http.ResponseWriter,r *http.Request){w.WriteHeader(204)})
   for _,path:=range []string{"/api/me/identity-link/confirm?userId=attacker","/api/me/fake-install-runs?userId=attacker","/api/me/clear-session?userId=attacker"}{
    w:=httptest.NewRecorder();next(w,httptest.NewRequest("POST",path,nil))
    if w.Code!=tc.want{t.Fatalf("%s: %d expected %d",path,w.Code,tc.want)}
    if tc.want==204&&tc.flag&&gate.seen!="trusted"{t.Fatalf("spoofed principal %q",gate.seen)}
   }
  })
 }
}
