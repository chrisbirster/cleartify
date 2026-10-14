package interop

import (
 "context"
 "database/sql"
 "errors"
 "crypto/ed25519"
 "time"
)

var ErrRuntimeDisabled = errors.New("authenticated local agent worker disabled")
var ErrRuntimeIdentity = errors.New("local caller identity not verified")

// VerifiedLocalCaller is supplied by trusted in-process authentication code,
// never constructed from a request body, header, DID claim or network peer.
type VerifiedLocalCaller struct {
 OwnerDID string
 AgentID string
 Verified bool
}

// DurableInbox records a delivery exactly once in the SAME authoritative SQL
// database as the outbox. A commit is the local sink side effect: no external
// callback, network transmission or arbitrary code execution is permitted.
type DurableInbox struct{DB *sql.DB}
func(s DurableInbox) Deliver(ctx context.Context,e Envelope)error{
 if s.DB==nil{return ErrOutboxUnavailable}
 tx,err:=s.DB.BeginTx(ctx,nil);if err!=nil{return err};defer tx.Rollback()
 now:=time.Now().UTC().Format(time.RFC3339Nano)
 _,err=tx.ExecContext(ctx,`INSERT INTO agent_local_inbox(message_id,sender_agent_id,recipient_agent_id,payload_digest,received_at)
 SELECT m.message_id,m.sender_agent_id,m.recipient_agent_id,m.payload_digest,?
 FROM agent_interop_messages m WHERE m.message_id=? AND m.sender_agent_id=? AND m.recipient_agent_id=? AND m.payload_digest=? AND m.status='pending' AND EXISTS (SELECT 1 FROM agent_interop_keys k JOIN agent_owner_bindings sb ON sb.agent_id=k.agent_id AND sb.revoked_at IS NULL JOIN agent_registrations sa ON sa.id=sb.agent_id AND sa.status='verified' JOIN agent_owner_bindings rb ON rb.agent_id=m.recipient_agent_id AND rb.revoked_at IS NULL JOIN agent_registrations ra ON ra.id=rb.agent_id AND ra.status='verified' JOIN agent_scoped_grants g ON g.agent_id=m.sender_agent_id AND g.grantee_agent_id=m.recipient_agent_id AND g.scope='message.preview' AND g.revoked_at IS NULL AND julianday(g.expires_at)>julianday(?) JOIN agent_message_consents c ON c.sender_agent_id=m.sender_agent_id AND c.recipient_agent_id=m.recipient_agent_id AND c.recipient_owner_did=rb.owner_did AND c.revoked_at IS NULL AND julianday(c.expires_at)>julianday(?) WHERE k.key_id=m.key_id AND k.agent_id=m.sender_agent_id AND k.revoked_at IS NULL)
 ON CONFLICT(message_id) DO NOTHING`,now,e.MessageID,e.SenderAgentID,e.RecipientAgentID,e.PayloadDigest,now,now)
 if err!=nil{return err}
 var id,sender,recipient,digest string
 err=tx.QueryRowContext(ctx,`SELECT message_id,sender_agent_id,recipient_agent_id,payload_digest FROM agent_local_inbox WHERE message_id=?`,e.MessageID).Scan(&id,&sender,&recipient,&digest)
 if err!=nil{return ErrOutboxUnavailable}
 if sender!=e.SenderAgentID||recipient!=e.RecipientAgentID||digest!=e.PayloadDigest{return ErrTransportDenied}
 return tx.Commit()
}

