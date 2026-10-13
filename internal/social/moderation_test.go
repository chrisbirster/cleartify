package social

import (
 "context"
 "testing"
 "github.com/chrisbirster/cleartify/internal/database"
 "github.com/chrisbirster/cleartify/internal/security"
)
func TestModerationOwnershipAndValidation(t *testing.T){
 ctx:=context.Background()
 db,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)};defer db.Close()
 if err=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 _,err=db.DB.ExecContext(ctx,`INSERT INTO social_profiles(did,handle,display_name,kind,updated_at) VALUES('did:plc:subject','subject.test','Subject','person','2026-10-13T12:00:00Z')`);if err!=nil{t.Fatal(err)}
 tokens,err:=security.NewTokens("01234567890123456789012345678901");if err!=nil{t.Fatal(err)}
 svc:=&Moderation{DB:db.DB,Tokens:tokens}
 if err:=svc.Block(ctx,"did:plc:actor","did:plc:subject",true);err!=nil{t.Fatal(err)}
 if err:=svc.Block(ctx,"did:plc:actor","did:plc:subject",true);err!=nil{t.Fatal(err)}
 var count int
 if err:=db.DB.QueryRowContext(ctx,"SELECT COUNT(*) FROM social_blocks WHERE actor_did='did:plc:actor'").Scan(&count);err!=nil{t.Fatal(err)}
 if count!=1{t.Fatalf("duplicate block %d",count)}
 if err:=svc.Block(ctx,"did:plc:someone","did:plc:subject",false);err!=nil{t.Fatal(err)}
 if err:=db.DB.QueryRowContext(ctx,"SELECT COUNT(*) FROM social_blocks WHERE actor_did='did:plc:actor'").Scan(&count);err!=nil{t.Fatal(err)}
 if count!=1{t.Fatal("foreign block deleted")}
 if err:=svc.Block(ctx,"did:plc:actor","did:plc:subject",false);err!=nil{t.Fatal(err)}
 if err:=db.DB.QueryRowContext(ctx,"SELECT COUNT(*) FROM social_blocks").Scan(&count);err!=nil{t.Fatal(err)}
 if count!=0{t.Fatal("not unblocked")}
 if _,err:=svc.Report(ctx,"did:plc:actor","did:plc:actor","impersonation");err==nil{t.Fatal("self-report")}
 if _,err:=svc.Report(ctx,"did:plc:actor","did:plc:subject","bad");err==nil{t.Fatal("unrecognized reason")}
 id,err:=svc.Report(ctx,"did:plc:actor","did:plc:subject","impersonation");if err!=nil||id==""{t.Fatalf("report %q %v",id,err)}
 if err:=db.DB.QueryRowContext(ctx,"SELECT COUNT(*) FROM social_reports WHERE reporter_did='did:plc:actor' AND reason='impersonation'").Scan(&count);err!=nil{t.Fatal(err)}
 if count!=1{t.Fatal("missing report")}
}
