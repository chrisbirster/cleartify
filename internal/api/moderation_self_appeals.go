package api
import(
 "net/http"
 "strings"
)
func(s *Server) moderationSelf(w http.ResponseWriter,r *http.Request)(string,bool){
 if s.socialAuth==nil||s.moderationDB==nil{writeJSON(w,503,map[string]string{"error":"moderation_unavailable"});return "",false}
 did,ok:=s.socialAuth.SocialSessionDID(r)
 if !ok||!strings.HasPrefix(did,"did:"){writeJSON(w,401,map[string]string{"error":"social_auth_required"});return "",false}
 return did,true
}
func(s *Server)myModerationHold(w http.ResponseWriter,r *http.Request){
 did,ok:=s.moderationSelf(w,r);if !ok{return}
 held,err:=s.subjectHeld(r,did)
 if err!=nil{writeJSON(w,503,map[string]string{"error":"moderation_unavailable"});return}
 var pending int
 err=s.moderationDB.QueryRowContext(r.Context(),`SELECT count(*) FROM social_moderation_cases WHERE subject_did=? AND kind='appeal' AND status='pending'`,did).Scan(&pending)
 if err!=nil{writeJSON(w,503,map[string]string{"error":"moderation_unavailable"});return}
 w.Header().Set("Cache-Control","no-store")
 writeJSON(w,200,map[string]any{"onHold":held,"appealPending":pending>0})
}
func(s *Server)myModerationAppeal(w http.ResponseWriter,r *http.Request){
 did,ok:=s.moderationSelf(w,r);if !ok{return}
 if s.appOrigin==""||r.Header.Get("Origin")!=s.appOrigin{writeJSON(w,403,map[string]string{"error":"invalid_origin"});return}
 var input struct{Reason string `json:"reason"`}
 if !decodeJSON(w,r,&input){return}
 input.Reason=strings.TrimSpace(input.Reason)
 if len(input.Reason)<5||len(input.Reason)>1000{writeJSON(w,400,map[string]string{"error":"invalid_appeal"});return}
 held,err:=s.subjectHeld(r,did)
 if err!=nil{writeJSON(w,503,map[string]string{"error":"moderation_unavailable"});return}
 if !held{writeJSON(w,409,map[string]string{"error":"no_active_hold"});return}
 if err=s.openModerationCase(r,did,"appeal",input.Reason);err!=nil{writeJSON(w,503,map[string]string{"error":"appeal_unavailable"});return}
 s.moderationMetrics.inc("self_appeal_submitted")
 w.Header().Set("Cache-Control","no-store");writeJSON(w,202,map[string]bool{"queued":true})
}
