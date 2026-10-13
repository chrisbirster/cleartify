package protocols

import (
 "context"
 "errors"
 "testing"
 "github.com/chrisbirster/cleartify/internal/database"
)
func TestDiscoveryOnlyPublishedVersions(t *testing.T){
 ctx:=context.Background()
 store,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)}
 defer store.Close()
 if err:=store.Migrate(ctx);err!=nil{t.Fatal(err)}
 for _,status:=range []string{"draft","published","withdrawn"}{
  _,err=store.DB.ExecContext(ctx,`INSERT INTO published_protocols(id,creator_did,slug,version,title,summary,license,spec_json,status,created_at) VALUES(?,?,?,?,?,?,?,?,?,?)`,status,"did:plc:ewvi7nxzyoun6zhxrhs64oiz",status,1,status,"description","MIT","{}",status,"2026-10-12")
  if err!=nil{t.Fatal(err)}
 }
 reader:=Reader{DB:store.DB}
 results,err:=reader.Discover(ctx,20);if err!=nil{t.Fatal(err)}
 if len(results)!=1||results[0].Slug!="published"{t.Fatalf("incorrect public discovery: %+v",results)}
 if _,err=reader.Version(ctx,"did:plc:ewvi7nxzyoun6zhxrhs64oiz","draft",1);!errors.Is(err,ErrNotFound){t.Fatalf("draft leaked: %v",err)}
 result,err:=reader.Version(ctx,"did:plc:ewvi7nxzyoun6zhxrhs64oiz","published",1);if err!=nil||result.Version!=1{t.Fatalf("published version: %+v %v",result,err)}
}
