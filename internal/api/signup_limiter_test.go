package api

import (
 "net/http"
 "net/http/httptest"
 "testing"
 "time"
)

func TestSignupLimiterBlocksExcessAndResets(t *testing.T){
 l:=newSignupLimiter()
 now:=time.Date(2026,10,12,10,0,0,0,time.UTC)
 l.now=func()time.Time{return now}
 for i:=0;i<signupLimitPerHour;i++{if !l.allow("192.0.2.1:1234"){t.Fatalf("attempt %d rejected",i)}}
 if l.allow("192.0.2.1:1234"){t.Fatal("excess accepted")}
 if !l.allow("192.0.2.2:1234"){t.Fatal("other IP blocked")}
 now=now.Add(time.Hour)
 if !l.allow("192.0.2.1:1234"){t.Fatal("window never reset")}
}
func TestSignupLimiterDoesNotTrustForwardedIP(t *testing.T){
 s:=&Server{signupLimiter:newSignupLimiter()}
 for i:=0;i<signupLimitPerHour;i++ {
  r:=httptest.NewRequest(http.MethodPost,"/api/waitlist",nil)
  r.RemoteAddr="192.0.2.5:1234"
  r.Header.Set("X-Forwarded-For","198.51.100.2")
  w:=httptest.NewRecorder()
  if !s.limitSignup(w,r){t.Fatalf("unexpected limit at %d",i)}
 }
 r:=httptest.NewRequest(http.MethodPost,"/api/waitlist",nil)
 r.RemoteAddr="192.0.2.5:1234"
 r.Header.Set("X-Forwarded-For","203.0.113.200")
 w:=httptest.NewRecorder()
 if s.limitSignup(w,r) || w.Code!=http.StatusTooManyRequests {t.Fatalf("forwarded IP bypassed: %d",w.Code)}
 if w.Header().Get("Retry-After")!="3600" {t.Fatal("missing retry header")}
}
