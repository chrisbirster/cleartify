package atproto

import (
 "bytes"
 "context"
 "encoding/base64"
 "encoding/binary"
 "fmt"
 "errors"
 "testing"

 "github.com/chrisbirster/cleartify/internal/database"
)

func SyntheticLocalKey(v byte)string{return base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{v},32))}
func TestVersionHeaderAuthenticated(t *testing.T){
 key:=SyntheticLocalKey(2)
 ring,err:=NewLocalOAuthKeyring(SyntheticLocalKey(1),map[uint32]string{1:key,2:key})
 if err!=nil{t.Fatal(err)}
 data,err:=ring.SealVersion([]byte("protected"),"request:state",1)
 if err!=nil{t.Fatal(err)}
 modified:=bytes.Clone(data)
 binary.BigEndian.PutUint32(modified[len(oauthCipherMagic):],2)
 if _,err=ring.Open(modified,"request:state");err==nil{t.Fatal("unauthenticated version change")}
 if _,err=ring.Open(data,"request:other");err==nil{t.Fatal("wrong associated identity accepted")}
 if plain,err:=ring.Open(data,"request:state");err!=nil||string(plain)!="protected"{t.Fatal("correct version failed",err)}
}
func TestLocalOAuthMigrationAtomicRollbackAndCompatibility(t *testing.T){
 ctx:=context.Background()
 store,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)}
 defer store.Close()
 if err=store.Migrate(ctx);err!=nil{t.Fatal(err)}
 old:=SyntheticLocalKey(10)
 ring,err:=NewLocalOAuthKeyring(old,map[uint32]string{1:SyntheticLocalKey(11),2:SyntheticLocalKey(12)})
 if err!=nil{t.Fatal(err)}
 original:=map[string][]byte{}
 for _,v:=range []struct{table,key,associated string}{
  {"atproto_oauth_requests","state-a","request:state-a"},
  {"atproto_oauth_requests","state-b","request:state-b"},
  {"atproto_oauth_sessions","did:plc:test:sid","session:did:plc:test:sid"},
 }{
  data,err:=ring.SealVersion([]byte("synthetic:"+v.key),v.associated,0);if err!=nil{t.Fatal(err)}
  original[v.key]=data
  if v.table=="atproto_oauth_requests"{
   _,err=store.DB.ExecContext(ctx,`INSERT INTO atproto_oauth_requests(state,payload_ciphertext,browser_nonce_hash,created_at,expires_at) VALUES(?,?,'hash','now','later')`,v.key,data)
  }else{
   _,err=store.DB.ExecContext(ctx,`INSERT INTO atproto_oauth_sessions(account_did,session_id,payload_ciphertext,created_at,updated_at) VALUES('did:plc:test','sid',?,'now','now')`,data)
  }
  if err!=nil{t.Fatal(err)}
 }
 get:=func(table,where string)[]byte{
  t.Helper()
  var b []byte
  if err:=store.DB.QueryRowContext(ctx,"SELECT payload_ciphertext FROM "+table+" WHERE "+where).Scan(&b);err!=nil{t.Fatal(err)}
  return b
 }
 m:=OAuthLocalMigrator{DB:store.DB,Ring:ring}
 if _,err=m.Migrate(ctx,0,1);!errors.Is(err,ErrOAuthMigrationDisabled){t.Fatal(err)}
 m.Enabled=true
 m.afterUpdate=func(n int)error{if n==3{return errors.New("simulated crash after session update")};return nil}
 if _,err=m.Migrate(ctx,0,1);err==nil{t.Fatal("failure injection not reached")}
 for _,item:=range []struct{table,where,key string}{
 {"atproto_oauth_requests","state='state-a'","state-a"},
 {"atproto_oauth_requests","state='state-b'","state-b"},
 {"atproto_oauth_sessions","session_id='sid'","did:plc:test:sid"},
 }{if !bytes.Equal(original[item.key],get(item.table,item.where)){t.Fatalf("partial commit after session update: %s",item.key)}}
 m.afterUpdate=nil
 n,err:=m.Migrate(ctx,0,1);if err!=nil||n!=3{t.Fatalf("rotation %d %v",n,err)}
 for _,item:=range []struct{table,where,ad,plain string}{
  {"atproto_oauth_requests","state='state-a'","request:state-a","synthetic:state-a"},
  {"atproto_oauth_requests","state='state-b'","request:state-b","synthetic:state-b"},
  {"atproto_oauth_sessions","session_id='sid'","session:did:plc:test:sid","synthetic:did:plc:test:sid"},
 }{
  data:=get(item.table,item.where)
  if bytes.Equal(data,original["state-a"]){t.Fatal("ciphertext not migrated")}
  out,err:=ring.Open(data,item.ad);if err!=nil||string(out)!=item.plain{t.Fatalf("unreadable %s %v",item.where,err)}
 }
 badRing,err:=NewLocalOAuthKeyring(old,map[uint32]string{1:SyntheticLocalKey(90)});if err!=nil{t.Fatal(err)}
 if _,err=badRing.Open(get("atproto_oauth_requests","state='state-a'"),"request:state-a");err==nil{t.Fatal("wrong key accepted")}
 if _,err=m.Migrate(ctx,0,2);!errors.Is(err,ErrOAuthMigrationConflict){t.Fatalf("mixed-version fail closed: %v",err)}
 // Mix real legacy request rows with a version-1 session row.
 sessionV1:=get("atproto_oauth_sessions","session_id='sid'")
 for _,key:=range []string{"state-a","state-b"}{
  if _,err=store.DB.ExecContext(ctx,"UPDATE atproto_oauth_requests SET payload_ciphertext=? WHERE state=?",original[key],key);err!=nil{t.Fatal(err)}
 }
 if _,err=m.Migrate(ctx,0,2);!errors.Is(err,ErrOAuthMigrationConflict){t.Fatalf("mixed 0/1 did not fail closed: %v",err)}
 for _,key:=range []string{"state-a","state-b"}{
  if !bytes.Equal(original[key],get("atproto_oauth_requests",fmt.Sprintf("state='%s'",key))){t.Fatal("mixed migration mutated request",key)}
 }
 if !bytes.Equal(sessionV1,get("atproto_oauth_sessions","session_id='sid'")){t.Fatal("mixed migration mutated session")}
 // Restore version 1 before testing wrong-key migration and rollback.
 for _,key:=range []string{"state-a","state-b"}{
  data,sealErr:=ring.SealVersion([]byte("synthetic:"+key),"request:"+key,1)
  if sealErr!=nil{t.Fatal(sealErr)}
  if _,err=store.DB.ExecContext(ctx,"UPDATE atproto_oauth_requests SET payload_ciphertext=? WHERE state=?",data,key);err!=nil{t.Fatal(err)}
 }
 beforeA:=get("atproto_oauth_requests","state='state-a'")
 beforeB:=get("atproto_oauth_requests","state='state-b'")
 beforeS:=get("atproto_oauth_sessions","session_id='sid'")
 badMigration:=OAuthLocalMigrator{DB:store.DB,Ring:badRing,Enabled:true}
 if _,err=badMigration.Migrate(ctx,1,0);err==nil{t.Fatal("wrong-key migration succeeded")}
 if !bytes.Equal(beforeA,get("atproto_oauth_requests","state='state-a'"))||!bytes.Equal(beforeB,get("atproto_oauth_requests","state='state-b'"))||!bytes.Equal(beforeS,get("atproto_oauth_sessions","session_id='sid'")){t.Fatal("wrong-key migration changed data")}
 // Migration back to version zero requires explicit opt-in and the old key.
 n,err=m.Migrate(ctx,1,0);if err!=nil||n!=3{t.Fatalf("rollback %d %v",n,err)}
 legacy,err:=NewSecretBox(old);if err!=nil{t.Fatal(err)}
 for _,item:=range []struct{table,where,ad,plain string}{
 {"atproto_oauth_requests","state='state-a'","request:state-a","synthetic:state-a"},
 {"atproto_oauth_requests","state='state-b'","request:state-b","synthetic:state-b"},
 {"atproto_oauth_sessions","session_id='sid'","session:did:plc:test:sid","synthetic:did:plc:test:sid"},
 }{
  plain,openErr:=legacy.Open(get(item.table,item.where),item.ad)
  if openErr!=nil||string(plain)!=item.plain{t.Fatalf("rollback unreadable %s: %v",item.where,openErr)}
 }
 versioned,err:=ring.SealVersion([]byte("unknown"),"test",2);if err!=nil{t.Fatal(err)}
 versioned[len(oauthCipherMagic)+3]=99
 if _,err=ring.Open(versioned,"test");!errors.Is(err,ErrUnknownOAuthCipherVersion){t.Fatalf("unknown version accepted: %v",err)}
}
