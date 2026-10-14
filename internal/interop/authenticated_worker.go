package interop

import (
 "context"
 "database/sql"
 "errors"
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
 _,err:=s.DB.ExecContext(ctx,`INSERT INTO agent_local_inbox(message_id,sender_agent_id,recipient_agent_id,payload_digest,received_at)
 SELECT m.message_id,m.sender_agent_id,m.recipient_agent_id,m.payload_digest,?
 FROM agent_interop_messages m WHERE m.message_id=? AND m.sender_agent_id=? AND m.recipient_agent_id=? AND m.payload_digest=? AND m.status='pending'
 ON CONFLICT(message_id) DO NOTHING`,time.Now().UTC().Format(time.RFC3339Nano),e.MessageID,e.SenderAgentID,e.RecipientAgentID,e.PayloadDigest)
 if err!=nil{return err}
 var id,sender,recipient,digest string
 err=s.DB.QueryRowContext(ctx,`SELECT message_id,sender_agent_id,recipient_agent_id,payload_digest FROM agent_local_inbox WHERE message_id=?`,e.MessageID).Scan(&id,&sender,&recipient,&digest)
 if err!=nil{return ErrOutboxUnavailable}
 if sender!=e.SenderAgentID||recipient!=e.RecipientAgentID||digest!=e.PayloadDigest{return ErrTransportDenied}
 return nil
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
 if w.DB==nil||w.Transport==nil||!w.Transport.enabled||!w.Outbox.Enabled||w.Outbox.DB!=w.DB{return ErrOutboxUnavailable}
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
 if err=w.Transport.Send(ctx,e);err!=nil{return err}
 return w.Outbox.Enqueue(ctx,e.MessageID)
}
// Tick executes at most one local delivery using the existing outbox lease,
// authorization/revocation checks, and durable idempotent inbox.
func(w AuthenticatedWorker) Tick(ctx context.Context)(bool,error){
 if err:=w.configured();err!=nil{return false,err}
 return w.Outbox.Tick(ctx)
}
