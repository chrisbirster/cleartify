package api

import (
 "context"
 "encoding/json"
 "net/http"
 "net/http/httptest"
 "testing"
 "github.com/chrisbirster/cleartify/internal/alpha"
 "github.com/chrisbirster/cleartify/internal/clearsession"
 "github.com/chrisbirster/cleartify/internal/database"
 "github.com/chrisbirster/cleartify/internal/protocols"
 "github.com/chrisbirster/cleartify/internal/security"
 "github.com/chrisbirster/cleartify/internal/social"
)

// One isolated local datastore exercises the M1 activated-user / M6 cohort
// gate alongside M3 moderation and M4 publishing/takedown without real PDS.
func TestCrossMilestoneLocalCohortHoldTakedownAndInstall(t *testing.T){
 ctx:=context.Background()
 store,e:=database.Open(ctx,database.Config{Path:":memory:"});if e!=nil{t.Fatal(e)};defer store.Close()
 if e=store.Migrate(ctx);e!=nil{t.Fatal(e)}
 exec:=func(q string,args ...any){t.Helper();if _,e:=store.DB.ExecContext(ctx,q,args...);e!=nil{t.Fatal(e)}}
 exec(`INSERT INTO users(id,display_name,primary_email,status,created_at,updated_at) VALUES('u','User','test@local.example','active','now','now')`)
 exec(`INSERT INTO waitlist_entries(id,name,email,email_normalized,phone_e164,created_at,updated_at) VALUES('w','User','test@local.example','test@local.example','+12125550123','now','now')`)
 exec(`INSERT INTO beta_invites(id,waitlist_entry_id,status,email_token_hash,email_token_expires_at,activated_user_id,activated_at,created_at,updated_at) VALUES('inv','w','activated','fakehash','later','u','now','now','now')`)
 exec(`INSERT INTO social_profiles(did,handle,display_name,kind,public_visible,claimed,updated_at) VALUES('did:plc:author','author.test','Author','person',1,1,'now')`)
 tokenCodec,e:=security.NewTokens("01234567890123456789012345678901");if e!=nil{t.Fatal(e)}
 sessions:=&clearsession.Service{Store:store,Tokens:tokenCodec}
 session,e:=sessions.Issue(ctx,"u");if e!=nil{t.Fatal(e)}
 gate:=alpha.Gate{DB:store.DB}
 if e=gate.Admit(ctx,"u");e!=nil{t.Fatal(e)}
 h:=New(Options{ClearSessionsEnabled:true,PrivateAlphaRoutesEnabled:true,ClearSessions:sessions,ClearPrincipalResolver:ClearSessionPrincipal(sessions),AlphaEligibility:gate,ModerationRuntimeEnabled:true,ModerationBroaderEnabled:true,ModerationEnforcementEnabled:true,ModerationDB:store.DB,Profiles:&social.Reader{DB:store.DB}})
 request:=func(path string,authenticated bool)int{
  t.Helper();req:=httptest.NewRequest(http.MethodGet,path,nil)
  if authenticated{req.AddCookie(&http.Cookie{Name:"clear_session",Value:session})}
  w:=httptest.NewRecorder();h.ServeHTTP(w,req);return w.Code
 }
 if status:=request("/api/me/alpha-eligibility",true);status!=200{t.Fatalf("cohort entry %d",status)}
 if status:=request("/api/me/alpha-eligibility",false);status!=401{t.Fatalf("unauthenticated cohort %d",status)}
 drafts:=protocols.DraftService{DB:store.DB,Tokens:tokenCodec}
 id,e:=drafts.Create(ctx,"did:plc:author",protocols.DraftInput{Slug:"local-demo",Version:1,Title:"Local demo",License:"MIT",Spec:json.RawMessage(`{"steps":[{"operation":"local.set","key":"result","value":"ok"}],"permissions":["local.read"]}`)})
 if e!=nil{t.Fatal(e)}
 for _,visibility:=range []string{"private","unlisted","public"}{
  if e=drafts.PublishDraft(ctx,"did:plc:author",id,visibility,"Local synthetic test provenance");e!=nil{t.Fatalf("publish %s: %v",visibility,e)}
 }
 manifestService:=protocols.InstallService{Store:store,StrictPermissions:true}
 manifest,e:=manifestService.Preview(ctx,id);if e!=nil{t.Fatal(e)}
 if _,e=manifestService.Request(ctx,"did:plc:author",id,"wrong");e==nil{t.Fatal("stale digest accepted")}
 installID,e:=manifestService.Request(ctx,"did:plc:author",id,manifest.ManifestSHA256);if e!=nil{t.Fatal(e)}
 if e=manifestService.Decide(ctx,installID,true);e!=nil{t.Fatal(e)}
 var approved int
 if e=store.DB.QueryRowContext(ctx,"SELECT count(*) FROM protocol_install_requests WHERE id=? AND status='approved'",installID).Scan(&approved);e!=nil||approved!=1{t.Fatalf("install not approved %d %v",approved,e)}
 takedown:=protocols.Takedowns{DB:store.DB}
 if _,e=takedown.Request(ctx,"did:plc:other",id,"rights_review");e==nil{t.Fatal("cross-creator takedown accepted")}
 reqID,e:=takedown.Request(ctx,"did:plc:author",id,"rights_review");if e!=nil{t.Fatal(e)}
 if e=takedown.Decide(ctx,reqID,"hidden");e!=nil{t.Fatal(e)}
 var count int
 if e=store.DB.QueryRowContext(ctx,`SELECT count(*) FROM creator_takedown_requests WHERE id=? AND status='hidden'`,reqID).Scan(&count);e!=nil||count!=1{t.Fatalf("takedown %d %v",count,e)}
 if e=takedown.Decide(ctx,reqID,"restored");e!=nil{t.Fatal(e)}
 if e=gate.Revoke(ctx,"u");e!=nil{t.Fatal(e)}
 if status:=request("/api/me/alpha-eligibility",true);status!=403{t.Fatalf("revoked cohort %d",status)}
 exec(`INSERT INTO social_impersonation_holds(subject_did,reason,created_at) VALUES('did:plc:author','lookalike','now')`)
 s:=&Server{moderationDB:store.DB,moderationBroader:true,moderationEnforcement:true,moderationMetrics:&moderationCounters{}}
 if s.enforceSubjectHold(httptest.NewRecorder(),httptest.NewRequest("POST","/api/me/protocol-drafts",nil),"did:plc:author"){t.Fatal("held creator allowed")}
 exec(`UPDATE social_impersonation_holds SET released_at='2026-10-13T00:00:00Z' WHERE subject_did='did:plc:author'`)
 if !s.enforceSubjectHold(httptest.NewRecorder(),httptest.NewRequest("POST","/api/me/protocol-drafts",nil),"did:plc:author"){t.Fatal("release not respected")}
 if e=sessions.Revoke(ctx,session);e!=nil{t.Fatal(e)}
 if status:=request("/api/me/alpha-eligibility",true);status!=401{t.Fatalf("revoked session %d",status)}
}
