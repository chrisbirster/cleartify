package monitoring
import("net/http";"net/http/httptest";"testing")
func TestDisabledDoesNotObserve(t *testing.T){
 c:=&Counter{}
 h:=http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){w.WriteHeader(503)})
 wrapped:=Wrap(h,false,c)
 wrapped.ServeHTTP(httptest.NewRecorder(),httptest.NewRequest("GET","/healthz",nil))
 if c.Snapshot().Requests!=0{t.Fatal("disabled hook ran")}
}
func TestOptInCountsErrors(t *testing.T){
 c:=&Counter{}
 h:=Wrap(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  if r.URL.Path=="/err"{http.Error(w,"failure",http.StatusInternalServerError);return}
  w.Write([]byte("ok"))
 }),true,c)
 for _,path:=range []string{"/","/err"}{h.ServeHTTP(httptest.NewRecorder(),httptest.NewRequest("GET",path,nil))}
 x:=c.Snapshot();if x.Requests!=2||x.Errors!=1||x.LastObservedAt.IsZero(){t.Fatalf("%+v",x)}
}
