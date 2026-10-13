package api
import(
 "context"
 "net/http/httptest"
 "testing"
 "github.com/chrisbirster/cleartify/internal/database"
)
func TestBroaderHoldGateAndReversibleRelease(t *testing.T){
 ctx:=context.Background();db,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)};defer db.Close()
 if err=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 _,err=db.DB.ExecContext(ctx,`INSERT INTO social_profiles(did,handle,display_name,kind,public_visible,claimed,updated_at) VALUES('did:plc:held','held.test','Held','person',1,1,'now')`);if err!=nil{t.Fatal(err)}
 _,err=db.DB.ExecContext(ctx,`INSERT INTO social_impersonation_holds(subject_did,reason,created_at) VALUES('did:plc:held','lookalike','now')`);if err!=nil{t.Fatal(err)}
 s:=&Server{moderationDB:db.DB,moderationEnforcement:true,moderationBroader:true,moderationMetrics:&moderationCounters{}}
 req:=httptest.NewRequest("POST","/api/me/protocol-drafts",nil)
 w:=httptest.NewRecorder()
 if s.enforceSubjectHold(w,req,"did:plc:held")||w.Code!=403{t.Fatalf("active hold not denied %d",w.Code)}
 _,err=db.DB.ExecContext(ctx,`UPDATE social_impersonation_holds SET released_at='2026-10-13T00:00:00Z' WHERE subject_did='did:plc:held'`);if err!=nil{t.Fatal(err)}
 w=httptest.NewRecorder()
 if !s.enforceSubjectHold(w,req,"did:plc:held"){t.Fatalf("released hold still denied %d",w.Code)}
 if s.moderationMetrics.snapshot()["hold_denied"]!=1{t.Fatal("missing denial metrics")}
 s.moderationEnforcement=false;s.moderationBroader=false
 if !s.enforceSubjectHold(httptest.NewRecorder(),req,"did:plc:held"){t.Fatal("flag-off blocked")}
}
func TestModerationSelfRoutesOff(t *testing.T){
 h:=New(Options{})
 for _,p:=range []string{"/api/me/moderation/hold","/api/me/moderation/appeal"}{
  w:=httptest.NewRecorder();h.ServeHTTP(w,httptest.NewRequest("GET",p,nil));if w.Code!=404{t.Fatalf("%s returned %d",p,w.Code)}
 }
}
