package atproto

import(
 "context"
 "errors"
 "strings"
 "sync"
 "time"
)
// FakePDS is an in-process test boundary, not an ATProto wire implementation.
// No HTTP client, listener, discovery, DNS, socket or real PDS call exists.
var ErrFakeLifecycle=errors.New("fake OAuth session invalid, expired or revoked")
type FakeLifecycleProvider interface{
 Refresh(context.Context,string,string)(string,time.Time,error)
 ResolveHandle(context.Context,string)(string,error)
 Revoke(context.Context,string)error
}
type fakeSession struct{did,refresh string;expiry time.Time;revoked bool}
type InProcessFakePDS struct{
 mu sync.Mutex
 sessions map[string]fakeSession
 handles map[string]string
 sequence uint64
 Now func()time.Time
}
func NewInProcessFakePDS()*InProcessFakePDS{return &InProcessFakePDS{sessions:map[string]fakeSession{},handles:map[string]string{}}}
func(p *InProcessFakePDS)clock()time.Time{if p.Now!=nil{return p.Now().UTC()};return time.Now().UTC()}
func(p *InProcessFakePDS)Issue(did,handle string,ttl time.Duration)(string,error){
 p.mu.Lock();defer p.mu.Unlock()
 if !strings.HasPrefix(did,"did:")||handle==""||ttl<=0||ttl>24*time.Hour{return "",ErrFakeLifecycle}
 p.sequence++;token:="fake-session-"+strings.Repeat("0",24)+"-"+decimal(p.sequence)
 p.sessions[token]=fakeSession{did:did,refresh:"fake-refresh-"+decimal(p.sequence),expiry:p.clock().Add(ttl)}
 p.handles[handle]=did
 return token,nil
}
func decimal(n uint64)string{
 if n==0{return "0"}
 out:="";for n>0{out=string(rune('0'+n%10))+out;n/=10};return out
}
func(p *InProcessFakePDS)Refresh(ctx context.Context,token,refresh string)(string,time.Time,error){
 if e:=ctx.Err();e!=nil{return "",time.Time{},e}
 p.mu.Lock();defer p.mu.Unlock()
 item,ok:=p.sessions[token]
 if !ok||item.revoked||refresh!=item.refresh||!p.clock().Before(item.expiry){return "",time.Time{},ErrFakeLifecycle}
 p.sequence++;newToken:="fake-session-"+strings.Repeat("0",24)+"-"+decimal(p.sequence)
 delete(p.sessions,token)
 item.refresh="fake-refresh-"+decimal(p.sequence)
 item.expiry=p.clock().Add(time.Hour)
 p.sessions[newToken]=item
 return newToken,item.expiry,nil
}
func(p *InProcessFakePDS)ResolveHandle(ctx context.Context,handle string)(string,error){
 if e:=ctx.Err();e!=nil{return "",e}
 p.mu.Lock();defer p.mu.Unlock()
 did,ok:=p.handles[handle];if !ok{return "",ErrFakeLifecycle}
 return did,nil
}
func(p *InProcessFakePDS)MoveHandle(handle,did string){
 p.mu.Lock();defer p.mu.Unlock();p.handles[handle]=did
}
func(p *InProcessFakePDS)Revoke(_ context.Context,token string)error{
 p.mu.Lock();defer p.mu.Unlock()
 item,ok:=p.sessions[token];if !ok{return ErrFakeLifecycle}
 item.revoked=true;p.sessions[token]=item
 return nil
}
// FakeLifecycle is disabled unless a local test explicitly enables it.
// Stable DID is never silently changed on handle migration or hijack.
type FakeLifecycle struct{Enabled bool;Provider FakeLifecycleProvider;Now func()time.Time}
func(s FakeLifecycle)Refresh(ctx context.Context,token,refresh string)(string,time.Time,error){
 if !s.Enabled||s.Provider==nil{return "",time.Time{},ErrFakeLifecycle}
 return s.Provider.Refresh(ctx,token,refresh)
}
func(s FakeLifecycle)VerifyHandle(ctx context.Context,expectedDID,handle string)error{
 if !s.Enabled||s.Provider==nil||!strings.HasPrefix(expectedDID,"did:"){return ErrFakeLifecycle}
 did,e:=s.Provider.ResolveHandle(ctx,handle)
 if e!=nil||did!=expectedDID{return ErrFakeLifecycle}
 return nil
}
func(s FakeLifecycle)Logout(ctx context.Context,token string)error{
 if !s.Enabled||s.Provider==nil{return ErrFakeLifecycle}
 return s.Provider.Revoke(ctx,token)
}
