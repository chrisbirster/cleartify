package agents

import (
 "context"
 "crypto/ed25519"
 "crypto/rand"
 "database/sql"
 "encoding/hex"
 "errors"
 "strings"
 "time"
)

var ErrOwnership=errors.New("owner binding proof or registration invalid")
var ErrGrant=errors.New("grant denied, expired or revoked")
const MaxGrantTTL=24*time.Hour
var AllowedScopes=map[string]bool{"message.preview":true,"protocol.preview":true,"local.read":true}
// OwnerProof confirms possession of the registered agent key by the claimed person DID.
// It does NOT independently validate a DID against a remote PDS. Caller must supply
// an authenticated, verified person DID from an existing server-owned session.
type GrantService struct{DB *sql.DB;Now func()time.Time}
func(s GrantService) now()time.Time{if s.Now!=nil{return s.Now().UTC()};return time.Now().UTC()}
func OwnerProofMessage(agentID,ownerDID string)[]byte{return []byte("cleartify-agent-owner-v1\n"+agentID+"\n"+ownerDID)}
func(s GrantService) Bind(ctx context.Context,authenticatedOwnerDID,agentID string,pub ed25519.PublicKey,sig []byte)error{
 if s.DB==nil||!strings.HasPrefix(authenticatedOwnerDID,"did:")||len(pub)!=ed25519.PublicKeySize||len(sig)!=ed25519.SignatureSize||!ed25519.Verify(pub,OwnerProofMessage(agentID,authenticatedOwnerDID),sig){return ErrOwnership}
 var owner,status string;var claimed int
 err:=s.DB.QueryRowContext(ctx,`SELECT a.owner_did,a.status,p.claimed FROM agent_registrations a JOIN social_profiles p ON p.did=a.owner_did WHERE a.id=? AND p.kind='person'`,agentID).Scan(&owner,&status,&claimed)
 if errors.Is(err,sql.ErrNoRows){return ErrOwnership};if err!=nil{return err}
 if owner!=authenticatedOwnerDID||claimed!=1||status=="disabled"{return ErrOwnership}
 now:=s.now().Format(time.RFC3339Nano)
 // Fail closed on rebind: an agent cannot silently swap owner keys.
 result,err:=s.DB.ExecContext(ctx,`INSERT INTO agent_owner_bindings(agent_id,owner_did,agent_public_key,proof_signature,verified_at) VALUES(?,?,?,?,?)`,agentID,authenticatedOwnerDID,[]byte(pub),sig,now)
 if err!=nil{return ErrOwnership};_,err=result.RowsAffected();return err
}
func(s GrantService) Grant(ctx context.Context,authenticatedOwnerDID,agentID,granteeID,scope string,ttl time.Duration)(string,error){
 if s.DB==nil||!AllowedScopes[scope]||ttl<=0||ttl>MaxGrantTTL||agentID==granteeID{return "",ErrGrant}
 var ok int
 err:=s.DB.QueryRowContext(ctx,`SELECT 1 FROM agent_owner_bindings b JOIN agent_registrations a ON a.id=b.agent_id WHERE b.agent_id=? AND b.owner_did=? AND b.revoked_at IS NULL AND a.status='verified' AND EXISTS(SELECT 1 FROM agent_owner_bindings rb JOIN agent_registrations ra ON ra.id=rb.agent_id WHERE rb.agent_id=? AND rb.revoked_at IS NULL AND ra.status='verified')`,agentID,authenticatedOwnerDID,granteeID).Scan(&ok)
 if errors.Is(err,sql.ErrNoRows){return "",ErrGrant};if err!=nil{return "",err}
 idBytes:=make([]byte,16);if _,err=rand.Read(idBytes);err!=nil{return "",err}
 id:=hex.EncodeToString(idBytes);now:=s.now()
 _,err=s.DB.ExecContext(ctx,`INSERT INTO agent_scoped_grants(id,owner_did,agent_id,grantee_agent_id,scope,created_at,expires_at) VALUES(?,?,?,?,?,?,?)`,id,authenticatedOwnerDID,agentID,granteeID,scope,now.Format(time.RFC3339Nano),now.Add(ttl).Format(time.RFC3339Nano))
 return id,err
}
func(s GrantService) Allows(ctx context.Context,agentID,granteeID,scope string)(bool,error){
 if s.DB==nil||!AllowedScopes[scope]{return false,nil}
 var count int
 err:=s.DB.QueryRowContext(ctx,`SELECT count(*) FROM agent_scoped_grants g JOIN agent_owner_bindings b ON b.agent_id=g.agent_id AND b.owner_did=g.owner_did JOIN agent_owner_bindings rb ON rb.agent_id=g.grantee_agent_id JOIN agent_registrations a ON a.id=g.agent_id JOIN agent_registrations ra ON ra.id=g.grantee_agent_id WHERE g.agent_id=? AND g.grantee_agent_id=? AND g.scope=? AND g.revoked_at IS NULL AND g.expires_at>? AND b.revoked_at IS NULL AND rb.revoked_at IS NULL AND a.status='verified' AND ra.status='verified'`,agentID,granteeID,scope,s.now().Format(time.RFC3339Nano)).Scan(&count)
 return count>0,err
}
func(s GrantService) Revoke(ctx context.Context,authenticatedOwnerDID,grantID string)error{
 if s.DB==nil||authenticatedOwnerDID==""||grantID==""{return ErrGrant}
 res,err:=s.DB.ExecContext(ctx,`UPDATE agent_scoped_grants SET revoked_at=? WHERE id=? AND owner_did=? AND revoked_at IS NULL`,s.now().Format(time.RFC3339Nano),grantID,authenticatedOwnerDID)
 if err!=nil{return err};n,err:=res.RowsAffected();if err!=nil{return err};if n!=1{return ErrGrant};return nil
}
func(s GrantService) Unbind(ctx context.Context,authenticatedOwnerDID,agentID string)error{
 if s.DB==nil||authenticatedOwnerDID==""||agentID==""{return ErrOwnership}
 res,err:=s.DB.ExecContext(ctx,`UPDATE agent_owner_bindings SET revoked_at=? WHERE agent_id=? AND owner_did=? AND revoked_at IS NULL`,s.now().Format(time.RFC3339Nano),agentID,authenticatedOwnerDID)
 if err!=nil{return err};n,err:=res.RowsAffected();if err!=nil{return err};if n!=1{return ErrOwnership};return nil
}
