package atproto
import("context";"errors";"net/http";"net/http/httptest";"strings";"testing";"time")
func TestFakeOAuthPDSLifecycleNoNetwork(t *testing.T){
 ctx:=context.Background()
 p:=NewInProcessFakePDS()
 now:=time.Date(2026,10,13,12,0,0,0,time.UTC)
 p.Now=func()time.Time{return now}
 token,e:=p.Issue("did:plc:alice","alice.test",time.Minute);if e!=nil{t.Fatal(e)}
 flow:=FakeLifecycle{Provider:p}
 if e=flow.VerifyHandle(ctx,"did:plc:alice","alice.test");!errors.Is(e,ErrFakeLifecycle){t.Fatalf("disabled allowed: %v",e)}
 flow.Enabled=true
 if e=flow.VerifyHandle(ctx,"did:plc:alice","alice.test");e!=nil{t.Fatal(e)}
 next,expiry,e:=flow.Refresh(ctx,token,"fake-refresh-1");if e!=nil||!expiry.Equal(now.Add(time.Hour)){t.Fatalf("refresh %v",e)}
 if _,_,e=flow.Refresh(ctx,token,"fake-refresh-1");!errors.Is(e,ErrFakeLifecycle){t.Fatal("old refresh reusable",e)}
 p.MoveHandle("alice.test","did:plc:other")
 if e=flow.VerifyHandle(ctx,"did:plc:alice","alice.test");!errors.Is(e,ErrFakeLifecycle){t.Fatal("handle reassigned but accepted",e)}
 if e=flow.Logout(ctx,next);e!=nil{t.Fatal(e)}
 if _,_,e=flow.Refresh(ctx,next,"fake-refresh-2");!errors.Is(e,ErrFakeLifecycle){t.Fatal("revoked refresh allowed",e)}
 token,e=p.Issue("did:plc:alice","new.test",time.Second);if e!=nil{t.Fatal(e)}
 now=now.Add(time.Second)
 if _,_,e=flow.Refresh(ctx,token,"fake-refresh-3");!errors.Is(e,ErrFakeLifecycle){t.Fatal("exact expiry allowed",e)}
}
func TestInProcessFakePDSHandler(t *testing.T){
 p:=NewInProcessFakePDS()
 token,e:=p.Issue("did:plc:alice","alice.test",time.Hour);if e!=nil{t.Fatal(e)}
 h:=FakePDSHandler(p)
 send:=func(method,path,body string)*httptest.ResponseRecorder{
  w:=httptest.NewRecorder();h.ServeHTTP(w,httptest.NewRequest(method,path,strings.NewReader(body)));return w
 }
 if w:=send("GET","/fake/pds/resolve?handle=alice.test","");w.Code!=200||!strings.Contains(w.Body.String(),"did:plc:alice"){t.Fatalf("resolve %d",w.Code)}
 if w:=send("POST","/fake/oauth/refresh",`{"token":"`+token+`","refresh":"fake-refresh-1"}`);w.Code!=200{t.Fatalf("refresh %d",w.Code)}
 if w:=send("POST","/fake/oauth/refresh",`{"token":"`+token+`","refresh":"fake-refresh-1"}`);w.Code!=401{t.Fatalf("replay %d",w.Code)}
 if w:=send("GET","/fake/pds/resolve?handle=unknown.test","");w.Code!=404{t.Fatalf("unknown handle %d",w.Code)}
 if w:=send("GET","/fake/oauth/refresh","");w.Code!=405{t.Fatalf("wrong method %d",w.Code)}
 _=http.MethodGet
}
