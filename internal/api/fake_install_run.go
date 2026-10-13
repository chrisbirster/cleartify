package api

import (
 "crypto/rand"
 "crypto/sha256"
 "database/sql"
 "encoding/hex"
 "encoding/json"
 "errors"
 "net/http"
 "time"

 "github.com/chrisbirster/cleartify/internal/runner"
)
func newRunID()(string,error){b:=make([]byte,18);if _,err:=rand.Read(b);err!=nil{return "",err};return hex.EncodeToString(b),nil}
func(s *Server) fakeRunPrincipal(w http.ResponseWriter,r *http.Request)(string,string,bool){
 user,err:=s.clearPrincipalResolver(r)
 if err!=nil||user==""{writeJSON(w,401,map[string]string{"error":"clear_auth_required"});return "","",false}
 if s.socialAuth==nil{writeJSON(w,503,map[string]string{"error":"social_auth_unavailable"});return "","",false}
 did,ok:=s.socialAuth.SocialSessionDID(r)
 if !ok{writeJSON(w,401,map[string]string{"error":"social_auth_required"});return "","",false}
 var linked int
 err=s.fakeInstallRunDB.QueryRowContext(r.Context(),`SELECT count(*) FROM identity_links WHERE clear_user_id=? AND social_did=? AND status='active'`,user,did).Scan(&linked)
 if err!=nil{writeJSON(w,503,map[string]string{"error":"identity_link_unavailable"});return "","",false}
 if linked!=1{writeJSON(w,403,map[string]string{"error":"active_identity_link_required"});return "","",false}
 return user,did,true
}
func(s *Server) startFakeInstallRun(w http.ResponseWriter,r *http.Request){
 if s.fakeInstallRunDB==nil||s.fakeInstallRunPush==nil{writeJSON(w,503,map[string]string{"error":"fake_run_unavailable"});return}
 user,did,ok:=s.fakeRunPrincipal(w,r);if !ok{return}
 var input struct{InstallRequestID string `json:"installRequestId"`}
 if !decodeJSON(w,r,&input){return}
 if len(input.InstallRequestID)!=36{writeJSON(w,400,map[string]string{"error":"invalid_install_request"});return}
 tx,err:=s.fakeInstallRunDB.BeginTx(r.Context(),nil)
 if err!=nil{writeJSON(w,503,map[string]string{"error":"fake_run_unavailable"});return};defer tx.Rollback()
 var pub,manifest,expectedSpec,actualSpec,spec,status,actor string
 var expectedVersion,actualVersion int
 err=tx.QueryRowContext(r.Context(),`SELECT i.publication_id,i.manifest_sha256,i.spec_sha256,i.protocol_version,i.status,i.actor_did,p.spec_json,p.version FROM protocol_install_requests i JOIN published_protocols p ON p.id=i.publication_id WHERE i.id=? AND p.status='published'`,input.InstallRequestID).Scan(&pub,&manifest,&expectedSpec,&expectedVersion,&status,&actor,&spec,&actualVersion)
 if errors.Is(err,sql.ErrNoRows){writeJSON(w,409,map[string]string{"error":"install_not_approved"});return}
 if err!=nil{writeJSON(w,503,map[string]string{"error":"fake_run_unavailable"});return}
 digest:=sha256.Sum256([]byte(spec));actualSpec=hex.EncodeToString(digest[:])
 if status!="approved"||actor!=did||expectedVersion!=actualVersion||expectedSpec!=actualSpec{writeJSON(w,409,map[string]string{"error":"install_not_approved_or_changed"});return}
 var visibility string
 err=tx.QueryRowContext(r.Context(),`SELECT visibility FROM creator_publications WHERE protocol_id=?`,pub).Scan(&visibility)
 if err!=nil||visibility=="private"{writeJSON(w,409,map[string]string{"error":"publication_not_accessible"});return}
 var plan runner.Plan
 if json.Unmarshal([]byte(spec),&plan)!=nil{writeJSON(w,400,map[string]string{"error":"invalid_fake_plan"});return}
 // Validate the complete fake plan before reserving the durable request.
 preview:=runner.New(true)
 snapshot,err:=preview.Start(plan)
 if err!=nil{writeJSON(w,400,map[string]string{"error":"invalid_fake_plan"});return}
 id,err:=newRunID();if err!=nil{writeJSON(w,503,map[string]string{"error":"fake_run_unavailable"});return}
 audit,err:=newRunID();if err!=nil{writeJSON(w,503,map[string]string{"error":"fake_run_unavailable"});return}
 now:=time.Now().UTC().Format(time.RFC3339Nano)
 _,err=tx.ExecContext(r.Context(),`INSERT INTO clear_fake_install_runs(id,install_request_id,user_id,actor_did,publication_id,manifest_sha256,status,created_at) VALUES(?,?,?,?,?,?,'reserved',?)`,id,input.InstallRequestID,user,did,pub,manifest,now)
 if err!=nil{writeJSON(w,409,map[string]string{"error":"install_already_used"});return}
 _,err=tx.ExecContext(r.Context(),`INSERT INTO clear_fake_install_audit(id,run_id,action,recorded_at) VALUES(?,?,'reserved',?)`,audit,id,now)
 if err!=nil||tx.Commit()!=nil{writeJSON(w,503,map[string]string{"error":"fake_run_unavailable"});return}
 if err=s.fakeInstallRunPush(r.Context());err!=nil{writeJSON(w,503,map[string]string{"error":"audit_persistence_unavailable"});return}
 // Local declarative in-memory simulation only: never a real executor.
 result,runErr:=preview.Run(r.Context(),snapshot.ID)
 finalStatus:="completed";if runErr!=nil{finalStatus="failed"}
 encoded,err:=json.Marshal(result);if err!=nil{writeJSON(w,503,map[string]string{"error":"audit_persistence_unavailable"});return}
 tx2,err:=s.fakeInstallRunDB.BeginTx(r.Context(),nil);if err!=nil{writeJSON(w,503,map[string]string{"error":"audit_persistence_unavailable"});return};defer tx2.Rollback()
 now=time.Now().UTC().Format(time.RFC3339Nano)
 res,err:=tx2.ExecContext(r.Context(),`UPDATE clear_fake_install_runs SET status=?,result_json=?,finished_at=? WHERE id=? AND status='reserved'`,finalStatus,string(encoded),now,id)
 if err!=nil{writeJSON(w,503,map[string]string{"error":"audit_persistence_unavailable"});return}
 n,err:=res.RowsAffected();if err!=nil||n!=1{writeJSON(w,503,map[string]string{"error":"audit_persistence_unavailable"});return}
 audit,err=newRunID();if err!=nil{writeJSON(w,503,map[string]string{"error":"audit_persistence_unavailable"});return}
 _,err=tx2.ExecContext(r.Context(),`INSERT INTO clear_fake_install_audit(id,run_id,action,recorded_at) VALUES(?,?,?,?)`,audit,id,finalStatus,now)
 if err!=nil||tx2.Commit()!=nil{writeJSON(w,503,map[string]string{"error":"audit_persistence_unavailable"});return}
 if err=s.fakeInstallRunPush(r.Context());err!=nil{writeJSON(w,503,map[string]string{"error":"audit_persistence_unavailable"});return}
 w.Header().Set("Cache-Control","no-store")
 writeJSON(w,201,map[string]any{"id":id,"status":finalStatus,"sandboxOnly":true,"realExecution":false})
}
func(s *Server) getFakeInstallRun(w http.ResponseWriter,r *http.Request){
 if s.fakeInstallRunDB==nil{writeJSON(w,503,map[string]string{"error":"fake_run_unavailable"});return}
 user,_,ok:=s.fakeRunPrincipal(w,r);if !ok{return}
 var status,result string
 err:=s.fakeInstallRunDB.QueryRowContext(r.Context(),`SELECT status,result_json FROM clear_fake_install_runs WHERE id=? AND user_id=?`,r.PathValue("id"),user).Scan(&status,&result)
 if errors.Is(err,sql.ErrNoRows){writeJSON(w,404,map[string]string{"error":"not_found"});return}
 if err!=nil{writeJSON(w,503,map[string]string{"error":"fake_run_unavailable"});return}
 w.Header().Set("Cache-Control","no-store")
 writeJSON(w,200,map[string]any{"status":status,"sandboxOnly":true,"realExecution":false,"snapshot":json.RawMessage(result)})
}
