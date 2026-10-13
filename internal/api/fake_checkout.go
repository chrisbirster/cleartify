package api

import (
 "errors"
 "net/http"
 "github.com/chrisbirster/cleartify/internal/marketplace"
)
func(s *Server)fakeCheckoutActor(w http.ResponseWriter,r *http.Request)(string,bool){
 if s.fakeCheckout==nil||s.socialAuth==nil{writeJSON(w,503,map[string]string{"error":"checkout_unavailable"});return "",false}
 did,ok:=s.socialAuth.SocialSessionDID(r)
 if !ok{writeJSON(w,401,map[string]string{"error":"atproto_login_required"});return "",false}
 return did,true
}
func(s *Server)fakeCheckoutPurchase(w http.ResponseWriter,r *http.Request){
 did,ok:=s.fakeCheckoutActor(w,r);if !ok{return}
 if s.appOrigin==""||r.Header.Get("Origin")!=s.appOrigin{writeJSON(w,403,map[string]string{"error":"invalid_origin"});return}
 r.Body=http.MaxBytesReader(w,r.Body,4096)
 var req struct{ListingID string `json:"listingId"`}
 if !decodeJSON(w,r,&req){return}
 if len(req.ListingID)==0||len(req.ListingID)>128{writeJSON(w,400,map[string]string{"error":"invalid_listing"});return}
 id,err:=s.fakeCheckout.Purchase(r.Context(),did,req.ListingID)
 if errors.Is(err,marketplace.ErrDenied)||errors.Is(err,marketplace.ErrPayment){writeJSON(w,409,map[string]string{"error":"fake_checkout_denied"});return}
 if err!=nil{writeJSON(w,409,map[string]string{"error":"checkout_conflict"});return}
 w.Header().Set("Cache-Control","no-store")
 writeJSON(w,201,map[string]any{"entitlementId":id,"provider":"fake","paid":false})
}
func(s *Server)fakeCheckoutAccess(w http.ResponseWriter,r *http.Request){
 did,ok:=s.fakeCheckoutActor(w,r);if !ok{return}
 id:=r.PathValue("id")
 if id==""||len(id)>128{writeJSON(w,404,map[string]string{"error":"not_found"});return}
 allowed,err:=s.fakeCheckout.Entitled(r.Context(),did,id)
 if err!=nil{writeJSON(w,503,map[string]string{"error":"entitlement_unavailable"});return}
 w.Header().Set("Cache-Control","no-store")
 if !allowed{writeJSON(w,403,map[string]string{"error":"purchase_required"});return}
 writeJSON(w,200,map[string]any{"listingId":id,"access":true,"provider":"fake"})
}
func(s *Server)fakeCheckoutRefund(w http.ResponseWriter,r *http.Request){
 did,ok:=s.fakeCheckoutActor(w,r);if !ok{return}
 if s.appOrigin==""||r.Header.Get("Origin")!=s.appOrigin{writeJSON(w,403,map[string]string{"error":"invalid_origin"});return}
 id:=r.PathValue("id");if id==""||len(id)>128{writeJSON(w,404,map[string]string{"error":"not_found"});return}
 if err:=s.fakeCheckout.Refund(r.Context(),did,id);err!=nil{
  if errors.Is(err,marketplace.ErrDenied)||errors.Is(err,marketplace.ErrPayment){writeJSON(w,403,map[string]string{"error":"refund_denied"});return}
  writeJSON(w,503,map[string]string{"error":"refund_unavailable"});return
 }
 w.Header().Set("Cache-Control","no-store")
 writeJSON(w,200,map[string]any{"status":"refunded","provider":"fake"})
}
