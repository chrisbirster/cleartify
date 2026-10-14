package atproto
import("context";"encoding/json";"net/http";"strings")
// FakePDSHandler is an in-process HTTP mux for httptest.ResponseRecorder only.
// Never register this handler with the real application or a network listener.
func FakePDSHandler(p *InProcessFakePDS)http.Handler{
 mux:=http.NewServeMux()
 mux.HandleFunc("POST /fake/oauth/refresh",func(w http.ResponseWriter,r *http.Request){
  var v struct{Token string `json:"token"`;Refresh string `json:"refresh"`}
  r.Body=http.MaxBytesReader(w,r.Body,2048)
  if json.NewDecoder(r.Body).Decode(&v)!=nil{http.Error(w,"invalid",400);return}
  token,expiry,e:=p.Refresh(r.Context(),v.Token,v.Refresh)
  if e!=nil{http.Error(w,"invalid",401);return}
  _=json.NewEncoder(w).Encode(map[string]any{"token":token,"expiresAt":expiry.UTC().Format("2006-01-02T15:04:05Z07:00")})
 })
 mux.HandleFunc("GET /fake/pds/resolve",func(w http.ResponseWriter,r *http.Request){
  handle:=r.URL.Query().Get("handle")
  if handle==""||len(handle)>253||strings.ContainsAny(handle," /\\"){http.Error(w,"invalid",400);return}
  did,e:=p.ResolveHandle(r.Context(),handle)
  if e!=nil{http.Error(w,"unavailable",404);return}
  _=json.NewEncoder(w).Encode(map[string]string{"did":did})
 })
 mux.HandleFunc("POST /fake/oauth/revoke",func(w http.ResponseWriter,r *http.Request){
  var v struct{Token string `json:"token"`}
  if json.NewDecoder(http.MaxBytesReader(w,r.Body,2048)).Decode(&v)!=nil{http.Error(w,"invalid",400);return}
  if p.Revoke(context.WithoutCancel(r.Context()),v.Token)!=nil{http.Error(w,"invalid",404);return}
  w.WriteHeader(http.StatusNoContent)
 })
 return mux
}
