package api

import (
 "bytes"
 "context"
 "encoding/json"
 "net/http"
 "net/http/cookiejar"
 "net/http/httptest"
 "strings"
 "testing"
 "time"

 "github.com/chrisbirster/cleartify/internal/clearsession"
 "github.com/chrisbirster/cleartify/internal/database"
 "github.com/chrisbirster/cleartify/internal/protocols"
 "github.com/chrisbirster/cleartify/internal/security"
)

// Uses local DB and notification capture; no real social endpoint or SMS network.
func TestRound6TwoPersonPublishLoginInstallIsolation(t *testing.T) {
 ctx:=context.Background()
 db,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)};defer db.Close()
 if err=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 tokens,err:=security.NewTokens(strings.Repeat("q",32));if err!=nil{t.Fatal(err)}
 stamp:=time.Now().UTC().Format(time.RFC3339Nano)
 for _,u:=range []struct{id,phone string}{{"a","+15550001111"},{"b","+15550002222"}}{
  _,err=db.DB.ExecContext(ctx,`INSERT INTO users(id,display_name,primary_email,status,created_at,updated_at) VALUES(?,?,?,'active',?,?)`,u.id,u.id,u.id+"@test.invalid",stamp,stamp);if err!=nil{t.Fatal(err)}
  _,err=db.DB.ExecContext(ctx,`INSERT INTO phone_identities(id,user_id,phone_e164,verified_at,created_at) VALUES(?,?,?,?,?)`,u.id,u.id,u.phone,stamp,stamp);if err!=nil{t.Fatal(err)}
 }
 draft:=protocols.DraftService{DB:db.DB,Tokens:tokens,GuideStepsEnabled:true}
 aDID:="did:plc:aaaaaaaaaaaaaaaaaaaaaaaa"
 pub,err:=draft.Create(ctx,aDID,protocols.DraftInput{Slug:"safe-guide",Version:1,Title:"Safe guide",Summary:"Test-only guide",License:"CC0-1.0",Spec:json.RawMessage(`{"steps":[{"operation":"guide","text":"Take one small step"}],"permissions":[]}`)})
 if err!=nil{t.Fatal(err)}
 if err=draft.PublishDraft(ctx,aDID,pub,"private","reviewed by creator");err!=nil{t.Fatal(err)}
 if err=draft.PublishDraft(ctx,aDID,pub,"public","reviewed for test");err!=nil{t.Fatal(err)}
 sender:=&loginCapture{}
 sessions:=&clearsession.Service{Store:db,Tokens:tokens}
 login:=&clearsession.Login{Sessions:sessions,Sender:sender}
 if err=login.Ready(ctx);err!=nil{t.Fatal(err)}
 flow:=&protocols.InstallService{Store:db}
 manifest,err:=flow.Preview(ctx,pub);if err!=nil{t.Fatal(err)}
 handler:=New(Options{AppOrigin:"http://localhost:3000",ClearSessionsEnabled:true,ClearSessions:sessions,ReturningLoginEnabled:true,ReturningLogin:login,SelfServeGuideInstallsEnabled:true,InstallFlow:flow})
 server:=httptest.NewServer(handler);defer server.Close()
 jar,err:=cookiejar.New(nil);if err!=nil{t.Fatal(err)}
 client:=&http.Client{Jar:jar}
 post:=func(path string,p any)int{
  t.Helper();body,_:=json.Marshal(p)
  req,_:=http.NewRequest(http.MethodPost,server.URL+path,bytes.NewReader(body))
  req.Header.Set("Origin","http://localhost:3000");req.Header.Set("Content-Type","application/json")
  resp,e:=client.Do(req);if e!=nil{t.Fatal(e)};defer resp.Body.Close();return resp.StatusCode
 }
 install:=func()int{return post("/api/me/self-serve-guide-installs",map[string]string{"publicationId":pub,"acknowledgedManifestSha256":manifest.ManifestSHA256})}
 if got:=install();got!=401{t.Fatalf("unauthenticated install %d",got)}
 if got:=post("/api/login/start",map[string]string{"phone":"+15550002222"});got!=202{t.Fatalf("start B %d",got)}
 if len(sender.codes)!=1{t.Fatal("missing B challenge")}
 if got:=post("/api/login/verify",map[string]string{"phone":"+15550002222","code":sender.codes[0]});got!=200{t.Fatalf("B verify %d",got)}
 if got:=install();got!=201{t.Fatalf("B install %d",got)}
 if got:=install();got!=200{t.Fatalf("duplicate %d",got)}
 var bCount,aCount int
 if err=db.DB.QueryRowContext(ctx,"SELECT COUNT(*) FROM protocol_installations WHERE user_id='b' AND protocol_key=?",pub).Scan(&bCount);err!=nil{t.Fatal(err)}
 if err=db.DB.QueryRowContext(ctx,"SELECT COUNT(*) FROM protocol_installations WHERE user_id='a'").Scan(&aCount);err!=nil{t.Fatal(err)}
 if bCount!=1||aCount!=0{t.Fatalf("isolation: B=%d A=%d",bCount,aCount)}
 // A's unpublished draft cannot be obtained through B's install endpoint.
 private,err:=draft.Create(ctx,aDID,protocols.DraftInput{Slug:"secret-guide",Version:1,Title:"Secret",Summary:"Private",License:"CC0-1.0",Spec:json.RawMessage(`{"steps":[{"operation":"guide","text":"Secret step"}],"permissions":[]}`)})
 if err!=nil{t.Fatal(err)}
 if got:=post("/api/me/self-serve-guide-installs",map[string]string{"publicationId":private,"acknowledgedManifestSha256":strings.Repeat("0",64)});got!=404{t.Fatalf("draft exposed to B %d",got)}
}
