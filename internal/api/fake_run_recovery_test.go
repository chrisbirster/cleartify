package api

import (
 "context"
 "database/sql"
 "errors"
 "path/filepath"
 "testing"
 "github.com/chrisbirster/cleartify/internal/database"
)

func openRecoveryFixture(t *testing.T,path string)*database.Store{
 t.Helper()
 store,err:=database.Open(context.Background(),database.Config{Path:path});if err!=nil{t.Fatal(err)}
 for _,stmt:=range []string{
  `CREATE TABLE IF NOT EXISTS clear_fake_install_runs(id TEXT PRIMARY KEY,status TEXT NOT NULL,result_json TEXT NOT NULL DEFAULT '{}',created_at TEXT NOT NULL,finished_at TEXT)`,
  `CREATE TABLE IF NOT EXISTS clear_fake_install_audit(id TEXT PRIMARY KEY,run_id TEXT NOT NULL,action TEXT NOT NULL,recorded_at TEXT NOT NULL)`,
 }{if _,err=store.DB.Exec(stmt);err!=nil{store.Close();t.Fatal(err)}}
 return store
}
func TestFakeRunCrashRecoveryPersistentAndIdempotent(t *testing.T){
 ctx:=context.Background();path:=filepath.Join(t.TempDir(),"fake-recovery.db")
 first:=openRecoveryFixture(t,path)
 for _,v:=range []struct{id,status string}{{"pending1","reserved"},{"done","completed"},{"pending2","reserved"},{"failed","failed"}}{
  if _,err:=first.DB.ExecContext(ctx,`INSERT INTO clear_fake_install_runs(id,status,created_at) VALUES(?,?,?)`,v.id,v.status,"2026-10-13T00:00:00Z");err!=nil{t.Fatal(err)}
 }
 if err:=first.Close();err!=nil{t.Fatal(err)}
 restarted:=openRecoveryFixture(t,path)
 pushes:=0;push:=func(context.Context)error{pushes++;return nil}
 n,err:=RecoverReservedFakeRuns(ctx,restarted.DB,push);if err!=nil||n!=2{t.Fatalf("recover %d %v",n,err)}
 if pushes!=1{t.Fatalf("pushes %d",pushes)}
 for _,tc:=range []struct{id,status string}{{"pending1","failed"},{"pending2","failed"},{"done","completed"},{"failed","failed"}}{
  var status,result string
  if err:=restarted.DB.QueryRowContext(ctx,"SELECT status,result_json FROM clear_fake_install_runs WHERE id=?",tc.id).Scan(&status,&result);err!=nil||status!=tc.status{t.Fatalf("%s %s %v",tc.id,status,err)}
  if tc.id=="pending1"&&result==""{t.Fatal("missing fail-safe diagnostic")}
 }
 var audited int
 if err:=restarted.DB.QueryRowContext(ctx,`SELECT count(*) FROM clear_fake_install_audit WHERE action='recovered_failed'`).Scan(&audited);err!=nil||audited!=2{t.Fatalf("audit %d %v",audited,err)}
 if err:=restarted.Close();err!=nil{t.Fatal(err)}
 again:=openRecoveryFixture(t,path);defer again.Close()
 n,err=RecoverReservedFakeRuns(ctx,again.DB,push);if err!=nil||n!=0||pushes!=1{t.Fatalf("idempotence %d %v, pushes %d",n,err,pushes)}
 if err:=again.DB.QueryRowContext(ctx,`SELECT count(*) FROM clear_fake_install_audit WHERE action='recovered_failed'`).Scan(&audited);err!=nil||audited!=2{t.Fatal(audited,err)}
}
func TestFakeRunRecoveryFailClosedAndTransactional(t *testing.T){
 ctx:=context.Background()
 if _,err:=RecoverReservedFakeRuns(ctx,nil,func(context.Context)error{return nil});err==nil{t.Fatal("missing DB accepted")}
 store:=openRecoveryFixture(t,filepath.Join(t.TempDir(),"transaction.db"));defer store.Close()
 if _,err:=RecoverReservedFakeRuns(ctx,store.DB,nil);err==nil{t.Fatal("missing durable push accepted")}
 if _,err:=store.DB.ExecContext(ctx,`INSERT INTO clear_fake_install_runs(id,status,created_at) VALUES('run1','reserved','now')`);err!=nil{t.Fatal(err)}
 // A failing audit insert must roll back the state transition.
 if _,err:=store.DB.ExecContext(ctx,`CREATE TRIGGER block_recovery_audit BEFORE INSERT ON clear_fake_install_audit BEGIN SELECT RAISE(ABORT,'injected audit error'); END`);err!=nil{t.Fatal(err)}
 if _,err:=RecoverReservedFakeRuns(ctx,store.DB,func(context.Context)error{return nil});err==nil{t.Fatal("audit failure not propagated")}
 var status string
 if err:=store.DB.QueryRowContext(ctx,`SELECT status FROM clear_fake_install_runs WHERE id='run1'`).Scan(&status);err!=nil||status!="reserved"{t.Fatalf("transition not rolled back: %s %v",status,err)}
 if _,err:=store.DB.ExecContext(ctx,"DROP TRIGGER block_recovery_audit");err!=nil{t.Fatal(err)}
 n,err:=RecoverReservedFakeRuns(ctx,store.DB,func(context.Context)error{return errors.New("push unavailable")})
 if n!=1||err==nil{t.Fatalf("push failure: %d %v",n,err)}
 if err:=store.DB.QueryRowContext(ctx,`SELECT status FROM clear_fake_install_runs WHERE id='run1'`).Scan(&status);err!=nil||status!="failed"{t.Fatalf("lost local durable audit: %s %v",status,err)}
}
var _ *sql.DB
