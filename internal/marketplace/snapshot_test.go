package marketplace

import (
 "context"
 "testing"
 "github.com/chrisbirster/cleartify/internal/database"
)
func TestImmutableFakeSaleSnapshot(t *testing.T){
 ctx:=context.Background()
 store,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)};defer store.Close()
 if err=store.Migrate(ctx);err!=nil{t.Fatal(err)}
 for _,q:=range []string{
  `INSERT INTO published_protocols(id,creator_did,slug,version,title,summary,license,spec_json,status,created_at) VALUES('p','did:plc:seller','slug',1,'Title','','MIT','{}','published','now')`,
  `INSERT INTO marketplace_listings(id,protocol_id,creator_did,title,pricing_kind,price_minor,currency,status,created_at) VALUES('l','p','did:plc:seller','Title','one_time',101,'USD','listed','now')`,
 } {if _,err=store.DB.ExecContext(ctx,q);err!=nil{t.Fatal(err)}}
 svc:=Service{DB:store.DB,Provider:FakeProvider{},Enabled:true}
 id,err:=svc.Purchase(ctx,"did:plc:buyer","l");if err!=nil{t.Fatal(err)}
 if _,err=store.DB.ExecContext(ctx,"UPDATE marketplace_listings SET price_minor=999 WHERE id='l'");err!=nil{t.Fatal(err)}
 a:=FakeAccounting{DB:store.DB,Enabled:true,FeeBPS:1000}
 n,err:=a.Reconcile(ctx);if err!=nil||n!=1{t.Fatalf("reconcile %d %v",n,err)}
 var gross int64;var currency string
 if err=store.DB.QueryRowContext(ctx,"SELECT gross_minor,currency FROM fake_seller_ledger WHERE entitlement_id=? AND entry_kind='sale'",id).Scan(&gross,&currency);err!=nil||gross!=101||currency!="USD"{t.Fatalf("historical price %d %s %v",gross,currency,err)}
 if _,err=store.DB.ExecContext(ctx,"UPDATE fake_sale_snapshots SET price_minor=999 WHERE entitlement_id=?",id);err==nil{t.Fatal("snapshot modified")}
 if _,err=store.DB.ExecContext(ctx,"DELETE FROM fake_sale_snapshots WHERE entitlement_id=?",id);err==nil{t.Fatal("snapshot deleted")}
 if _,err=svc.Purchase(ctx,"did:plc:buyer","l");err!=nil{t.Fatal("idempotent purchase",err)}
 if _,err=store.DB.ExecContext(ctx,"UPDATE marketplace_listings SET price_minor=0 WHERE id='l'");err!=nil{t.Fatal(err)}
 if _,err=svc.Purchase(ctx,"did:plc:newbuyer","l");err==nil{t.Fatal("zero price accepted")}
 if _,err=store.DB.ExecContext(ctx,"UPDATE marketplace_listings SET price_minor=-1 WHERE id='l'");err==nil{t.Fatal("negative price accepted")}
 if _,err=store.DB.ExecContext(ctx,"UPDATE marketplace_listings SET currency='EUR' WHERE id='l'");err==nil{t.Fatal("currency mismatch accepted")}
 if err=a.OpenDispute(ctx,"did:plc:intruder",id,"other");err==nil{t.Fatal("unauthorized dispute accepted")}
 if err=a.OpenDispute(ctx,"did:plc:buyer",id,"other");err!=nil{t.Fatal(err)}
 if err=a.OpenDispute(ctx,"did:plc:buyer",id,"other");err==nil{t.Fatal("duplicate dispute accepted")}
 if err=a.ResolveDispute(ctx,"local-admin",id,true);err!=nil{t.Fatal(err)}
 if err=a.ResolveDispute(ctx,"local-admin",id,true);err==nil{t.Fatal("duplicate refund accepted")}
 n,err=a.Reconcile(ctx);if err!=nil||n!=1{t.Fatalf("reversal %d %v",n,err)}
 var total int64;if err=store.DB.QueryRowContext(ctx,"SELECT SUM(net_minor) FROM fake_seller_ledger WHERE entitlement_id=?",id).Scan(&total);err!=nil||total!=0{t.Fatalf("reversal total %d %v",total,err)}
}
