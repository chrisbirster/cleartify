package api

import (
 "bytes"
 "context"
 "encoding/json"
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"

 "github.com/chrisbirster/cleartify/internal/clearsession"
 "github.com/chrisbirster/cleartify/internal/database"
 "github.com/chrisbirster/cleartify/internal/followups"
 "github.com/chrisbirster/cleartify/internal/security"
)
func TestClearTextConversationSessionIsolation(t *testing.T){
 ctx:=context.Background()
 db,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)};defer db.Close()
 if err=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 for _,id:=range []string{"a","b"}{
  _,err=db.DB.ExecContext(ctx,`INSERT INTO users(id,display_name,primary_email,status,created_at,updated_at) VALUES(?,?,?,'active','now','now')`,id,id,id+"@test.invalid")
  if err!=nil{t.Fatal(err)}
 }
 tokens,err:=security.NewTokens(strings.Repeat("z",32));if err!=nil{t.Fatal(err)}
 sessions:=&clearsession.Service{Store:db,Tokens:tokens}
 a,err:=sessions.Issue(ctx,"a");if err!=nil{t.Fatal(err)}
 b,err:=sessions.Issue(ctx,"b");if err!=nil{t.Fatal(err)}
 svc:=&followups.Service{DB:db.DB};if err=svc.Ready(ctx);err!=nil{t.Fatal(err)}
 handler:=New(Options{AppOrigin:"http://localhost:3000",ClearSessionsEnabled:true,ClearSessions:sessions,ClearConversationEnabled:true,Followups:svc})
 send:=func(cookie,text string)(int,map[string]any){
  t.Helper();raw,_:=json.Marshal(map[string]string{"text":text})
  req:=httptest.NewRequest("POST","/api/me/clear/messages",bytes.NewReader(raw))
  req.Header.Set("Origin","http://localhost:3000")
  if cookie!=""{req.AddCookie(&http.Cookie{Name:"clear_session",Value:cookie})}
  rec:=httptest.NewRecorder();handler.ServeHTTP(rec,req)
  out:=map[string]any{};_ =json.Unmarshal(rec.Body.Bytes(),&out)
  return rec.Code,out
 }
 if status,_:=send("","remind me to buy groceries");status!=401{t.Fatalf("missing auth %d",status)}
 status,out:=send(b,"remind me to review recipe");if status!=201{t.Fatalf("create %d %+v",status,out)}
 followup,ok:=out["followup"].(map[string]any);if !ok{t.Fatal(out)}
 id,_:=followup["id"].(string)
 read:=func(cookie string)string{
  req:=httptest.NewRequest("GET","/api/me/clear/followups",nil)
  req.AddCookie(&http.Cookie{Name:"clear_session",Value:cookie})
  rec:=httptest.NewRecorder();handler.ServeHTTP(rec,req)
  if rec.Code!=200{t.Fatalf("get followups %d",rec.Code)}
  return rec.Body.String()
 }
 if strings.Contains(read(a),id){t.Fatal("A saw B followup")}
 if !strings.Contains(read(b),id){t.Fatal("B cannot see own followup")}
 if status,_:=send(a,"done "+id);status!=404{t.Fatalf("A completed B followup %d",status)}
 if status,_:=send(b,"done "+id);status!=200{t.Fatalf("B completion %d",status)}
}
