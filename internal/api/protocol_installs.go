package api

import(
 "errors"
 "net/http"

 "github.com/chrisbirster/cleartify/internal/protocols"
)
func(s *Server) installManifest(w http.ResponseWriter,r *http.Request){
 m,err:=s.installFlow.Preview(r.Context(),r.PathValue("id"))
 if err!=nil{
  if errors.Is(err,protocols.ErrInstallUnavailable){writeJSON(w,404,map[string]string{"error":"manifest_unavailable"});return}
  writeJSON(w,500,map[string]string{"error":"internal_error"});return
 }
 w.Header().Set("Cache-Control","no-store")
 writeJSON(w,200,m)
}
func(s *Server) requestProtocolInstall(w http.ResponseWriter,r *http.Request){
 if s.socialAuth==nil||s.installFlow==nil{writeJSON(w,503,map[string]string{"error":"unavailable"});return}
 if s.appOrigin==""||r.Header.Get("Origin")!=s.appOrigin{writeJSON(w,403,map[string]string{"error":"invalid_origin"});return}
 did,ok:=s.socialAuth.SocialSessionDID(r);if !ok{writeJSON(w,401,map[string]string{"error":"atproto_login_required"});return}
 r.Body=http.MaxBytesReader(w,r.Body,4096)
 var req struct{PublicationID string `json:"publicationId"`;AcknowledgedManifestSHA256 string `json:"acknowledgedManifestSha256"`}
 if !decodeJSON(w,r,&req){return}
 if s.paidEnforcement {
  allowed,e:=s.fakeCheckout.PublicationAccess(r.Context(),did,req.PublicationID)
  if e!=nil{writeJSON(w,503,map[string]string{"error":"entitlement_unavailable"});return}
  if !allowed{writeJSON(w,403,map[string]string{"error":"purchase_required"});return}
 }
 id,err:=s.installFlow.Request(r.Context(),did,req.PublicationID,req.AcknowledgedManifestSHA256)
 if err!=nil{
  switch {
  case errors.Is(err,protocols.ErrManifestChanged):writeJSON(w,409,map[string]string{"error":"manifest_changed_repreview_required"})
  case errors.Is(err,protocols.ErrInstallUnavailable):writeJSON(w,400,map[string]string{"error":"protocol_unavailable"})
  default:writeJSON(w,500,map[string]string{"error":"internal_error"})
  }
  return
 }
 w.Header().Set("Cache-Control","no-store")
 writeJSON(w,202,map[string]any{"id":id,"status":"pending_approval","installed":false,"executed":false})
}
func(s *Server) decideProtocolInstall(w http.ResponseWriter,r *http.Request){
 if s.adminToken==""{writeJSON(w,503,map[string]string{"error":"admin_not_configured"});return}
 if !secureBearerMatch(r.Header.Get("Authorization"),s.adminToken){writeJSON(w,401,map[string]string{"error":"unauthorized"});return}
 r.Body=http.MaxBytesReader(w,r.Body,1024)
 var req struct{Approve bool `json:"approve"`}
 if !decodeJSON(w,r,&req){return}
 if err:=s.installFlow.Decide(r.Context(),r.PathValue("id"),req.Approve);err!=nil{
  if errors.Is(err,protocols.ErrInstallDecision){writeJSON(w,409,map[string]string{"error":"not_pending_or_version_changed"});return}
  writeJSON(w,500,map[string]string{"error":"internal_error"});return
 }
 state:="rejected";if req.Approve{state="approved"}
 writeJSON(w,200,map[string]any{"status":state,"installed":false,"executed":false})
}
