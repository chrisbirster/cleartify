package protocols

import(
 "context"
 "crypto/rand"
 "database/sql"
 "encoding/hex"
 "errors"
 "time"
)
var ErrTakedownDenied=errors.New("creator takedown denied")
func takedownID()(string,error){var b [16]byte;_,err:=rand.Read(b[:]);return hex.EncodeToString(b[:]),err}
type Takedowns struct{DB *sql.DB}
func(t Takedowns) Request(ctx context.Context,actor,protocol,reasonKey string)(string,error){
 if t.DB==nil||len(actor)<5||actor[:4]!="did:"||protocol==""||reasonKey!="rights_review"&&reasonKey!="identity_review"&&reasonKey!="other_review"{return "",ErrTakedownDenied}
 tx,err:=t.DB.BeginTx(ctx,nil);if err!=nil{return "",err};defer tx.Rollback()
 var owner string
 if err=tx.QueryRowContext(ctx,"SELECT creator_did FROM published_protocols WHERE id=?",protocol).Scan(&owner);err!=nil||owner!=actor{return "",ErrTakedownDenied}
 id,err:=takedownID();if err!=nil{return "",err};audit,err:=takedownID();if err!=nil{return "",err}
 now:=time.Now().UTC().Format(time.RFC3339Nano)
 if _,err=tx.ExecContext(ctx,`INSERT INTO creator_takedown_requests(id,protocol_id,requester_did,reason_key,status,created_at) VALUES(?,?,?,?,'pending',?)`,id,protocol,actor,reasonKey,now);err!=nil{return "",ErrTakedownDenied}
 if _,err=tx.ExecContext(ctx,`INSERT INTO creator_takedown_audit(id,request_id,action,actor_did,recorded_at) VALUES(?,?,'requested',?,?)`,audit,id,actor,now);err!=nil{return "",err}
 return id,tx.Commit()
}
// Decide is only reachable after the caller's independent admin authentication.
// Hiding/restoring never destroys the protocol, original visibility or lineage.
func(t Takedowns) Decide(ctx context.Context,id,decision string)error{
 if t.DB==nil||id==""||decision!="hidden"&&decision!="rejected"&&decision!="restored"{return ErrTakedownDenied}
 tx,err:=t.DB.BeginTx(ctx,nil);if err!=nil{return err};defer tx.Rollback()
 var previous string
 err=tx.QueryRowContext(ctx,"SELECT status FROM creator_takedown_requests WHERE id=?",id).Scan(&previous)
 if err!=nil{return ErrTakedownDenied}
 if !(previous=="pending"&&(decision=="hidden"||decision=="rejected")||previous=="hidden"&&decision=="restored"){return ErrTakedownDenied}
 audit,err:=takedownID();if err!=nil{return err};now:=time.Now().UTC().Format(time.RFC3339Nano)
 res,err:=tx.ExecContext(ctx,"UPDATE creator_takedown_requests SET status=?,decided_at=? WHERE id=? AND status=?",decision,now,id,previous)
 if err!=nil{return err};n,err:=res.RowsAffected();if err!=nil||n!=1{return ErrTakedownDenied}
 if _,err=tx.ExecContext(ctx,`INSERT INTO creator_takedown_audit(id,request_id,action,actor_did,recorded_at) VALUES(?,?,?,'admin',?)`,audit,id,decision,now);err!=nil{return err}
 return tx.Commit()
}
func(t Takedowns) Lineage(ctx context.Context,id string)([]string,error){
 if t.DB==nil||id==""{return nil,ErrTakedownDenied}
 out:=[]string{};seen:=map[string]bool{}
 for len(out)<32{
  if seen[id]{return nil,ErrTakedownDenied};seen[id]=true
  var parent string
  err:=t.DB.QueryRowContext(ctx,"SELECT parent_protocol_id FROM creator_remix_lineage WHERE protocol_id=?",id).Scan(&parent)
  if errors.Is(err,sql.ErrNoRows){return out,nil};if err!=nil{return nil,err}
  out=append(out,parent);id=parent
 }
 return nil,ErrTakedownDenied
}
