package atproto

import (
 "crypto/aes"
 "crypto/cipher"
 "crypto/rand"
 "encoding/base64"
 "errors"
 "io"
)

var ErrInvalidOAuthKey = errors.New("ATPROTO_SESSION_KEY_B64 must decode to 32 bytes")

// SecretBox encrypts OAuth verifier, DPoP key and refresh token material before
// any Turso persistence. Key material comes from host-managed env/secrets only.
type SecretBox struct { aead cipher.AEAD }

func NewSecretBox(encoded string) (*SecretBox,error){
 key,err:=base64.StdEncoding.DecodeString(encoded)
 if err!=nil || len(key)!=32 {return nil,ErrInvalidOAuthKey}
 block,err:=aes.NewCipher(key);if err!=nil{return nil,err}
 a,err:=cipher.NewGCM(block);if err!=nil{return nil,err}
 return &SecretBox{aead:a},nil
}
func (b *SecretBox) Seal(plaintext []byte, associatedID string) ([]byte,error){
 nonce:=make([]byte,b.aead.NonceSize())
 if _,err:=io.ReadFull(rand.Reader,nonce);err!=nil{return nil,err}
 return b.aead.Seal(nonce,nonce,plaintext,[]byte(associatedID)),nil
}
func (b *SecretBox) Open(ciphertext []byte,associatedID string)([]byte,error){
 n:=b.aead.NonceSize()
 if len(ciphertext)<n {return nil,errors.New("invalid OAuth ciphertext")}
 return b.aead.Open(nil,ciphertext[:n],ciphertext[n:],[]byte(associatedID))
}
