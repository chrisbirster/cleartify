package protocols
import("context";"errors";"testing";"github.com/chrisbirster/cleartify/internal/database")
func TestManifestStaticEscapeCases(t *testing.T){
 for _,raw:=range []string{
  `{"permissions":["network.send"],"steps":[]}`,
  `{"permissions":["local.read"],"steps":[{"operation":"shell.exec","key":"a"}]}`,
  `{"permissions":["local.read"],"steps":[{"operation":"local.set","key":"a","command":"rm -rf /"}]}`,
  `{"permissions":["local.read"],"steps":[{"operation":"local.set","key":"../secret","value":"x"}]}`,
  `{"permissions":["local.read"],"steps":[{"operation":"env.get","key":"SECRET"}]}`,
  `{"permissions":["local.read"],"steps":[{"operation":"network.fetch","key":"a"}]}`,
  `{"permissions":["local.read"],"steps":[{"operation":"fs.read","key":"a"}]}`,
  `{"permissions":["local.read"],"steps":[],"script":"unsafe"}`,
 }{
  if err:=AnalyzeManifest([]byte(raw));!errors.Is(err,ErrUnsafePermissions){t.Fatalf("accepted unsafe manifest %s: %v",raw,err)}
 }
 if err:=AnalyzeManifest([]byte(`{"permissions":["local.read"],"steps":[{"operation":"local.set","key":"ok","value":"x"}]}`));err!=nil{t.Fatal(err)}
 d:=DiffPermissions([]string{"local.read","notify.preview"},[]string{"local.write.preview","local.read"})
 if !d.Changed||len(d.Added)!=1||d.Added[0]!="local.write.preview"||len(d.Removed)!=1||d.Removed[0]!="notify.preview"{t.Fatalf("diff %+v",d)}
}
func TestPermissionChangeRequiresFreshDigest(t *testing.T){
 ctx:=context.Background();db,e:=database.Open(ctx,database.Config{Path:":memory:"});if e!=nil{t.Fatal(e)};defer db.Close()
 if e=db.Migrate(ctx);e!=nil{t.Fatal(e)}
 for _,q:=range []string{
 `INSERT INTO published_protocols(id,creator_did,slug,version,title,summary,license,spec_json,status,created_at) VALUES('v1','did:plc:creator','demo',1,'Demo','','MIT','{"permissions":["local.read"],"steps":[]}','published','now')`,
 `INSERT INTO published_protocols(id,creator_did,slug,version,title,summary,license,spec_json,status,created_at) VALUES('v2','did:plc:creator','demo',2,'Demo','','MIT','{"permissions":["local.read","notify.preview"],"steps":[]}','published','now')`,
 }{if _,e=db.DB.ExecContext(ctx,q);e!=nil{t.Fatal(e)}}
 service:=InstallService{Store:db,PermissionDiffEnabled:true}
 old,e:=service.Preview(ctx,"v1");if e!=nil{t.Fatal(e)}
 next,e:=service.Preview(ctx,"v2");if e!=nil{t.Fatal(e)}
 if next.PermissionChanges==nil||!next.PermissionChanges.Changed||len(next.PermissionChanges.Added)!=1||next.PermissionChanges.Added[0]!="notify.preview"{t.Fatalf("missing diff %+v",next)}
 if _,e=service.Request(ctx,"did:plc:buyer","v2",old.ManifestSHA256);!errors.Is(e,ErrManifestChanged){t.Fatalf("old approval reused: %v",e)}
 id,e:=service.Request(ctx,"did:plc:buyer","v2",next.ManifestSHA256);if e!=nil{t.Fatal(e)}
 if e=service.Decide(ctx,id,true);e!=nil{t.Fatal(e)}
 service.PermissionDiffEnabled=false
 without,e:=service.Preview(ctx,"v2");if e!=nil||without.PermissionChanges!=nil{t.Fatalf("flag-off diff %+v %v",without,e)}
}
