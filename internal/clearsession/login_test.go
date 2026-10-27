package clearsession

import (
 "context"
 "errors"
 "path/filepath"
 "strings"
 "testing"
 "time"
 "github.com/chrisbirster/cleartify/internal/database"
 "github.com/chrisbirster/cleartify/internal/notifications"
 "github.com/chrisbirster/cleartify/internal/security"
)
type captureLoginSMS struct{ sms []notifications.SMS }
func(s *captureLoginSMS)SendSMS(_ context.Context,m notifications.SMS)error{s.sms=append(s.sms,m);return nil}
func(s *captureLoginSMS)SendEmail(context.Context,notifications.Email)error{return nil}
func TestReturningLoginReplayExpiryIsolationRestart(t *testing.T){
 ctx:=context.Background();path:=filepath.Join(t.TempDir(),"login.db")
 db,err:=database.Open(ctx,database.Config{Path:path});if err!=nil{t.Fatal(err)}
 if err=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 now:=time.Date(2026,10,27,12,0,0,0,time.UTC)
 for _,u:=range []struct{id,phone string}{{"a","+15550000001"},{"b","+15550000002"}} {
  _,err=db.DB.ExecContext(ctx,`INSERT INTO users(id,display_name,primary_email,status,created_at,updated_at) VALUES(?,?,?,'active',?,?)`,u.id,u.id,u.id+"@example.com",now.Format(time.RFC3339Nano),now.Format(time.RFC3339Nano));if err!=nil{t.Fatal(err)}
  _,err=db.DB.ExecContext(ctx,`INSERT INTO phone_identities(id,user_id,phone_e164,verified_at,created_at) VALUES(?,?,?,?,?)`,u.id,u.id,u.phone,now.Format(time.RFC3339Nano),now.Format(time.RFC3339Nano));if err!=nil{t.Fatal(err)}
 }
 tokens,err:=security.NewTokens(strings.Repeat("x",32));if err!=nil{t.Fatal(err)}
 sender:=&captureLoginSMS{}
 sessions:=&Service{Store:db,Tokens:tokens,Now:func()time.Time{return now}}
 login:=&Login{Sessions:sessions,Sender:sender,Now:func()time.Time{return now}}
 if err=login.Ready(ctx);err!=nil{t.Fatal(err)}
 if err=login.Start(ctx,"+15550000001");err!=nil{t.Fatal(err)}
 if len(sender.sms)!=1{t.Fatal("missing notification")}
 code:=strings.TrimSuffix(strings.Split(sender.sms[0].Text," ")[5],".");if len(code)!=5{t.Fatalf("bad code: %q",code)}
 if _,err=login.Verify(ctx,"+15550000002",code);!errors.Is(err,ErrChallenge){t.Fatal("cross-account accepted",err)}
 if _,err=login.Verify(ctx,"+15550000001","00000");!errors.Is(err,ErrChallenge){t.Fatal("wrong code accepted",err)}
 if err=db.Close();err!=nil{t.Fatal(err)}
 db,err=database.Open(ctx,database.Config{Path:path});if err!=nil{t.Fatal(err)};defer db.Close()
 sessions.Store=db
 token,err:=login.Verify(ctx,"+15550000001",code);if err!=nil{t.Fatal("restart verification",err)}
 id,err:=sessions.Authenticate(ctx,token);if err!=nil||id!="a"{t.Fatal("wrong user",id,err)}
 if _,err=login.Verify(ctx,"+15550000001",code);!errors.Is(err,ErrChallenge){t.Fatal("replay accepted",err)}
 now=now.Add(time.Minute*11)
 if err=login.Start(ctx,"+15550000002");err!=nil{t.Fatal(err)}
 code2:=strings.TrimSuffix(strings.Split(sender.sms[len(sender.sms)-1].Text," ")[5],".")
 now=now.Add(time.Minute*11)
 if _,err=login.Verify(ctx,"+15550000002",code2);!errors.Is(err,ErrChallenge){t.Fatal("expired accepted",err)}
}

func TestLoginThrottleSurvivesResendAndPhoneNormalization(t *testing.T){
 ctx:=context.Background();db,err:=database.Open(ctx,database.Config{Path:filepath.Join(t.TempDir(),"limits.db")});if err!=nil{t.Fatal(err)};defer db.Close()
 if err=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 now:=time.Date(2026,10,27,12,0,0,0,time.UTC)
 stamp:=now.Format(time.RFC3339Nano)
 if _,err=db.DB.ExecContext(ctx,`INSERT INTO users(id,display_name,primary_email,status,created_at,updated_at) VALUES('limit','Limit','limit@test.invalid','active',?,?)`,stamp,stamp);err!=nil{t.Fatal(err)}
 if _,err=db.DB.ExecContext(ctx,`INSERT INTO phone_identities(id,user_id,phone_e164,verified_at,created_at) VALUES('limit','limit','+15550000099',?,?)`,stamp,stamp);err!=nil{t.Fatal(err)}
 tokens,err:=security.NewTokens(strings.Repeat("t",32));if err!=nil{t.Fatal(err)}
 sender:=&captureLoginSMS{};login:=&Login{Sessions:&Service{Store:db,Tokens:tokens,Now:func()time.Time{return now}},Sender:sender,Now:func()time.Time{return now}}
 if err=login.Ready(ctx);err!=nil{t.Fatal(err)}
 for _,invalid:=range []string{"5550000099","+05550000099","+1-5550000099","+1a555000099","+1234567890123456"}{
  if _,err=NormalizePhone(invalid);err==nil{t.Fatalf("accepted %q",invalid)}
 }
 phone,err:=NormalizePhone(" +15550000099 ");if err!=nil||phone!="+15550000099"{t.Fatalf("normalize %q %v",phone,err)}
 if err=login.Start(ctx,phone);err!=nil{t.Fatal(err)}
 for i:=0;i<4;i++ {if _,err=login.Verify(ctx,phone,"xxxxx");!errors.Is(err,ErrChallenge){t.Fatal(err)}}
 now=now.Add(2*time.Minute)
 if err=login.Start(ctx,phone);err!=nil{t.Fatal(err)}
 if len(sender.sms)!=2{t.Fatal("resend not attempted")}
 // Resend must not forgive previously failed attempts.
 if _,err=login.Verify(ctx,phone,"xxxxx");!errors.Is(err,ErrChallenge){t.Fatal(err)}
 good:=strings.TrimSuffix(strings.Split(sender.sms[1].Text," ")[5],".")
 if _,err=login.Verify(ctx,phone,good);!errors.Is(err,ErrChallenge){t.Fatal("lockout bypassed",err)}
 for i:=0;i<8;i++{now=now.Add(time.Minute);if err=login.Start(ctx,phone);err!=nil{t.Fatal(err)}}
 if len(sender.sms)!=2{t.Fatal("lockout bypassed via resend")}
}
