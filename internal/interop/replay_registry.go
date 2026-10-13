package interop

import (
 "context"
 "crypto/ed25519"
 "database/sql"
 "errors"
 "time"
 "github.com/chrisbirster/cleartify/internal/database"
)
var ErrReplay=errors.New("agent envelope message ID already recorded")
var ErrKeyUnavailable=errors.New("agent signing key absent or revoked")
type ReplayRegistry struct{Store *database.Store;Now func()time.Time}
func (r ReplayRegistry) clock()time.Time{
 if r.Now!=nil{return r.Now().UTC()}
 return time.Now().UTC()
}
// VerifyAndRecord performs local signature and replay checks only.
// It intentionally never delivers, accepts or executes a message;
// recipient consent and scoped capability grants are separate mandatory gates.
func(r ReplayRegistry) VerifyAndRecord(ctx context.Context,e Envelope)error{
 if r.Store==nil{return ErrKeyUnavailable}
 var raw []byte
 err:=r.Store.DB.QueryRowContext(ctx,`SELECT public_key FROM agent_interop_keys WHERE key_id=? AND agent_id=? AND revoked_at IS NULL AND algorithm='Ed25519'`,e.KeyID,e.SenderAgentID).Scan(&raw)
 if errors.Is(err,sql.ErrNoRows){return ErrKeyUnavailable}
 if err!=nil{return err}
 if err:=VerifyEnvelope(e,ed25519.PublicKey(raw),r.clock());err!=nil{return err}
 _,err=r.Store.DB.ExecContext(ctx,`INSERT INTO agent_interop_messages(message_id,sender_agent_id,recipient_agent_id,key_id,payload_digest,signature,issued_at,expires_at,status) VALUES(?,?,?,?,?,?,?,?,'pending')`,e.MessageID,e.SenderAgentID,e.RecipientAgentID,e.KeyID,e.PayloadDigest,e.Signature,e.IssuedAt,e.ExpiresAt)
 if err!=nil{return ErrReplay}
 return r.Store.Push(ctx)
}
