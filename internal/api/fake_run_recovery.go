package api

import (
 "context"
 "crypto/rand"
 "database/sql"
 "encoding/hex"
 "errors"
 "fmt"
 "time"
)

// RecoverReservedFakeRuns is a startup-only fail-safe reconciliation.
// A reserved run may have executed before a crash, so we must NEVER rerun it.
// Each interrupted run is failed with a durable audit row in the SAME transaction.
// Calling again is idempotent; completed/failed runs are untouched.
// The caller must finish this before accepting any new run request.
func RecoverReservedFakeRuns(ctx context.Context, db *sql.DB, push func(context.Context) error) (int,error) {
 if db==nil||push==nil{return 0,errors.New("fake run recovery dependencies unavailable")}
 tx,err:=db.BeginTx(ctx,nil);if err!=nil{return 0,err}
 defer tx.Rollback()
 rows,err:=tx.QueryContext(ctx,`SELECT id FROM clear_fake_install_runs WHERE status='reserved' ORDER BY created_at,id`)
 if err!=nil{return 0,fmt.Errorf("enumerate reserved fake runs: %w",err)}
 ids:=[]string{}
 for rows.Next(){var id string;if err=rows.Scan(&id);err!=nil{rows.Close();return 0,err};ids=append(ids,id)}
 if err=rows.Err();err!=nil{rows.Close();return 0,err}
 if err=rows.Close();err!=nil{return 0,err}
 now:=time.Now().UTC().Format(time.RFC3339Nano)
 for _,id:=range ids {
  res,e:=tx.ExecContext(ctx,`UPDATE clear_fake_install_runs SET status='failed',result_json=?,finished_at=? WHERE id=? AND status='reserved'`,`{"recovery":"interrupted_before_terminal_audit","executed":false,"outcome":"unknown"}`,now,id)
  if e!=nil{return 0,fmt.Errorf("fail interrupted fake run: %w",e)}
  affected,e:=res.RowsAffected();if e!=nil{return 0,e};if affected!=1{return 0,fmt.Errorf("reserved fake run changed during recovery: %s",id)}
  var b [16]byte;if _,e=rand.Read(b[:]);e!=nil{return 0,e}
  if _,e=tx.ExecContext(ctx,`INSERT INTO clear_fake_install_audit(id,run_id,action,recorded_at) VALUES(?,?,?,?)`,hex.EncodeToString(b[:]),id,"recovered_failed",now);e!=nil{return 0,fmt.Errorf("audit interrupted fake run: %w",e)}
 }
 if err=tx.Commit();err!=nil{return 0,err}
 if len(ids)>0 {if err=push(ctx);err!=nil{return len(ids),fmt.Errorf("push fake run recovery audit: %w",err)}}
 return len(ids),nil
}
