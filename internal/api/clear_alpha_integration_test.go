package api
import(
 "context"
 "net/http"
 "net/http/httptest"
 "testing"
 "github.com/chrisbirster/cleartify/internal/alpha"
 "github.com/chrisbirster/cleartify/internal/clearsession"
 "github.com/chrisbirster/cleartify/internal/database"
 "github.com/chrisbirster/cleartify/internal/security"
)
func TestRealClearSessionsAlphaAccessAndCrossUser(t *testing.T){
 ctx:=context.Background()
 db,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)};defer db.Close()
 if err=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 for _,u:=range []struct{id,email string}{{"u1","u1@example.com"},{"u2","u2@example.com"}}{
  if _,err=db.DB.ExecContext(ctx,`INSERT INTO users(id,display_name,primary_email,status,created_at,updated_at) VALUES(?,'Test',?,'active','now','now')`,u.id,u.email);err!=nil{t.Fatal(err)}
 }
 _,err=db.DB.ExecContext(ctx,`INSERT INTO waitlist_entries(id,name,email,email_normalized,phone_e164,created_at,updated_at) VALUES('w','Test','u1@example.com','u1@example.com','+17175551212','now','now')`);if err!=nil{t.Fatal(err)}
 _,err=db.DB.ExecContext(ctx,`INSERT INTO beta_invites(id,waitlist_entry_id,status,email_token_hash,email_token_expires_at,activated_user_id,activated_at,created_at,updated_at) VALUES('i','w','activated','hash','later','u1','now','now','now')`);if err!=nil{t.Fatal(err)}
 gate:=alpha.Gate{DB:db.DB};if err:=gate.Admit(ctx,"u1");err!=nil{t.Fatal(err)}
 tok,err:=security.NewTokens("01234567890123456789012345678901");if err!=nil{t.Fatal(err)}
 sessions:=&clearsession.Service{Store:db,Tokens:tok}
 s1,err:=sessions.Issue(ctx,"u1");if err!=nil{t.Fatal(err)}
 s2,err:=sessions.Issue(ctx,"u2");if err!=nil{t.Fatal(err)}
 h:=New(Options{ClearSessionsEnabled:true,PrivateAlphaRoutesEnabled:true,ClearSessions:sessions,ClearPrincipalResolver:ClearSessionPrincipal(sessions),AlphaEligibility:gate})
 check:=func(token,path string,want int){
  t.Helper()
  req:=httptest.NewRequest("GET",path,nil)
  if token!=""{req.AddCookie(&http.Cookie{Name:"clear_session",Value:token})}
  w:=httptest.NewRecorder();h.ServeHTTP(w,req)
  if w.Code!=want{t.Fatalf("%s => %d want %d: %s",path,w.Code,want,w.Body.String())}
 }
 check("","/api/me/clear-session",401)
 check(s2,"/api/me/clear-session?userId=u1",403)
 check(s1,"/api/me/clear-session?userId=u2",200)
 check(s2,"/api/me/alpha-eligibility?userId=u1",403)
 check(s1,"/api/me/alpha-eligibility",200)
 if err:=gate.Revoke(ctx,"u1");err!=nil{t.Fatal(err)}
 check(s1,"/api/me/clear-session",403)
 check(s1,"/api/me/alpha-eligibility",403)
 if err:=sessions.Revoke(ctx,s1);err!=nil{t.Fatal(err)}
 check(s1,"/api/me/clear-session",401)
 hOff:=New(Options{ClearSessionsEnabled:true,ClearSessions:sessions,ClearPrincipalResolver:ClearSessionPrincipal(sessions),AlphaEligibility:gate})
 w:=httptest.NewRecorder();hOff.ServeHTTP(w,httptest.NewRequest("GET","/api/me/alpha-eligibility",nil))
 if w.Code!=404{t.Fatalf("gate flag should hide alpha route: %d",w.Code)}
}
func TestAlphaRouteFailsClosedWithoutValidator(t *testing.T){
 h:=New(Options{ClearSessionsEnabled:true,PrivateAlphaRoutesEnabled:true,ClearSessions:&clearsession.Service{},AlphaEligibility:&eligibilityStub{allow:true}})
 for _,path:=range []string{"/api/me/clear-session","/api/me/alpha-eligibility"}{
  w:=httptest.NewRecorder();h.ServeHTTP(w,httptest.NewRequest("GET",path,nil))
  if w.Code!=503{t.Fatalf("%s fail-open %d",path,w.Code)}
 }
}
