package protocols

import(
 "context"
 "errors"
 "testing"

 "github.com/chrisbirster/cleartify/internal/database"
)
func TestPinAndAcknowledgeManifestBeforeApproval(t *testing.T){
 ctx:=context.Background()
 db,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)}
 defer db.Close()
 if err=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 _,err=db.DB.ExecContext(ctx,`INSERT INTO published_protocols(id,creator_did,slug,version,title,summary,license,spec_json,status,created_at)
 VALUES('p1','did:plc:aaaaaaaaaaaaaaaaaaaaaaaa','weekly',3,'Weekly','summary','MIT','{"permissions":["read:tasks"]}','published','now')`)
 if err!=nil{t.Fatal(err)}
 s:=InstallService{Store:db}
 m,err:=s.Preview(ctx,"p1");if err!=nil{t.Fatal(err)}
 if m.Version!=3||len(m.Permissions)!=1||m.Permissions[0]!="read:tasks"||len(m.ManifestSHA256)!=64{t.Fatalf("bad preview %+v",m)}
 did:="did:plc:bbbbbbbbbbbbbbbbbbbbbbbb"
 if _,err=s.Request(ctx,did,"p1","");!errors.Is(err,ErrManifestChanged){t.Fatalf("missing acknowledgement accepted: %v",err)}
 id,err:=s.Request(ctx,did,"p1",m.ManifestSHA256);if err!=nil{t.Fatal(err)}
 if err=s.Decide(ctx,id,true);err!=nil{t.Fatal(err)}
 if err=s.Decide(ctx,id,true);!errors.Is(err,ErrInstallDecision){t.Fatalf("replay approved: %v",err)}
 var version int;var status,digest string
 if err=db.DB.QueryRowContext(ctx,`SELECT protocol_version,status,manifest_sha256 FROM protocol_install_requests WHERE id=?`,id).Scan(&version,&status,&digest);err!=nil{t.Fatal(err)}
 if version!=3||status!="approved"||digest!=m.ManifestSHA256{t.Fatalf("lost pinned version: %d %s %s",version,status,digest)}
 var n int
 if err=db.DB.QueryRowContext(ctx,"SELECT COUNT(*) FROM protocol_installations").Scan(&n);err!=nil||n!=0{t.Fatalf("approved request unexpectedly installed %d %v",n,err)}
}
func TestChangedPublishedSpecCannotReuseApproval(t *testing.T){
 ctx:=context.Background()
 db,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)}
 defer db.Close()
 if err=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 _,err=db.DB.ExecContext(ctx,`INSERT INTO published_protocols(id,creator_did,slug,version,title,summary,license,spec_json,status,created_at)
 VALUES('p1','did:plc:aaaaaaaaaaaaaaaaaaaaaaaa','weekly',1,'Weekly','summary','MIT','{"permissions":[]}','published','now')`)
 if err!=nil{t.Fatal(err)}
 s:=InstallService{Store:db}
 m,err:=s.Preview(ctx,"p1");if err!=nil{t.Fatal(err)}
 id,err:=s.Request(ctx,"did:plc:bbbbbbbbbbbbbbbbbbbbbbbb","p1",m.ManifestSHA256);if err!=nil{t.Fatal(err)}
 _,err=db.DB.ExecContext(ctx,`UPDATE published_protocols SET spec_json='{"permissions":["send:mail"]}' WHERE id='p1'`);if err!=nil{t.Fatal(err)}
 if err=s.Decide(ctx,id,true);!errors.Is(err,ErrInstallDecision){t.Fatalf("modified spec approved: %v",err)}
 if _,err=s.Request(ctx,"did:plc:bbbbbbbbbbbbbbbbbbbbbbbb","p1",m.ManifestSHA256);!errors.Is(err,ErrManifestChanged){t.Fatalf("stale digest accepted: %v",err)}
}
