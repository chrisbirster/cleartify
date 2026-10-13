package runner
import("context";"sync";"testing";"time")
func TestConcurrentScheduleIdempotencyAndPlanCopy(t *testing.T){
 now:=time.Date(2026,10,13,0,0,0,0,time.UTC)
 s:=&Scheduler{Enabled:true,Executor:LocalSandbox{Enabled:true,Limits:DefaultLimits()},MaxQueued:2,MaxRetries:1,Backoff:time.Second,Now:func()time.Time{return now}}
 plan:=Plan{Steps:[]Step{{Operation:"local.set",Key:"state",Value:"original"}}}
 var wg sync.WaitGroup;errs:=make([]error,32)
 for i:=range errs{wg.Add(1);go func(i int){defer wg.Done();errs[i]=s.Submit("same-id",plan)}(i)};wg.Wait()
 for _,e:=range errs{if e!=nil{t.Fatal(e)}}
 plan.Steps[0].Value="mutated"
 jobs:=make([]bool,16)
 for i:=range jobs{wg.Add(1);go func(i int){defer wg.Done();did,e:=s.Tick(context.Background());if e!=nil{t.Error(e)};jobs[i]=did}(i)};wg.Wait()
 count:=0;for _,v:=range jobs{if v{count++}}
 if count!=1{t.Fatalf("duplicate execution count=%d",count)}
 job,ok:=s.Status("same-id");if !ok||!job.Done||job.Attempts!=1||job.Plan.Steps[0].Value!="original"{t.Fatalf("lost plan copy %+v",job)}
}
