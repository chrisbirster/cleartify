package api
import(
 "net/http"
 "sync"
)
type moderationCounters struct{mu sync.Mutex;values map[string]uint64}
func(m *moderationCounters)inc(key string){if m==nil{return};m.mu.Lock();defer m.mu.Unlock();if m.values==nil{m.values=map[string]uint64{}};m.values[key]++}
func(m *moderationCounters)snapshot()map[string]uint64{m.mu.Lock();defer m.mu.Unlock();out:=map[string]uint64{};for k,v:=range m.values{out[k]=v};return out}
func(s *Server)localModerationMetrics(w http.ResponseWriter,r *http.Request){if !s.moderationAdmin(w,r){return};w.Header().Set("Cache-Control","no-store");writeJSON(w,200,map[string]any{"counters":s.moderationMetrics.snapshot(),"scope":"local_process"})}
// enforceSubjectHold is called at user-facing write boundaries; fail closed on
// datastore failures. The feature is entirely inert unless explicitly enabled.
func(s *Server)enforceSubjectHold(w http.ResponseWriter,r *http.Request,did string)bool{
 if !s.moderationEnforcement && !s.moderationBroader{return true}
 if s.moderationDB==nil{writeJSON(w,503,map[string]string{"error":"enforcement_unavailable"});return false}
 held,err:=s.subjectHeld(r,did)
 if err!=nil{writeJSON(w,503,map[string]string{"error":"enforcement_unavailable"});s.moderationMetrics.inc("hold_check_error");return false}
 if held{writeJSON(w,403,map[string]string{"error":"subject_on_hold"});s.moderationMetrics.inc("hold_denied");return false}
 s.moderationMetrics.inc("hold_allowed");return true
}
