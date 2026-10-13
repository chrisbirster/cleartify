package api

import(
 "context"
 "encoding/json"
 "net/http/httptest"
 "strings"
 "testing"
 "github.com/chrisbirster/cleartify/internal/database"
 "github.com/chrisbirster/cleartify/internal/social"
)
func TestModerationRuntimeReviewAndAppeal(t *testing.T){
 ctx:=context.Background()
 db,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)};defer db.Close()
 if err=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 _,err=db.DB.ExecContext(ctx,`INSERT INTO social_profiles(did,handle,display_name,kind,public_visible,updated_at) VALUES('did:plc:alice','admin.test','Test','person',1,'now')`)
 if err!=nil{t.Fatal(err)}
 options:=Options{AppOrigin:"http://localhost:3000",AdminToken:"test-admin-secret",ModerationRuntimeEnabled:true,ModerationDB:db.DB,Profiles:&social.Reader{DB:db.DB}}
 h:=New(options)
 request:=func(method,path,body,token,origin string)(int,string){
  t.Helper();req:=httptest.NewRequest(method,path,strings.NewReader(body));req.Header.Set("Content-Type","application/json")
  if token!=""{req.Header.Set("Authorization","Bearer "+token)}
  if origin!=""{req.Header.Set("Origin",origin)}
  w:=httptest.NewRecorder();h.ServeHTTP(w,req);return w.Code,w.Body.String()
 }
 if code,_:=request("POST","/api/admin/moderation/subjects/did:plc:alice/evaluate","", "",options.AppOrigin);code!=401{t.Fatalf("admin bypass: %d",code)}
 if code,_:=request("POST","/api/admin/moderation/subjects/did:plc:alice/evaluate","",options.AdminToken,"https://attacker.test");code!=403{t.Fatalf("CSRF bypass: %d",code)}
 if code,body:=request("POST","/api/admin/moderation/subjects/did:plc:alice/evaluate","",options.AdminToken,options.AppOrigin);code!=200{t.Fatalf("evaluate %d: %s",code,body)}
 if code,_:=request("GET","/api/profiles/admin.test","", "", "");code!=404{t.Fatalf("held profile exposed: %d",code)}
 if code,_:=request("POST","/api/admin/moderation/subjects/did:plc:alice/appeal",`{"reason":"The handle was assigned to me"}`,options.AdminToken,options.AppOrigin);code!=202{t.Fatalf("appeal: %d",code)}
 code,body:=request("GET","/api/admin/moderation/cases","",options.AdminToken,"")
 if code!=200{t.Fatalf("case list %d: %s",code,body)}
 var result struct{Cases []struct{ID,Kind,Status string}}
 if err=json.Unmarshal([]byte(body),&result);err!=nil{t.Fatal(err)}
 if len(result.Cases)!=2{t.Fatalf("cases: %+v",result)}
 var reviewID string
 for _,cs:=range result.Cases{if cs.Kind=="review"{reviewID=cs.ID}}
 if reviewID==""{t.Fatal("review missing")}
 if code,body=request("POST","/api/admin/moderation/cases/"+reviewID+"/decision",`{"decision":"released","reason":"manual review"}`,options.AdminToken,options.AppOrigin);code!=200{t.Fatalf("decision %d: %s",code,body)}
 if code,_:=request("GET","/api/profiles/admin.test","","","");code!=200{t.Fatalf("released still hidden: %d",code)}
 if code,_:=request("POST","/api/admin/moderation/cases/"+reviewID+"/decision",`{"decision":"released","reason":"again again"}`,options.AdminToken,options.AppOrigin);code!=409{t.Fatalf("replay: %d",code)}
 var audits int
 if err=db.DB.QueryRowContext(ctx,"SELECT count(*) FROM social_moderation_audit").Scan(&audits);err!=nil{t.Fatal(err)}
 if audits!=3{t.Fatalf("unexpected audit count %d",audits)}
 off:=New(Options{AdminToken:options.AdminToken,Profiles:options.Profiles,ModerationDB:db.DB})
 w:=httptest.NewRecorder();off.ServeHTTP(w,httptest.NewRequest("GET","/api/profiles/admin.test",nil))
 if w.Code!=200{t.Fatalf("flag-off public changed: %d",w.Code)}
 if code,_:=request("GET","/api/admin/moderation/cases","","","");code!=401{t.Fatalf("list unauthorized: %d",code)}
}
