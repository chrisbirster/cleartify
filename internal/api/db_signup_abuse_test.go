package api
import(
 "net/http/httptest"
 "testing"
 "github.com/chrisbirster/cleartify/internal/database"
 "context"
)
func TestDBSignupAbuseAcrossInstances(t *testing.T){
 ctx:=context.Background()
 db,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)};defer db.Close()
 if err=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 secret:="01234567890123456789012345678901"
 a:=&Server{dbAbuseEnabled:true,dbAbuseDB:db.DB,dbAbuseSecret:secret}
 b:=&Server{dbAbuseEnabled:true,dbAbuseDB:db.DB,dbAbuseSecret:secret}
 for i:=0;i<6;i++{
  s:=a;if i%2==1{s=b}
  r:=httptest.NewRequest("POST","/api/waitlist",nil);r.RemoteAddr="198.51.100.10:1234";r.Header.Set("X-Forwarded-For","192.0.2.1")
  w:=httptest.NewRecorder()
  allowed:=s.allowDBSignup(w,r)
  if allowed!=(i<5){t.Fatalf("request %d allowed=%v status=%d",i,allowed,w.Code)}
  if i==5&&w.Code!=429{t.Fatalf("expected 429 got %d",w.Code)}
 }
 r:=httptest.NewRequest("POST","/api/waitlist",nil);r.RemoteAddr="198.51.100.11:1234"
 if !a.allowDBSignup(httptest.NewRecorder(),r){t.Fatal("different peer limited")}
 if !(&Server{}).allowDBSignup(httptest.NewRecorder(),r){t.Fatal("disabled guard blocked")}
 if (&Server{dbAbuseEnabled:true}).allowDBSignup(httptest.NewRecorder(),r){t.Fatal("enabled missing DB must fail closed")}
}
