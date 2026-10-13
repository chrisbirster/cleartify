package clearsession
import("context";"errors";"testing";"time";"github.com/chrisbirster/cleartify/internal/database";"github.com/chrisbirster/cleartify/internal/security")
func TestSessionIdleExpiryRevocationAndFreshnessBoundary(t *testing.T){
 ctx:=context.Background();d,e:=database.Open(ctx,database.Config{Path:":memory:"});if e!=nil{t.Fatal(e)};defer d.Close()
 if e=d.Migrate(ctx);e!=nil{t.Fatal(e)}
 _,e=d.DB.ExecContext(ctx,`INSERT INTO users(id,display_name,primary_email,status,created_at,updated_at) VALUES('u','User','u@local.test','active','now','now')`);if e!=nil{t.Fatal(e)}
 tokens,e:=security.NewTokens("01234567890123456789012345678901");if e!=nil{t.Fatal(e)}
 now:=time.Date(2026,10,13,12,0,0,0,time.UTC)
 s:=&Service{Store:d,Tokens:tokens,Now:func()time.Time{return now}}
 raw,e:=s.Issue(ctx,"u");if e!=nil{t.Fatal(e)}
 if _,e=s.AuthenticateFresh(ctx,raw,10*time.Minute);e!=nil{t.Fatal(e)}
 now=now.Add(10*time.Minute+time.Nanosecond)
 if _,e=s.AuthenticateFresh(ctx,raw,10*time.Minute);!errors.Is(e,ErrUnauthenticated){t.Fatalf("aged authentication accepted: %v",e)}
 now=now.Add(IdleLimit+time.Nanosecond)
 if _,e=s.Authenticate(ctx,raw);!errors.Is(e,ErrUnauthenticated){t.Fatalf("idle expired token accepted: %v",e)}
 fresh,e:=s.Issue(ctx,"u");if e!=nil{t.Fatal(e)}
 if e=s.RevokeUser(ctx,"u");e!=nil{t.Fatal(e)}
 if _,e=s.Authenticate(ctx,fresh);!errors.Is(e,ErrUnauthenticated){t.Fatalf("revoked user session accepted: %v",e)}
}
