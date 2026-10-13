package api

import (
 "net"
 "net/http"
 "sync"
 "time"
)

// signupLimiter is a bounded, process-local abuse guard for the initial single
// backend writer. Multi-instance deployments must move limits to shared storage.
type signupLimiter struct {
 mu sync.Mutex
 windows map[string]signupWindow
 now func() time.Time
}
type signupWindow struct { starts time.Time; count int }
const signupLimitPerHour = 10
const signupMaxTrackedIPs = 10000

func newSignupLimiter() *signupLimiter {
 return &signupLimiter{windows: make(map[string]signupWindow), now:time.Now}
}

func (l *signupLimiter) allow(remoteAddr string) bool {
 ip, _, err := net.SplitHostPort(remoteAddr)
 if err != nil { return false }
 parsed := net.ParseIP(ip)
 if parsed == nil { return false }
 key := parsed.String()
 now:=l.now()
 l.mu.Lock()
 defer l.mu.Unlock()
 win,ok:=l.windows[key]
 if ok && now.Sub(win.starts)>=time.Hour {delete(l.windows,key);ok=false}
 if !ok {
  if len(l.windows)>=signupMaxTrackedIPs {
   for k,v:=range l.windows { if now.Sub(v.starts)>=time.Hour {delete(l.windows,k)} }
  }
  if len(l.windows)>=signupMaxTrackedIPs { return false }
  win=signupWindow{starts:now}
 }
 if win.count>=signupLimitPerHour {return false}
 win.count++
 l.windows[key]=win
 return true
}

func (s *Server) limitSignup(w http.ResponseWriter,r *http.Request) bool {
 if s.signupLimiter.allow(r.RemoteAddr) {return true}
 w.Header().Set("Retry-After","3600")
 writeJSON(w,http.StatusTooManyRequests,map[string]string{"error":"rate_limited"})
 return false
}
