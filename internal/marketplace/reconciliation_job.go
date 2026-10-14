package marketplace

import (
 "context"
 "database/sql"
 "errors"
 "fmt"
 "time"
)

var ErrReconciliationJobDisabled = errors.New("fake reconciliation job disabled")
var ErrReconciliationLeaseLost = errors.New("fake reconciliation lease lost")
var ErrReconciliationJobConfig = errors.New("invalid fake reconciliation job configuration")

// ReconciliationJob is an explicitly invoked, fake-only worker. Nothing is
// scheduled in the application runtime; callers must opt in and call Tick.
// Fencing guards job state transitions. Accounting writes are independently
// transactional and idempotent by immutable ledger (entitlement,entry_kind).
type ReconciliationJob struct {
 Accounting FakeAccounting
 Enabled bool
 Lease time.Duration
 Backoff time.Duration
 MaxAttempts int
 Now func()time.Time
}
func(j ReconciliationJob) now()time.Time {
 if j.Now!=nil{return j.Now().UTC()}
 return time.Now().UTC()
}
func(j ReconciliationJob) check()error{
 if !j.Enabled{return ErrReconciliationJobDisabled}
 if j.Accounting.check()!=nil || j.Lease<=0 || j.Lease>time.Minute || j.Backoff<=0 || j.Backoff>time.Minute || j.MaxAttempts<1 || j.MaxAttempts>10 {return ErrReconciliationJobConfig}
 return nil
}
func(j ReconciliationJob) db()*sql.DB{return j.Accounting.DB}
func stamp(t time.Time)string{return t.UTC().Format(time.RFC3339Nano)}

// Enqueue schedules a single named local reconciliation request. Repeated
// enqueue is a no-op, including after success: callers use a fresh ID for a
// later pass. The job ID is not used in financial ledger identity.
func(j ReconciliationJob) Enqueue(ctx context.Context,id string)error{
 if err:=j.check();err!=nil{return err}
 if id==""||len(id)>128{return ErrReconciliationJobConfig}
 now:=stamp(j.now())
 _,err:=j.db().ExecContext(ctx,`INSERT INTO fake_reconciliation_jobs(id,state,attempts,available_at,updated_at)
 VALUES(?,'pending',0,?,?) ON CONFLICT(id) DO NOTHING`,id,now,now)
 return err
}

// Tick claims at most one due job and invokes the existing fake accounting.
// An expired lease can be reclaimed; stale workers cannot finalize or replace
// a newer claim. Reconciliation remains idempotent even after a worker crash
// between ledger commit and job acknowledgement.
func(j ReconciliationJob) Tick(ctx context.Context)(bool,error){
 if err:=j.check();err!=nil{return false,err}
 now:=j.now()
 token,err:=randomID();if err!=nil{return false,err}
 tx,err:=j.db().BeginTx(ctx,nil);if err!=nil{return false,err}
 defer tx.Rollback()
 var id string
 err=tx.QueryRowContext(ctx,`SELECT id FROM fake_reconciliation_jobs
 WHERE (state='pending' AND available_at<=?) OR (state='leased' AND lease_until<=?)
 ORDER BY available_at,id LIMIT 1`,stamp(now),stamp(now)).Scan(&id)
 if errors.Is(err,sql.ErrNoRows){return false,nil}
 if err!=nil{return false,err}
 res,err:=tx.ExecContext(ctx,`UPDATE fake_reconciliation_jobs SET state='leased',lease_token=?,lease_until=?,attempts=attempts+1,updated_at=?
 WHERE id=? AND ((state='pending' AND available_at<=?) OR (state='leased' AND lease_until<=?))`,
 token,stamp(now.Add(j.Lease)),stamp(now),id,stamp(now),stamp(now))
 if err!=nil{return false,err}
 n,err:=res.RowsAffected();if err!=nil{return false,err}
 if n!=1{return false,nil}
 var attempts int
 if err=tx.QueryRowContext(ctx,`SELECT attempts FROM fake_reconciliation_jobs WHERE id=? AND lease_token=?`,id,token).Scan(&attempts);err!=nil{return false,err}
 if err=tx.Commit();err!=nil{return false,err}
 // Check the lease again before accounting; final acknowledgement is fenced.
 // This is not a substitute for the ledger's transaction/idempotency guards.
 var current string
 err=j.db().QueryRowContext(ctx,`SELECT lease_token FROM fake_reconciliation_jobs WHERE id=? AND state='leased' AND lease_until>?`,id,stamp(j.now())).Scan(&current)
 if err!=nil||current!=token{return true,ErrReconciliationLeaseLost}
 inserted,runErr:=j.Accounting.RunReconciliationJob(ctx)
 state:="done"
 next:=now
 failure:=""
 if runErr!=nil{
  failure="reconciliation_failed"
  if errors.Is(runErr,ErrMissingSaleSnapshot){failure="missing_sale_snapshot"}
  if attempts>=j.MaxAttempts{state="dead"}else{
   state="pending"
   delay:=j.Backoff
   for i:=1;i<attempts;i++{
    if delay>=time.Hour/2{delay=time.Hour;break}
    delay*=2
   }
   if delay>time.Hour{delay=time.Hour}
   next=j.now().Add(delay)
  }
 }
 res,err=j.db().ExecContext(ctx,`UPDATE fake_reconciliation_jobs
 SET state=?,available_at=?,lease_token=NULL,lease_until=NULL,last_error_key=?,last_inserted=?,updated_at=?
 WHERE id=? AND state='leased' AND lease_token=? AND lease_until>?`,
 state,stamp(next),failure,inserted,stamp(j.now()),id,token,stamp(j.now()))
 if err!=nil{return true,err}
 n,err=res.RowsAffected();if err!=nil{return true,err}
 if n!=1{return true,fmt.Errorf("%w: %s",ErrReconciliationLeaseLost,id)}
 return true,runErr
}
