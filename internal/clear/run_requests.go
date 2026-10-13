package clear

import (
 "context"
 "crypto/rand"
 "database/sql"
 "encoding/hex"
 "errors"
 "time"
)
var ErrNotPublished=errors.New("protocol is not published")
var ErrNotPending=errors.New("run request is not pending")
type Service struct{DB *sql.DB}
func(s Service) Request(ctx context.Context,userID,publicationID string)(string,error){
 if userID==""||publicationID=="" {return "",ErrNotPublished}
 var exists int
 err:=s.DB.QueryRowContext(ctx,`SELECT 1 FROM published_protocols WHERE id=? AND status='published'`,publicationID).Scan(&exists)
 if errors.Is(err,sql.ErrNoRows){return "",ErrNotPublished};if err!=nil{return "",err}
 var installed int
 err=s.DB.QueryRowContext(ctx,`SELECT 1 FROM protocol_installations WHERE user_id=? AND enabled=1 LIMIT 1`,userID).Scan(&installed)
 if errors.Is(err,sql.ErrNoRows){return "",ErrNotPublished};if err!=nil{return "",err}
 idBytes:=make([]byte,18);if _,err=rand.Read(idBytes);err!=nil{return "",err}
 id:=hex.EncodeToString(idBytes)
 _,err=s.DB.ExecContext(ctx,`INSERT INTO clear_run_requests(id,user_id,publication_id,status,requested_at) VALUES(?,?,?,'pending_approval',?)`,id,userID,publicationID,time.Now().UTC().Format(time.RFC3339Nano))
 return id,err
}
func(s Service) Decide(ctx context.Context,userID,requestID string,approve bool)error{
 status:="rejected";if approve {status="approved"}
 result,err:=s.DB.ExecContext(ctx,`UPDATE clear_run_requests SET status=?,decided_at=? WHERE id=? AND user_id=? AND status='pending_approval'`,status,time.Now().UTC().Format(time.RFC3339Nano),requestID,userID)
 if err!=nil{return err}
 n,err:=result.RowsAffected();if err!=nil{return err};if n!=1{return ErrNotPending}
 return nil
}
