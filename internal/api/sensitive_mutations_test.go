package api
import(
 "bytes"
 "context"
 "encoding/base64"
 "net/http"
 "net/http/httptest"
 "testing"
 "time"
 "github.com/chrisbirster/cleartify/internal/atproto"
 "github.com/chrisbirster/cleartify/internal/database"
 "github.com/chrisbirster/cleartify/internal/social"
)
func TestSensitiveMutationLimiterWindowAndNoForwardedSpoofing(t *testing.T){
 now:=time.Date(2026,10,13,8,0,0,0,time.UTC)
 l:=newMutationLimiter();l.now=func()time.Time{return now}
 for i:=0;i<12;i++{if !l.allow("192.0.2.1|POST|/api/social/accounts"){t.Fatalf("blocked request %d",i)}}
 if l.allow("192.0.2.1|POST|/api/social/accounts"){t.Fatal("13th request not throttled")}
 now=now.Add(61*time.Second)
 if !l.allow("192.0.2.1|POST|/api/social/accounts"){t.Fatal("window never reset")}
 req:=httptest.NewRequest("POST","/api/social/accounts",nil)
 req.RemoteAddr="192.0.2.1:1122"
 req.Header.Set("X-Forwarded-For","203.0.113.123")
 if sourceIP(req)!="192.0.2.1"{t.Fatalf("trusted spoofed header: %q",sourceIP(req))}
}
func TestSocialEndpointsCSRFAuthAndRateLimit(t *testing.T){
 ctx:=context.Background()
 db,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)}
 defer db.Close()
 if err=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 origin:="https://app.example"
 key:=base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{6},32))
 oauth,err:=atproto.NewOAuthHandler(db,"https://api.app.example",origin,key);if err!=nil{t.Fatal(err)}
 handler:=New(Options{AppOrigin:origin,OAuth:oauth,SocialAccounts:&social.Accounts{Store:db},SocialAccountRegistrationEnabled:true,SocialActions:&social.Actions{Store:db},SocialProfileActionsEnabled:true})
 mk:=func(method,path,origin string)*http.Request{r:=httptest.NewRequest(method,path,nil);r.Header.Set("Origin",origin);return r}
 for _,tc:=range []struct{method,path string}{{"POST","/api/social/accounts"},{"POST","/api/me/profile/claim"},{"PUT","/api/me/following/chris"},{"DELETE","/api/me/following/chris"}}{
  bad:=mk(tc.method,tc.path,"https://evil.example");w:=httptest.NewRecorder();handler.ServeHTTP(w,bad)
  if w.Code!=403{t.Fatalf("%s %s wrong Origin %d",tc.method,tc.path,w.Code)}
  cross:=mk(tc.method,tc.path,origin);cross.Header.Set("Sec-Fetch-Site","cross-site")
  w=httptest.NewRecorder();handler.ServeHTTP(w,cross);if w.Code!=403{t.Fatalf("%s %s cross-site metadata %d",tc.method,tc.path,w.Code)}
  good:=mk(tc.method,tc.path,origin);good.AddCookie(&http.Cookie{Name:"cleartify_atproto_session",Value:"forged-token"})
  w=httptest.NewRecorder();handler.ServeHTTP(w,good);if w.Code!=401{t.Fatalf("%s %s forged session %d",tc.method,tc.path,w.Code)}
 }
 for i:=0;i<11;i++{
  r:=mk("POST","/api/social/accounts",origin);r.Header.Set("X-Forwarded-For","203.0.113."+string(rune('A'+i)))
  w:=httptest.NewRecorder();handler.ServeHTTP(w,r)
  if i<11&&w.Code!=401{t.Fatalf("request %d code %d",i,w.Code)}
 }
 r:=mk("POST","/api/social/accounts",origin)
 w:=httptest.NewRecorder();handler.ServeHTTP(w,r)
 if w.Code!=429||w.Header().Get("Retry-After")==""{t.Fatalf("rate limit=%d retry=%s",w.Code,w.Header().Get("Retry-After"))}
 var count int
 if err=db.DB.QueryRowContext(ctx,"SELECT COUNT(*) FROM social_profiles").Scan(&count);err!=nil||count!=0{t.Fatalf("unauthenticated profiles persisted %d %v",count,err)}
}
func TestDisabledSocialRoutesStayHidden(t *testing.T){
 s:=New(Options{AppOrigin:"https://app.example"})
 r:=httptest.NewRequest("POST","/api/social/accounts",nil)
 r.Header.Set("Origin","https://evil.example")
 w:=httptest.NewRecorder();s.ServeHTTP(w,r)
 if w.Code!=404{t.Fatalf("off route leaked as %d",w.Code)}
}
