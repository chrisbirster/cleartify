package atproto
import("bytes";"encoding/base64";"testing")
func TestOAuthKeyRotationRequiresMigration(t *testing.T){
 a,e:=NewSecretBox(base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{1},32)));if e!=nil{t.Fatal(e)}
 b,e:=NewSecretBox(base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{2},32)));if e!=nil{t.Fatal(e)}
 cipher,e:=a.Seal([]byte("synthetic-session-token"),"oauth-row");if e!=nil{t.Fatal(e)}
 if _,e=b.Open(cipher,"oauth-row");e==nil{t.Fatal("new key unexpectedly decrypted old session")}
 if _,e=a.Open(cipher,"other-row");e==nil{t.Fatal("associated row switched")}
 plain,e:=a.Open(cipher,"oauth-row");if e!=nil||string(plain)!="synthetic-session-token"{t.Fatalf("original key failed: %v",e)}
 if _,e=NewSecretBox("broken");e==nil{t.Fatal("invalid key accepted")}
}
