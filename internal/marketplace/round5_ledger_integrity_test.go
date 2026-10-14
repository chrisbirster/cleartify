package marketplace

import (
 "context"
 "errors"
 "testing"

 "github.com/chrisbirster/cleartify/internal/database"
)

func TestRound5ExistingFakeLedgerIntegrityCrossBoundary(t *testing.T){
 ctx:=context.Background()
 d,err:=database.Open(ctx,database.Config{Path:t.TempDir()+"/ledger.db"})
 if err!=nil{t.Fatal(err)};defer d.Close()
 if err=d.Migrate(ctx);err!=nil{t.Fatal(err)}
 for _,q:=range []string{
  `INSERT INTO published_protocols(id,creator_did,slug,version,title,summary,license,spec_json,status,created_at) VALUES('p','did:plc:s','s',1,'S','','MIT','{}','published','now')`,
  `INSERT INTO marketplace_listings(id,protocol_id,creator_did,title,pricing_kind,price_minor,currency,status,created_at) VALUES('l','p','did:plc:s','S','one_time',101,'USD','listed','now')`,
 }{if _,err=d.DB.ExecContext(ctx,q);err!=nil{t.Fatal(err)}}
 svc:=Service{DB:d.DB,Provider:FakeProvider{},Enabled:true}
 id,err:=svc.Purchase(ctx,"did:plc:buyer","l");if err!=nil{t.Fatal(err)}
 a:=FakeAccounting{DB:d.DB,Enabled:true,FeeBPS:1000}
 if n,err:=a.RunReconciliationJob(ctx);err!=nil||n!=1{t.Fatal(n,err)}
 // Later configured rates must not rewrite historically committed fees.
 a.FeeBPS=2500
 if n,err:=a.RunReconciliationJob(ctx);err!=nil||n!=0{t.Fatal("fee change",n,err)}
 // A conflicting existing sale ledger row cannot silently pass idempotent insert.
 _,err=d.DB.ExecContext(ctx,"UPDATE fake_seller_ledger SET net_minor=net_minor+1 WHERE entitlement_id=? AND entry_kind='sale'",id)
 if err!=nil{t.Fatal(err)}
 if _,err=a.RunReconciliationJob(ctx);!errors.Is(err,ErrDenied){t.Fatal("corrupt active sale accepted",err)}
 _,err=d.DB.ExecContext(ctx,"UPDATE fake_seller_ledger SET net_minor=gross_minor-fee_minor WHERE entitlement_id=? AND entry_kind='sale'",id)
 if err!=nil{t.Fatal(err)}
 if err=svc.Refund(ctx,"did:plc:buyer",id);err!=nil{t.Fatal(err)}
 if n,err:=a.RunReconciliationJob(ctx);err!=nil||n!=1{t.Fatal("refund reversal",n,err)}
 _,err=d.DB.ExecContext(ctx,"UPDATE fake_seller_ledger SET fee_minor=fee_minor+1 WHERE entitlement_id=? AND entry_kind='reversal'",id)
 if err!=nil{t.Fatal(err)}
 if _,err=a.RunReconciliationJob(ctx);!errors.Is(err,ErrDenied){t.Fatal("corrupt reversal accepted",err)}
 var n int
 if err=d.DB.QueryRowContext(ctx,"SELECT COUNT(*) FROM fake_seller_ledger WHERE entitlement_id=?",id).Scan(&n);err!=nil||n!=2{t.Fatal("ledger row loss or duplication",n,err)}
}
