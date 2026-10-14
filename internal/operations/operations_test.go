package operations
import(
 "bytes"
 "context"
 "log/slog"
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
 "github.com/chrisbirster/cleartify/internal/database"
)
func TestPreflightOnlyNames(t *testing.T){
 secret:="sensitive-hidden-example-1234567890-long"
 env:=map[string]string{"APP_SECRET":secret,"APP_ORIGIN":"http://bad.example","DATABASE_PATH":":memory:","NOTIFICATION_MODE":"debug","ATPROTO_OAUTH_ENABLED":"true"}
 findings:=Preflight(env)
 if len(findings)<3{t.Fatalf("expected findings, got %d",len(findings))}
 for _,f:=range findings{
  if strings.Contains(f.Name,secret)||strings.Contains(f.Problem,secret)||strings.Contains(f.Problem,"bad.example"){t.Fatal("configuration value leaked")}
 }
 valid:=map[string]string{"APP_SECRET":secret,"APP_ORIGIN":"https://app.example","DATABASE_PATH":":memory:","PORT":"8080"}
 if got:=Preflight(valid);len(got)!=0{t.Fatalf("valid config failed: %+v",got)}
}
func TestReadinessAndDrain(t *testing.T){
 d,e:=database.Open(context.Background(),database.Config{Path:":memory:"});if e!=nil{t.Fatal(e)};defer d.Close()
 r:=&Readiness{DB:d.DB}
 app:=WithReadiness(http.HandlerFunc(func(w http.ResponseWriter,_ *http.Request){w.WriteHeader(200)}),r)
 request:=func(path string)int{w:=httptest.NewRecorder();app.ServeHTTP(w,httptest.NewRequest(http.MethodGet,path,nil));return w.Code}
 if code:=request("/readyz");code!=200{t.Fatal(code)}
 if code:=request("/healthz");code!=200{t.Fatal(code)}
 r.Drain()
 if code:=request("/readyz");code!=503{t.Fatal(code)}
 if code:=request("/healthz");code!=200{t.Fatal(code)}
}
func TestStructuredRedaction(t *testing.T){
 var output bytes.Buffer
 logger:=slog.New(RedactingHandler{Next:slog.NewJSONHandler(&output,nil)})
 secret:="sensitive-hidden-example-1234567890-long"
 logger.Info("safe_event", "access_token",secret,"authorization",secret,"error",secret,"event_code","SAFE","nested",slog.GroupValue(slog.String("api_key",secret)))
 if strings.Contains(output.String(),secret){t.Fatal("secret present in structured output")}
 if !strings.Contains(output.String(),"[REDACTED]")||!strings.Contains(output.String(),"SAFE"){t.Fatal("redaction or event code absent")}
}
