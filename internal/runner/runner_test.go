package runner

import (
 "context"
 "errors"
 "testing"
)
func TestDisabledAndValidation(t *testing.T){
 r:=New(false)
 if _,err:=r.Start(Plan{Steps:[]Step{{Operation:"local.set",Key:"a",Value:"b"}}});!errors.Is(err,ErrDisabled){t.Fatal(err)}
 r=New(true)
 for _,p:=range []Plan{{},{Steps:[]Step{{Operation:"http.get",Key:"a"}}},{Steps:[]Step{{Operation:"local.set",Key:"a/b"}}}}{
  if _,err:=r.Start(p);!errors.Is(err,ErrInvalid){t.Fatalf("expected invalid for %+v: %v",p,err)}
 }
}
func TestRunAuditIsolationAndCancel(t *testing.T){
 r:=New(true)
 s,err:=r.Start(Plan{Steps:[]Step{{Operation:"local.set",Key:"answer",Value:"42"},{Operation:"local.assert",Key:"answer",Value:"42"}}})
 if err!=nil{t.Fatal(err)}
 done,err:=r.Run(context.Background(),s.ID);if err!=nil||done.Status!="completed"||done.Values["answer"]!="42"||len(done.Audit)!=5{t.Fatalf("%+v %v",done,err)}
 done.Values["answer"]="tampered";done.Audit[0].Action="tampered"
 again,_:=r.Get(s.ID);if again.Values["answer"]!="42"||again.Audit[0].Action!="created"{t.Fatal("snapshot leaked internal state")}
 if _,err:=r.Cancel(s.ID);!errors.Is(err,ErrState){t.Fatal(err)}
 pending,_:=r.Start(Plan{Steps:[]Step{{Operation:"local.set",Key:"x",Value:"y"}}})
 cancelled,err:=r.Cancel(pending.ID);if err!=nil||cancelled.Status!="cancelled"{t.Fatalf("%+v %v",cancelled,err)}
 if _,err:=r.Run(context.Background(),pending.ID);!errors.Is(err,ErrState){t.Fatal(err)}
}
func TestRetriesBoundedAndContextCancellation(t *testing.T){
 r:=New(true)
 s,_:=r.Start(Plan{Steps:[]Step{{Operation:"local.assert",Key:"missing",Value:"present"}}})
 for attempt:=1;attempt<=MaxAttempts;attempt++{
  snap,err:=r.Run(context.Background(),s.ID)
  if !errors.Is(err,ErrState)||snap.Status!="failed"||snap.Attempts!=attempt{t.Fatalf("%+v %v",snap,err)}
  if attempt<MaxAttempts{if _,err=r.Retry(s.ID);err!=nil{t.Fatal(err)}}
 }
 if _,err:=r.Retry(s.ID);!errors.Is(err,ErrState){t.Fatal("unbounded retry",err)}
 s,_=r.Start(Plan{Steps:[]Step{{Operation:"local.set",Key:"x",Value:"y"}}})
 ctx,cancel:=context.WithCancel(context.Background());cancel()
 snap,err:=r.Run(ctx,s.ID)
 if !errors.Is(err,context.Canceled)||snap.Status!="cancelled"||len(snap.Values)!=0{t.Fatalf("%+v %v",snap,err)}
}
