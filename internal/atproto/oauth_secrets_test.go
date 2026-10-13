package atproto

import (
 "bytes"
 "encoding/base64"
 "testing"
)
func TestOAuthEncryptionRoundtripAndAuthentication(t *testing.T){
 b,err:=NewSecretBox(base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{7},32)))
 if err!=nil{t.Fatal(err)}
 p:=[]byte("oauth-refresh-token-placeholder")
 a,err:=b.Seal(p,"state-1");if err!=nil{t.Fatal(err)}
 c,err:=b.Seal(p,"state-1");if err!=nil{t.Fatal(err)}
 if bytes.Equal(a,c)||bytes.Contains(a,p){t.Fatal("ciphertext leaked or deterministic")}
 out,err:=b.Open(a,"state-1");if err!=nil || !bytes.Equal(out,p){t.Fatalf("roundtrip %q %v",out,err)}
 if _,err:=b.Open(a,"state-2");err==nil{t.Fatal("context substitution accepted")}
 a[len(a)-1]^=1
 if _,err:=b.Open(a,"state-1");err==nil{t.Fatal("tampered ciphertext accepted")}
}
func TestOAuthKeyValidation(t *testing.T){
 if _,err:=NewSecretBox("");err==nil{t.Fatal("empty key accepted")}
 if _,err:=NewSecretBox(base64.StdEncoding.EncodeToString([]byte("short")));err==nil{t.Fatal("short key accepted")}
}
