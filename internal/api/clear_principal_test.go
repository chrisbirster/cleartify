package api

import (
 "context"
 "net/http"
 "net/http/httptest"
 "testing"
 "time"
 "github.com/chrisbirster/cleartify/internal/alpha"
 "github.com/chrisbirster/cleartify/internal/clearsession"
 "github.com/chrisbirster/cleartify/internal/database"
 "github.com/chrisbirster/cleartify/internal/security"
)
func TestRealClearSessionAlphaGate(t *testing.T){
 ctx:=context.Background()
 db,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)};defer db.Close()
 if err=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 stamp:=time.Date(2026,10,13,12,0,0,0,time.UTC).Format(time.RFC3339Nano)
 for _,id:=range []string{"u1","u2"}{
  if _,err=db.DB.ExecContext(ctx,"INSERT INTO users(id,display_name,primary_email,status,created_at,updated_at) VALUES(?,?,?,'active',?,?)",id,id,id+"@test.example",stamp,stamp);err!=nil{t.Fatal(err)}
  if _,err=db.DB.ExecContext(ctx,"INSERT INTO waitlist_entries(id,name,email,email_normalized,phone_e164,status,created_at,updated_at) VALUES(?,?,?,?,?,'activated',?,?)",id,id,id+"@test.example",id+"@test.example","+1555000000"+id[1:],stamp,stamp);err!=nil{t.Fatal(err)}
  if _,err=db.DB.ExecContext(ctx,"INSERT INTO beta_invites(id,waitlist_entry_id,status,email_token_hash,email_token_expires_at,activated_user_id,activated_at,created_at,updated_at) VALUES(?,?,'activated',?,?,?,?,?,?)","invite-"+id,id,"hash-"+id,stamp,id,stamp,stamp,stamp);err!=nil{t.Fatal(err)}
 }
 gate:=alpha.Gate{DB:db.DB}
 if err:=gate.Admit(ctx,"u1");err!=nil{t.Fatal(err)}
 tokens,err:=security.NewTokens("01234567890123456789012345678901");if err!=nil{t.Fatal(err)}
 session:=&clearsession.Service{Store:db,Tokens:tokens}
 h:=New(Options{PrivateAlphaRoutesEnabled:true,AlphaEligibility:gate,ClearPrincipalResolver:ClearSessionPrincipal(session)})
 request:=func(raw string)int{
  w:=httptest.NewRecorder();req:=httptest.NewRequest(http.MethodGet,"/api/me/alpha-eligibility?userId=u1",nil)
  if raw!=""{req.AddCookie(&http.Cookie{Name:"clear_session",Value:raw})}
  h.ServeHTTP(w,req);return w.Code
 }
 if got:=request("");got!=401{t.Fatalf("anonymous %d",got)}
 if got:=request("forged");got!=401{t.Fatalf("forged %d",got)}
 second,err:=session.Issue(ctx,"u2");if err!=nil{t.Fatal(err)}
 if got:=request(second);got!=403{t.Fatalf("cross-user alpha %d",got)}
 first,err:=session.Issue(ctx,"u1");if err!=nil{t.Fatal(err)}
 if got:=request(first);got!=200{t.Fatalf("eligible %d",got)}
 if err:=gate.Revoke(ctx,"u1");err!=nil{t.Fatal(err)}
 if got:=request(first);got!=403{t.Fatalf("revoked membership %d",got)}
 if err:=session.Revoke(ctx,first);err!=nil{t.Fatal(err)}
 if got:=request(first);got!=401{t.Fatalf("revoked session %d",got)}
}
