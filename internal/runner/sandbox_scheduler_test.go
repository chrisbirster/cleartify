package runner
import("context";"errors";"testing";"time")
func TestLocalSandboxBoundaries(t *testing.T){
 s:=LocalSandbox{Enabled:true,Limits:DefaultLimits()}
 safe:=Plan{Steps:[]Step{{Operation:"local.set",Key:"a",Value:"ok"},{Operation:"local.assert",Key:"a",Value:"ok"}}}
 values,err:=s.Execute(context.Background(),safe);if err!=nil||values["a"]!="ok"{t.Fatalf("safe %v %v",values,err)}
 for _,op:=range []string{"shell.exec","network.fetch","http.get","fs.read","env.get","secret.read","local.eval","process.spawn"}{
  _,err=s.Execute(context.Background(),Plan{Steps:[]Step{{Operation:op,Key:"x",Value:"/etc/passwd"}}})
  if !errors.Is(err,ErrInvalid){t.Fatalf("escape %q: %v",op,err)}
 }
 small:=LocalSandbox{Enabled:true,Limits:Limits{MaxSteps:1,MaxBytes:3,Timeout:time.Second}}
 if _,err=small.Execute(context.Background(),safe);!errors.Is(err,ErrQuota){t.Fatalf("steps %v",err)}
 if _,err=small.Execute(context.Background(),Plan{Steps:[]Step{{Operation:"local.set",Key:"x",Value:"large"}}});!errors.Is(err,ErrQuota){t.Fatalf("bytes %v",err)}
 ctx,cancel:=context.WithCancel(context.Background());cancel()
 if _,err=s.Execute(ctx,safe);!errors.Is(err,context.Canceled){t.Fatalf("cancellation %v",err)}
 ctx2,cancel2:=context.WithDeadline(context.Background(),time.Now().Add(-time.Second));defer cancel2()
 if _,err=s.Execute(ctx2,safe);!errors.Is(err,context.DeadlineExceeded){t.Fatalf("deadline %v",err)}
 s.Enabled=false;if _,err=s.Execute(context.Background(),safe);!errors.Is(err,ErrDisabled){t.Fatal("off flag ignored")}
}
type flakyExecutor struct{calls int}
func(f *flakyExecutor)Execute(_ context.Context,_ Plan)(map[string]string,error){f.calls++;if f.calls==1{return nil,ErrAssertion};return map[string]string{"ok":"yes"},nil}
func TestSchedulerRetryBackoffAndFlag(t *testing.T){
 now:=time.Date(2026,10,13,0,0,0,0,time.UTC)
 f:=&flakyExecutor{}
 s:=&Scheduler{Enabled:true,Executor:f,MaxQueued:2,MaxRetries:2,Backoff:time.Second,Now:func()time.Time{return now}}
 p:=Plan{Steps:[]Step{{Operation:"local.set",Key:"ok",Value:"yes"}}}
 if err:=s.Submit("j1",p);err!=nil{t.Fatal(err)}
 if err:=s.Submit("j1",p);err!=nil{t.Fatal(err)}
 if did,err:=s.Tick(context.Background());!did||!errors.Is(err,ErrAssertion){t.Fatalf("first %v %v",did,err)}
 if did,err:=s.Tick(context.Background());did||err!=nil{t.Fatalf("backoff ignored %v %v",did,err)}
 now=now.Add(time.Second)
 if did,err:=s.Tick(context.Background());!did||err!=nil{t.Fatalf("retry %v %v",did,err)}
 j,ok:=s.Status("j1");if !ok||!j.Done||j.Attempts!=2{t.Fatalf("job %+v",j)}
 s.Enabled=false;if err:=s.Submit("j2",p);!errors.Is(err,ErrDisabled){t.Fatalf("flag %v",err)}
}
