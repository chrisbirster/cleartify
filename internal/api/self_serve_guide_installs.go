package api

import (
 "bytes"
 "crypto/rand"
 "crypto/sha256"
 "database/sql"
 "encoding/hex"
 "encoding/json"
 "errors"
 "net/http"
 "strings"
 "strconv"
 "time"
 "github.com/chrisbirster/cleartify/internal/protocols"
)

func(s *Server) selfServeGuideInstall(w http.ResponseWriter,r *http.Request){
 w.Header().Set("Cache-Control","no-store")
 if s.installFlow==nil||s.installFlow.Store==nil||s.clearSessions==nil {writeJSON(w,503,map[string]string{"error":"unavailable"});return}
 if s.appOrigin==""||r.Header.Get("Origin")!=s.appOrigin{writeJSON(w,403,map[string]string{"error":"invalid_origin"});return}
 user,ok:=s.clearPrincipal(w,r);if !ok{return}
 var req struct{PublicationID string `json:"publicationId"`;AcknowledgedManifestSHA256 string `json:"acknowledgedManifestSha256"`;CreateFollowups bool `json:"createFollowups"`;FollowupDueAt string `json:"followupDueAt"`}
 if !decodeJSON(w,r,&req){return}
 if len(req.PublicationID)<1||len(req.PublicationID)>128||len(req.AcknowledgedManifestSHA256)!=64{writeJSON(w,400,map[string]string{"error":"invalid_request"});return}
 var available int
 err:=s.installFlow.Store.DB.QueryRowContext(r.Context(),`SELECT COUNT(*) FROM published_protocols p WHERE p.id=? AND p.status='published' AND EXISTS (SELECT 1 FROM creator_publications c WHERE c.protocol_id=p.id AND c.visibility='public') AND NOT EXISTS (SELECT 1 FROM creator_takedown_requests t WHERE t.protocol_id=p.id AND t.status='hidden')`,req.PublicationID).Scan(&available)
 if err!=nil{writeJSON(w,503,map[string]string{"error":"unavailable"});return}
 if available!=1{writeJSON(w,404,map[string]string{"error":"publication_unavailable"});return}
 manifest,err:=s.installFlow.Preview(r.Context(),req.PublicationID)
 if err!=nil{if errors.Is(err,protocols.ErrInstallUnavailable){writeJSON(w,404,map[string]string{"error":"publication_unavailable"});return};writeJSON(w,503,map[string]string{"error":"unavailable"});return}
 if req.AcknowledgedManifestSHA256!=manifest.ManifestSHA256 {writeJSON(w,409,map[string]string{"error":"manifest_changed"});return}
 if len(manifest.Permissions)!=0{writeJSON(w,403,map[string]string{"error":"approval_required"});return}
 tx,err:=s.installFlow.Store.DB.BeginTx(r.Context(),nil);if err!=nil{writeJSON(w,503,map[string]string{"error":"unavailable"});return};defer tx.Rollback()
 var slug,spec string
 var version int
 err=tx.QueryRowContext(r.Context(),`SELECT slug,version,spec_json FROM published_protocols p
 WHERE p.id=? AND p.status='published'
 AND EXISTS (SELECT 1 FROM creator_publications c WHERE c.protocol_id=p.id AND c.visibility='public')
 AND NOT EXISTS (SELECT 1 FROM creator_takedown_requests t WHERE t.protocol_id=p.id AND t.status='hidden')`,req.PublicationID).Scan(&slug,&version,&spec)
 if errors.Is(err,sql.ErrNoRows){writeJSON(w,404,map[string]string{"error":"publication_unavailable"});return}
 if err!=nil{writeJSON(w,503,map[string]string{"error":"unavailable"});return}
 digest:=sha256.Sum256([]byte(spec))
 if hex.EncodeToString(digest[:])!=manifest.SpecSHA256 || version!=manifest.Version || strings.TrimSpace(slug)=="" {writeJSON(w,409,map[string]string{"error":"manifest_changed"});return}
 // Only allow text-only, permissionless guide steps, not any executable operation.
 var body struct{Steps []json.RawMessage `json:"steps"`;Permissions []string `json:"permissions"`;RemixOf string `json:"remixOf"`;Attribution string `json:"attribution"`}
 dec:=json.NewDecoder(bytes.NewReader([]byte(spec)));dec.DisallowUnknownFields()
 if err=dec.Decode(&body);err!=nil||len(body.Steps)==0||len(body.Steps)>32||len(body.Permissions)!=0 {writeJSON(w,403,map[string]string{"error":"approval_required"});return}
 var guideSteps []struct{Text string;DueAt string}
 for _,raw:=range body.Steps{
  var step struct{Operation string `json:"operation"`;Text string `json:"text"`;DueAt string `json:"dueAt"`}
  d:=json.NewDecoder(bytes.NewReader(raw));d.DisallowUnknownFields()
  if d.Decode(&step)!=nil||step.Operation!="guide"||strings.TrimSpace(step.Text)==""||len(step.Text)>4096 {writeJSON(w,403,map[string]string{"error":"approval_required"});return}
   guideSteps=append(guideSteps,struct{Text string;DueAt string}{step.Text,step.DueAt})
 }
 // Idempotency is per private Clear user, publication, and pinned version.
 var existing string
 err=tx.QueryRowContext(r.Context(),`SELECT id FROM protocol_installations WHERE user_id=? AND protocol_key=? AND version=? LIMIT 1`,user,req.PublicationID,version).Scan(&existing)
 if err==nil {if err=tx.Commit();err!=nil{writeJSON(w,503,map[string]string{"error":"unavailable"});return};writeJSON(w,200,map[string]any{"id":existing,"installed":true,"existing":true});return}
 if !errors.Is(err,sql.ErrNoRows){writeJSON(w,503,map[string]string{"error":"unavailable"});return}
 var b [18]byte;if _,err=rand.Read(b[:]);err!=nil{writeJSON(w,503,map[string]string{"error":"unavailable"});return}
 id:=hex.EncodeToString(b[:])
 _,err=tx.ExecContext(r.Context(),`INSERT INTO protocol_installations(id,user_id,protocol_key,version,source,enabled,created_at) VALUES(?,?,?,?, 'self_serve_guide',1,?)`,id,user,req.PublicationID,version,time.Now().UTC().Format(time.RFC3339Nano))
 if err!=nil{
  if strings.Contains(strings.ToLower(err.Error()),"unique constraint")||strings.Contains(strings.ToLower(err.Error()),"constraint failed"){writeJSON(w,409,map[string]string{"error":"install_already_exists_or_version_conflict"});return}
  writeJSON(w,503,map[string]string{"error":"unavailable"});return
 }
 followupCount:=0;defaulted:=false
 if req.CreateFollowups {
  if s.followups==nil{writeJSON(w,503,map[string]string{"error":"followups_disabled"});return}
  if len(guideSteps)>20{writeJSON(w,429,map[string]string{"error":"followup_limit"});return}
  now:=s.followups.NowUTC()
  var existingCount int
  err=tx.QueryRowContext(r.Context(),"SELECT COUNT(*) FROM clear_followups WHERE user_id=? AND created_at>=?",user,now.Add(-24*time.Hour).Format(time.RFC3339Nano)).Scan(&existingCount)
  if err!=nil{writeJSON(w,503,map[string]string{"error":"unavailable"});return}
  if existingCount+len(guideSteps)>20{writeJSON(w,429,map[string]string{"error":"followup_limit"});return}
  for i,step:=range guideSteps{
   due:=now.Add(24*time.Hour)
   if step.DueAt!="" {parsed,e:=time.Parse(time.RFC3339,step.DueAt);if e!=nil{writeJSON(w,400,map[string]string{"error":"invalid_due_at"});return};due=parsed.UTC()} else if req.FollowupDueAt!=""{parsed,e:=time.Parse(time.RFC3339,req.FollowupDueAt);if e!=nil{writeJSON(w,400,map[string]string{"error":"invalid_due_at"});return};due=parsed.UTC()}else{defaulted=true}
   _,err=tx.ExecContext(r.Context(),`INSERT INTO clear_followups(id,user_id,text,due_at,status,created_at) VALUES(?,?,?,?,'open',?)`,id+"-step-"+strconv.Itoa(i+1),user,step.Text,due.Format(time.RFC3339Nano),now.Format(time.RFC3339Nano))
   if err!=nil{writeJSON(w,503,map[string]string{"error":"unavailable"});return}
   followupCount++
  }
 }
 _,err=tx.ExecContext(r.Context(),`INSERT INTO guide_install_audit(id,user_id,publication_id,version,manifest_sha256,action,created_at)
 VALUES(?,?,?,?,?,'installed',?)`,id,user,req.PublicationID,version,req.AcknowledgedManifestSHA256,time.Now().UTC().Format(time.RFC3339Nano))
 if err!=nil{writeJSON(w,503,map[string]string{"error":"unavailable"});return}
 if err=tx.Commit();err!=nil{writeJSON(w,503,map[string]string{"error":"unavailable"});return}
 if err=s.installFlow.Store.Push(r.Context());err!=nil{writeJSON(w,503,map[string]string{"error":"unavailable"});return}
 channel:="none";delivery:="not_requested"
 if followupCount>0 {channel=s.followups.ChannelName();delivery="pending"}
 writeJSON(w,201,map[string]any{"id":id,"installed":true,"executed":false,"followupsCreated":followupCount,"dueTimeDefaulted":defaulted,"deliveryChannel":channel,"deliveryStatus":delivery})
}
