package interop
import("crypto/ed25519";"crypto/rand";"crypto/sha256";"encoding/hex";"testing";"time")
func TestSignedEnvelope(t *testing.T){
 pub,priv,err:=ed25519.GenerateKey(rand.Reader);if err!=nil{t.Fatal(err)}
 now:=time.Date(2026,10,13,0,0,0,0,time.UTC)
 e:=Envelope{MessageID:"m1",SenderAgentID:"a",RecipientAgentID:"b",KeyID:"k",PayloadDigest:hex.EncodeToString(make([]byte,32)),IssuedAt:now.Format(time.RFC3339),ExpiresAt:now.Add(time.Minute).Format(time.RFC3339)}
 canonical,_:=e.canonical();digest:=sha256.Sum256(canonical);e.Signature=ed25519.Sign(priv,digest[:])
 if err:=VerifyEnvelope(e,pub,now);err!=nil{t.Fatal(err)}
 e.RecipientAgentID="c"
 if err:=VerifyEnvelope(e,pub,now);err==nil{t.Fatal("tamper accepted")}
 e.RecipientAgentID="b"
 if err:=VerifyEnvelope(e,pub,now.Add(2*time.Minute));err==nil{t.Fatal("expired accepted")}
}
