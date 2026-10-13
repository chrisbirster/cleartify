package database

import (
 "context"
 "testing"
)
func TestMigrateCreatesOAuthPersistenceTables(t *testing.T){
 ctx:=context.Background()
 db,err:=Open(ctx,Config{Path:":memory:"})
 if err!=nil{t.Fatal(err)}
 defer db.Close()
 for i:=0;i<2;i++ {
  if err:=db.Migrate(ctx);err!=nil{t.Fatalf("migration pass %d: %v",i,err)}
 }
 for _,name:=range []string{"atproto_oauth_requests","atproto_oauth_sessions","atproto_browser_sessions"}{
  var got string
  if err:=db.DB.QueryRowContext(ctx,"SELECT name FROM sqlite_master WHERE type='table' AND name = ?",name).Scan(&got);err!=nil{t.Fatal(err)}
  if got!=name{t.Fatalf("table %s not found",name)}
 }
}
