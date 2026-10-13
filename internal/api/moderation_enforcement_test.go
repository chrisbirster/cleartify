package api
import(
 "context"
 "net/http/httptest"
 "testing"
 "github.com/chrisbirster/cleartify/internal/database"
)
func TestModerationHoldEnforcement(t *testing.T){
 ctx:=context.Background();db,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)};defer db.Close()
 if err=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 s:=&Server{moderationDB:db.DB,moderationEnforcement:true,moderationMetrics:&moderationCounters{}}
 req:=httptest.NewRequest("POST","/test",nil)
 if !s.enforceSubjectHold(httptest.NewRecorder(),req,"did:plc:held"){t.Fatal("unheld denied")}
 _,err=db.DB.ExecContext(ctx,`INSERT INTO social_profiles(did,handle,display_name,kind,public_visible,claimed,updated_at) VALUES('did:plc:held','held.test','Held','person',1,1,'now')`);if err!=nil{t.Fatal(err)}
 _,err=db.DB.ExecContext(ctx,`INSERT INTO social_impersonation_holds(subject_did,reason,created_at) VALUES('did:plc:held','lookalike','now')`);if err!=nil{t.Fatal(err)}
 w:=httptest.NewRecorder();if s.enforceSubjectHold(w,req,"did:plc:held")||w.Code!=403{t.Fatalf("held allowed: %d",w.Code)}
 _,err=db.DB.ExecContext(ctx,`UPDATE social_impersonation_holds SET released_at='now' WHERE subject_did='did:plc:held'`);if err!=nil{t.Fatal(err)}
 if !s.enforceSubjectHold(httptest.NewRecorder(),req,"did:plc:held"){t.Fatal("release ignored")}
 if got:=s.moderationMetrics.snapshot()["hold_denied"];got!=1{t.Fatalf("denied counter %d",got)}
 s.moderationEnforcement=false;if !s.enforceSubjectHold(httptest.NewRecorder(),req,"did:plc:held"){t.Fatal("off blocked")}
 s.moderationEnforcement=true;s.moderationDB=nil;w=httptest.NewRecorder();if s.enforceSubjectHold(w,req,"did:plc:held")||w.Code!=503{t.Fatal("DB failure not closed")}
}
func TestLocalMetricsRouteDisabled(t *testing.T){
 w:=httptest.NewRecorder();New(Options{}).ServeHTTP(w,httptest.NewRequest("GET","/api/admin/moderation/local-metrics",nil));if w.Code!=404{t.Fatalf("route present while off: %d",w.Code)}
}
