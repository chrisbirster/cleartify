package agents

import(
 "context"
 "crypto/ed25519"
 "crypto/rand"
 "errors"
 "testing"
 "time"
 "github.com/chrisbirster/cleartify/internal/database"
)
func TestAgentOwnerBindingAndScopedRevocableGrants(t *testing.T){
 ctx:=context.Background()
 db,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)};defer db.Close()
 if err=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 owner:="did:plc:alice";other:="did:plc:bob"
 for _,a:=range []struct{did,handle string}{{owner,"alice.test"},{other,"bob.test"}}{
  if _,err=db.DB.ExecContext(ctx,`INSERT INTO social_profiles(did,handle,display_name,kind,public_visible,claimed,updated_at) VALUES(?,?,?,'person',1,1,'now')`,a.did,a.handle,a.handle);err!=nil{t.Fatal(err)}
 }
 a,err:=Service{DB:db.DB}.Register(ctx,owner,"A","");if err!=nil{t.Fatal(err)}
 b,err:=Service{DB:db.DB}.Register(ctx,other,"B","");if err!=nil{t.Fatal(err)}
 clock:=time.Date(2026,10,13,12,0,0,0,time.UTC)
 s:=GrantService{DB:db.DB,Now:func()time.Time{return clock}}
 pub,priv,_:=ed25519.GenerateKey(rand.Reader)
 sign:=func(id,did string)[]byte{return ed25519.Sign(priv,OwnerProofMessage(id,did))}
 if err:=s.Bind(ctx,other,a.ID,pub,sign(a.ID,other));!errors.Is(err,ErrOwnership){t.Fatal("cross-owner bind",err)}
 if err:=s.Bind(ctx,owner,a.ID,pub,sign(a.ID,owner));err!=nil{t.Fatal(err)}
 if err:=s.Bind(ctx,owner,a.ID,pub,sign(a.ID,owner));!errors.Is(err,ErrOwnership){t.Fatal("duplicate binding",err)}
 pubB,privB,_:=ed25519.GenerateKey(rand.Reader)
 if err:=s.Bind(ctx,other,b.ID,pubB,ed25519.Sign(privB,OwnerProofMessage(b.ID,other)));err!=nil{t.Fatal(err)}
 if _,err:=s.Grant(ctx,owner,a.ID,b.ID,"network.send",time.Hour);!errors.Is(err,ErrGrant){t.Fatal(err)}
 if _,err:=s.Grant(ctx,other,a.ID,b.ID,"message.preview",time.Hour);!errors.Is(err,ErrGrant){t.Fatal(err)}
 if _,err:=s.Grant(ctx,owner,a.ID,b.ID,"message.preview",25*time.Hour);!errors.Is(err,ErrGrant){t.Fatal(err)}
 for _,id:=range []string{a.ID,b.ID}{if _,err=db.DB.ExecContext(ctx,`UPDATE agent_registrations SET status='verified' WHERE id=?`,id);err!=nil{t.Fatal(err)}}
 id,err:=s.Grant(ctx,owner,a.ID,b.ID,"message.preview",time.Hour);if err!=nil{t.Fatal(err)}
 allowed,err:=s.Allows(ctx,a.ID,b.ID,"message.preview");if err!=nil||!allowed{t.Fatalf("%v %v",allowed,err)}
 allowed,_=s.Allows(ctx,a.ID,b.ID,"protocol.preview");if allowed{t.Fatal("scope escalated")}
 if err:=s.Revoke(ctx,other,id);!errors.Is(err,ErrGrant){t.Fatal("other owner revoke",err)}
 if err:=s.Revoke(ctx,owner,id);err!=nil{t.Fatal(err)}
 allowed,_=s.Allows(ctx,a.ID,b.ID,"message.preview");if allowed{t.Fatal("revoked grant allowed")}
 id,err=s.Grant(ctx,owner,a.ID,b.ID,"message.preview",time.Hour);if err!=nil{t.Fatal(err)}
 clock=clock.Add(2*time.Hour)
 allowed,_=s.Allows(ctx,a.ID,b.ID,"message.preview");if allowed{t.Fatal("expired grant allowed")}
 clock=clock.Add(-2*time.Hour)
 if err:=s.Unbind(ctx,owner,a.ID);err!=nil{t.Fatal(err)}
 allowed,_=s.Allows(ctx,a.ID,b.ID,"message.preview");if allowed{t.Fatal("unbound owner allowed")}
 _=id
}
