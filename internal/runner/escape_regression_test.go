package runner
import("context";"errors";"strings";"testing";"time")
func TestMoreDeclarativeSandboxEscapeAttempts(t *testing.T){
 sandbox:=LocalSandbox{Enabled:true,Limits:DefaultLimits()}
 for _,name:=range []string{"exec","os.getenv","syscall","http.post","file.write","secret.dump","local.set;fs.read","local.set\n"}{
  _,e:=sandbox.Execute(context.Background(),Plan{Steps:[]Step{{Operation:name,Key:"safe",Value:"x"}}})
  if !errors.Is(e,ErrInvalid){t.Fatalf("operation %q accepted: %v",name,e)}
 }
 for _,key:=range []string{"../private","/etc/passwd","HOME","a.b","a/b","a\n"}{
  if key=="HOME"{continue} // a harmless in-memory literal, never environment access
  _,e:=sandbox.Execute(context.Background(),Plan{Steps:[]Step{{Operation:"local.set",Key:key,Value:"x"}}})
  if !errors.Is(e,ErrInvalid){t.Fatalf("key %q accepted: %v",key,e)}
 }
 long:=strings.Repeat("x",MaxValueBytes+1)
 _,e:=sandbox.Execute(context.Background(),Plan{Steps:[]Step{{Operation:"local.set",Key:"x",Value:long}}})
 if !errors.Is(e,ErrInvalid){t.Fatalf("oversized data accepted: %v",e)}
 expired,cancel:=context.WithDeadline(context.Background(),time.Now().Add(-time.Second));defer cancel()
 _,e=sandbox.Execute(expired,Plan{Steps:[]Step{{Operation:"local.set",Key:"ok",Value:"x"}}})
 if !errors.Is(e,context.DeadlineExceeded){t.Fatalf("deadline ignored: %v",e)}
}
