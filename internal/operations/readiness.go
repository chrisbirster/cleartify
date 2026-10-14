package operations
import("context";"database/sql";"net/http";"sync/atomic";"time")
type Readiness struct{DB *sql.DB;draining atomic.Bool}
func(r *Readiness)Drain(){r.draining.Store(true)}
func(r *Readiness)Handler(w http.ResponseWriter,req *http.Request){
 w.Header().Set("Content-Type","application/json")
 w.Header().Set("Cache-Control","no-store")
 if r.draining.Load()||r.DB==nil{w.WriteHeader(http.StatusServiceUnavailable);_,_=w.Write([]byte(`{"ready":false}`));return}
 ctx,cancel:=context.WithTimeout(req.Context(),300*time.Millisecond);defer cancel()
 if err:=r.DB.PingContext(ctx);err!=nil{w.WriteHeader(http.StatusServiceUnavailable);_,_=w.Write([]byte(`{"ready":false}`));return}
 _,_=w.Write([]byte(`{"ready":true}`))
}
// WithReadiness adds opt-in health endpoints around existing application mux.
// Existing /healthz continues to mean process liveness, not dependency health.
func WithReadiness(next http.Handler,ready *Readiness)http.Handler{
 return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  if r.Method==http.MethodGet&&r.URL.Path=="/readyz"{ready.Handler(w,r);return}
  next.ServeHTTP(w,r)
 })
}