// AuthenticatedWorker is an explicitly invoked local-only entrypoint. All
// flags default to false; no listener or background goroutine is created.
type AuthenticatedWorker struct{
 Enabled bool
 Transport *LocalTransport
 Outbox Outbox
 DB *sql.DB
}
func(w AuthenticatedWorker) configured()error{
 if !w.Enabled{return ErrRuntimeDisabled}
 if w.DB==nil||w.Transport==nil||!w.Transport.enabled||w.Transport.db!=w.DB||!w.Outbox.Enabled||w.Outbox.DB!=w.DB{return ErrOutboxUnavailable}
 if _,ok:=w.Outbox.Sink.(DurableInbox);!ok{return ErrOutboxUnavailable}
 if w.Outbox.Sink.(DurableInbox).DB!=w.DB{return ErrOutboxUnavailable}
 return w.Outbox.config()
}
// Submit binds a verified in-process caller's agent to the signed envelope.
// Existing Send verifies Ed25519 signatures, ownership, grant, consent and
// replay; only then may the durable outbox be enqueued.
func(w AuthenticatedWorker) Submit(ctx context.Context,caller VerifiedLocalCaller,e Envelope)error{
 if err:=w.configured();err!=nil{return err}
 if !caller.Verified||caller.OwnerDID==""||caller.AgentID==""||caller.AgentID!=e.SenderAgentID{return ErrRuntimeIdentity}
 var count int
 err:=w.DB.QueryRowContext(ctx,`SELECT count(*) FROM agent_owner_bindings b JOIN agent_registrations a ON a.id=b.agent_id
 WHERE b.agent_id=? AND b.owner_did=? AND b.revoked_at IS NULL AND a.status='verified'`,caller.AgentID,caller.OwnerDID).Scan(&count)
 if err!=nil{return err}
 if count!=1{return ErrRuntimeIdentity}
 // Record and enqueue atomically. Unlike Send followed by Enqueue, a failed
 // transaction leaves no committed message stranded or replay-blocked.
 if err=w.Outbox.config();err!=nil{return err}
 if err=w.Outbox.validateTimeRows(ctx);err!=nil{return err}
 now:=w.Transport.now().UTC()
 tx,err:=w.DB.BeginTx(ctx,nil);if err!=nil{return err};defer tx.Rollback()
 var pub []byte
 err=tx.QueryRowContext(ctx,`SELECT k.public_key FROM agent_interop_keys k
 JOIN agent_owner_bindings sb ON sb.agent_id=k.agent_id AND sb.revoked_at IS NULL
 JOIN agent_registrations sa ON sa.id=sb.agent_id AND sa.status='verified'
 JOIN agent_owner_bindings rb ON rb.agent_id=? AND rb.revoked_at IS NULL
 JOIN agent_registrations ra ON ra.id=rb.agent_id AND ra.status='verified'
 WHERE k.key_id=? AND k.agent_id=? AND k.algorithm='Ed25519' AND k.revoked_at IS NULL`,
 e.RecipientAgentID,e.KeyID,e.SenderAgentID).Scan(&pub)
 if err!=nil{return ErrTransportDenied}
 if err=VerifyEnvelope(e,ed25519.PublicKey(pub),now);err!=nil{return err}
 var authorized int
 err=tx.QueryRowContext(ctx,`SELECT count(*) FROM agent_scoped_grants g
 JOIN agent_owner_bindings b ON b.agent_id=g.agent_id AND b.owner_did=g.owner_did AND b.revoked_at IS NULL
 JOIN agent_message_consents c ON c.sender_agent_id=g.agent_id AND c.recipient_agent_id=g.grantee_agent_id AND c.revoked_at IS NULL
 JOIN agent_owner_bindings rb ON rb.agent_id=c.recipient_agent_id AND rb.owner_did=c.recipient_owner_did AND rb.revoked_at IS NULL
 WHERE g.agent_id=? AND g.grantee_agent_id=? AND g.scope='message.preview' AND g.revoked_at IS NULL
 AND julianday(g.expires_at)>julianday(?) AND julianday(c.expires_at)>julianday(?)`,
 e.SenderAgentID,e.RecipientAgentID,now.Format(time.RFC3339Nano),now.Format(time.RFC3339Nano)).Scan(&authorized)
 if err!=nil||authorized!=1{return ErrTransportDenied}
 res,err:=tx.ExecContext(ctx,`INSERT INTO agent_interop_messages(message_id,sender_agent_id,recipient_agent_id,key_id,payload_digest,signature,issued_at,expires_at,status)
 VALUES(?,?,?,?,?,?,?,?,'pending')`,e.MessageID,e.SenderAgentID,e.RecipientAgentID,e.KeyID,e.PayloadDigest,e.Signature,e.IssuedAt,e.ExpiresAt)
 if err!=nil{return ErrReplay}
 if _,err=res.RowsAffected();err!=nil{return err}
 at:=outboxStamp(w.Outbox.clock())
 _,err=tx.ExecContext(ctx,`INSERT INTO agent_local_outbox(message_id,available_at,updated_at) VALUES(?,?,?)`,e.MessageID,at,at)
 if err!=nil{return err}
 if err=tx.Commit();err!=nil{return err}
 return w.Transport.push(ctx)
}
// Tick executes at most one local delivery using the existing outbox lease,
// authorization/revocation checks, and durable idempotent inbox.
func(w AuthenticatedWorker) Tick(ctx context.Context)(bool,error){
 if err:=w.configured();err!=nil{return false,err}
 return w.Outbox.Tick(ctx)
}
