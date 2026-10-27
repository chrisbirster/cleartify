package api

import (
 "net/http"
 "strings"
 "time"
 "errors"
 "github.com/chrisbirster/cleartify/internal/followups"
)
func(s *Server) clearMessage(w http.ResponseWriter,r *http.Request){
 w.Header().Set("Cache-Control","no-store")
 if s.appOrigin==""||r.Header.Get("Origin")!=s.appOrigin{writeJSON(w,403,map[string]string{"error":"invalid_origin"});return}
 user,ok:=s.clearPrincipal(w,r);if !ok{return}
 var req struct{Text string `json:"text"`}
 if !decodeJSON(w,r,&req){return}
 text:=strings.TrimSpace(req.Text)
 if len(text)==0||len(text)>4096{writeJSON(w,400,map[string]string{"error":"invalid_message"});return}
 if strings.HasPrefix(text,"done "){
  id:=strings.TrimSpace(strings.TrimPrefix(text,"done "))
  if err:=s.followups.Done(r.Context(),user,id);err!=nil{
   if errors.Is(err,followups.ErrNotFound){writeJSON(w,404,map[string]string{"error":"followup_not_found"});return}
   writeJSON(w,503,map[string]string{"error":"unavailable"});return
  }
  writeJSON(w,200,map[string]any{"reply":"Marked done.","completed":id});return
 }
 if strings.HasPrefix(strings.ToLower(text),"remind me to "){
  body:=strings.TrimSpace(text[len("remind me to "):])
  if body==""{writeJSON(w,400,map[string]string{"error":"missing_task"});return}
  f,err:=s.followups.Create(r.Context(),user,body,time.Now().UTC().Add(24*time.Hour))
  if err!=nil{writeJSON(w,400,map[string]string{"error":"invalid_followup"});return}
  writeJSON(w,201,map[string]any{"reply":"I'll follow up tomorrow. Reply done "+f.ID+" when it's finished.","followup":f});return
 }
 writeJSON(w,200,map[string]any{"reply":"I can follow through. Try: remind me to [task], or done [followup ID]."})
}
func(s *Server) listClearFollowups(w http.ResponseWriter,r *http.Request){
 user,ok:=s.clearPrincipal(w,r);if !ok{return}
 items,err:=s.followups.List(r.Context(),user);if err!=nil{writeJSON(w,503,map[string]string{"error":"unavailable"});return}
 w.Header().Set("Cache-Control","no-store");writeJSON(w,200,map[string]any{"followups":items})
}
func(s *Server) listClearNudges(w http.ResponseWriter,r *http.Request){
 user,ok:=s.clearPrincipal(w,r);if !ok{return}
 items,err:=s.followups.Inbox(r.Context(),user);if err!=nil{writeJSON(w,503,map[string]string{"error":"unavailable"});return}
 w.Header().Set("Cache-Control","no-store");writeJSON(w,200,map[string]any{"nudges":items})
}
