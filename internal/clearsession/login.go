package clearsession

import (
 "context"
 "database/sql"
 "errors"
 "regexp"
 "strings"
 "time"

 "github.com/chrisbirster/cleartify/internal/notifications"
)

var ErrChallenge = errors.New("invalid or expired login challenge")
var ErrLoginThrottled = errors.New("login temporarily unavailable")
const challengeLifetime = 10*time.Minute
const throttleWindow = time.Hour
const maxAttempts = 5
const maxSends = 5

var e164 = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)

// NormalizePhone accepts canonical E.164 only, with optional surrounding space.
// No punctuation, national prefixes, or ambiguous formatting is accepted.
func NormalizePhone(input string)(string,error){
 phone:=strings.TrimSpace(input)
 if !e164.MatchString(phone){return "",ErrChallenge}
 return phone,nil
}

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
 if err!=nil{return err}
 _,err=l.Sessions.Store.DB.ExecContext(ctx,`CREATE TABLE IF NOT EXISTS clear_login_throttles (
  phone_digest TEXT PRIMARY KEY, user_id TEXT, window_start TEXT NOT NULL,
  failed_attempts INTEGER NOT NULL DEFAULT 0, send_count INTEGER NOT NULL DEFAULT 0,
  last_sent_at TEXT, locked_until TEXT
 )`)
 return err
}
func (l *Login) phoneDigest(phone string)string{return l.Sessions.Tokens.Digest("clear-login-phone:"+phone)}
func(l *Login) throttle(ctx context.Context,phone,user string,send bool)error{
 now:=l.clock()
 // Persistent shared counters survive resend, challenge consumption and restart.
 tx,err:=l.Sessions.Store.DB.BeginTx(ctx,nil);if err!=nil{return err};defer tx.Rollback()
 var start string
 var sent sql.NullString
 var locked sql.NullString
 var fails,sends int
 digest:=l.phoneDigest(phone)
 err=tx.QueryRowContext(ctx,`SELECT window_start,failed_attempts,send_count,last_sent_at,locked_until
 FROM clear_login_throttles WHERE phone_digest=?`,digest).Scan(&start,&fails,&sends,&sent,&locked)
 if err!=nil&&!errors.Is(err,sql.ErrNoRows){return err}
 startTime,_:=time.Parse(time.RFC3339Nano,start)
 if start==""||now.Sub(startTime)>=throttleWindow {start=now.Format(time.RFC3339Nano);fails=0;sends=0;sent=sql.NullString{}}
 if locked.Valid {
  until,e:=time.Parse(time.RFC3339Nano,locked.String)
  if e!=nil{return ErrLoginThrottled}
  if now.Before(until){return ErrLoginThrottled}
 }
 if user!="" {
  var userFails,userSends int
  err=tx.QueryRowContext(ctx,`SELECT COALESCE(SUM(failed_attempts),0),COALESCE(SUM(send_count),0) FROM clear_login_throttles WHERE user_id=? AND window_start>?`,user,now.Add(-throttleWindow).Format(time.RFC3339Nano)).Scan(&userFails,&userSends)
  if err!=nil{return err}
  if userFails>=maxAttempts||userSends>=maxSends{return ErrLoginThrottled}
 }
 if fails>=maxAttempts||sends>=maxSends {return ErrLoginThrottled}
 if send && sent.Valid {
  prev,e:=time.Parse(time.RFC3339Nano,sent.String)
  if e==nil&&now.Sub(prev)<time.Minute{return ErrLoginThrottled}
 }
 if send{sends++}else{fails++}
 nextLock:=sql.NullString{}
 if fails>=maxAttempts {nextLock=sql.NullString{String:now.Add(throttleWindow).Format(time.RFC3339Nano),Valid:true}}
 _,err=tx.ExecContext(ctx,`INSERT INTO clear_login_throttles(phone_digest,user_id,window_start,failed_attempts,send_count,last_sent_at,locked_until)
 VALUES(?,?,?,?,?,?,?) ON CONFLICT(phone_digest) DO UPDATE SET
 user_id=excluded.user_id,window_start=excluded.window_start,
 failed_attempts=excluded.failed_attempts,send_count=excluded.send_count,
 last_sent_at=excluded.last_sent_at,locked_until=excluded.locked_until`,
 digest,user,start,fails,sends,func()any{if send{return now.Format(time.RFC3339Nano)};if sent.Valid{return sent.String};return nil}(),func()any{if nextLock.Valid{return nextLock.String};return nil}())
 if err!=nil{return err}
 if err=tx.Commit();err!=nil{return err}
 return l.Sessions.Store.Push(ctx)
}
func(l *Login) Start(ctx context.Context,rawPhone string) error{
 phone,err:=NormalizePhone(rawPhone);if err!=nil{return err}
 var user string
 err=l.Sessions.Store.DB.QueryRowContext(ctx,`SELECT p.user_id FROM phone_identities p JOIN users u ON u.id=p.user_id WHERE p.phone_e164=? AND u.status='active'`,phone).Scan(&user)
 if err!=nil&&!errors.Is(err,sql.ErrNoRows){return err}
 // Unknown phones go through identical persisted throttling, without sending.
 if err=l.throttle(ctx,phone,user,true);err!=nil{
  if errors.Is(err,ErrLoginThrottled){return nil}
  return err
 }
 if user==""{return nil}
 if l.Sender==nil{return notifications.ErrUnavailable}
 code,err:=l.Sessions.Tokens.FiveDigitCode();if err!=nil{return err}
 now:=l.clock()
 _,err=l.Sessions.Store.DB.ExecContext(ctx,`INSERT INTO clear_login_challenges(user_id,code_digest,expires_at,attempts,created_at,consumed_at)
 VALUES(?,?,?,0,?,NULL) ON CONFLICT(user_id) DO UPDATE SET code_digest=excluded.code_digest,
 expires_at=excluded.expires_at,attempts=clear_login_challenges.attempts,created_at=excluded.created_at,consumed_at=NULL`,
 user,l.Sessions.Tokens.Digest(code),now.Add(challengeLifetime).Format(time.RFC3339Nano),now.Format(time.RFC3339Nano))
 if err!=nil{return err}
 if err=l.Sessions.Store.Push(ctx);err!=nil{return err}
 return l.Sender.SendSMS(ctx,notifications.SMS{To:phone,Text:"Your Cleartify sign-in code is "+code+". It expires in 10 minutes."})
}
func(l *Login) Verify(ctx context.Context,rawPhone,code string)(string,error){
 phone,err:=NormalizePhone(rawPhone);if err!=nil{return "",ErrChallenge}
 var user string
 err=l.Sessions.Store.DB.QueryRowContext(ctx,`SELECT p.user_id FROM phone_identities p JOIN users u ON u.id=p.user_id WHERE p.phone_e164=? AND u.status='active'`,phone).Scan(&user)
 if err!=nil&&!errors.Is(err,sql.ErrNoRows){return "",err}
 // Every attempt, including unknown users and expired challenges, counts.
 if err=l.throttle(ctx,phone,user,false);err!=nil{
  if errors.Is(err,ErrLoginThrottled){return "",ErrChallenge}
  return "",err
 }
 if user==""{return "",ErrChallenge}
 tx,err:=l.Sessions.Store.DB.BeginTx(ctx,nil);if err!=nil{return "",err};defer tx.Rollback()
 var digest,expires string
 var attempts int
 err=tx.QueryRowContext(ctx,`SELECT code_digest,expires_at,attempts FROM clear_login_challenges WHERE user_id=? AND consumed_at IS NULL`,user).Scan(&digest,&expires,&attempts)
 if err!=nil{return "",ErrChallenge}
 now:=l.clock();expiry,e:=time.Parse(time.RFC3339Nano,expires)
 if e!=nil||!now.Before(expiry)||attempts>=maxAttempts{return "",ErrChallenge}
 valid:=len(code)==5&&l.Sessions.Tokens.Matches(code,digest)
 if !valid {
  if _,err=tx.ExecContext(ctx,`UPDATE clear_login_challenges SET attempts=attempts+1 WHERE user_id=? AND consumed_at IS NULL`,user);err!=nil{return "",err}
  if err=tx.Commit();err!=nil{return "",err}
  if err=l.Sessions.Store.Push(ctx);err!=nil{return "",err}
  return "",ErrChallenge
 }
 result,err:=tx.ExecContext(ctx,`UPDATE clear_login_challenges SET consumed_at=? WHERE user_id=? AND consumed_at IS NULL AND code_digest=? AND attempts<? AND expires_at>?`,now.Format(time.RFC3339Nano),user,digest,maxAttempts,now.Format(time.RFC3339Nano))
 if err!=nil{return "",err}
 affected,err:=result.RowsAffected();if err!=nil||affected!=1{return "",ErrChallenge}
 if err=tx.Commit();err!=nil{return "",err}
 if err=l.Sessions.Store.Push(ctx);err!=nil{return "",err}
 return l.Sessions.Issue(ctx,user)
}
