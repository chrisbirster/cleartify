package alerts

import (
 "context"
 "errors"
 "sync"
 "time"
)

// Event is an internal typed signal, never an outbound notification.
type Event struct {
 Key string
 Stage string
 At time.Time
}
type Sink interface { Emit(context.Context, Event) error }
type Noop struct{}
func(Noop)Emit(context.Context,Event)error{return nil}
var ErrInvalidEvent=errors.New("invalid local alert event")
var ErrFull=errors.New("local alert buffer full")
type Local struct { mu sync.Mutex; events []Event; limit int }
func NewLocal(limit int)*Local{return &Local{limit:limit}}
func(l *Local)Emit(ctx context.Context,e Event)error{
 if err:=ctx.Err();err!=nil{return err}
 if e.Key==""||e.Stage==""||e.At.IsZero(){return ErrInvalidEvent}
 l.mu.Lock();defer l.mu.Unlock()
 if l.limit<=0||len(l.events)>=l.limit{return ErrFull}
 l.events=append(l.events,e)
 return nil
}
func(l *Local)Snapshot()[]Event{l.mu.Lock();defer l.mu.Unlock();return append([]Event(nil),l.events...)}
