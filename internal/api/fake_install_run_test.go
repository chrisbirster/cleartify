package api
import(
 "net/http"
 "net/http/httptest"
 "testing"
 "github.com/chrisbirster/cleartify/internal/clearsession"
)
func TestFakeInstallRunFailClosed(t *testing.T){
 for _,tc:=range []struct{name string;enabled bool;want int}{
  {"flag_off",false,404},{"missing_dependencies",true,404},
 }{
  t.Run(tc.name,func(t *testing.T){
   h:=New(Options{FakeInstallRunEnabled:tc.enabled,ClearSessionsEnabled:true,ClearSessions:&clearsession.Service{}})
   for _,v:=range []struct{method,path string}{{"POST","/api/me/fake-install-runs"},{"GET","/api/me/fake-install-runs/run-id"}}{
    w:=httptest.NewRecorder();h.ServeHTTP(w,httptest.NewRequest(v.method,v.path,nil))
    if w.Code!=tc.want{t.Fatalf("%s %s: got %d want %d",v.method,v.path,w.Code,tc.want)}
   }
  })
 }
}
