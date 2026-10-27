package followups

import (
 "context"
 "path/filepath"
 "testing"
 "time"
 "github.com/chrisbirster/cleartify/internal/database"
)
func TestRestartIdempotentTickAndPrivateDone(t *testing.T){
 ctx:=context.Background();path:=filepath.Join(t.TempDir(),"followups.db")
 db,err:=database.Open(ctx,database.Config{Path:path});if err!=nil{t.Fatal(err)}
 if err=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 now:=time.Date(2026,10,27,12,0,0,0,time.UTC)
 for _,id:=range []string{"a","b"}{
  _,err=db.DB.ExecContext(ctx,`INSERT INTO users(id,display_name,primary_email,status,created_at,updated_at) VALUES(?,?,?,'active',?,?)`,id,id,id+"@test.invalid",now.Format(time.RFC3339Nano),now.Format(time.RFC3339Nano))
  if err!=nil{t.Fatal(err)}
 }
 s:=Service{DB:db.DB,Now:func()time.Time{return now}}
 if err=s.Ready(ctx);err!=nil{t.Fatal(err)}
 f,err:=s.Create(ctx,"b","Review recipe step",now.Add(time.Minute));if err!=nil{t.Fatal(err)}
 if err=s.Done(ctx,"a",f.ID);err!=ErrNotFound{t.Fatalf("cross-user completion: %v",err)}
 n,err:=s.Tick(ctx);if err!=nil||n!=0{t.Fatalf("early tick %d %v",n,err)}
 now=now.Add(2*time.Minute)
 n,err=s.Tick(ctx);if err!=nil||n!=1{t.Fatalf("due tick %d %v",n,err)}
 n,err=s.Tick(ctx);if err!=nil||n!=0{t.Fatalf("duplicate tick %d %v",n,err)}
 if err=db.Close();err!=nil{t.Fatal(err)}
 db,err=database.Open(ctx,database.Config{Path:path});if err!=nil{t.Fatal(err)};defer db.Close()
 s.DB=db.DB
 if err=s.Ready(ctx);err!=nil{t.Fatal(err)}
 inbox,err:=s.Inbox(ctx,"b");if err!=nil||len(inbox)!=1{t.Fatalf("restart inbox: %+v %v",inbox,err)}
 a,err:=s.Inbox(ctx,"a");if err!=nil||len(a)!=0{t.Fatalf("cross-user inbox: %+v %v",a,err)}
 n,err=s.Tick(ctx);if err!=nil||n!=0{t.Fatalf("restart duplicate %d %v",n,err)}
 if err=s.Done(ctx,"b",f.ID);err!=nil{t.Fatal(err)}
 inbox,err=s.Inbox(ctx,"b");if err!=nil||len(inbox)!=0{t.Fatalf("done still nudged %+v %v",inbox,err)}
 n,err=s.Tick(ctx);if err!=nil||n!=0{t.Fatalf("done requeued %d %v",n,err)}
}
