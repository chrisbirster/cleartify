package clearsession

import (
 "context"
 "database/sql"
 "errors"
 "time"

 "github.com/chrisbirster/cleartify/internal/database"
 "github.com/chrisbirster/cleartify/internal/security"
)

var ErrUnauthenticated = errors.New("Clear session invalid or expired")
const IdleLimit = 30 * 24 * time.Hour

// Service stores only keyed token digests. A session is never an admin credential.
type Service struct {
 Store *database.Store
 Tokens *security.Tokens
 Now func() time.Time
}

func (s *Service) clock() time.Time {
 if s.Now != nil {return s.Now().UTC()}
 return time.Now().UTC()
}
func (s *Service) Issue(ctx context.Context,userID string)(string,error){
 var status string
 if err:=s.Store.DB.QueryRowContext(ctx,"SELECT status FROM users WHERE id=?",userID).Scan(&status);err!=nil{return "",err}
 if status!="active"{return "",ErrUnauthenticated}
 raw,err:=s.Tokens.RandomURLToken(32);if err!=nil{return "",err}
 now:=s.clock()
 _,err=s.Store.DB.ExecContext(ctx,`INSERT INTO clear_user_sessions(token_digest,user_id,created_at,last_seen_at,expires_at) VALUES(?,?,?,?,?)`,s.Tokens.Digest(raw),userID,now.Format(time.RFC3339Nano),now.Format(time.RFC3339Nano),now.Add(IdleLimit).Format(time.RFC3339Nano))
 if err!=nil{return "",err}
 if err=s.Store.Push(ctx);err!=nil{return "",err}
 return raw,nil
}
func(s *Service) Authenticate(ctx context.Context,raw string)(string,error){
 if len(raw)<32{return "",ErrUnauthenticated}
 now:=s.clock()
 var id,expires string
 err:=s.Store.DB.QueryRowContext(ctx,`SELECT s.user_id,s.expires_at FROM clear_user_sessions s JOIN users u ON u.id=s.user_id WHERE s.token_digest=? AND s.revoked_at IS NULL AND u.status='active'`,s.Tokens.Digest(raw)).Scan(&id,&expires)
 if errors.Is(err,sql.ErrNoRows){return "",ErrUnauthenticated}
 if err!=nil{return "",err}
 until,err:=time.Parse(time.RFC3339Nano,expires);if err!=nil||!now.Before(until){return "",ErrUnauthenticated}
 // Conditional update guarantees a concurrent revocation cannot silently extend a session.
 res,err:=s.Store.DB.ExecContext(ctx,`UPDATE clear_user_sessions SET last_seen_at=?,expires_at=? WHERE token_digest=? AND revoked_at IS NULL AND expires_at>?`,now.Format(time.RFC3339Nano),now.Add(IdleLimit).Format(time.RFC3339Nano),s.Tokens.Digest(raw),now.Format(time.RFC3339Nano))
 if err!=nil{return "",err}
 n,err:=res.RowsAffected();if err!=nil{return "",err};if n!=1{return "",ErrUnauthenticated}
 if err:=s.Store.Push(ctx);err!=nil{return "",err}
 return id,nil
}
func(s *Service) Revoke(ctx context.Context,raw string)error{
 if raw==""{return nil}
 now:=s.clock().Format(time.RFC3339Nano)
 _,err:=s.Store.DB.ExecContext(ctx,`UPDATE clear_user_sessions SET revoked_at=? WHERE token_digest=? AND revoked_at IS NULL`,now,s.Tokens.Digest(raw))
 if err!=nil{return err}
 return s.Store.Push(ctx)
}
func(s *Service) RevokeUser(ctx context.Context,userID string)error{
 now:=s.clock().Format(time.RFC3339Nano)
 _,err:=s.Store.DB.ExecContext(ctx,`UPDATE clear_user_sessions SET revoked_at=? WHERE user_id=? AND revoked_at IS NULL`,now,userID)
 if err!=nil{return err}
 return s.Store.Push(ctx)
}

 // AuthenticateFresh requires a newly issued session, not one merely extended by idle activity.
func(s *Service) AuthenticateFresh(ctx context.Context,raw string,maxAge time.Duration)(string,error){
 id,err:=s.Authenticate(ctx,raw);if err!=nil{return "",err}
 var issued string
 err=s.Store.DB.QueryRowContext(ctx,`SELECT created_at FROM clear_user_sessions WHERE token_digest=? AND user_id=? AND revoked_at IS NULL`,s.Tokens.Digest(raw),id).Scan(&issued)
 if err!=nil{return "",ErrUnauthenticated}
 at,err:=time.Parse(time.RFC3339Nano,issued)
 now:=s.clock()
 if err!=nil||at.After(now)||now.Sub(at)>maxAge{return "",ErrUnauthenticated}
 return id,nil
}
