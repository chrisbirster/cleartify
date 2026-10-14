package interop

import (
 "context"
 "crypto/ed25519"
 "crypto/rand"
 "crypto/sha256"
 "encoding/hex"
 "errors"
 "sync"
 "testing"
 "time"

 "github.com/chrisbirster/cleartify/internal/agents"
 "github.com/chrisbirster/cleartify/internal/database"
)

func TestAuthenticatedWorkerLocalDurableDelivery(t *testing.T){
 ctx:=context.Background()
 store,err:=database.Open(ctx,database.Config{Path:t.TempDir()+"/interop.db"})
 if err!=nil{t.Fatal(err)};defer store.Close()
 if err=store.Migrate(ctx);err!=nil{t.Fatal(err)}
 now:=time.Now().UTC().Truncate(time.Second)
 for _,u:=range []string{"alice","bob"}{
  _,err=store.DB.ExecContext(ctx,`INSERT INTO social_profiles(did,handle,display_name,kind,public_visible,claimed,updated_at) VALUES(?,?,?,'person',1,1,'now')`,"did:plc:"+u,u+".test",u)
  if err!=nil{t.Fatal(err)}
 }
 a,err:=(agents.Service{DB:store.DB}).Register(ctx,"did:plc:alice","A","");if err!=nil{t.Fatal(err)}
 b,err:=(agents.Service{DB:store.DB}).Register(ctx,"did:plc:bob","B","");if err!=nil{t.Fatal(err)}
 grant:=agents.GrantService{DB:store.DB,Now:func()time.Time{return now}}
 pub,priv,_:=ed25519.GenerateKey(rand.Reader)
 bpub,bpriv,_:=ed25519.GenerateKey(rand.Reader)
 if err=grant.Bind(ctx,"did:plc:alice",a.ID,pub,ed25519.Sign(priv,agents.OwnerProofMessage(a.ID,"did:plc:alice")));err!=nil{t.Fatal(err)}
 if err=grant.Bind(ctx,"did:plc:bob",b.ID,bpub,ed25519.Sign(bpriv,agents.OwnerProofMessage(b.ID,"did:plc:bob")));err!=nil{t.Fatal(err)}
 for _,id:=range []string{a.ID,b.ID}{if _,err=store.DB.ExecContext(ctx,"UPDATE agent_registrations SET status='verified' WHERE id=?",id);err!=nil{t.Fatal(err)}}
 if _,err=store.DB.ExecContext(ctx,`INSERT INTO agent_interop_keys(key_id,agent_id,algorithm,public_key,created_at) VALUES('key',?,'Ed25519',?,'now')`,a.ID,[]byte(pub));err!=nil{t.Fatal(err)}
 if _,err=grant.Grant(ctx,"did:plc:alice",a.ID,b.ID,"message.preview",time.Hour);err!=nil{t.Fatal(err)}
 local:=LocalService{Store:store,Now:func()time.Time{return now}}
 if err=local.Consent(ctx,"did:plc:bob",b.ID,a.ID,time.Hour);err!=nil{t.Fatal(err)}
 transport:=NewLocalTransport(true,store.DB,store.Push);transport.now=func()time.Time{return now}
 inbox:=DurableInbox{DB:store.DB}
 outbox:=Outbox{DB:store.DB,Enabled:true,Sink:inbox,Now:func()time.Time{return now},Lease:10*time.Second,Backoff:time.Second,MaxAttempts:3}
 worker:=AuthenticatedWorker{Transport:transport,Outbox:outbox,DB:store.DB}
 caller:=VerifiedLocalCaller{OwnerDID:"did:plc:alice",AgentID:a.ID,Verified:true}
 envelope:=func(id string)Envelope{
  dig:=sha256.Sum256([]byte("test-payload"))
  e:=Envelope{MessageID:id,SenderAgentID:a.ID,RecipientAgentID:b.ID,KeyID:"key",PayloadDigest:hex.EncodeToString(dig[:]),IssuedAt:now.Format(time.RFC3339),ExpiresAt:now.Add(time.Minute).Format(time.RFC3339)}
  raw,_:=e.canonical();hash:=sha256.Sum256(raw);e.Signature=ed25519.Sign(priv,hash[:]);return e
 }
 if err=worker.Submit(ctx,caller,envelope("one"));!errors.Is(err,ErrRuntimeDisabled){t.Fatalf("default enabled: %v",err)}
 worker.Enabled=true
 if err=worker.Submit(ctx,VerifiedLocalCaller{OwnerDID:caller.OwnerDID,AgentID:caller.AgentID},envelope("one"));!errors.Is(err,ErrRuntimeIdentity){t.Fatalf("unverified caller: %v",err)}
 if err=worker.Submit(ctx,VerifiedLocalCaller{OwnerDID:"did:plc:bob",AgentID:a.ID,Verified:true},envelope("one"));!errors.Is(err,ErrRuntimeIdentity){t.Fatalf("wrong owner: %v",err)}
 if err=worker.Submit(ctx,caller,envelope("one"));err!=nil{t.Fatal(err)}
 if err=worker.Submit(ctx,caller,envelope("one"));!errors.Is(err,ErrReplay){t.Fatal("replay",err)}
 var wg sync.WaitGroup
 errs:=make(chan error,4)
 successes:=make(chan bool,4)
 for i:=0;i<4;i++{
  wg.Add(1);go func(){defer wg.Done();done,e:=worker.Tick(ctx);if e!=nil{errs<-e};if done&&e==nil{successes<-true}}()
 }
 wg.Wait();close(errs);close(successes)
 // Snapshot-contention errors must be surfaced, not treated as successful delivery.
 for e:=range errs {
  if e!=nil && e.Error()!="turso: error: database snapshot is stale, rollback and retry the transaction"{t.Fatalf("unexpected worker error: %v",e)}
 }
 if len(successes)!=1{t.Fatalf("completed deliveries %d",len(successes))}
 count:=func()int{t.Helper();var n int;if e:=store.DB.QueryRowContext(ctx,"SELECT COUNT(*) FROM agent_local_inbox").Scan(&n);e!=nil{t.Fatal(e)};return n}
 if count()!=1{t.Fatal("duplicate sink delivery")}
 // Crash after the durable sink commit but before acknowledgement. Restart
 // reclaims an expired lease and observes the same inbox row (no duplicate).
 if err=worker.Submit(ctx,caller,envelope("crash"));err!=nil{t.Fatal(err)}
 if err=inbox.Deliver(ctx,envelope("crash"));err!=nil{t.Fatal(err)}
 _,err=store.DB.ExecContext(ctx,`UPDATE agent_local_outbox SET state='leased',lease_token='stale',lease_until=? WHERE message_id='crash'`,now.Add(-time.Second).Format(time.RFC3339Nano))
 if err!=nil{t.Fatal(err)}
 if done,e:=worker.Tick(ctx);!done||e!=nil{t.Fatalf("restart %v %v",done,e)}
 if count()!=2{t.Fatalf("non-idempotent crash recovery %d",count())}
 res,err:=store.DB.ExecContext(ctx,`UPDATE agent_local_outbox SET state='delivered' WHERE message_id='crash' AND state='leased' AND lease_token='stale'`)
 if err!=nil{t.Fatal(err)};n,err:=res.RowsAffected();if err!=nil||n!=0{t.Fatal("stale ack accepted",n,err)}
 // Revocation after enqueue but before delivery prevents side effects.
 if err=worker.Submit(ctx,caller,envelope("revoked"));err!=nil{t.Fatal(err)}
 if err=local.RevokeConsent(ctx,"did:plc:bob",b.ID,a.ID);err!=nil{t.Fatal(err)}
 if done,e:=worker.Tick(ctx);!done||!errors.Is(e,ErrTransportDenied){t.Fatalf("revoked delivery %v %v",done,e)}
 if count()!=2{t.Fatalf("revoked message reached sink: %d",count())}
 if err=worker.Submit(ctx,caller,envelope("blocked"));!errors.Is(err,ErrTransportDenied){t.Fatalf("revoked enqueue accepted: %v",err)}
}
