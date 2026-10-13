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
 "github.com/chrisbirster/cleartify/internal/database"
)
func TestOfflineReplayRegistry(t *testing.T){
 ctx:=context.Background()
 db,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)};defer db.Close()
 if err:=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 pub,priv,err:=ed25519.GenerateKey(rand.Reader);if err!=nil{t.Fatal(err)}
 now:=time.Date(2026,10,13,12,0,0,0,time.UTC)
 _,err=db.DB.ExecContext(ctx,`INSERT INTO agent_interop_keys(key_id,agent_id,algorithm,public_key,created_at) VALUES('key-a','agent-a','Ed25519',?,?)`,[]byte(pub),now.Format(time.RFC3339));if err!=nil{t.Fatal(err)}
 r:=ReplayRegistry{Store:db,Now:func()time.Time{return now}}
 makeEnvelope:=func(message string)Envelope{
  e:=Envelope{MessageID:message,SenderAgentID:"agent-a",RecipientAgentID:"agent-b",KeyID:"key-a",PayloadDigest:hex.EncodeToString(make([]byte,32)),IssuedAt:now.Format(time.RFC3339),ExpiresAt:now.Add(time.Minute).Format(time.RFC3339)}
  payload,_:=e.canonical();hash:=sha256.Sum256(payload);e.Signature=ed25519.Sign(priv,hash[:]);return e
 }
 e:=makeEnvelope("message-1")
 if err:=r.VerifyAndRecord(ctx,e);err!=nil{t.Fatal(err)}
 if err:=r.VerifyAndRecord(ctx,e);!errors.Is(err,ErrReplay){t.Fatalf("duplicate accepted: %v",err)}
 var status string
 if err:=db.DB.QueryRowContext(ctx,"SELECT status FROM agent_interop_messages WHERE message_id='message-1'").Scan(&status);err!=nil{t.Fatal(err)}
 if status!="pending"{t.Fatalf("record auto-accepted: %s",status)}
 e2:=makeEnvelope("message-2")
 e2.SenderAgentID="other"
 if err:=r.VerifyAndRecord(ctx,e2);!errors.Is(err,ErrKeyUnavailable){t.Fatalf("wrong agent key accepted: %v",err)}
 e2=makeEnvelope("message-2");e2.RecipientAgentID="tampered"
 if err:=r.VerifyAndRecord(ctx,e2);!errors.Is(err,ErrInvalidEnvelope){t.Fatalf("tamper accepted: %v",err)}
 now=now.Add(2*time.Minute)
 if err:=r.VerifyAndRecord(ctx,makeEnvelope("message-3"));err!=nil{t.Fatalf("new valid at advanced clock: %v",err)}
 if err:=r.VerifyAndRecord(ctx,e2);!errors.Is(err,ErrInvalidEnvelope){t.Fatalf("expired or tampered: %v",err)}
 if _,err=db.DB.ExecContext(ctx,"UPDATE agent_interop_keys SET revoked_at=? WHERE key_id='key-a'",now.Format(time.RFC3339));err!=nil{t.Fatal(err)}
 if err:=r.VerifyAndRecord(ctx,makeEnvelope("message-4"));!errors.Is(err,ErrKeyUnavailable){t.Fatalf("revoked key allowed: %v",err)}
}
