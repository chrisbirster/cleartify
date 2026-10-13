// Package runner provides a deterministic, local-only M5 preview runner.
// It has no network, filesystem, process, payment or external side-effect capability.
package runner

import (
 "context"
 "errors"
 "fmt"
 "sync"
 "time"
)

var (
 ErrDisabled=errors.New("runner disabled")
 ErrInvalid=errors.New("invalid declarative plan")
 ErrNotFound=errors.New("run not found")
 ErrState=errors.New("invalid run state")
)
const MaxSteps=32
const MaxAttempts=3
const MaxValueBytes=4096

type Step struct {
 Operation string `json:"operation"`
 Key string `json:"key,omitempty"`
 Value string `json:"value,omitempty"`
}
type Plan struct {Steps []Step `json:"steps"`}
type AuditEvent struct {
 Sequence int `json:"sequence"`
 Action string `json:"action"`
 Step int `json:"step"`
 Attempt int `json:"attempt"`
 At time.Time `json:"at"`
}
type Snapshot struct {
 ID string `json:"id"`
 Status string `json:"status"`
 NextStep int `json:"nextStep"`
 Attempts int `json:"attempts"`
 Values map[string]string `json:"values"`
 Audit []AuditEvent `json:"audit"`
}
type record struct {
 plan Plan
 snapshot Snapshot
}
type Runner struct {
 mu sync.Mutex
 enabled bool
 now func()time.Time
 next uint64
 runs map[string]*record
}
// New is disabled by default. Only explicitly flagged local tests may enable it.
func New(enabled bool)*Runner{return &Runner{enabled:enabled,now:time.Now,runs:make(map[string]*record)}}
func validate(p Plan)error{
 if len(p.Steps)==0||len(p.Steps)>MaxSteps{return ErrInvalid}
 for _,s:=range p.Steps{
  if len(s.Key)==0||len(s.Key)>64||len(s.Value)>MaxValueBytes{return ErrInvalid}
  for _,c:=range s.Key {if !(c>='a'&&c<='z'||c>='A'&&c<='Z'||c>='0'&&c<='9'||c=='_'||c=='-'){return ErrInvalid}}
  switch s.Operation{
  case "local.set": // fake in-memory write only
  case "local.assert": // compare in-memory state only
  default:return ErrInvalid
  }
 }
 return nil
}
func(r *Runner)event(rec *record,action string){
 rec.snapshot.Audit=append(rec.snapshot.Audit,AuditEvent{Sequence:len(rec.snapshot.Audit)+1,Action:action,Step:rec.snapshot.NextStep,Attempt:rec.snapshot.Attempts,At:r.now().UTC()})
}
func clone(s Snapshot)Snapshot{
 out:=s
 out.Values=make(map[string]string,len(s.Values))
 for k,v:=range s.Values{out.Values[k]=v}
 out.Audit=append([]AuditEvent(nil),s.Audit...)
 return out
}
func(r *Runner)Start(p Plan)(Snapshot,error){
 r.mu.Lock();defer r.mu.Unlock()
 if !r.enabled{return Snapshot{},ErrDisabled}
 if err:=validate(p);err!=nil{return Snapshot{},err}
 r.next++
 id:=fmt.Sprintf("preview-%d",r.next)
 steps:=append([]Step(nil),p.Steps...)
 rec:=&record{plan:Plan{Steps:steps},snapshot:Snapshot{ID:id,Status:"pending",Values:map[string]string{}}}
 r.event(rec,"created");r.runs[id]=rec
 return clone(rec.snapshot),nil
}
func(r *Runner)Run(ctx context.Context,id string)(Snapshot,error){
 r.mu.Lock();defer r.mu.Unlock()
 if !r.enabled{return Snapshot{},ErrDisabled}
 rec,ok:=r.runs[id];if !ok{return Snapshot{},ErrNotFound}
 if rec.snapshot.Status!="pending"{return clone(rec.snapshot),ErrState}
 rec.snapshot.Status="running";r.event(rec,"started")
 for rec.snapshot.NextStep<len(rec.plan.Steps){
  if err:=ctx.Err();err!=nil{
   rec.snapshot.Status="cancelled";r.event(rec,"cancelled");return clone(rec.snapshot),err
  }
  step:=rec.plan.Steps[rec.snapshot.NextStep]
  switch step.Operation{
  case "local.set":rec.snapshot.Values[step.Key]=step.Value
  case "local.assert":
   if rec.snapshot.Values[step.Key]!=step.Value{
    rec.snapshot.Status="failed";rec.snapshot.Attempts++;r.event(rec,"assertion_failed")
    return clone(rec.snapshot),ErrState
   }
  }
  rec.snapshot.NextStep++;r.event(rec,"step_completed")
 }
 rec.snapshot.Status="completed";r.event(rec,"completed")
 return clone(rec.snapshot),nil
}
func(r *Runner)Cancel(id string)(Snapshot,error){
 r.mu.Lock();defer r.mu.Unlock()
 if !r.enabled{return Snapshot{},ErrDisabled}
 rec,ok:=r.runs[id];if !ok{return Snapshot{},ErrNotFound}
 if rec.snapshot.Status!="pending"&&rec.snapshot.Status!="failed"{return clone(rec.snapshot),ErrState}
 rec.snapshot.Status="cancelled";r.event(rec,"cancelled")
 return clone(rec.snapshot),nil
}
func(r *Runner)Retry(id string)(Snapshot,error){
 r.mu.Lock();defer r.mu.Unlock()
 if !r.enabled{return Snapshot{},ErrDisabled}
 rec,ok:=r.runs[id];if !ok{return Snapshot{},ErrNotFound}
 if rec.snapshot.Status!="failed"||rec.snapshot.Attempts>=MaxAttempts{return clone(rec.snapshot),ErrState}
 rec.snapshot.Status="pending";r.event(rec,"retry_queued")
 return clone(rec.snapshot),nil
}
func(r *Runner)Get(id string)(Snapshot,error){
 r.mu.Lock();defer r.mu.Unlock()
 if !r.enabled{return Snapshot{},ErrDisabled}
 rec,ok:=r.runs[id];if !ok{return Snapshot{},ErrNotFound}
 return clone(rec.snapshot),nil
}
