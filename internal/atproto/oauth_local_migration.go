package atproto

import (
 "bytes"
 "context"
 "database/sql"
 "encoding/base64"
 "encoding/binary"
 "errors"
 "fmt"
)

var ErrUnknownOAuthCipherVersion=errors.New("unknown OAuth ciphertext version")
var ErrOAuthMigrationDisabled=errors.New("OAuth key migration requires explicit opt-in")
var ErrOAuthMigrationConflict=errors.New("OAuth key migration encountered unexpected data")

// Version 0 is the existing, unprefixed AES-GCM nonce+ciphertext layout.
// Versioned payloads use a fixed magic and uint32 version before the legacy
// nonce+ciphertext envelope. No keys or decrypted data are persisted by this type.
var oauthCipherMagic=[]byte{0x43,0x4c,0x54,0x46,0x59,0x4b,0x45,0x59,0x00,0x01}

type OAuthKeyring struct{
 Legacy *SecretBox
 Versions map[uint32]*SecretBox
}
func NewLocalOAuthKeyring(legacyEncoded string, versions map[uint32]string)(OAuthKeyring,error){
 legacy,err:=NewSecretBox(legacyEncoded);if err!=nil{return OAuthKeyring{},err}
 result:=OAuthKeyring{Legacy:legacy,Versions:make(map[uint32]*SecretBox,len(versions))}
 for id,key:=range versions{
  if id==0{return OAuthKeyring{},ErrUnknownOAuthCipherVersion}
  box,err:=NewSecretBox(key);if err!=nil{return OAuthKeyring{},err}
  result.Versions[id]=box
 }
 return result,nil
}
func (r OAuthKeyring) version(ciphertext []byte)(uint32,error){
 if len(ciphertext)>=len(oauthCipherMagic)&&bytes.Equal(ciphertext[:len(oauthCipherMagic)],oauthCipherMagic){
  if len(ciphertext)<len(oauthCipherMagic)+4{return 0,ErrUnknownOAuthCipherVersion}
  id:=binary.BigEndian.Uint32(ciphertext[len(oauthCipherMagic):])
  if id==0||r.Versions[id]==nil{return 0,ErrUnknownOAuthCipherVersion}
  return id,nil
 }
 // Strictly reject damaged/truncated version prefixes rather than
 // silently attempting legacy decryption on a known prefix.
 if len(ciphertext)>0&&len(ciphertext)<len(oauthCipherMagic)&&bytes.Equal(ciphertext,oauthCipherMagic[:len(ciphertext)]){return 0,ErrUnknownOAuthCipherVersion}
 if len(ciphertext)>=len(oauthCipherMagic)-1&&bytes.Equal(ciphertext[:len(oauthCipherMagic)-1],oauthCipherMagic[:len(oauthCipherMagic)-1]){return 0,ErrUnknownOAuthCipherVersion}
 return 0,nil
}
func (r OAuthKeyring) Open(ciphertext []byte,associatedID string)([]byte,error){
 id,err:=r.version(ciphertext);if err!=nil{return nil,err}
 if id==0{
  if r.Legacy==nil{return nil,ErrUnknownOAuthCipherVersion}
  return r.Legacy.Open(ciphertext,associatedID)
 }
 return r.Versions[id].Open(ciphertext[len(oauthCipherMagic)+4:],associatedID)
}
func (r OAuthKeyring) SealVersion(plaintext []byte,associatedID string,id uint32)([]byte,error){
 box:=r.Versions[id]
 if id==0{box=r.Legacy}
 if box==nil{return nil,ErrUnknownOAuthCipherVersion}
 raw,err:=box.Seal(plaintext,associatedID);if err!=nil{return nil,err}
 if id==0{return raw,nil}
 out:=make([]byte,len(oauthCipherMagic)+4,len(oauthCipherMagic)+4+len(raw))
 copy(out,oauthCipherMagic)
 binary.BigEndian.PutUint32(out[len(oauthCipherMagic):],id)
 return append(out,raw...),nil
}

type OAuthLocalMigrator struct{
 DB *sql.DB
 Ring OAuthKeyring
 Enabled bool
 // Test-only failure injection: return an error after the Nth update.
 afterUpdate func(int)error
}
// Migrate performs an all-or-nothing local conversion of BOTH persisted OAuth
// request and session ciphertext columns. It never calls Push, providers,
// browsers or remote services. Rollback is a separate explicitly requested
// reverse migration while both key versions are still available.
func(m OAuthLocalMigrator) Migrate(ctx context.Context,from,to uint32)(int,error){
 if !m.Enabled{return 0,ErrOAuthMigrationDisabled}
 if m.DB==nil||from==to{return 0,ErrOAuthMigrationConflict}
 if from!=0&&m.Ring.Versions[from]==nil{return 0,ErrUnknownOAuthCipherVersion}
 if to!=0&&m.Ring.Versions[to]==nil{return 0,ErrUnknownOAuthCipherVersion}
 if m.Ring.Legacy==nil{return 0,ErrUnknownOAuthCipherVersion}
 tx,err:=m.DB.BeginTx(ctx,nil);if err!=nil{return 0,err}
 defer tx.Rollback()
 total:=0
 for _,table:=range []struct{name,key,context string}{
  {"atproto_oauth_requests","state","request:"},
  {"atproto_oauth_sessions","account_did,session_id","session:"},
 }{
  query:=""
  if table.name=="atproto_oauth_requests"{query="SELECT state,payload_ciphertext FROM atproto_oauth_requests ORDER BY state"
  }else{query="SELECT account_did,session_id,payload_ciphertext FROM atproto_oauth_sessions ORDER BY account_did,session_id"}
  rows,err:=tx.QueryContext(ctx,query);if err!=nil{return 0,err}
  type entry struct{first,second string;data []byte}
  entries:=[]entry{}
  for rows.Next(){
   var e entry
   if table.name=="atproto_oauth_requests"{err=rows.Scan(&e.first,&e.data)}else{err=rows.Scan(&e.first,&e.second,&e.data)}
   if err!=nil{rows.Close();return 0,err}
   entries=append(entries,e)
  }
  err=rows.Err();rows.Close();if err!=nil{return 0,err}
  for _,e:=range entries{
   actual,err:=m.Ring.version(e.data);if err!=nil{return 0,err}
   if actual!=from{return 0,fmt.Errorf("%w: mixed version in %s",ErrOAuthMigrationConflict,table.name)}
   associated:=table.context+e.first
   if e.second!=""{associated+=":"+e.second}
   plain,err:=m.Ring.Open(e.data,associated);if err!=nil{return 0,err}
   sealed,err:=m.Ring.SealVersion(plain,associated,to);if err!=nil{return 0,err}
   var res sql.Result
   if e.second==""{res,err=tx.ExecContext(ctx,"UPDATE "+table.name+" SET payload_ciphertext=? WHERE "+table.key+"=? AND payload_ciphertext=?",sealed,e.first,e.data)
   }else{res,err=tx.ExecContext(ctx,"UPDATE "+table.name+" SET payload_ciphertext=? WHERE account_did=? AND session_id=? AND payload_ciphertext=?",sealed,e.first,e.second,e.data)}
   if err!=nil{return 0,err}
   n,err:=res.RowsAffected();if err!=nil||n!=1{return 0,ErrOAuthMigrationConflict}
   total++
   if m.afterUpdate!=nil{if err=m.afterUpdate(total);err!=nil{return 0,err}}
  }
 }
 if err=tx.Commit();err!=nil{return 0,err}
 return total,nil
}
func SyntheticLocalKey(value byte)string{
 // Convenience only for tests: never call this in production config.
 return base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{value},32))
}
