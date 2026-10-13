package api

import (
 "context"
 "net/http"
)

// ClearPrincipal resolves a server-authenticated private Clear user ID.
// No HTTP body/query user ID is ever trusted as a principal.
type ClearPrincipal func(*http.Request)(string,error)

type alphaEligibility interface {
 Eligible(context.Context,string)(bool,error)
}

// privateAlpha enforces authentication and alpha eligibility for any private
// handler wrapped by it. Fail closed on missing dependencies and DB errors.
func (s *Server) privateAlpha(next http.HandlerFunc) http.HandlerFunc {
 return func(w http.ResponseWriter,r *http.Request){
  w.Header().Set("Cache-Control","no-store")
  if s.clearPrincipalResolver==nil || s.alphaEligibility==nil {
   writeJSON(w,http.StatusServiceUnavailable,map[string]string{"error":"private_auth_unavailable"});return
  }
  userID,err:=s.clearPrincipalResolver(r)
  if err!=nil || userID=="" {
   writeJSON(w,http.StatusUnauthorized,map[string]string{"error":"unauthenticated"});return
  }
  eligible,err:=s.alphaEligibility.Eligible(r.Context(),userID)
  if err!=nil {writeJSON(w,http.StatusServiceUnavailable,map[string]string{"error":"alpha_check_unavailable"});return}
  if !eligible {writeJSON(w,http.StatusForbidden,map[string]string{"error":"alpha_access_denied"});return}
  next(w,r)
 }
}
func (s *Server) alphaEligibilityStatus(w http.ResponseWriter,r *http.Request){
 writeJSON(w,http.StatusOK,map[string]bool{"eligible":true})
}
