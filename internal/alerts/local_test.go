package alerts
import("context";"errors";"testing";"time")
func TestLocalHooks(t *testing.T){
 ctx:=context.Background();l:=NewLocal(1)
 event:=Event{Key:"failure",Stage:"local",At:time.Unix(1,0)}
 if err:=(Noop{}).Emit(ctx,event);err!=nil{t.Fatal(err)}
 if err:=l.Emit(ctx,Event{});!errors.Is(err,ErrInvalidEvent){t.Fatal("invalid accepted")}
 if err:=l.Emit(ctx,event);err!=nil{t.Fatal(err)}
 if err:=l.Emit(ctx,event);!errors.Is(err,ErrFull){t.Fatal("unbounded")}
 snap:=l.Snapshot();snap[0].Key="edited"
 if l.Snapshot()[0].Key!="failure"{t.Fatal("snapshot alias")}
 canceled,cancel:=context.WithCancel(ctx);cancel()
 if !errors.Is(l.Emit(canceled,event),context.Canceled){t.Fatal("cancel ignored")}
}
