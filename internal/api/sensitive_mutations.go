package api

import(
 "net"
 "net/http"
 "strings"
 "sync"
 "time"
)

type mutationHit struct {start time.Time; count int}
type mutationLimiter struct {
 mu sync.Mutex
 hits map[string]mutationHit
 now func() time.Time
}
func newMutationLimiter()*mutationLimiter{return &mutationLimiter{hits:make(map[string]mutationHit),now:time.Now}}
func(l *mutationLimiter) allow(key string)bool{
 now:=l.now()
 l.mu.Lock();defer l.mu.Unlock()
 entry:=l.hits[key]
 if entry.start.IsZero()||now.Sub(entry.start)>=time.Minute {entry=mutationHit{start:now}}
 if entry.count>=12 {return false}
 entry.count++
 l.hits[key]=entry
 if len(l.hits)>4096 {
  for k,v:=range l.hits {if now.Sub(v.start)>=time.Minute{delete(l.hits,k)}}
  for len(l.hits)>4096 {for k:=range l.hits{delete(l.hits,k);break}}
 }
 return true
}
func sourceIP(r *http.Request) string{
 host,_,err:=net.SplitHostPort(r.RemoteAddr)
 if err!=nil||host=="" {return r.RemoteAddr}
 return host // deliberately ignore untrusted Forwarded/X-Forwarded-For
}
func requiresProtectedMutation(r *http.Request) (bool,bool){
 if r.Method!=http.MethodPost&&r.Method!=http.MethodPut&&r.Method!=http.MethodPatch&&r.Method!=http.MethodDelete{return false,false}
 p:=r.URL.Path
 if strings.HasPrefix(p,"/api/me/")||p=="/api/social/accounts"{return true,true}
 if p=="/api/admin/agents"||strings.HasPrefix(p,"/api/admin/clear/run-requests")||strings.HasPrefix(p,"/api/admin/protocol-install-requests")||p=="/api/admin/invites" {return true,false}
 return false,false
}
// protectSensitive wraps only registered routes. Disabled feature flags
// continue returning 404 without leaking their existence.
func(s *Server) protectSensitive(mux *http.ServeMux)http.Handler {
 return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  protected,needsOrigin:=requiresProtectedMutation(r)
  if !protected {mux.ServeHTTP(w,r);return}
  _,pattern:=mux.Handler(r)
  if pattern=="" {mux.ServeHTTP(w,r);return}
  if site:=r.Header.Get("Sec-Fetch-Site");site=="cross-site"{
   writeJSON(w,http.StatusForbidden,map[string]string{"error":"cross_site_forbidden"});return
  }
  origin:=r.Header.Get("Origin")
  if (needsOrigin&&(s.appOrigin==""||origin!=s.appOrigin))||(!needsOrigin&&origin!=""&&origin!=s.appOrigin){
   writeJSON(w,http.StatusForbidden,map[string]string{"error":"invalid_origin"});return
  }
  if !s.mutationLimiter.allow(sourceIP(r)+"|"+r.Method+"|"+pattern){
   w.Header().Set("Retry-After","60")
   writeJSON(w,http.StatusTooManyRequests,map[string]string{"error":"rate_limited"});return
  }
  w.Header().Set("Cache-Control","no-store")
  mux.ServeHTTP(w,r)
 })
}
