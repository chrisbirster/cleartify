package api
import(
 "net/http/httptest"
 "testing"
 "github.com/chrisbirster/cleartify/internal/marketplace"
)
func TestFakeCheckoutRoutesDisabled(t *testing.T){
 for _,path:=range []struct{method,path string}{
  {"POST","/api/me/fake-checkout"},
  {"GET","/api/me/marketplace-listings/test/access"},
  {"POST","/api/me/fake-entitlements/test/refund"},
 }{
  for _,opts:=range []Options{{},{FakeCheckoutEnabled:true,FakeCheckout:&marketplace.Service{Enabled:true}}}{
   w:=httptest.NewRecorder();New(opts).ServeHTTP(w,httptest.NewRequest(path.method,path.path,nil))
   if w.Code!=404{t.Fatalf("%s %s flag/dependency fail closed: %d",path.method,path.path,w.Code)}
  }
 }
}
