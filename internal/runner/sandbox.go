package runner

import (
 "context"
 "errors"
 "time"
)

// Executor accepts only the closed declarative Step vocabulary. It never
// interprets shell, scripts, URLs, paths, environment names or arbitrary code.
type Executor interface { Execute(context.Context,Plan)(map[string]string,error) }
var ErrQuota=errors.New("local sandbox quota exceeded")
var ErrAssertion=errors.New("local assertion failed")
type Limits struct {MaxSteps int;MaxBytes int;Timeout time.Duration}
func DefaultLimits() Limits{return Limits{MaxSteps:32,MaxBytes:16*1024,Timeout:time.Second}}
type LocalSandbox struct{Enabled bool;Limits Limits}
func(s LocalSandbox)Execute(parent context.Context,p Plan)(map[string]string,error){
 if !s.Enabled{return nil,ErrDisabled}
 l:=s.Limits
 if l.MaxSteps<=0||l.MaxSteps>MaxSteps||l.MaxBytes<=0||l.MaxBytes>1024*1024||l.Timeout<=0||l.Timeout>30*time.Second{return nil,ErrInvalid}
 if err:=validate(p);err!=nil{return nil,err}
 if len(p.Steps)>l.MaxSteps{return nil,ErrQuota}
 ctx,cancel:=context.WithTimeout(parent,l.Timeout);defer cancel()
 values:=make(map[string]string)
 bytes:=0
 for _,step:=range p.Steps{
  if err:=ctx.Err();err!=nil{return nil,err}
  switch step.Operation{
  case "local.set":
   previous,exists:=values[step.Key]
   next:=bytes+len(step.Key)+len(step.Value)
   if exists{next=bytes-len(step.Key)-len(previous)+len(step.Key)+len(step.Value)}
   if next>l.MaxBytes{return nil,ErrQuota}
   bytes=next;values[step.Key]=step.Value
  case "local.assert":
   if values[step.Key]!=step.Value{return nil,ErrAssertion}
  default:return nil,ErrInvalid
  }
 }
 return values,nil
}
