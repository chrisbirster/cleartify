package marketplace

import (
 "context"
 "database/sql"
 "errors"
 "sync"
 "testing"
 "time"

 "github.com/chrisbirster/cleartify/internal/database"
)

func setupReconciliationJobTest(t *testing.T)(*sql.DB,string){
 t.Helper()
 ctx:=context.Background()
 store,err:=database.Open(ctx,database.Config{Path:t.TempDir()+"/local.db"})
 if err!=nil{t.Fatal(err)}
 t.Cleanup(func(){_ = store.Close()})
 if err=store.Migrate(ctx);err!=nil{t.Fatal(err)}
 for _,q:=range []string{
  `INSERT INTO published_protocols(id,creator_did,slug,version,title,summary,license,spec_json,status,created_at) VALUES('p','did:plc:seller','slug',1,'Title','','MIT','{}','published','now')`,
  `INSERT INTO marketplace_listings(id,protocol_id,creator_did,title,pricing_kind,price_minor,currency,status,created_at) VALUES('l','p','did:plc:seller','Title','one_time',101,'USD','listed','now')`,
 }{if _,err=store.DB.ExecContext(ctx,q);err!=nil{t.Fatal(err)}}
 svc:=Service{DB:store.DB,Provider:FakeProvider{},Enabled:true}
 id,err:=svc.Purchase(ctx,"did:plc:buyer","l");if err!=nil{t.Fatal(err)}
 return store.DB,id
}
func newJob(db *sql.DB,now *time.Time)ReconciliationJob{
 return ReconciliationJob{Accounting:FakeAccounting{DB:db,Enabled:true,FeeBPS:1000},
 Enabled:true,Lease:20*time.Second,Backoff:time.Second,MaxAttempts:3,Now:func()time.Time{return *now}}
}
func TestReconciliationJobDisabledAndIdempotentSettlement(t *testing.T){
 db,id:=setupReconciliationJobTest(t);ctx:=context.Background()
 now:=time.Now().UTC()
 j:=newJob(db,&now)
 j.Enabled=false
 if err:=j.Enqueue(ctx,"once");!errors.Is(err,ErrReconciliationJobDisabled){t.Fatal(err)}
 if _,err:=j.Tick(ctx);!errors.Is(err,ErrReconciliationJobDisabled){t.Fatal(err)}
 j.Enabled=true
 if err:=j.Enqueue(ctx,"once");err!=nil{t.Fatal(err)}
 if err:=j.Enqueue(ctx,"once");err!=nil{t.Fatal(err)}
 if worked,err:=j.Tick(ctx);!worked||err!=nil{t.Fatal(worked,err)}
 if worked,err:=j.Tick(ctx);worked||err!=nil{t.Fatal("duplicate run",worked,err)}
 var n int
 if err:=db.QueryRowContext(ctx,`SELECT COUNT(*) FROM fake_seller_ledger WHERE entitlement_id=?`,id).Scan(&n);err!=nil||n!=1{t.Fatal(n,err)}
 svc:=Service{DB:db,Provider:FakeProvider{},Enabled:true}
 if err:=svc.Refund(ctx,"did:plc:buyer",id);err!=nil{t.Fatal(err)}
 j.Accounting.FeeBPS=2000
 if err:=j.Enqueue(ctx,"after-refund");err!=nil{t.Fatal(err)}
 if worked,err:=j.Tick(ctx);!worked||err!=nil{t.Fatal(worked,err)}
 var gross,fees,net int64
 if err:=db.QueryRowContext(ctx,`SELECT COALESCE(SUM(gross_minor),0),COALESCE(SUM(fee_minor),0),COALESCE(SUM(net_minor),0) FROM fake_seller_ledger WHERE entitlement_id=?`,id).Scan(&gross,&fees,&net);err!=nil||gross!=0||fees!=0||net!=0{t.Fatalf("settlement invariants %d/%d/%d %v",gross,fees,net,err)}
}
func TestReconciliationJobLeaseRecoveryAndFence(t *testing.T){
 db,_:=setupReconciliationJobTest(t);ctx:=context.Background()
 now:=time.Now().UTC();j:=newJob(db,&now)
 if err:=j.Enqueue(ctx,"crash");err!=nil{t.Fatal(err)}
 _,err:=db.ExecContext(ctx,`UPDATE fake_reconciliation_jobs SET state='leased',lease_token='old',lease_until=? WHERE id='crash'`,stamp(now.Add(-time.Second)))
 if err!=nil{t.Fatal(err)}
 if worked,err:=j.Tick(ctx);!worked||err!=nil{t.Fatal(worked,err)}
 var state,token string
 if err:=db.QueryRowContext(ctx,`SELECT state,COALESCE(lease_token,'') FROM fake_reconciliation_jobs WHERE id='crash'`).Scan(&state,&token);err!=nil||state!="done"||token!=""{t.Fatal(state,token,err)}
 // The old worker cannot change a reclaimed/completed record.
 res,err:=db.ExecContext(ctx,`UPDATE fake_reconciliation_jobs SET state='dead' WHERE id='crash' AND state='leased' AND lease_token='old'`)
 if err!=nil{t.Fatal(err)}
 n,err:=res.RowsAffected();if err!=nil||n!=0{t.Fatal("stale lease committed",n,err)}
}
func TestReconciliationJobMissingSnapshotRetryAndDead(t *testing.T){
 db,_:=setupReconciliationJobTest(t);ctx:=context.Background()
 _,err:=db.ExecContext(ctx,`INSERT INTO marketplace_entitlements(id,listing_id,buyer_did,provider,provider_receipt,status,created_at) VALUES('legacy','l','did:plc:legacy','fake','fake-legacy','active','now')`)
 if err!=nil{t.Fatal(err)}
 _,err=db.ExecContext(ctx,`INSERT INTO marketplace_payment_audit(id,entitlement_id,event,actor_did,recorded_at) VALUES('audit-legacy','legacy','fake_purchase','did:plc:legacy','now')`)
 if err!=nil{t.Fatal(err)}
 now:=time.Now().UTC();j:=newJob(db,&now);j.MaxAttempts=2
 if err:=j.Enqueue(ctx,"legacy");err!=nil{t.Fatal(err)}
 for attempt:=0;attempt<2;attempt++{
  worked,err:=j.Tick(ctx)
  if !worked||!errors.Is(err,ErrMissingSaleSnapshot){t.Fatal(worked,err)}
  var n int
  if err:=db.QueryRowContext(ctx,`SELECT COUNT(*) FROM fake_seller_ledger`).Scan(&n);err!=nil||n!=0{t.Fatal("partial ledger",n,err)}
  now=now.Add(2*time.Second)
 }
 var state string
 if err:=db.QueryRowContext(ctx,`SELECT state FROM fake_reconciliation_jobs WHERE id='legacy'`).Scan(&state);err!=nil||state!="dead"{t.Fatal(state,err)}
}
func TestReconciliationJobConcurrentWorkers(t *testing.T){
 db,id:=setupReconciliationJobTest(t);ctx:=context.Background()
 now:=time.Now().UTC()
 j:=newJob(db,&now)
 if err:=j.Enqueue(ctx,"parallel");err!=nil{t.Fatal(err)}
 var wg sync.WaitGroup
 var mu sync.Mutex
 succeeded:=0
 for i:=0;i<4;i++{
  wg.Add(1)
  go func(){
   defer wg.Done()
   worked,err:=j.Tick(ctx)
   if worked&&err==nil{mu.Lock();succeeded++;mu.Unlock()}
  }()
 }
 wg.Wait()
 if succeeded!=1{t.Fatalf("workers finished %d jobs; want one",succeeded)}
 var n int
 if err:=db.QueryRowContext(ctx,`SELECT COUNT(*) FROM fake_seller_ledger WHERE entitlement_id=? AND entry_kind='sale'`,id).Scan(&n);err!=nil||n!=1{t.Fatal("duplicate sale",n,err)}
}
