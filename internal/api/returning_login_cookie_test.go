package api

import (
 "bytes"
 "context"
 "encoding/json"
 "net/http"
 "net/http/cookiejar"
 "net/http/httptest"
 "testing"
 "time"
 "strings"
 "github.com/chrisbirster/cleartify/internal/clearsession"
 "github.com/chrisbirster/cleartify/internal/database"
 "github.com/chrisbirster/cleartify/internal/notifications"
 "github.com/chrisbirster/cleartify/internal/security"
)
type loginSocialDID struct{}
func(loginSocialDID)SocialSessionDID(*http.Request)(string,bool){return "did:plc:aaaaaaaaaaaaaaaaaaaaaaaa",true}
type loginCapture struct {codes []string}
func(s *loginCapture)SendEmail(context.Context,notifications.Email)error{return nil}
func(s *loginCapture)SendSMS(_ context.Context,m notifications.SMS)error{s.codes=append(s.codes,strings.TrimSuffix(strings.Split(m.Text," ")[5],"."));return nil}
func TestLoginCookieJarPrivateRoutes(t *testing.T){
 ctx:=context.Background()
 db,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)};defer db.Close()
 if err=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 now:=time.Now().UTC().Format(time.RFC3339Nano)
 _,err=db.DB.ExecContext(ctx,`INSERT INTO users(id,display_name,primary_email,status,created_at,updated_at) VALUES('user-a','A','a@test.invalid','active',?,?)`,now,now);if err!=nil{t.Fatal(err)}
 _,err=db.DB.ExecContext(ctx,`INSERT INTO phone_identities(id,user_id,phone_e164,verified_at,created_at) VALUES('phone-a','user-a','+15550000123',?,?)`,now,now);if err!=nil{t.Fatal(err)}
 tokens,err:=security.NewTokens(strings.Repeat("p",32));if err!=nil{t.Fatal(err)}
 sess:=&clearsession.Service{Store:db,Tokens:tokens}
 sender:=&loginCapture{}
 login:=&clearsession.Login{Sessions:sess,Sender:sender}
 if err=login.Ready(ctx);err!=nil{t.Fatal(err)}
 handler:=New(Options{ReturningLoginEnabled:true,ReturningLogin:login,ClearSessionsEnabled:true,ClearSessions:sess,AppOrigin:"http://localhost:3000",IdentityLinkingEnabled:true,LinkSocial:loginSocialDID{},IdentityLinks:&LinkPersistence{DB:db.DB,Push:db.Push}})
 server:=httptest.NewServer(handler);defer server.Close()
 jar,err:=cookiejar.New(nil);if err!=nil{t.Fatal(err)}
 client:=&http.Client{Jar:jar}
 call:=func(path string,payload any)*http.Response{
  t.Helper()
  body,err:=json.Marshal(payload);if err!=nil{t.Fatal(err)}
  req,err:=http.NewRequest("POST",server.URL+path,bytes.NewReader(body));if err!=nil{t.Fatal(err)}
  req.Header.Set("Content-Type","application/json")
  req.Header.Set("Origin","http://localhost:3000")
  resp,err:=client.Do(req);if err!=nil{t.Fatal(err)}
  return resp
 }
 r:=call("/api/login/start",map[string]string{"phone":"+15550000123"});r.Body.Close()
 if r.StatusCode!=202||len(sender.codes)!=1{t.Fatalf("start status=%d codes=%d",r.StatusCode,len(sender.codes))}
 r=call("/api/login/verify",map[string]string{"phone":"+15550000123","code":sender.codes[0]});r.Body.Close()
 if r.StatusCode!=200{t.Fatalf("verify %d",r.StatusCode)}
 res,err:=client.Get(server.URL+"/api/me/clear-session");if err!=nil{t.Fatal(err)};res.Body.Close()
 if res.StatusCode!=200{t.Fatalf("private cookie not sent: %d",res.StatusCode)}
 link:=call("/api/me/identity-link/start",map[string]any{});link.Body.Close()
 if link.StatusCode!=201{t.Fatalf("identity link did not accept freshly issued Clear cookie: %d",link.StatusCode)}
 preview,err:=client.Get(server.URL+"/api/me/identity-link/preview");if err!=nil{t.Fatal(err)};preview.Body.Close()
 if preview.StatusCode!=200{t.Fatalf("identity-link preview %d",preview.StatusCode)}
 // Cookie should reach both /api/me/identity-link and /api/me/protocol-install-requests paths,
 // not /api/login. Those features require separately configured OAuth/services.
 req,err:=http.NewRequest("POST",server.URL+"/api/me/identity-link/start",nil);if err!=nil{t.Fatal(err)}
 for _,c:=range jar.Cookies(req.URL){if c.Name=="clear_session"{goto found}}
 t.Fatal("identity-link request does not receive Clear cookie")
 found:
 req,err=http.NewRequest("POST",server.URL+"/api/me/protocol-install-requests",nil);if err!=nil{t.Fatal(err)}
 seen:=false;for _,c:=range jar.Cookies(req.URL){if c.Name=="clear_session"{seen=true}}
 if !seen{t.Fatal("install request does not receive Clear cookie")}
 req,err=http.NewRequest("POST",server.URL+"/api/login/verify",nil);if err!=nil{t.Fatal(err)}
 for _,c:=range jar.Cookies(req.URL){if c.Name=="clear_session"{t.Fatal("login path unexpectedly receives private cookie")}}
}
