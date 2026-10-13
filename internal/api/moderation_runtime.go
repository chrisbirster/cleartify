package api

import (
 "crypto/rand"
 "database/sql"
 "encoding/hex"
 "errors"
 "net/http"
 "strings"
 "time"
 "github.com/chrisbirster/cleartify/internal/social"
)
func moderationID()string{b:=make([]byte,16);if _,err:=rand.Read(b);err!=nil{return ""};return hex.EncodeToString(b)}
func(s *Server) moderationAdmin(w http.ResponseWriter,r *http.Request)bool{
 if s.adminToken==""||s.moderationDB==nil{writeJSON(w,503,map[string]string{"error":"not_configured"});return false}
 if !secureBearerMatch(r.Header.Get("Authorization"),s.adminToken){writeJSON(w,401,map[string]string{"error":"unauthorized"});return false}
 return true
}
func(s *Server) subjectHeld(r *http.Request,did string)(bool,error){
 if s.moderationDB==nil{return false,errors.New("missing moderation db")}
 var n int
 err:=s.moderationDB.QueryRowContext(r.Context(),"SELECT count(*) FROM social_impersonation_holds WHERE subject_did=? AND released_at IS NULL",did).Scan(&n)
 return n>0,err
}
func(s *Server) evaluateModeration(w http.ResponseWriter,r *http.Request){
 if !s.moderationAdmin(w,r){return}
 did:=r.PathValue("did")
 if !strings.HasPrefix(did,"did:")||len(did)>253{writeJSON(w,400,map[string]string{"error":"invalid_did"});return}
 label,err:=(social.Impersonation{DB:s.moderationDB,Enabled:true}).EvaluateRecords(r.Context(),did)
 if err!=nil{writeJSON(w,503,map[string]string{"error":"evaluation_unavailable"});return}
 if label.OnHold{
  if err:=s.openModerationCase(r,did,"review","automated hold evaluation");err!=nil{writeJSON(w,503,map[string]string{"error":"case_unavailable"});return}
 }
 if s.moderationEnforcement {s.moderationMetrics.inc("evaluation_completed");if label.OnHold{s.moderationMetrics.inc("hold_evaluated")}}
 writeJSON(w,200,label)
}
func(s *Server) openModerationCase(r *http.Request,did,kind,reason string)error{
 tx,err:=s.moderationDB.BeginTx(r.Context(),nil);if err!=nil{return err};defer tx.Rollback()
 id:=moderationID();audit:=moderationID();if id==""||audit==""{return errors.New("random unavailable")}
 now:=time.Now().UTC().Format(time.RFC3339Nano)
 result,err:=tx.ExecContext(r.Context(),`INSERT INTO social_moderation_cases(id,subject_did,kind,status,rationale,created_at) VALUES(?,?,?,'pending',?,?) ON CONFLICT DO NOTHING`,id,did,kind,reason,now)
 if err!=nil{return err}
 n,err:=result.RowsAffected();if err!=nil{return err}
 if n==1{
  _,err=tx.ExecContext(r.Context(),`INSERT INTO social_moderation_audit(id,case_id,actor,action,recorded_at) VALUES(?,?,'admin','opened',?)`,audit,id,now)
  if err!=nil{return err}
 }
 return tx.Commit()
}
func(s *Server) appealModeration(w http.ResponseWriter,r *http.Request){
 if !s.moderationAdmin(w,r){return}
 did:=r.PathValue("did")
 var input struct{Reason string `json:"reason"`}
 if !decodeJSON(w,r,&input){return}
 if !strings.HasPrefix(did,"did:")||len(did)>253||len(strings.TrimSpace(input.Reason))<5||len(input.Reason)>1000{writeJSON(w,400,map[string]string{"error":"invalid_appeal"});return}
 held,err:=s.subjectHeld(r,did);if err!=nil{writeJSON(w,503,map[string]string{"error":"moderation_unavailable"});return};if !held{writeJSON(w,409,map[string]string{"error":"no_active_hold"});return}
 if err=s.openModerationCase(r,did,"appeal",input.Reason);err!=nil{writeJSON(w,503,map[string]string{"error":"case_unavailable"});return}
 if s.moderationEnforcement{s.moderationMetrics.inc("appeal_queued")}
 writeJSON(w,202,map[string]bool{"queued":true})
}
func(s *Server) listModerationCases(w http.ResponseWriter,r *http.Request){
 if !s.moderationAdmin(w,r){return}
 rows,err:=s.moderationDB.QueryContext(r.Context(),`SELECT id,subject_did,kind,status,rationale,created_at FROM social_moderation_cases ORDER BY created_at DESC LIMIT 100`)
 if err!=nil{writeJSON(w,503,map[string]string{"error":"cases_unavailable"});return};defer rows.Close()
 type record struct{ID string `json:"id"`;DID string `json:"did"`;Kind string `json:"kind"`;Status string `json:"status"`;Rationale string `json:"rationale"`;CreatedAt string `json:"createdAt"`}
 out:=[]record{}
 for rows.Next(){var rec record;if err=rows.Scan(&rec.ID,&rec.DID,&rec.Kind,&rec.Status,&rec.Rationale,&rec.CreatedAt);err!=nil{writeJSON(w,503,map[string]string{"error":"cases_unavailable"});return};out=append(out,rec)}
 if rows.Err()!=nil{writeJSON(w,503,map[string]string{"error":"cases_unavailable"});return}
 writeJSON(w,200,map[string]any{"cases":out})
}
func(s *Server) decideModeration(w http.ResponseWriter,r *http.Request){
 if !s.moderationAdmin(w,r){return}
 var input struct{Decision string `json:"decision"`;Reason string `json:"reason"`}
 if !decodeJSON(w,r,&input){return}
 if input.Decision!="upheld"&&input.Decision!="released"&&input.Decision!="rejected"||len(strings.TrimSpace(input.Reason))<5||len(input.Reason)>1000 {writeJSON(w,400,map[string]string{"error":"invalid_decision"});return}
 tx,err:=s.moderationDB.BeginTx(r.Context(),nil);if err!=nil{writeJSON(w,503,map[string]string{"error":"case_unavailable"});return};defer tx.Rollback()
 var did string
 err=tx.QueryRowContext(r.Context(),`SELECT subject_did FROM social_moderation_cases WHERE id=? AND status='pending'`,r.PathValue("id")).Scan(&did)
 if errors.Is(err,sql.ErrNoRows){writeJSON(w,409,map[string]string{"error":"case_not_pending"});return};if err!=nil{writeJSON(w,503,map[string]string{"error":"case_unavailable"});return}
 now:=time.Now().UTC().Format(time.RFC3339Nano)
 res,err:=tx.ExecContext(r.Context(),`UPDATE social_moderation_cases SET status=?,decided_at=?,decision_reason=? WHERE id=? AND status='pending'`,input.Decision,now,input.Reason,r.PathValue("id"))
 if err!=nil{writeJSON(w,503,map[string]string{"error":"case_unavailable"});return}
 n,err:=res.RowsAffected();if err!=nil||n!=1{writeJSON(w,409,map[string]string{"error":"case_not_pending"});return}
 if input.Decision=="released"{
  _,err=tx.ExecContext(r.Context(),`UPDATE social_impersonation_holds SET released_at=? WHERE subject_did=? AND released_at IS NULL`,now,did)
  if err!=nil{writeJSON(w,503,map[string]string{"error":"case_unavailable"});return}
 }
 _,err=tx.ExecContext(r.Context(),`INSERT INTO social_moderation_audit(id,case_id,actor,action,recorded_at) VALUES(?,?,'admin',?,?)`,moderationID(),r.PathValue("id"),input.Decision,now)
 if err!=nil||tx.Commit()!=nil{writeJSON(w,503,map[string]string{"error":"case_unavailable"});return}
 if s.moderationEnforcement{s.moderationMetrics.inc("decision_"+input.Decision)}
 writeJSON(w,200,map[string]any{"status":input.Decision})
}
