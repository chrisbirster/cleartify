package interop

import(
 "context"
 "crypto/ed25519"
 "database/sql"
 "errors"
 "time"
 "github.com/chrisbirster/cleartify/internal/agents"
 "github.com/chrisbirster/cleartify/internal/database"
)
var ErrConsent=errors.New("recipient consent or scoped sender grant missing")
var ErrRotation=errors.New("key rotation rejected")
type LocalService struct{Store *database.Store;Now func()time.Time}
func(s LocalService) clock()time.Time{if s.Now!=nil{return s.Now().UTC()};return time.Now().UTC()}
func(s LocalService) Consent(ctx context.Context,recipientOwner,recipient,sender string,ttl time.Duration)error{
 if s.Store==nil||ttl<=0||ttl>24*time.Hour||recipient==sender{return ErrConsent}
 var owner string
 err:=s.Store.DB.QueryRowContext(ctx,`SELECT b.owner_did FROM agent_owner_bindings b JOIN agent_registrations a ON a.id=b.agent_id WHERE b.agent_id=? AND b.revoked_at IS NULL AND a.status='verified'`,recipient).Scan(&owner)
 if err!=nil||owner!=recipientOwner{return ErrConsent}
 _,err=s.Store.DB.ExecContext(ctx,`INSERT INTO agent_message_consents(sender_agent_id,recipient_agent_id,recipient_owner_did,granted_at,expires_at,revoked_at) VALUES(?,?,?,?,?,NULL) ON CONFLICT(sender_agent_id,recipient_agent_id) DO UPDATE SET recipient_owner_did=excluded.recipient_owner_did,granted_at=excluded.granted_at,expires_at=excluded.expires_at,revoked_at=NULL`,sender,recipient,recipientOwner,s.clock().Format(time.RFC3339Nano),s.clock().Add(ttl).Format(time.RFC3339Nano))
 return err
}
func(s LocalService) RevokeConsent(ctx context.Context,recipientOwner,recipient,sender string)error{
 if s.Store==nil{return ErrConsent}
 res,err:=s.Store.DB.ExecContext(ctx,`UPDATE agent_message_consents SET revoked_at=? WHERE sender_agent_id=? AND recipient_agent_id=? AND recipient_owner_did=? AND revoked_at IS NULL`,s.clock().Format(time.RFC3339Nano),sender,recipient,recipientOwner)
 if err!=nil{return err};n,err:=res.RowsAffected();if err!=nil||n!=1{return ErrConsent};return nil
}
// RotateKey requires a signature by the currently active key over the next key,
// preserving continuity of control. It atomically revokes the old key and adds
// the new public key; no private keys are ever stored.
func(s LocalService) RotateKey(ctx context.Context,agentID,oldID,newID string,newPub ed25519.PublicKey,proof []byte)error{
 if s.Store==nil||oldID==newID||newID==""||len(newPub)!=ed25519.PublicKeySize{return ErrRotation}
 tx,err:=s.Store.DB.BeginTx(ctx,nil);if err!=nil{return err};defer tx.Rollback()
 var old []byte
 err=tx.QueryRowContext(ctx,`SELECT k.public_key FROM agent_interop_keys k JOIN agent_owner_bindings b ON b.agent_id=k.agent_id JOIN agent_registrations a ON a.id=k.agent_id WHERE k.key_id=? AND k.agent_id=? AND k.revoked_at IS NULL AND b.revoked_at IS NULL AND a.status='verified'`,oldID,agentID).Scan(&old)
 if err!=nil||!ed25519.Verify(ed25519.PublicKey(old),[]byte("cleartify-rotate-v1\n"+agentID+"\n"+oldID+"\n"+newID+"\n"+string(newPub)),proof){return ErrRotation}
 now:=s.clock().Format(time.RFC3339Nano)
 res,err:=tx.ExecContext(ctx,`UPDATE agent_interop_keys SET revoked_at=? WHERE key_id=? AND agent_id=? AND revoked_at IS NULL`,now,oldID,agentID);if err!=nil{return err}
 n,err:=res.RowsAffected();if err!=nil||n!=1{return ErrRotation}
 if _,err=tx.ExecContext(ctx,`INSERT INTO agent_interop_keys(key_id,agent_id,algorithm,public_key,created_at) VALUES(?,?,'Ed25519',?,?)`,newID,agentID,[]byte(newPub),now);err!=nil{return ErrRotation}
 return tx.Commit()
}
// VerifyAndRecordConsented enforces BOTH grant and recipient consent before any
// local pending-message storage. It deliberately performs no delivery.
func(s LocalService) VerifyAndRecordConsented(ctx context.Context,e Envelope)error{
 if s.Store==nil{return ErrConsent}
 now:=s.clock()
 grant:=agents.GrantService{DB:s.Store.DB,Now:func()time.Time{return now}}
 allowed,err:=grant.Allows(ctx,e.SenderAgentID,e.RecipientAgentID,"message.preview")
 if err!=nil{return err};if !allowed{return ErrConsent}
 var count int
 err=s.Store.DB.QueryRowContext(ctx,`SELECT count(*) FROM agent_message_consents c JOIN agent_owner_bindings b ON b.agent_id=c.recipient_agent_id AND b.owner_did=c.recipient_owner_did WHERE c.sender_agent_id=? AND c.recipient_agent_id=? AND c.revoked_at IS NULL AND c.expires_at>? AND b.revoked_at IS NULL`,e.SenderAgentID,e.RecipientAgentID,now.Format(time.RFC3339Nano)).Scan(&count)
 if err!=nil{return err};if count!=1{return ErrConsent}
 return (ReplayRegistry{Store:s.Store,Now:func()time.Time{return now}}).VerifyAndRecord(ctx,e)
}
// Storage layer is local-only, no delivery, no transport, no acceptance.
var _ *sql.DB
