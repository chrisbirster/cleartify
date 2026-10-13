package api

import (
 "database/sql"
 "errors"
 "net/http"
 "time"
)

// identityReset atomically revokes all live Clear sessions, active DID links,
// and outstanding linking challenges. Caller must have a fresh authenticated
// Clear session; disconnect additionally requires the currently linked social DID.
func(s *Server) identityReset(w http.ResponseWriter,r *http.Request,requireSocial bool){
 if s.clearSessions==nil||s.identityLinks==nil||s.identityLinks.DB==nil||s.identityLinks.Push==nil{
  writeJSON(w,503,map[string]string{"error":"recovery_unavailable"});return
 }
 if s.appOrigin==""||r.Header.Get("Origin")!=s.appOrigin{writeJSON(w,403,map[string]string{"error":"invalid_origin"});return}
 cookie,err:=r.Cookie("clear_session");if err!=nil{writeJSON(w,401,map[string]string{"error":"clear_reauthentication_required"});return}
 user,err:=s.clearSessions.AuthenticateFresh(r.Context(),cookie.Value,linkTTL)
 if err!=nil{writeJSON(w,401,map[string]string{"error":"clear_reauthentication_required"});return}
 var input struct{Confirm bool `json:"confirm"`}
 if !decodeJSON(w,r,&input){return}
 if !input.Confirm{writeJSON(w,400,map[string]string{"error":"confirmation_required"});return}
 did:=""
 if requireSocial {
  if s.linkSocial==nil{writeJSON(w,503,map[string]string{"error":"social_auth_unavailable"});return}
  var ok bool;did,ok=s.linkSocial.SocialSessionDID(r)
  if !ok||did==""{writeJSON(w,401,map[string]string{"error":"social_reauthentication_required"});return}
 }
 tx,err:=s.identityLinks.DB.BeginTx(r.Context(),nil);if err!=nil{writeJSON(w,503,map[string]string{"error":"recovery_unavailable"});return};defer tx.Rollback()
 var linkedDID string
 err=tx.QueryRowContext(r.Context(),`SELECT social_did FROM identity_links WHERE clear_user_id=? AND status='active'`,user).Scan(&linkedDID)
 if errors.Is(err,sql.ErrNoRows){writeJSON(w,409,map[string]string{"error":"no_active_link"});return}
 if err!=nil{writeJSON(w,503,map[string]string{"error":"recovery_unavailable"});return}
 if requireSocial&&did!=linkedDID{writeJSON(w,403,map[string]string{"error":"linked_social_did_mismatch"});return}
 event,_,err:=challengeToken();if err!=nil{writeJSON(w,503,map[string]string{"error":"recovery_unavailable"});return}
 now:=time.Now().UTC().Format(time.RFC3339Nano)
 result,err:=tx.ExecContext(r.Context(),`UPDATE identity_links SET status='revoked',revoked_at=? WHERE clear_user_id=? AND social_did=? AND status='active'`,now,user,linkedDID)
 if err!=nil{writeJSON(w,503,map[string]string{"error":"recovery_unavailable"});return}
 affected,err:=result.RowsAffected();if err!=nil||affected!=1{writeJSON(w,409,map[string]string{"error":"link_changed"});return}
 if _,err=tx.ExecContext(r.Context(),`UPDATE identity_link_challenges SET state='cancelled' WHERE clear_user_id=? AND state='pending'`,user);err!=nil{writeJSON(w,503,map[string]string{"error":"recovery_unavailable"});return}
 if _,err=tx.ExecContext(r.Context(),`UPDATE clear_user_sessions SET revoked_at=? WHERE user_id=? AND revoked_at IS NULL`,now,user);err!=nil{writeJSON(w,503,map[string]string{"error":"recovery_unavailable"});return}
 if _,err=tx.ExecContext(r.Context(),`INSERT INTO identity_link_events(id,clear_user_id,social_did,action,recorded_at) VALUES(?,?,?,'unlinked',?)`,event,user,linkedDID,now);err!=nil{writeJSON(w,503,map[string]string{"error":"recovery_unavailable"});return}
 if err=tx.Commit();err!=nil{writeJSON(w,503,map[string]string{"error":"recovery_unavailable"});return}
 if err=s.identityLinks.Push(r.Context());err!=nil{writeJSON(w,503,map[string]string{"error":"sync_unavailable"});return}
 http.SetCookie(w,&http.Cookie{Name:"clear_session",Value:"",Path:"/api/me/",HttpOnly:true,Secure:r.TLS!=nil,SameSite:http.SameSiteStrictMode,MaxAge:-1})
 s.linkCookieSet(w,r,"",-1)
 w.Header().Set("Cache-Control","no-store")
 writeJSON(w,200,map[string]any{"linked":false,"clearSessionsRevoked":true,"localRecoveryOnly":true})
}
func(s *Server)linkDisconnectAll(w http.ResponseWriter,r *http.Request){s.identityReset(w,r,true)}
func(s *Server)linkRecoverLocal(w http.ResponseWriter,r *http.Request){s.identityReset(w,r,false)}
