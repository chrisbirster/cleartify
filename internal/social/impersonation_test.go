package social
import(
 "context"
 "testing"
 "github.com/chrisbirster/cleartify/internal/database"
)
func TestReservedAndLookalikeHandles(t *testing.T){
 for _,h:=range []string{"admin","support.example","cleartify","supp0rt.example","0fficial","secur1ty"}{
  if _,ok:=ScreenHandle(h);ok{t.Fatalf("unsafe handle allowed: %s",h)}
 }
 for _,h:=range []string{"reader.example","chris","normal-person.test"}{
  if reason,ok:=ScreenHandle(h);!ok{t.Fatalf("safe handle rejected %q: %q",h,reason)}
 }
}
func TestFlaggedImpersonationReportHoldsAndLabels(t *testing.T){
 ctx:=context.Background()
 db,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)};defer db.Close()
 if err=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 subject:="did:plc:subject"
 _,err=db.DB.ExecContext(ctx,`INSERT INTO social_profiles(did,handle,display_name,kind,public_visible,claimed,updated_at) VALUES(?,'reader.example','Reader','person',1,1,'now')`,subject)
 if err!=nil{t.Fatal(err)}
 off:=Impersonation{DB:db.DB}
 label,err:=off.Label(ctx,subject);if err!=nil||label.Verification!="unverified"{t.Fatal(label,err)}
 if _,err:=off.EvaluateRecords(ctx,subject);err!=ErrImpersonationReview{t.Fatal(err)}
 on:=Impersonation{DB:db.DB,Enabled:true}
 label,err=on.Label(ctx,subject);if err!=nil||label.Verification!="locally_claimed_not_independently_verified"{t.Fatal(label,err)}
 _,err=db.DB.ExecContext(ctx,`INSERT INTO social_reports(id,reporter_did,subject_did,reason,created_at) VALUES('r1','did:plc:reporter1',?,'impersonation','now'),('r2','did:plc:reporter1',?,'impersonation','now')`,subject,subject)
 if err!=nil{t.Fatal(err)}
 label,err=on.EvaluateRecords(ctx,subject);if err!=nil||label.OnHold{t.Fatal("single unique reporter should not trigger",label,err)}
 _,err=db.DB.ExecContext(ctx,`INSERT INTO social_reports(id,reporter_did,subject_did,reason,created_at) VALUES('r3','did:plc:reporter2',?,'impersonation','now')`,subject)
 if err!=nil{t.Fatal(err)}
 label,err=on.EvaluateRecords(ctx,subject);if err!=nil||!label.OnHold||label.HoldReason!="reports"||label.Verification!="unverified"{t.Fatal(label,err)}
 if err=on.ReleaseHold(ctx,subject);err!=nil{t.Fatal(err)}
 label,err=on.Label(ctx,subject);if err!=nil||label.OnHold{t.Fatal(label,err)}
}
