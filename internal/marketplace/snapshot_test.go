package marketplace

import (
 "context"
 "errors"
 "testing"
 "github.com/chrisbirster/cleartify/internal/database"
)
func TestImmutableFakeSaleSnapshot(t *testing.T){
 ctx:=context.Background()
 store,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)};defer store.Close()
 if err=store.Migrate(ctx);err!=nil{t.Fatal(err)}
 for _,q:=range []string{
  `INSERT INTO social_profiles(did,handle,display_name,kind,public_visible,claimed,updated_at) VALUES('did:plc:seller','seller.test','Seller','person',1,1,'now')`,
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
 a.FeeBPS=2000
 if err=a.OpenDispute(ctx,"did:plc:intruder",id,"other");err==nil{t.Fatal("unauthorized dispute accepted")}
 if err=a.OpenDispute(ctx,"did:plc:buyer",id,"other");err!=nil{t.Fatal(err)}
 if err=a.OpenDispute(ctx,"did:plc:buyer",id,"other");err==nil{t.Fatal("duplicate dispute accepted")}
 if err=a.ResolveDispute(ctx,"local-admin",id,true);err!=nil{t.Fatal(err)}
 if err=a.ResolveDispute(ctx,"local-admin",id,true);err==nil{t.Fatal("duplicate refund accepted")}
 n,err=a.Reconcile(ctx);if err!=nil||n!=1{t.Fatalf("reversal %d %v",n,err)}
 var saleGross,saleFee,saleNet,reverseGross,reverseFee,reverseNet int64
 err=store.DB.QueryRowContext(ctx,"SELECT gross_minor,fee_minor,net_minor FROM fake_seller_ledger WHERE entitlement_id=? AND entry_kind='sale'",id).Scan(&saleGross,&saleFee,&saleNet)
 if err!=nil{t.Fatal(err)}
 err=store.DB.QueryRowContext(ctx,"SELECT gross_minor,fee_minor,net_minor FROM fake_seller_ledger WHERE entitlement_id=? AND entry_kind='reversal'",id).Scan(&reverseGross,&reverseFee,&reverseNet)
 if err!=nil||saleGross!=-reverseGross||saleFee!=-reverseFee||saleNet!=-reverseNet{t.Fatalf("reversal does not negate sale: sale=%d/%d/%d reversal=%d/%d/%d err=%v",saleGross,saleFee,saleNet,reverseGross,reverseFee,reverseNet,err)}
 var total int64;if err=store.DB.QueryRowContext(ctx,"SELECT SUM(net_minor) FROM fake_seller_ledger WHERE entitlement_id=?",id).Scan(&total);err!=nil||total!=0{t.Fatalf("reversal total %d %v",total,err)}
}

func TestReconciliationFailsClosedForLegacyMissingSnapshot(t *testing.T) {
 ctx := context.Background()
 store, err := database.Open(ctx, database.Config{Path: ":memory:"})
 if err != nil { t.Fatal(err) }
 defer store.Close()
 if err := store.Migrate(ctx); err != nil { t.Fatal(err) }
 for _, q := range []string{
  `INSERT INTO social_profiles(did,handle,display_name,kind,public_visible,claimed,updated_at) VALUES('did:plc:seller','seller.test','Seller','person',1,1,'now')`,
  `INSERT INTO published_protocols(id,creator_did,slug,version,title,summary,license,spec_json,status,created_at) VALUES('p','did:plc:seller','slug',1,'Title','','MIT','{}','published','now')`,
  `INSERT INTO marketplace_listings(id,protocol_id,creator_did,title,pricing_kind,price_minor,currency,status,created_at) VALUES('l','p','did:plc:seller','Title','one_time',101,'USD','listed','now')`,
 } { if _, err := store.DB.ExecContext(ctx,q); err != nil { t.Fatal(err) } }
 svc := Service{DB:store.DB,Provider:FakeProvider{},Enabled:true}
 id,err := svc.Purchase(ctx,"did:plc:buyer","l"); if err != nil { t.Fatal(err) }
 // Simulate a pre-snapshot entitlement without relying on mutable listing prices.
 _,err=store.DB.ExecContext(ctx,`INSERT INTO marketplace_entitlements(id,listing_id,buyer_did,provider,provider_receipt,status,created_at) VALUES('legacy','l','did:plc:legacy','fake','fake-legacy','active','now')`)
 if err != nil {t.Fatal(err)}
 _,err=store.DB.ExecContext(ctx,`INSERT INTO marketplace_payment_audit(id,entitlement_id,event,actor_did,recorded_at) VALUES('legacy-audit','legacy','fake_purchase','did:plc:legacy','now')`)
 if err != nil {t.Fatal(err)}
 a:=FakeAccounting{DB:store.DB,Enabled:true,FeeBPS:1000}
 n,err:=a.Reconcile(ctx)
 if n!=0||!errors.Is(err,ErrMissingSaleSnapshot){t.Fatalf("expected missing snapshot, got %d %v",n,err)}
 var ledger int
 if err=store.DB.QueryRowContext(ctx,`SELECT count(*) FROM fake_seller_ledger`).Scan(&ledger);err!=nil||ledger!=0{t.Fatalf("partial ledger %d %v",ledger,err)}
 // Existing valid entitlement remains intact and is still unbooked.
 var snapshot int
 if err=store.DB.QueryRowContext(ctx,`SELECT count(*) FROM fake_sale_snapshots WHERE entitlement_id=?`,id).Scan(&snapshot);err!=nil||snapshot!=1{t.Fatalf("valid snapshot %d %v",snapshot,err)}
}
