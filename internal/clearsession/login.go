package clearsession

import (
 "context"
 "database/sql"
 "errors"
 "time"
 "github.com/chrisbirster/cleartify/internal/notifications"
)

var ErrChallenge = errors.New("invalid or expired login challenge")
const challengeLifetime = 10*time.Minute

// Login is an opt-in, existing-account phone reauthentication service.
// The challenge is never returned to an HTTP client or stored in plaintext.
type Login struct {
 Sessions *Service
 Sender notifications.Sender
 Now func()time.Time
}
func (l *Login) clock()time.Time {if l.Now!=nil{return l.Now().UTC()};return time.Now().UTC()}
func (l *Login) Ready(ctx context.Context)error {
 _,err:=l.Sessions.Store.DB.ExecContext(ctx,`CREATE TABLE IF NOT EXISTS clear_login_challenges (
  user_id TEXT PRIMARY KEY, code_digest TEXT NOT NULL, expires_at TEXT NOT NULL,
  attempts INTEGER NOT NULL, created_at TEXT NOT NULL, consumed_at TEXT
 )`)
 return err
}
func(l *Login) Start(ctx context.Context,phone string) error{
 // Do not expose account existence to HTTP clients. Lookup is server-side.
 var user string
 err:=l.Sessions.Store.DB.QueryRowContext(ctx,`SELECT p.user_id FROM phone_identities p JOIN users u ON u.id=p.user_id
 WHERE p.phone_e164=? AND u.status='active'`,phone).Scan(&user)
 if errors.Is(err,sql.ErrNoRows){return nil};if err!=nil{return err}
 now:=l.clock()
 var created string
 e:=l.Sessions.Store.DB.QueryRowContext(ctx,`SELECT created_at FROM clear_login_challenges WHERE user_id=?`,user).Scan(&created)
 if e==nil {at,parseErr:=time.Parse(time.RFC3339Nano,created);if parseErr==nil&&now.Sub(at)<time.Minute{return nil}}
 if e!=nil&&!errors.Is(e,sql.ErrNoRows){return e}
 code,err:=l.Sessions.Tokens.FiveDigitCode();if err!=nil{return err}
 // Send first: on delivery failure a previously valid challenge remains unchanged.
 if err=l.Sender.SendSMS(ctx,notifications.SMS{To:phone,Text:"Your Cleartify sign-in code is "+code+". It expires in 10 minutes."});err!=nil{return err}
 stamp:=now.Format(time.RFC3339Nano)
 _,err=l.Sessions.Store.DB.ExecContext(ctx,`INSERT INTO clear_login_challenges(user_id,code_digest,expires_at,attempts,created_at,consumed_at)
 VALUES(?,?,?,0,?,NULL) ON CONFLICT(user_id) DO UPDATE SET code_digest=excluded.code_digest,
 expires_at=excluded.expires_at,attempts=0,created_at=excluded.created_at,consumed_at=NULL`,
 user,l.Sessions.Tokens.Digest(code),now.Add(challengeLifetime).Format(time.RFC3339Nano),stamp)
 if err!=nil{return err}
 return l.Sessions.Store.Push(ctx)
}
func(l *Login) Verify(ctx context.Context,phone,code string)(string,error){
 var user string
 err:=l.Sessions.Store.DB.QueryRowContext(ctx,`SELECT p.user_id FROM phone_identities p JOIN users u ON u.id=p.user_id WHERE p.phone_e164=? AND u.status='active'`,phone).Scan(&user)
 if err!=nil{return "",ErrChallenge}
 tx,err:=l.Sessions.Store.DB.BeginTx(ctx,nil);if err!=nil{return "",err}
 defer tx.Rollback()
 var digest,expires string
 var attempts int
 err=tx.QueryRowContext(ctx,`SELECT code_digest,expires_at,attempts FROM clear_login_challenges WHERE user_id=? AND consumed_at IS NULL`,user).Scan(&digest,&expires,&attempts)
 if err!=nil{return "",ErrChallenge}
 now:=l.clock();expiry,err:=time.Parse(time.RFC3339Nano,expires)
 if err!=nil||!now.Before(expiry)||attempts>=5{return "",ErrChallenge}
 valid:=len(code)==5&&l.Sessions.Tokens.Matches(code,digest)
 if !valid {
  if _,err=tx.ExecContext(ctx,`UPDATE clear_login_challenges SET attempts=attempts+1 WHERE user_id=? AND consumed_at IS NULL`,user);err!=nil{return "",err}
  if err=tx.Commit();err!=nil{return "",err}
  _=l.Sessions.Store.Push(ctx)
  return "",ErrChallenge
 }
 result,err:=tx.ExecContext(ctx,`UPDATE clear_login_challenges SET consumed_at=? WHERE user_id=? AND consumed_at IS NULL AND code_digest=? AND attempts<5 AND expires_at>?`,now.Format(time.RFC3339Nano),user,digest,now.Format(time.RFC3339Nano))
 if err!=nil{return "",err}
 affected,err:=result.RowsAffected();if err!=nil||affected!=1{return "",ErrChallenge}
 if err=tx.Commit();err!=nil{return "",err}
 if err=l.Sessions.Store.Push(ctx);err!=nil{return "",err}
 return l.Sessions.Issue(ctx,user)
}
