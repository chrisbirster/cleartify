package followups

import(
 "context"
 "database/sql"
 "github.com/chrisbirster/cleartify/internal/database"
 "errors"
 "fmt"
 "testing"
 "time"
 "github.com/chrisbirster/cleartify/internal/notifications"
)
type failingSender struct{calls int}
func(f *failingSender)SendSMS(context.Context,notifications.SMS)error{f.calls++;return errors.New("secret provider error")}
func(f *failingSender)SendEmail(context.Context,notifications.Email)error{return nil}
func TestWorkerFakeClockGracefulStopRetryAndLimits(t *testing.T){
 ctx:=context.Background()
 s,db:=testWorkerStore(t)
 clock:=time.Date(2026,10,27,12,0,0,0,time.UTC)
 s.Now=func()time.Time{return clock}
 sender:=&failingSender{}
 s.Sender=sender
 s.Channel="sms"
 for i:=0;i<MaxCreatedPerDay;i++{_,err:=s.Create(ctx,"b","step",clock);if err!=nil{t.Fatal(err)}}
 if _,err:=s.Create(ctx,"b","too many",clock);!errors.Is(err,ErrRateLimit){t.Fatal("missing creation throttle",err)}
 ticks:=make(chan time.Time)
 workCtx,cancel:=context.WithCancel(ctx)
 stopped:=make(chan error,1)
 go func(){stopped<-s.Run(workCtx,ticks)}()
 ticks<-clock
 // Verify processing is scheduled, irrespective of async worker timing.
 cancel()
 select{case err:=<-stopped:if err!=nil{t.Fatal(err)};case <-time.After(3*time.Second):t.Fatal("worker ignored cancellation")}
 count,err:=s.Tick(ctx);if err!=nil{t.Fatal(err)}
 if count<0{t.Fatal("bad count")}
 _,err=s.DeliverDue(ctx);if err!=nil{t.Fatal(err)}
 if sender.calls==0{t.Fatal("Sender never invoked")}
 var status string;var attempts int
 if err=db.QueryRowContext(ctx,"SELECT status,attempts FROM clear_followup_delivery LIMIT 1").Scan(&status,&attempts);err!=nil{t.Fatal(err)}
 if status!="retry"||attempts!=1{t.Fatalf("retry state: %s %d",status,attempts)}
 clock=clock.Add(3*time.Minute)
 _,err=s.DeliverDue(ctx);if err!=nil{t.Fatal(err)}
 if sender.calls<2{t.Fatal("backoff retry missing")}
 for i:=2;i<MaxDeliveryAttempts;i++{
  clock=clock.Add(2*time.Hour)
  if _,err=s.DeliverDue(ctx);err!=nil{t.Fatal(err)}
 }
 var dead int
 if err=db.QueryRowContext(ctx,"SELECT COUNT(*) FROM clear_followup_delivery WHERE status='dead' AND attempts=?",MaxDeliveryAttempts).Scan(&dead);err!=nil{t.Fatal(err)}
 if dead!=MaxCreatedPerDay{t.Fatalf("expected all terminal dead deliveries, got %d",dead)}
 calls:=sender.calls
 clock=clock.Add(48*time.Hour)
 if _,err=s.DeliverDue(ctx);err!=nil{t.Fatal(err)}
 if sender.calls!=calls{t.Fatal("terminal dead message sent again")}
}
// store helper

func testWorkerStore(t *testing.T)(Service,*sql.DB){
 t.Helper()
 ctx:=context.Background()
 db,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)}
 t.Cleanup(func(){_ =db.Close()})
 if err=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 _,err=db.DB.ExecContext(ctx,`INSERT INTO users(id,display_name,primary_email,status,created_at,updated_at) VALUES('b','B','b@test.invalid','active','now','now')`)
 if err!=nil{t.Fatal(err)}
 _,err=db.DB.ExecContext(ctx,`INSERT INTO phone_identities(id,user_id,phone_e164,verified_at,created_at) VALUES('b-phone','b','+15550110000','2026-10-27T12:00:00Z','2026-10-27T12:00:00Z')`)
 if err!=nil{t.Fatal(err)}
 s:=Service{DB:db.DB}
 if err=s.Ready(ctx);err!=nil{t.Fatal(err)}
 return s,db.DB
}

func TestThirtyNudgesAllowedThirtyFirstRejected(t *testing.T){
 ctx:=context.Background()
 s,db:=testWorkerStore(t)
 now:=time.Date(2026,10,27,12,0,0,0,time.UTC)
 s.Now=func()time.Time{return now}
 stamp:=now.Format(time.RFC3339Nano)
 for i:=0;i<31;i++{
  _,err:=db.ExecContext(ctx,`INSERT INTO clear_followups(id,user_id,text,due_at,status,created_at) VALUES(?, 'b','fixture',?,'open',?)`,fmt.Sprintf("boundary-%02d",i),stamp,stamp)
  if err!=nil{t.Fatal(err)}
 }
 admitted,err:=s.Tick(ctx);if err!=nil{t.Fatal(err)}
 if admitted!=30{t.Fatalf("30 should be allowed, got %d",admitted)}
 var count int
 if err=db.QueryRowContext(ctx,"SELECT COUNT(*) FROM clear_followup_nudges").Scan(&count);err!=nil{t.Fatal(err)}
 if count!=30{t.Fatalf("expected 30, got %d",count)}
 admitted,err=s.Tick(ctx);if err!=nil{t.Fatal(err)}
 if admitted!=0{t.Fatalf("31st should be rejected; inserted %d",admitted)}
 if err=db.QueryRowContext(ctx,"SELECT COUNT(*) FROM clear_followups f WHERE f.id NOT IN (SELECT followup_id FROM clear_followup_nudges)").Scan(&count);err!=nil{t.Fatal(err)}
 if count!=1{t.Fatalf("expected exactly one deferred nudge, got %d",count)}
}
