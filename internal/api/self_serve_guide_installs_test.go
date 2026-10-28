package api

import (
 "bytes"
 "context"
 "encoding/json"
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
 "time"
 "github.com/chrisbirster/cleartify/internal/clearsession"
 "github.com/chrisbirster/cleartify/internal/clear"
 "github.com/chrisbirster/cleartify/internal/database"
 "github.com/chrisbirster/cleartify/internal/protocols"
 "github.com/chrisbirster/cleartify/internal/security"
)
func TestSelfServeGuideInstallOnlyVisiblePublicAndOwnerScoped(t *testing.T){
 ctx:=context.Background();db,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)};defer db.Close()
 if err=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 now:=time.Now().UTC().Format(time.RFC3339Nano)
 for _,user:=range []string{"a","b"}{
  if _,err=db.DB.ExecContext(ctx,`INSERT INTO users(id,display_name,primary_email,status,created_at,updated_at) VALUES(?,?,?,'active',?,?)`,user,user,user+"@test.invalid",now,now);err!=nil{t.Fatal(err)}
 }
 spec:=`{"steps":[{"operation":"guide","text":"First step"}],"permissions":[]}`
 for _,p:=range []struct{id,status,spec string}{{"pub","published",spec},{"draft","draft",spec},{"hidden","published",spec},{"exec","published",`{"steps":[{"operation":"local.set","key":"a","value":"b"}],"permissions":[]}`}}{
  _,err=db.DB.ExecContext(ctx,`INSERT INTO published_protocols(id,creator_did,slug,version,title,summary,license,spec_json,status,created_at) VALUES(?,'did:plc:aaaaaaaaaaaaaaaaaaaaaaaa',?,1,'Recipe','summary','MIT',?,?,?)`,p.id,p.id,p.spec,p.status,now);if err!=nil{t.Fatal(err)}
 }
 for _,id:=range []string{"pub","hidden","exec"}{
  _,err=db.DB.ExecContext(ctx,`INSERT INTO creator_publications(protocol_id,visibility,original_source_id,provenance_note,approved_at,updated_at) VALUES(?,'public','','reviewed',?,?)`,id,now,now);if err!=nil{t.Fatal(err)}
 }
 _,err=db.DB.ExecContext(ctx,`INSERT INTO creator_takedown_requests(id,protocol_id,requester_did,reason_key,status,created_at) VALUES('hide1','hidden','did:plc:aaaaaaaaaaaaaaaaaaaaaaaa','rights_review','hidden',?)`,now);if err!=nil{t.Fatal(err)}
 tok,err:=security.NewTokens(strings.Repeat("x",32));if err!=nil{t.Fatal(err)}
 sessions:=&clearsession.Service{Store:db,Tokens:tok}
 b,err:=sessions.Issue(ctx,"b");if err!=nil{t.Fatal(err)}
 flow:=&protocols.InstallService{Store:db}
 origin:="http://localhost:3000"
 off:=New(Options{ClearSessionsEnabled:true,ClearSessions:sessions,InstallFlow:flow,AppOrigin:origin})
 on:=New(Options{ClearSessionsEnabled:true,ClearSessions:sessions,InstallFlow:flow,AppOrigin:origin,SelfServeGuideInstallsEnabled:true,Protocols:&protocols.Reader{DB:db.DB}})
 listing:=httptest.NewRecorder()
 on.ServeHTTP(listing,httptest.NewRequest("GET","/api/protocols",nil))
 if listing.Code!=200{t.Fatalf("discovery %d",listing.Code)}
 for _,secret:=range []string{`"id":"draft"`,`"id":"hidden"`} {
  if bytes.Contains(listing.Body.Bytes(),[]byte(secret)){t.Fatalf("nonpublic recipe discoverable: %s",secret)}
 }
 post:=func(h http.Handler,id,digest string,auth bool)int{
  t.Helper()
  payload,_:=json.Marshal(map[string]string{"publicationId":id,"acknowledgedManifestSha256":digest})
  req:=httptest.NewRequest("POST","/api/me/self-serve-guide-installs",bytes.NewReader(payload))
  req.Header.Set("Origin",origin)
  if auth{req.AddCookie(&http.Cookie{Name:"clear_session",Value:b})}
  w:=httptest.NewRecorder();h.ServeHTTP(w,req);return w.Code
 }
 m,err:=flow.Preview(ctx,"pub");if err!=nil{t.Fatal(err)}
 if status:=post(off,"pub",m.ManifestSHA256,true);status!=404{t.Fatalf("flag off: %d",status)}
 if status:=post(on,"pub",m.ManifestSHA256,false);status!=401{t.Fatalf("without user: %d",status)}
 if status:=post(on,"pub",m.ManifestSHA256,true);status!=201{t.Fatalf("install: %d",status)}
 if status:=post(on,"pub",m.ManifestSHA256,true);status!=200{t.Fatalf("idempotent: %d",status)}
 for _,id:=range []string{"draft","hidden"}{
  if status:=post(on,id,strings.Repeat("0",64),true);status!=404{t.Fatalf("%s accessible: %d",id,status)}
 }
 bad,err:=flow.Preview(ctx,"exec");if err!=nil{t.Fatal(err)}
 if status:=post(on,"exec",bad.ManifestSHA256,true);status!=403{t.Fatalf("executable accepted: %d",status)}
 var bCount,aCount int
 if err=db.DB.QueryRowContext(ctx,"SELECT COUNT(*) FROM protocol_installations WHERE user_id='b' AND protocol_key='pub'").Scan(&bCount);err!=nil{t.Fatal(err)}
 if err=db.DB.QueryRowContext(ctx,"SELECT COUNT(*) FROM protocol_installations WHERE user_id='a'").Scan(&aCount);err!=nil{t.Fatal(err)}
 if _,err= (clear.Service{DB:db.DB}).Request(ctx,"b","pub");err!=nil{t.Fatal("guide installation cannot authorize its publication",err)}
 if bCount!=1||aCount!=0{t.Fatalf("leak or duplicate: B=%d A=%d",bCount,aCount)}
}
