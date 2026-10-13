package clearsession

import (
 "context"
 "errors"
 "testing"
 "time"
 "github.com/chrisbirster/cleartify/internal/database"
 "github.com/chrisbirster/cleartify/internal/security"
)
func TestLifecycleAndActiveUser(t *testing.T){
 ctx:=context.Background()
 db,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)};defer db.Close()
 if err=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 now:=time.Date(2026,10,13,12,0,0,0,time.UTC)
 _,err=db.DB.ExecContext(ctx,`INSERT INTO users(id,display_name,primary_email,status,created_at,updated_at) VALUES('user1','Test','test@example.com','active',?,?)`,now.Format(time.RFC3339Nano),now.Format(time.RFC3339Nano));if err!=nil{t.Fatal(err)}
 tok,err:=security.NewTokens("01234567890123456789012345678901");if err!=nil{t.Fatal(err)}
 svc:=&Service{Store:db,Tokens:tok,Now:func()time.Time{return now}}
 raw,err:=svc.Issue(ctx,"user1");if err!=nil{t.Fatal(err)}
 if raw==""{t.Fatal("empty token")}
 var stored string
 if err=db.DB.QueryRowContext(ctx,"SELECT token_digest FROM clear_user_sessions").Scan(&stored);err!=nil{t.Fatal(err)}
 if raw==stored{t.Fatal("raw token stored")}
 if _,err=svc.Authenticate(ctx,"forged");!errors.Is(err,ErrUnauthenticated){t.Fatalf("forged: %v",err)}
 id,err:=svc.Authenticate(ctx,raw);if err!=nil||id!="user1"{t.Fatalf("auth %s %v",id,err)}
 now=now.Add(IdleLimit+time.Second)
 if _,err=svc.Authenticate(ctx,raw);!errors.Is(err,ErrUnauthenticated){t.Fatalf("expired: %v",err)}
 raw,err=svc.Issue(ctx,"user1");if err!=nil{t.Fatal(err)}
 if err=svc.Revoke(ctx,raw);err!=nil{t.Fatal(err)}
 if _,err=svc.Authenticate(ctx,raw);!errors.Is(err,ErrUnauthenticated){t.Fatalf("revoked: %v",err)}
 raw,err=svc.Issue(ctx,"user1");if err!=nil{t.Fatal(err)}
 if err=svc.RevokeUser(ctx,"user1");err!=nil{t.Fatal(err)}
 if _,err=svc.Authenticate(ctx,raw);!errors.Is(err,ErrUnauthenticated){t.Fatalf("revoked user: %v",err)}
}
