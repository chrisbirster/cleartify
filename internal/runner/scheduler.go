package runner
import("context";"errors";"sync";"time")
var ErrQueue=errors.New("local scheduler queue full or invalid")
type Job struct{ID string;Plan Plan;Attempts int;NextAt time.Time;Done bool;Failed bool}
type Scheduler struct{
 mu sync.Mutex
 Enabled bool
 Executor Executor
 MaxQueued int
 MaxRetries int
 Backoff time.Duration
 Now func()time.Time
 jobs map[string]*Job
 order []string
}
func(s *Scheduler)clock()time.Time{if s.Now!=nil{return s.Now().UTC()};return time.Now().UTC()}
func(s *Scheduler)Submit(id string,p Plan)error{
 s.mu.Lock();defer s.mu.Unlock()
 if !s.Enabled{return ErrDisabled}
 if s.Executor==nil||s.MaxQueued<=0||s.MaxQueued>1024||s.MaxRetries<0||s.MaxRetries>5||s.Backoff<=0||s.Backoff>time.Minute||id==""{return ErrQueue}
 if err:=validate(p);err!=nil{return err}
 if s.jobs==nil{s.jobs=map[string]*Job{}}
 if prior:=s.jobs[id];prior!=nil{return nil} // idempotent job key
 if len(s.order)>=s.MaxQueued{return ErrQueue}
 cloned:=Plan{Steps:append([]Step(nil),p.Steps...)}
 s.jobs[id]=&Job{ID:id,Plan:cloned,NextAt:s.clock()}
 s.order=append(s.order,id)
 return nil
}
// Tick performs at most one due job; there is no goroutine, timer, network
// listener or automatic production scheduler. Caller controls polling.
func(s *Scheduler)Tick(ctx context.Context)(bool,error){
 s.mu.Lock();defer s.mu.Unlock()
 if !s.Enabled{return false,ErrDisabled}
 if s.Executor==nil{return false,ErrQueue}
 now:=s.clock()
 for _,id:=range s.order{
  j:=s.jobs[id]
  if j.Done||j.Failed||now.Before(j.NextAt){continue}
  _,err:=s.Executor.Execute(ctx,j.Plan);j.Attempts++
  if err==nil{j.Done=true;return true,nil}
  if errors.Is(err,ErrInvalid)||errors.Is(err,ErrQuota)||errors.Is(err,ErrDisabled)||errors.Is(err,context.Canceled)||errors.Is(err,context.DeadlineExceeded)||j.Attempts>s.MaxRetries{
   j.Failed=true;return true,err
  }
  delay:=s.Backoff
  for i:=1;i<j.Attempts;i++{delay*=2}
  j.NextAt=now.Add(delay)
  return true,err
 }
 return false,nil
}
func(s *Scheduler)Status(id string)(Job,bool){
 s.mu.Lock();defer s.mu.Unlock()
 j:=s.jobs[id];if j==nil{return Job{},false}
 cp:=*j;cp.Plan.Steps=append([]Step(nil),j.Plan.Steps...);return cp,true
}
