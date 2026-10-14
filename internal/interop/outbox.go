package interop

import (
 "context"
 "crypto/rand"
 "database/sql"
 "encoding/hex"
 "errors"
 "time"
)

// LocalDelivery has no network address, URL, filesystem or subprocess access.
// Implementations must be local no-op/in-memory handlers only.
type LocalDelivery interface{ Deliver(context.Context,Envelope) error }
type NoopDelivery struct{}
func(NoopDelivery)Deliver(context.Context,Envelope)error{return nil}
var ErrOutboxDisabled=errors.New("local outbox disabled")
var ErrOutboxUnavailable=errors.New("local outbox unavailable")
type Outbox struct{
 DB *sql.DB
 Enabled bool
 Sink LocalDelivery
 Now func()time.Time
 Lease time.Duration
 MaxAttempts int
 Backoff time.Duration
}
func(o Outbox)clock()time.Time{if o.Now!=nil{return o.Now().UTC()};return time.Now().UTC()}
func(o Outbox)config()error{
 if !o.Enabled{return ErrOutboxDisabled}
 if o.DB==nil||o.Sink==nil||o.Lease<=0||o.Lease>time.Minute||o.MaxAttempts<1||o.MaxAttempts>10||o.Backoff<=0||o.Backoff>time.Minute{return ErrOutboxUnavailable}
 return nil
}
const outboxTimeLayout="2006-01-02T15:04:05.000000000Z"
func outboxStamp(t time.Time)string{return t.UTC().Format(outboxTimeLayout)}
var ErrOutboxLegacyTime=errors.New("legacy outbox timestamp requires migration")
func(o Outbox) validateTimeRows(ctx context.Context)error{
 var n int
 err:=o.DB.QueryRowContext(ctx,`SELECT COUNT(*) FROM agent_local_outbox WHERE length(available_at)!=30 OR substr(available_at,20,1)!='.' OR substr(available_at,30,1)!='Z' OR (lease_until IS NOT NULL AND (length(lease_until)!=30 OR substr(lease_until,20,1)!='.' OR substr(lease_until,30,1)!='Z'))`).Scan(&n)
 if err!=nil{return err};if n!=0{return ErrOutboxLegacyTime};return nil
}
func randomLease()(string,error){b:=make([]byte,16);if _,err:=rand.Read(b);err!=nil{return "",err};return hex.EncodeToString(b),nil}
// Enqueue accepts only previously authenticated/signed and authorized local
// transport records. It never inserts an unverified envelope.
func(o Outbox)Enqueue(ctx context.Context,messageID string)error{
 if err:=o.config();err!=nil{return err}
 if messageID==""{return ErrOutboxUnavailable}
 if err:=o.validateTimeRows(ctx);err!=nil{return err}
 now:=outboxStamp(o.clock())
 result,err:=o.DB.ExecContext(ctx,`INSERT INTO agent_local_outbox(message_id,available_at,updated_at)
 SELECT message_id,?,? FROM agent_interop_messages WHERE message_id=? AND status='pending' AND julianday(expires_at)>julianday(?)
 ON CONFLICT(message_id) DO NOTHING`,now,now,messageID,now)
 if err!=nil{return err}
 n,err:=result.RowsAffected();if err!=nil{return err}
 if n==0{
  var existing int
  if err=o.DB.QueryRowContext(ctx,"SELECT COUNT(*) FROM agent_local_outbox WHERE message_id=?",messageID).Scan(&existing);err!=nil{return err}
  if existing!=1{return ErrOutboxUnavailable}
 }
 return nil
}
type claimed struct{envelope Envelope; token string; attempts int}
// Tick claims at most one due message. Conditional UPDATE with a random lease
// is safe across processes sharing the same SQL database; leases bound crash
// recovery but delivery remains at-least-once on lease expiry.
func(o Outbox)Tick(ctx context.Context)(bool,error){
 if err:=o.config();err!=nil{return false,err}
 if err:=o.validateTimeRows(ctx);err!=nil{return false,err}
 now:=o.clock()
 token,err:=randomLease();if err!=nil{return false,err}
 var c claimed
 tx,err:=o.DB.BeginTx(ctx,nil);if err!=nil{return false,err}
 defer tx.Rollback()
 var id string
 var priorAttempts int
 err=tx.QueryRowContext(ctx,`SELECT message_id,attempts FROM agent_local_outbox
 WHERE (state='pending' AND available_at<=?) OR (state='leased' AND lease_until<=?)
 ORDER BY available_at,message_id LIMIT 1`,outboxStamp(now),outboxStamp(now)).Scan(&id,&priorAttempts)
 if errors.Is(err,sql.ErrNoRows){return false,nil}
 if err!=nil{return false,err}
 if priorAttempts>=o.MaxAttempts {
  res,err:=tx.ExecContext(ctx,`UPDATE agent_local_outbox SET state='dead',lease_token=NULL,lease_until=NULL,last_error_key='attempts_exhausted',updated_at=? WHERE message_id=? AND attempts>=? AND ((state='pending' AND available_at<=?) OR (state='leased' AND lease_until<=?))`,outboxStamp(now),id,o.MaxAttempts,outboxStamp(now),outboxStamp(now))
  if err!=nil{return false,err}
  n,err:=res.RowsAffected();if err!=nil{return false,err};if n!=1{return false,nil}
  if err=tx.Commit();err!=nil{return false,err}
  return true,nil
 }
 res,err:=tx.ExecContext(ctx,`UPDATE agent_local_outbox SET state='leased',lease_token=?,lease_until=?,attempts=attempts+1,updated_at=?
 WHERE message_id=? AND attempts<? AND ((state='pending' AND available_at<=?) OR (state='leased' AND lease_until<=?))`,token,outboxStamp(now.Add(o.Lease)),outboxStamp(now),id,o.MaxAttempts,outboxStamp(now),outboxStamp(now))
 if err!=nil{return false,err}
 n,err:=res.RowsAffected();if err!=nil{return false,err};if n!=1{return false,nil}
 c.token=token
 err=tx.QueryRowContext(ctx,`SELECT m.message_id,m.sender_agent_id,m.recipient_agent_id,m.key_id,m.payload_digest,m.signature,m.issued_at,m.expires_at,o.attempts
 FROM agent_local_outbox o JOIN agent_interop_messages m ON m.message_id=o.message_id WHERE o.message_id=?`,id).Scan(
 &c.envelope.MessageID,&c.envelope.SenderAgentID,&c.envelope.RecipientAgentID,&c.envelope.KeyID,&c.envelope.PayloadDigest,&c.envelope.Signature,&c.envelope.IssuedAt,&c.envelope.ExpiresAt,&c.attempts)
 if err!=nil{return false,err}
 if err=tx.Commit();err!=nil{return false,err}
 // Do not deliver expired envelopes or when previously authorized grant,
 // owner binding, key or recipient consent has been revoked.
 var live string
 leaseErr:=o.DB.QueryRowContext(ctx,`SELECT lease_token FROM agent_local_outbox WHERE message_id=? AND state='leased' AND lease_until>?`,id,outboxStamp(o.clock())).Scan(&live)
 if leaseErr!=nil||live!=c.token{return true,ErrOutboxUnavailable}
 deliveryErr:=o.authorized(ctx,c.envelope,o.clock())
 if deliveryErr==nil{deliveryErr=o.Sink.Deliver(ctx,c.envelope)}
 state:="delivered";next:=now;reason:=""
 if deliveryErr!=nil{
  reason="delivery_failed"
  if errors.Is(deliveryErr,ErrTransportDenied){reason="authorization_revoked"}
  if errors.Is(deliveryErr,ErrInvalidEnvelope){reason="expired"}
  if c.attempts>=o.MaxAttempts||reason!="delivery_failed"{state="dead"}else{
   state="pending";delay:=o.Backoff
   for i:=1;i<c.attempts;i++{delay*=2}
   if delay>time.Hour{delay=time.Hour}
   next=now.Add(delay)
  }
 }
 // Fencing token ensures an old worker cannot complete a newly leased job.
 result,err:=o.DB.ExecContext(ctx,`UPDATE agent_local_outbox SET state=?,available_at=?,lease_token=NULL,lease_until=NULL,last_error_key=?,updated_at=? WHERE message_id=? AND state='leased' AND lease_token=? AND lease_until>?`,
 state,outboxStamp(next),reason,outboxStamp(o.clock()),id,c.token,outboxStamp(o.clock()))
 if err!=nil{return true,err}
 n,err=result.RowsAffected();if err!=nil{return true,err}
 if n!=1{return true,ErrOutboxUnavailable}
 return true,deliveryErr
}
func(o Outbox)authorized(ctx context.Context,e Envelope,now time.Time)error{
 deadline,err:=time.Parse(time.RFC3339,e.ExpiresAt)
 if err!=nil||!now.Before(deadline){return ErrInvalidEnvelope}
 var n int
 err=o.DB.QueryRowContext(ctx,`SELECT count(*) FROM agent_interop_keys k
 JOIN agent_owner_bindings sb ON sb.agent_id=k.agent_id AND sb.revoked_at IS NULL
 JOIN agent_registrations sa ON sa.id=sb.agent_id AND sa.status='verified'
 JOIN agent_owner_bindings rb ON rb.agent_id=?
 JOIN agent_registrations ra ON ra.id=rb.agent_id AND ra.status='verified'
 JOIN agent_scoped_grants g ON g.agent_id=k.agent_id AND g.owner_did=sb.owner_did AND g.grantee_agent_id=rb.agent_id AND g.scope='message.preview' AND g.revoked_at IS NULL AND julianday(g.expires_at)>julianday(?)
 JOIN agent_message_consents c ON c.sender_agent_id=k.agent_id AND c.recipient_agent_id=rb.agent_id AND c.revoked_at IS NULL AND julianday(c.expires_at)>julianday(?)
 WHERE k.key_id=? AND k.agent_id=? AND k.algorithm='Ed25519' AND k.revoked_at IS NULL AND rb.revoked_at IS NULL AND c.recipient_owner_did=rb.owner_did`,
 e.RecipientAgentID,now.Format(time.RFC3339Nano),now.Format(time.RFC3339Nano),e.KeyID,e.SenderAgentID).Scan(&n)
 if err!=nil{return err};if n!=1{return ErrTransportDenied}
 return nil
}
