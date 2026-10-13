package interop

import (
 "crypto/ed25519"
 "crypto/sha256"
 "encoding/json"
 "errors"
 "time"
)
var ErrInvalidEnvelope=errors.New("invalid signed agent envelope")
type Envelope struct {
 MessageID string `json:"messageId"`
 SenderAgentID string `json:"senderAgentId"`
 RecipientAgentID string `json:"recipientAgentId"`
 KeyID string `json:"keyId"`
 PayloadDigest string `json:"payloadDigest"`
 IssuedAt string `json:"issuedAt"`
 ExpiresAt string `json:"expiresAt"`
 Signature []byte `json:"signature"`
}
func(e Envelope) canonical()([]byte,error){
 return json.Marshal(struct{MessageID,SenderAgentID,RecipientAgentID,KeyID,PayloadDigest,IssuedAt,ExpiresAt string}{
 e.MessageID,e.SenderAgentID,e.RecipientAgentID,e.KeyID,e.PayloadDigest,e.IssuedAt,e.ExpiresAt})
}
// VerifyEnvelope checks signature and expiry only. Caller must separately
// resolve a trusted, unrevoked key, replay record, ownership and approval.
func VerifyEnvelope(e Envelope,public ed25519.PublicKey,now time.Time)error{
 if len(public)!=ed25519.PublicKeySize||e.MessageID==""||e.SenderAgentID==""||e.RecipientAgentID==""||e.KeyID==""||e.PayloadDigest==""||len(e.PayloadDigest)!=64||e.SenderAgentID==e.RecipientAgentID{return ErrInvalidEnvelope}
 issued,err:=time.Parse(time.RFC3339,e.IssuedAt);if err!=nil{return ErrInvalidEnvelope}
 expires,err:=time.Parse(time.RFC3339,e.ExpiresAt);if err!=nil||!expires.After(issued)||!now.Before(expires)||issued.After(now.Add(2*time.Minute))||expires.Sub(issued)>10*time.Minute{return ErrInvalidEnvelope}
 b,err:=e.canonical();if err!=nil{return err}
 digest:=sha256.Sum256(b)
 if !ed25519.Verify(public,digest[:],e.Signature){return ErrInvalidEnvelope}
 return nil
}
