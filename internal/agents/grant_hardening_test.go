package agents
import("context";"crypto/ed25519";"crypto/rand";"testing";"time";"github.com/chrisbirster/cleartify/internal/database")
func TestGrantExactExpiryAndRecipientKeyRevocation(t *testing.T){
 ctx:=context.Background();d,e:=database.Open(ctx,database.Config{Path:":memory:"});if e!=nil{t.Fatal(e)};defer d.Close()
 if e=d.Migrate(ctx);e!=nil{t.Fatal(e)}
 for _,x:=range []struct{did,handle string}{{"did:plc:owner","owner.test"},{"did:plc:other","other.test"}}{
  _,e=d.DB.ExecContext(ctx,`INSERT INTO social_profiles(did,handle,display_name,kind,public_visible,claimed,updated_at) VALUES(?,?,?,'person',1,1,'now')`,x.did,x.handle,x.handle);if e!=nil{t.Fatal(e)}
 }
 a,e:=(Service{DB:d.DB}).Register(ctx,"did:plc:owner","A","");if e!=nil{t.Fatal(e)}
 b,e:=(Service{DB:d.DB}).Register(ctx,"did:plc:other","B","");if e!=nil{t.Fatal(e)}
 now:=time.Date(2026,10,13,12,0,0,0,time.UTC)
 s:=GrantService{DB:d.DB,Now:func()time.Time{return now}}
 for _,v:=range []struct{id,did string}{{a.ID,"did:plc:owner"},{b.ID,"did:plc:other"}}{
  pub,priv,e:=ed25519.GenerateKey(rand.Reader);if e!=nil{t.Fatal(e)}
  if e=s.Bind(ctx,v.did,v.id,pub,ed25519.Sign(priv,OwnerProofMessage(v.id,v.did)));e!=nil{t.Fatal(e)}
  if _,e=d.DB.ExecContext(ctx,"UPDATE agent_registrations SET status='verified' WHERE id=?",v.id);e!=nil{t.Fatal(e)}
 }
 if _,e=s.Grant(ctx,"did:plc:owner",a.ID,b.ID,"message.preview",0);e==nil{t.Fatal("zero duration")}
 if _,e=s.Grant(ctx,"did:plc:owner",a.ID,b.ID,"message.preview",MaxGrantTTL+time.Nanosecond);e==nil{t.Fatal("over max")}
 _,e=s.Grant(ctx,"did:plc:owner",a.ID,b.ID,"message.preview",time.Second);if e!=nil{t.Fatal(e)}
 now=now.Add(time.Second)
 if yes,e:=s.Allows(ctx,a.ID,b.ID,"message.preview");e!=nil||yes{t.Fatalf("exact expiry accepted %v %v",yes,e)}
 now=now.Add(-time.Second)
 if e=s.Unbind(ctx,"did:plc:other",b.ID);e!=nil{t.Fatal(e)}
 if yes,e:=s.Allows(ctx,a.ID,b.ID,"message.preview");e!=nil||yes{t.Fatalf("recipient binding revoked but grant allowed %v %v",yes,e)}
}
