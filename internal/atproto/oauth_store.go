package atproto

import (
 "context"
 "crypto/sha256"
 "database/sql"
 "encoding/hex"
 "encoding/json"
 "errors"
 "fmt"
 "time"

 "github.com/bluesky-social/indigo/atproto/auth/oauth"
 "github.com/bluesky-social/indigo/atproto/syntax"
 "github.com/chrisbirster/cleartify/internal/database"
)

// OAuthStore is a durable implementation of the Indigo ClientAuthStore.
// Tokens, DPoP keys, and PKCE verifiers are encrypted before entering Turso.
type OAuthStore struct {
 DB *database.Store
 Box *SecretBox
}
var _ oauth.ClientAuthStore = (*OAuthStore)(nil)
var ErrInvalidState = errors.New("OAuth request expired, consumed, or not from this browser")

type nonceContextKey struct{}
func bindNonce(ctx context.Context, nonce string) context.Context {return context.WithValue(ctx,nonceContextKey{},nonce)}
func nonceFrom(ctx context.Context) (string,bool) {v,ok:=ctx.Value(nonceContextKey{}).(string);return v,ok&&len(v)>=32}
func digest(value string) string {h:=sha256.Sum256([]byte(value));return hex.EncodeToString(h[:])}

func (s *OAuthStore) SaveAuthRequestInfo(ctx context.Context, info oauth.AuthRequestData) error {
 nonce,ok:=nonceFrom(ctx);if !ok{return ErrInvalidState}
 raw,err:=json.Marshal(info);if err!=nil{return err}
 data,err:=s.Box.Seal(raw,"request:"+info.State);if err!=nil{return err}
 now:=time.Now().UTC()
 _,err=s.DB.DB.ExecContext(ctx,`INSERT INTO atproto_oauth_requests(state,payload_ciphertext,browser_nonce_hash,created_at,expires_at) VALUES(?,?,?,?,?)`,
 info.State,data,digest(nonce),now.Format(time.RFC3339Nano),now.Add(10*time.Minute).Format(time.RFC3339Nano))
 if err!=nil{return err}
 return s.DB.Push(ctx)
}

// GetAuthRequestInfo atomically consumes the browser-bound request before token exchange.
// A second callback, mismatched cookie, expired state or concurrent replay cannot succeed.
func(s *OAuthStore) GetAuthRequestInfo(ctx context.Context, state string)(*oauth.AuthRequestData,error){
 nonce,ok:=nonceFrom(ctx);if !ok{return nil,ErrInvalidState}
 tx,err:=s.DB.DB.BeginTx(ctx,nil);if err!=nil{return nil,err}
 defer tx.Rollback()
 var encrypted []byte;var expires string
 err=tx.QueryRowContext(ctx,`SELECT payload_ciphertext,expires_at FROM atproto_oauth_requests WHERE state=? AND browser_nonce_hash=?`,state,digest(nonce)).Scan(&encrypted,&expires)
 if errors.Is(err,sql.ErrNoRows){return nil,ErrInvalidState};if err!=nil{return nil,err}
 expiresAt,err:=time.Parse(time.RFC3339Nano,expires);if err!=nil||!time.Now().Before(expiresAt){return nil,ErrInvalidState}
 result,err:=tx.ExecContext(ctx,`DELETE FROM atproto_oauth_requests WHERE state=? AND browser_nonce_hash=?`,state,digest(nonce));if err!=nil{return nil,err}
 n,err:=result.RowsAffected();if err!=nil||n!=1{return nil,ErrInvalidState}
 if err:=tx.Commit();err!=nil{return nil,err}
 if err:=s.DB.Push(ctx);err!=nil{return nil,err}
 raw,err:=s.Box.Open(encrypted,"request:"+state);if err!=nil{return nil,err}
 var info oauth.AuthRequestData
 if err:=json.Unmarshal(raw,&info);err!=nil{return nil,err}
 if info.State!=state{return nil,ErrInvalidState}
 return &info,nil
}
func(s *OAuthStore) DeleteAuthRequestInfo(ctx context.Context,state string) error{
 _,err:=s.DB.DB.ExecContext(ctx,`DELETE FROM atproto_oauth_requests WHERE state=?`,state)
 if err!=nil{return err}
 return s.DB.Push(ctx)
}
func(s *OAuthStore) SaveSession(ctx context.Context,sess oauth.ClientSessionData) error{
 raw,err:=json.Marshal(sess);if err!=nil{return err}
 key:=sess.AccountDID.String()+":"+sess.SessionID
 data,err:=s.Box.Seal(raw,"session:"+key);if err!=nil{return err}
 now:=time.Now().UTC().Format(time.RFC3339Nano)
 _,err=s.DB.DB.ExecContext(ctx,`INSERT INTO atproto_oauth_sessions(account_did,session_id,payload_ciphertext,created_at,updated_at) VALUES(?,?,?,?,?) ON CONFLICT(account_did,session_id) DO UPDATE SET payload_ciphertext=excluded.payload_ciphertext,updated_at=excluded.updated_at`,sess.AccountDID.String(),sess.SessionID,data,now,now)
 if err!=nil{return err}
 return s.DB.Push(ctx)
}
func(s *OAuthStore) GetSession(ctx context.Context,did syntax.DID,id string)(*oauth.ClientSessionData,error){
 var ciphertext []byte
 err:=s.DB.DB.QueryRowContext(ctx,`SELECT payload_ciphertext FROM atproto_oauth_sessions WHERE account_did=? AND session_id=?`,did.String(),id).Scan(&ciphertext);if err!=nil{return nil,err}
 raw,err:=s.Box.Open(ciphertext,"session:"+did.String()+":"+id);if err!=nil{return nil,err}
 var sess oauth.ClientSessionData;if err:=json.Unmarshal(raw,&sess);err!=nil{return nil,err}
 if sess.AccountDID.String()!=did.String()||sess.SessionID!=id{return nil,fmt.Errorf("OAuth session identity mismatch")}
 return &sess,nil
}
func(s *OAuthStore) DeleteSession(ctx context.Context,did syntax.DID,id string) error{
 tx,err:=s.DB.DB.BeginTx(ctx,nil);if err!=nil{return err}
 defer tx.Rollback()
 if _,err=tx.ExecContext(ctx,`DELETE FROM atproto_browser_sessions WHERE account_did=? AND oauth_session_id=?`,did.String(),id);err!=nil{return err}
 if _,err=tx.ExecContext(ctx,`DELETE FROM atproto_oauth_sessions WHERE account_did=? AND session_id=?`,did.String(),id);err!=nil{return err}
 if err=tx.Commit();err!=nil{return err}
 return s.DB.Push(ctx)
}
