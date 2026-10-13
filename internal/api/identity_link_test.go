package api

import (
 "context"
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"

 "github.com/chrisbirster/cleartify/internal/clearsession"
 "github.com/chrisbirster/cleartify/internal/database"
 "github.com/chrisbirster/cleartify/internal/security"
)

type fakeLinkSocial struct{did string; valid bool}
func(f fakeLinkSocial)SocialSessionDID(*http.Request)(string,bool){return f.did,f.valid}

func TestIdentityLinkCeremonyFakeOAuth(t *testing.T){
 ctx:=context.Background()
 db,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)};defer db.Close()
 if err=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 for _,id:=range []string{"u1","u2"}{
  if _,err=db.DB.ExecContext(ctx,`INSERT INTO users(id,display_name,primary_email,status,created_at,updated_at) VALUES(?,?,?,'active','now','now')`,id,id,id+"@example.com");err!=nil{t.Fatal(err)}
 }
 tokens,err:=security.NewTokens("01234567890123456789012345678901");if err!=nil{t.Fatal(err)}
 sessions:=&clearsession.Service{Store:db,Tokens:tokens}
 u1,err:=sessions.Issue(ctx,"u1");if err!=nil{t.Fatal(err)}
 u2,err:=sessions.Issue(ctx,"u2");if err!=nil{t.Fatal(err)}
 opts:=Options{AppOrigin:"https://cleartify.example",IdentityLinkingEnabled:true,ClearSessionsEnabled:true,ClearSessions:sessions,IdentityLinks:&LinkPersistence{DB:db.DB,Push:db.Push},LinkSocial:fakeLinkSocial{"did:plc:alice",true}}
 h:=New(opts)
 call:=func(method,path,body,token,challenge string)(int,*httptest.ResponseRecorder){
  t.Helper()
  req:=httptest.NewRequest(method,path,strings.NewReader(body))
  req.Header.Set("Origin",opts.AppOrigin)
  req.Header.Set("Content-Type","application/json")
  if token!=""{req.AddCookie(&http.Cookie{Name:"clear_session",Value:token})}
  if challenge!=""{req.AddCookie(&http.Cookie{Name:linkCookie,Value:challenge})}
  w:=httptest.NewRecorder();h.ServeHTTP(w,req);return w.Code,w
 }
 if code,_:=call("POST","/api/me/identity-link/start","{}", "", "");code!=401{t.Fatalf("missing Clear auth %d",code)}
 code,w:=call("POST","/api/me/identity-link/start","{}",u1,"");if code!=201{t.Fatalf("start %d %s",code,w.Body)}
 var nonce string
 for _,c:=range w.Result().Cookies(){if c.Name==linkCookie{nonce=c.Value;if !c.HttpOnly||!c.Secure{t.Fatal("unsafe challenge cookie")}}}
 if nonce==""{t.Fatal("missing nonce cookie")}
 if code,_=call("GET","/api/me/identity-link/preview","",u2,nonce);code!=409{t.Fatalf("cross-user preview %d",code)}
 if code,_=call("GET","/api/me/identity-link/preview","",u1,nonce);code!=200{t.Fatalf("preview %d",code)}
 if code,_=call("POST","/api/me/identity-link/confirm",`{"consent":false,"consentVersion":"internal-m2-v1"}`,u1,nonce);code!=400{t.Fatalf("consent bypass %d",code)}
 if code,_=call("POST","/api/me/identity-link/confirm",`{"consent":true,"consentVersion":"internal-m2-v1"}`,u1,nonce);code!=200{t.Fatalf("confirm %d",code)}
 if code,_=call("POST","/api/me/identity-link/confirm",`{"consent":true,"consentVersion":"internal-m2-v1"}`,u1,nonce);code!=409{t.Fatalf("replay %d",code)}
 if code,_=call("POST","/api/me/identity-link/start","{}",u2,"");code!=201{t.Fatalf("second start %d",code)}
 var nonce2 string
 _,w=call("POST","/api/me/identity-link/start","{}",u2,"")
 for _,c:=range w.Result().Cookies(){if c.Name==linkCookie{nonce2=c.Value}}
 if code,_=call("POST","/api/me/identity-link/confirm",`{"consent":true,"consentVersion":"internal-m2-v1"}`,u2,nonce2);code!=409{t.Fatalf("duplicate DID %d",code)}
 if code,_=call("DELETE","/api/me/identity-link",`{"confirm":true}`,u1,"");code!=200{t.Fatalf("unlink %d",code)}
 if code,_=call("POST","/api/me/identity-link/start","{}",u2,"");code!=201{t.Fatalf("relink start %d",code)}
 opts.IdentityLinkingEnabled=false
 off:=New(opts)
 w=httptest.NewRecorder();off.ServeHTTP(w,httptest.NewRequest("POST","/api/me/identity-link/start",nil))
 if w.Code!=404{t.Fatalf("default-off route exposed: %d",w.Code)}
 if err=sessions.Revoke(ctx,u1);err!=nil{t.Fatal(err)}
 if code,_=call("POST","/api/me/identity-link/start","{}",u1,"");code!=401{t.Fatalf("revoked Clear session %d",code)}
}
func TestIdentityLinkFailsClosedWithoutSocialVerifier(t *testing.T){
 h:=New(Options{IdentityLinkingEnabled:true,ClearSessionsEnabled:true,ClearSessions:&clearsession.Service{}})
 w:=httptest.NewRecorder();h.ServeHTTP(w,httptest.NewRequest("POST","/api/me/identity-link/start",nil))
 if w.Code!=404{t.Fatalf("unconfigured route enabled: %d",w.Code)}
}
