package interop
import(
 "context"
 "crypto/ed25519"
 "crypto/rand"
 "crypto/sha256"
 "encoding/hex"
 "errors"
 "testing"
 "time"
 "github.com/chrisbirster/cleartify/internal/agents"
 "github.com/chrisbirster/cleartify/internal/database"
)
func TestTransactionalLocalTransport(t *testing.T){
 ctx:=context.Background();store,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)};defer store.Close()
 if err=store.Migrate(ctx);err!=nil{t.Fatal(err)}
 now:=time.Date(2026,10,13,12,0,0,0,time.UTC)
 for _,u:=range []string{"alice","bob"}{
  _,err=store.DB.ExecContext(ctx,`INSERT INTO social_profiles(did,handle,display_name,kind,public_visible,claimed,updated_at) VALUES(?,?,?,'person',1,1,'now')`,"did:plc:"+u,u+".test",u);if err!=nil{t.Fatal(err)}
 }
 a,err:=(agents.Service{DB:store.DB}).Register(ctx,"did:plc:alice","A","");if err!=nil{t.Fatal(err)}
 b,err:=(agents.Service{DB:store.DB}).Register(ctx,"did:plc:bob","B","");if err!=nil{t.Fatal(err)}
 grant:=agents.GrantService{DB:store.DB,Now:func()time.Time{return now}}
 aPub,aPriv,_:=ed25519.GenerateKey(rand.Reader);bPub,bPriv,_:=ed25519.GenerateKey(rand.Reader)
 if err=grant.Bind(ctx,"did:plc:alice",a.ID,aPub,ed25519.Sign(aPriv,agents.OwnerProofMessage(a.ID,"did:plc:alice")));err!=nil{t.Fatal(err)}
 if err=grant.Bind(ctx,"did:plc:bob",b.ID,bPub,ed25519.Sign(bPriv,agents.OwnerProofMessage(b.ID,"did:plc:bob")));err!=nil{t.Fatal(err)}
 for _,id:=range []string{a.ID,b.ID}{if _,err=store.DB.ExecContext(ctx,"UPDATE agent_registrations SET status='verified' WHERE id=?",id);err!=nil{t.Fatal(err)}}
 if _,err=store.DB.ExecContext(ctx,`INSERT INTO agent_interop_keys(key_id,agent_id,algorithm,public_key,created_at) VALUES('a-key',?,'Ed25519',?,'now')`,a.ID,[]byte(aPub));err!=nil{t.Fatal(err)}
 env:=func(id string)Envelope{d:=sha256.Sum256([]byte("payload"));e:=Envelope{MessageID:id,SenderAgentID:a.ID,RecipientAgentID:b.ID,KeyID:"a-key",PayloadDigest:hex.EncodeToString(d[:]),IssuedAt:now.Format(time.RFC3339),ExpiresAt:now.Add(time.Minute).Format(time.RFC3339)};raw,_:=e.canonical();hash:=sha256.Sum256(raw);e.Signature=ed25519.Sign(aPriv,hash[:]);return e}
 transport:=NewLocalTransport(false,store.DB,store.Push);transport.now=func()time.Time{return now}
 if err=transport.Send(ctx,env("m1"));!errors.Is(err,ErrTransportDisabled){t.Fatal("flag-off bypass",err)}
 transport.enabled=true
 if err=transport.Send(ctx,env("m1"));!errors.Is(err,ErrTransportDenied){t.Fatal("missing grant",err)}
 if _,err=grant.Grant(ctx,"did:plc:alice",a.ID,b.ID,"message.preview",time.Hour);err!=nil{t.Fatal(err)}
 if err=transport.Send(ctx,env("m1"));!errors.Is(err,ErrTransportDenied){t.Fatal("missing consent",err)}
 svc:=LocalService{Store:store,Now:func()time.Time{return now}}
 if err=svc.Consent(ctx,"did:plc:bob",b.ID,a.ID,time.Hour);err!=nil{t.Fatal(err)}
 if _,err=grant.Grant(ctx,"did:plc:alice",a.ID,b.ID,"message.preview",time.Hour);err!=nil{t.Fatal(err)}
 if err=transport.Send(ctx,env("m1"));err!=nil{t.Fatal(err)}
 if err=transport.Send(ctx,env("m1"));!errors.Is(err,ErrReplay){t.Fatal("replay",err)}
 if _,err=store.DB.ExecContext(ctx,`INSERT INTO social_blocks(actor_did,subject_did,created_at) VALUES('did:plc:bob','did:plc:alice','now')`);err!=nil{t.Fatal(err)}
 if err=transport.Send(ctx,env("blocked-message"));!errors.Is(err,ErrTransportDenied){t.Fatal("owner block bypassed",err)}
 if _,err=store.DB.ExecContext(ctx,`DELETE FROM social_blocks`);err!=nil{t.Fatal(err)}
 if err=transport.RevokeConsent(ctx,"did:plc:alice",b.ID,a.ID);!errors.Is(err,ErrTransportDenied){t.Fatal("wrong owner revoked",err)}
 if err=transport.RevokeConsent(ctx,"did:plc:bob",b.ID,a.ID);err!=nil{t.Fatal(err)}
 if err=transport.Send(ctx,env("m2"));!errors.Is(err,ErrTransportDenied){t.Fatal("revoked consent bypass",err)}
 var n int;if err=store.DB.QueryRowContext(ctx,"SELECT count(*) FROM agent_interop_messages").Scan(&n);err!=nil||n!=1{t.Fatalf("messages %d %v",n,err)}
}
