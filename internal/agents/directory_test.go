package agents

import(
 "context"
 "testing"
 "github.com/chrisbirster/cleartify/internal/database"
)
func TestDirectoryMinimumProjection(t *testing.T){
 ctx:=context.Background()
 db,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)};defer db.Close()
 if err:=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 _,err=db.DB.ExecContext(ctx,`INSERT INTO social_profiles(did,handle,display_name,kind,updated_at) VALUES('did:plc:x','x.test','X','person','2026-10-13T00:00:00Z'),('did:plc:y','y.test','Y','person','2026-10-13T00:00:00Z'),('did:plc:z','z.test','Z','person','2026-10-13T00:00:00Z')`)
 if err!=nil{t.Fatal(err)}
 _,err=db.DB.ExecContext(ctx,`INSERT INTO agent_registrations(id,owner_did,display_name,description,status,created_at) VALUES('unverified','did:plc:x','Unverified','private','unverified','2026-10-13T00:00:00Z'),('verified','did:plc:y','Reviewed','safe summary','verified','2026-10-13T00:00:01Z'),('disabled','did:plc:z','Disabled','hidden','disabled','2026-10-13T00:00:02Z')`)
 if err!=nil{t.Fatal(err)}
 out,err:=Directory(ctx,db.DB,20);if err!=nil{t.Fatal(err)}
 if len(out)!=1||out[0].ID!="verified"{t.Fatalf("unsafe directory: %+v",out)}
 if out[0].OwnerVerification!="not_independently_verified"||len(out[0].Capabilities)!=0{t.Fatalf("unsupported owner/capability claims: %+v",out)}
}
