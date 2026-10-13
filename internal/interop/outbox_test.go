package interop

import (
 "context"
 "errors"
 "sync"
 "testing"
 "time"
 "github.com/chrisbirster/cleartify/internal/database"
)
type countSink struct{mu sync.Mutex;n int}
func(c *countSink)Deliver(context.Context,Envelope)error{c.mu.Lock();defer c.mu.Unlock();c.n++;return nil}
func TestOutboxFlagOffEnqueueAndDeadLetter(t *testing.T){
 ctx:=context.Background();db,e:=database.Open(ctx,database.Config{Path:":memory:"});if e!=nil{t.Fatal(e)};defer db.Close()
 if e=db.Migrate(ctx);e!=nil{t.Fatal(e)}
 now:=time.Now().UTC().Truncate(time.Second)
 sink:=&countSink{}
 o:=Outbox{DB:db.DB,Sink:sink,Now:func()time.Time{return now},Lease:10*time.Second,Backoff:time.Second,MaxAttempts:2}
 if e=o.Enqueue(ctx,"unknown");!errors.Is(e,ErrOutboxDisabled){t.Fatalf("disabled %v",e)}
 o.Enabled=true
 if e=o.Enqueue(ctx,"unknown");!errors.Is(e,ErrOutboxUnavailable){t.Fatalf("unverified enqueue %v",e)}
 _,e=db.DB.ExecContext(ctx,`INSERT INTO agent_interop_keys(key_id,agent_id,algorithm,public_key,created_at) VALUES('key','agent-a','Ed25519',?,'now')`,make([]byte,32));if e!=nil{t.Fatal(e)}
 _,e=db.DB.ExecContext(ctx,`INSERT INTO agent_interop_messages(message_id,sender_agent_id,recipient_agent_id,key_id,payload_digest,signature,issued_at,expires_at,status) VALUES('m1','agent-a','agent-b','key',?,? ,?,?,'pending')`, "0000000000000000000000000000000000000000000000000000000000000000",[]byte("fake"),now.Format(time.RFC3339),now.Add(time.Minute).Format(time.RFC3339));if e!=nil{t.Fatal(e)}
 if e=o.Enqueue(ctx,"m1");e!=nil{t.Fatal(e)}
 if e=o.Enqueue(ctx,"m1");e!=nil{t.Fatalf("duplicate enqueue %v",e)}
 done,e:=o.Tick(ctx);if !done||!errors.Is(e,ErrTransportDenied){t.Fatalf("revoked ownership should dead letter %v %v",done,e)}
 var state string;var tries int
 e=db.DB.QueryRowContext(ctx,"SELECT state,attempts FROM agent_local_outbox WHERE message_id='m1'").Scan(&state,&tries)
 if e!=nil||state!="dead"||tries!=1{t.Fatalf("dead-letter %s %d %v",state,tries,e)}
 if done,e=o.Tick(ctx);done||e!=nil{t.Fatalf("dead letter replay %v %v",done,e)}
 if sink.n!=0{t.Fatal("unauthorized message delivered")}
}
// Simulates another process taking over an expired lease. The fencing token
// must prevent the previous process from changing the new worker's outcome.
func TestOutboxConcurrentClaimLeaseAndFencing(t *testing.T){
 ctx:=context.Background();db,e:=database.Open(ctx,database.Config{Path:":memory:"});if e!=nil{t.Fatal(e)};defer db.Close()
 if e=db.Migrate(ctx);e!=nil{t.Fatal(e)}
 now:=time.Now().UTC().Truncate(time.Second)
 _,e=db.DB.ExecContext(ctx,`INSERT INTO agent_interop_keys(key_id,agent_id,algorithm,public_key,created_at) VALUES('key','a','Ed25519',?,'now')`,make([]byte,32));if e!=nil{t.Fatal(e)}
 _,e=db.DB.ExecContext(ctx,`INSERT INTO agent_interop_messages(message_id,sender_agent_id,recipient_agent_id,key_id,payload_digest,signature,issued_at,expires_at,status) VALUES('m1','a','b','key',?,?,?,?,'pending')`,"0000000000000000000000000000000000000000000000000000000000000000",[]byte("fake"),now.Format(time.RFC3339),now.Add(time.Minute).Format(time.RFC3339));if e!=nil{t.Fatal(e)}
 sink:=&countSink{}
 makeWorker:=func()Outbox{return Outbox{DB:db.DB,Enabled:true,Sink:sink,Now:func()time.Time{return now},Lease:time.Second,Backoff:time.Second,MaxAttempts:2}}
 a,b:=makeWorker(),makeWorker()
 if e=a.Enqueue(ctx,"m1");e!=nil{t.Fatal(e)}
 // Lease held by another worker is never claimable until expiration.
 _,e=db.DB.ExecContext(ctx,`UPDATE agent_local_outbox SET state='leased',lease_token='older',lease_until=? WHERE message_id='m1'`,now.Add(time.Second).Format(time.RFC3339Nano));if e!=nil{t.Fatal(e)}
 if claimed,e:=b.Tick(ctx);claimed||e!=nil{t.Fatalf("claimed live lease %v %v",claimed,e)}
 now=now.Add(2*time.Second)
 if claimed,e:=b.Tick(ctx);!claimed||!errors.Is(e,ErrTransportDenied){t.Fatalf("lease takeover %v %v",claimed,e)}
 res,e:=db.DB.ExecContext(ctx,`UPDATE agent_local_outbox SET state='delivered' WHERE message_id='m1' AND state='leased' AND lease_token='older'`);if e!=nil{t.Fatal(e)}
 n,e:=res.RowsAffected();if e!=nil||n!=0{t.Fatalf("stale lease completed %d %v",n,e)}
 if sink.n!=0{t.Fatal("unauthorized delivery")}
}
