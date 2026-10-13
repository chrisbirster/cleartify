package api
import(
 "errors"
 "net/http"
 "github.com/chrisbirster/cleartify/internal/protocols"
)
func(s *Server) updateCreatorVisibility(w http.ResponseWriter,r *http.Request){
 if !s.creatorPublishingEnabled||s.creatorDrafts==nil||s.socialAuth==nil{writeJSON(w,503,map[string]string{"error":"publishing_unavailable"});return}
 did,ok:=s.socialAuth.SocialSessionDID(r)
 if !ok{writeJSON(w,401,map[string]string{"error":"atproto_login_required"});return}
 if s.moderationBroader && !s.enforceSubjectHold(w,r,did){return}
 var input struct{Visibility string `json:"visibility"`;Provenance string `json:"provenance"`}
 if !decodeJSON(w,r,&input){return}
 err:=s.creatorDrafts.PublishDraft(r.Context(),did,r.PathValue("id"),input.Visibility,input.Provenance)
 if errors.Is(err,protocols.ErrPublishingInvalid){writeJSON(w,400,map[string]string{"error":"license_or_provenance_invalid"});return}
 if err!=nil{writeJSON(w,409,map[string]string{"error":"publication_conflict"});return}
 w.Header().Set("Cache-Control","no-store")
 writeJSON(w,200,map[string]any{"visibility":input.Visibility,"published":input.Visibility!="private"})
}
