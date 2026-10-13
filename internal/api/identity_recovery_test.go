package api
import(
 "context"
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
 "github.com/chrisbirster/cleartify/internal/clearsession"
 "github.com/chrisbirster/cleartify/internal/database"
 "github.com/chrisbirster/cleartify/internal/security"
)
func TestLocalIdentityLinkedE2ERecoveryAndDisconnect(t *testing.T){
 ctx:=context.Background();db,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)};defer db.Close()
 if err=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 for _,u:=range []string{"alice","bob"}{
  if _,err=db.DB.ExecContext(ctx,`INSERT INTO users(id,display_name,primary_email,status,created_at,updated_at) VALUES(?,?,?,'active','now','now')`,u,u,u+"@example.test");err!=nil{t.Fatal(err)}
 }
 tok,err:=security.NewTokens("01234567890123456789012345678901");if err!=nil{t.Fatal(err)}
 sessions:=&clearsession.Service{Store:db,Tokens:tok}
 first,err:=sessions.Issue(ctx,"alice");if err!=nil{t.Fatal(err)}
 second,err:=sessions.Issue(ctx,"alice");if err!=nil{t.Fatal(err)}
 bob,err:=sessions.Issue(ctx,"bob");if err!=nil{t.Fatal(err)}
 opts:=Options{AppOrigin:"https://app.test",IdentityLinkingEnabled:true,IdentityRecoveryEnabled:true,ClearSessionsEnabled:true,ClearSessions:sessions,IdentityLinks:&LinkPersistence{DB:db.DB,Push:db.Push},LinkSocial:fakeLinkSocial{"did:plc:alice",true}}
 handler:=New(opts)
 call:=func(method,path,body,token,nonce string)(int,*httptest.ResponseRecorder){
  req:=httptest.NewRequest(method,path,strings.NewReader(body));req.Header.Set("Origin",opts.AppOrigin);req.Header.Set("Content-Type","application/json")
  if token!=""{req.AddCookie(&http.Cookie{Name:"clear_session",Value:token})}
  if nonce!=""{req.AddCookie(&http.Cookie{Name:linkCookie,Value:nonce})}
  w:=httptest.NewRecorder();handler.ServeHTTP(w,req);return w.Code,w
 }
 code,w:=call("POST","/api/me/identity-link/start","{}",first,"");if code!=201{t.Fatalf("start %d %s",code,w.Body)}
 nonce:="";for _,cookie:=range w.Result().Cookies(){if cookie.Name==linkCookie{nonce=cookie.Value}}
 if nonce==""{t.Fatal("missing challenge")}
 if n,_:=call("POST","/api/me/identity-link/confirm",`{"consent":true,"consentVersion":"internal-m2-v1"}`,first,nonce);n!=200{t.Fatalf("confirm %d",n)}
 if n,_:=call("POST","/api/me/identity-link/disconnect",`{"confirm":true}`,bob,"");n!=409{t.Fatalf("cross-account %d",n)}
 if n,_:=call("POST","/api/me/identity-link/disconnect",`{"confirm":false}`,first,"");n!=400{t.Fatalf("missing consent %d",n)}
 if n,_:=call("POST","/api/me/identity-link/disconnect",`{"confirm":true}`,first,"");n!=200{t.Fatalf("disconnect %d",n)}
 for _,raw:=range []string{first,second}{if _,err=sessions.Authenticate(ctx,raw);err==nil{t.Fatal("Clear session survived disconnect")}}
 var n int
 if err=db.DB.QueryRowContext(ctx,`SELECT count(*) FROM identity_links WHERE clear_user_id='alice' AND status='active'`).Scan(&n);err!=nil||n!=0{t.Fatalf("link survived %d %v",n,err)}
 if _,err=sessions.Authenticate(ctx,bob);err!=nil{t.Fatal("cross-user session incorrectly revoked",err)}
 renewed,err:=sessions.Issue(ctx,"alice");if err!=nil{t.Fatal(err)}
 code,w=call("POST","/api/me/identity-link/start","{}",renewed,"");if code!=201{t.Fatalf("relink start %d",code)}
 nonce="";for _,cookie:=range w.Result().Cookies(){if cookie.Name==linkCookie{nonce=cookie.Value}}
 if n,_:=call("POST","/api/me/identity-link/confirm",`{"consent":true,"consentVersion":"internal-m2-v1"}`,renewed,nonce);n!=200{t.Fatalf("relink %d",n)}
 if n,_:=call("POST","/api/me/identity-link/recover",`{"confirm":true}`,renewed,"");n!=200{t.Fatalf("local recovery %d",n)}
 if _,err=sessions.Authenticate(ctx,renewed);err==nil{t.Fatal("recovery did not revoke session")}
 var audits int;if err=db.DB.QueryRowContext(ctx,`SELECT count(*) FROM identity_link_events WHERE clear_user_id='alice' AND action='unlinked'`).Scan(&audits);err!=nil||audits!=2{t.Fatalf("disconnect/recovery audit %d %v",audits,err)}
 opts.IdentityRecoveryEnabled=false
 off:=New(opts);resp:=httptest.NewRecorder();off.ServeHTTP(resp,httptest.NewRequest("POST","/api/me/identity-link/recover",nil))
 if resp.Code!=404{t.Fatalf("recovery enabled by default: %d",resp.Code)}
}
