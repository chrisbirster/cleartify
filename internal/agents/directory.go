package agents

import (
 "context"
 "database/sql"
)
type DirectoryEntry struct{
 ID string `json:"id"`
 DisplayName string `json:"displayName"`
 Description string `json:"description"`
 OwnerVerification string `json:"ownerVerification"`
 Capabilities []string `json:"capabilities"`
}
// Directory returns a minimum public projection. Registration is not identity
// verification. Unverified/disabled agents must never appear.
func Directory(ctx context.Context,db *sql.DB,limit int)([]DirectoryEntry,error){
 if limit<=0||limit>50{limit=20}
 rows,err:=db.QueryContext(ctx,`SELECT id,display_name,description FROM agent_registrations WHERE status='verified' ORDER BY created_at DESC,id DESC LIMIT ?`,limit)
 if err!=nil{return nil,err}
 defer rows.Close()
 out:=make([]DirectoryEntry,0)
 for rows.Next(){
  var e DirectoryEntry
  if err:=rows.Scan(&e.ID,&e.DisplayName,&e.Description);err!=nil{return nil,err}
  e.OwnerVerification="not_independently_verified"
  e.Capabilities=[]string{}
  out=append(out,e)
 }
 return out,rows.Err()
}
