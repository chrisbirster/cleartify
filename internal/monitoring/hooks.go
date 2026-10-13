// Package monitoring contains opt-in, local-only observation hooks.
// It deliberately makes no HTTP requests or calls to alerting providers.
package monitoring

import (
 "net/http"
 "sync"
 "time"
)
type Snapshot struct{Requests uint64;Errors uint64;LastObservedAt time.Time}
type Counter struct{mu sync.Mutex;value Snapshot}
func(c *Counter) Record(code int){
 if c==nil{return}
 c.mu.Lock();defer c.mu.Unlock()
 c.value.Requests++;if code>=500{c.value.Errors++}
 c.value.LastObservedAt=time.Now().UTC()
}
func(c *Counter) Snapshot()Snapshot{
 if c==nil{return Snapshot{}}
 c.mu.Lock();defer c.mu.Unlock()
 return c.value
}
type captureWriter struct{http.ResponseWriter;status int}
func(w *captureWriter) WriteHeader(status int){if w.status!=0{return};w.status=status;w.ResponseWriter.WriteHeader(status)}
func(w *captureWriter) Write(p []byte)(int,error){if w.status==0{w.WriteHeader(http.StatusOK)};return w.ResponseWriter.Write(p)}
// Wrap is disabled by default. It adds no paths, exporters, outbound calls or labels.
func Wrap(next http.Handler,enabled bool,c *Counter)http.Handler{
 if !enabled||c==nil{return next}
 return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  out:=&captureWriter{ResponseWriter:w}
  defer func(){if out.status==0{out.status=http.StatusOK};c.Record(out.status)}()
  next.ServeHTTP(out,r)
 })
}
