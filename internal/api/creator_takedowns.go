package api
import(
 "net/http"
)
func(s *Server)creatorTakedownRequest(w http.ResponseWriter,r *http.Request){
 did,ok:=s.socialAuth.SocialSessionDID(r);if !ok{writeJSON(w,401,map[string]string{"error":"social_auth_required"});return}
 if s.appOrigin==""||r.Header.Get("Origin")!=s.appOrigin{writeJSON(w,403,map[string]string{"error":"invalid_origin"});return}
 var input struct{ReasonKey string `json:"reasonKey"`}
 if !decodeJSON(w,r,&input){return}
 id,err:=s.creatorTakedowns.Request(r.Context(),did,r.PathValue("id"),input.ReasonKey)
 if err!=nil{writeJSON(w,409,map[string]string{"error":"request_denied"});return}
 w.Header().Set("Cache-Control","no-store");writeJSON(w,202,map[string]any{"requestId":id,"status":"pending"})
}
func(s *Server)creatorTakedownDecision(w http.ResponseWriter,r *http.Request){
 if s.adminToken==""{writeJSON(w,503,map[string]string{"error":"not_configured"});return}
 if !secureBearerMatch(r.Header.Get("Authorization"),s.adminToken){writeJSON(w,401,map[string]string{"error":"unauthorized"});return}
 var input struct{Decision string `json:"decision"`}
 if !decodeJSON(w,r,&input){return}
 if err:=s.creatorTakedowns.Decide(r.Context(),r.PathValue("id"),input.Decision);err!=nil{writeJSON(w,409,map[string]string{"error":"transition_denied"});return}
 writeJSON(w,200,map[string]string{"status":input.Decision})
}
func(s *Server)creatorLineage(w http.ResponseWriter,r *http.Request){
 if _,ok:=s.socialAuth.SocialSessionDID(r);!ok{writeJSON(w,401,map[string]string{"error":"social_auth_required"});return}
 chain,err:=s.creatorTakedowns.Lineage(r.Context(),r.PathValue("id"));if err!=nil{writeJSON(w,503,map[string]string{"error":"lineage_unavailable"});return}
 w.Header().Set("Cache-Control","no-store");writeJSON(w,200,map[string]any{"parents":chain})
}
