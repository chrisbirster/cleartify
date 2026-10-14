package api

import (
 "errors"
 "net/http"
 "github.com/chrisbirster/cleartify/internal/marketplace"
)

// Fake operator actions require an explicitly enabled fake accounting service,
// an administrator bearer token, and the configured same-origin mutation check.
func(s *Server) fakeOperatorAuthorized(w http.ResponseWriter,r *http.Request) bool {
 if s.fakeAccounting==nil||!s.fakeAccounting.Enabled||s.fakeCheckout==nil||!s.fakeCheckout.Enabled||s.adminToken=="" {
  writeJSON(w,503,map[string]string{"error":"fake_operator_unavailable"});return false
 }
 if !secureBearerMatch(r.Header.Get("Authorization"),s.adminToken) {
  writeJSON(w,401,map[string]string{"error":"unauthorized"});return false
 }
 if s.appOrigin==""||r.Header.Get("Origin")!=s.appOrigin {
  writeJSON(w,403,map[string]string{"error":"invalid_origin"});return false
 }
 return true
}
func(s *Server) fakeOperatorDispute(w http.ResponseWriter,r *http.Request){
 if !s.fakeOperatorAuthorized(w,r){return}
 id:=r.PathValue("id");if id==""||len(id)>128{writeJSON(w,400,map[string]string{"error":"invalid_entitlement"});return}
 r.Body=http.MaxBytesReader(w,r.Body,4096)
 var req struct {BuyerDID string `json:"buyerDid"`;Reason string `json:"reason"`}
 if !decodeJSON(w,r,&req){return}
 // A bearer-token operator cannot impersonate a buyer to open a dispute.
 // Buyer-owned dispute initiation needs an independently verified OAuth session.
 writeJSON(w,403,map[string]string{"error":"buyer_session_required"})
}
func(s *Server) fakeOperatorResolve(w http.ResponseWriter,r *http.Request){
 if !s.fakeOperatorAuthorized(w,r){return}
 id:=r.PathValue("id");if id==""||len(id)>128{writeJSON(w,400,map[string]string{"error":"invalid_entitlement"});return}
 r.Body=http.MaxBytesReader(w,r.Body,4096)
 var req struct{Refund bool `json:"refund"`}
 if !decodeJSON(w,r,&req){return}
 if err:=s.fakeAccounting.ResolveDispute(r.Context(),"local-admin",id,req.Refund);err!=nil{
  writeJSON(w,409,map[string]string{"error":"dispute_denied"});return
 }
 w.Header().Set("Cache-Control","no-store")
 writeJSON(w,200,map[string]string{"status":"resolved","provider":"fake"})
}
func(s *Server) fakeOperatorRefund(w http.ResponseWriter,r *http.Request){
 if !s.fakeOperatorAuthorized(w,r){return}
 id:=r.PathValue("id");if id==""||len(id)>128{writeJSON(w,400,map[string]string{"error":"invalid_entitlement"});return}
 // Operator refund is confined to existing open disputes: no arbitrary buyer/seller override.
 if err:=s.fakeAccounting.ResolveDispute(r.Context(),"local-admin",id,true);err!=nil{
  if errors.Is(err,marketplace.ErrDenied){writeJSON(w,409,map[string]string{"error":"refund_denied"});return}
  writeJSON(w,503,map[string]string{"error":"refund_unavailable"});return
 }
 w.Header().Set("Cache-Control","no-store")
 writeJSON(w,200,map[string]string{"status":"refunded","provider":"fake"})
}
