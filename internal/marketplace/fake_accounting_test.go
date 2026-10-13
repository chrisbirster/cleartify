package marketplace
import(
 "context"
 "testing"
 "github.com/chrisbirster/cleartify/internal/database"
)
func TestFakeAccountingReconcileDisputeAndRefund(t *testing.T){
 ctx:=context.Background()
 db,e:=database.Open(ctx,database.Config{Path:":memory:"});if e!=nil{t.Fatal(e)};defer db.Close()
 if e=db.Migrate(ctx);e!=nil{t.Fatal(e)}
 for _,q:=range []string{
 `INSERT INTO social_profiles(did,handle,display_name,kind,public_visible,claimed,updated_at) VALUES('did:plc:seller','seller.test','Seller','person',1,1,'now')`,
 `INSERT INTO published_protocols(id,creator_did,slug,version,title,summary,license,spec_json,status,created_at) VALUES('p','did:plc:seller','test',1,'Test','','MIT','{}','published','now')`,
 `INSERT INTO marketplace_listings(id,protocol_id,creator_did,title,pricing_kind,price_minor,currency,status,created_at) VALUES('l','p','did:plc:seller','Listing','one_time',101,'USD','listed','now')`,
 }{if _,e=db.DB.ExecContext(ctx,q);e!=nil{t.Fatal(e)}}
 base:=Service{DB:db.DB,Provider:FakeProvider{},Enabled:true}
 id,e:=base.Purchase(ctx,"did:plc:buyer","l");if e!=nil{t.Fatal(e)}
 a:=FakeAccounting{DB:db.DB,Enabled:true,FeeBPS:1000}
 n,e:=a.Reconcile(ctx);if e!=nil||n!=1{t.Fatalf("first ledger %d %v",n,e)}
 bal,e:=a.SellerBalance(ctx,"did:plc:seller");if e!=nil||bal!=91{t.Fatalf("fee rounding %d %v",bal,e)}
 n,e=a.Reconcile(ctx);if e!=nil||n!=0{t.Fatalf("idempotent ledger %d %v",n,e)}
 if e=a.OpenDispute(ctx,"did:plc:intruder",id,"other");e==nil{t.Fatal("cross-buyer dispute accepted")}
 if e=a.OpenDispute(ctx,"did:plc:buyer",id,"invalid");e==nil{t.Fatal("invalid reason accepted")}
 if e=a.OpenDispute(ctx,"did:plc:buyer",id,"not_received");e!=nil{t.Fatal(e)}
 if e=a.OpenDispute(ctx,"did:plc:buyer",id,"not_received");e==nil{t.Fatal("duplicate dispute accepted")}
 if e=a.ResolveDispute(ctx,"local-operator",id,true);e!=nil{t.Fatal(e)}
 if e=a.ResolveDispute(ctx,"local-operator",id,true);e==nil{t.Fatal("replayed decision accepted")}
 active,e:=base.PublicationAccess(ctx,"did:plc:buyer","p");if e!=nil||active{t.Fatal("refund did not revoke access",e)}
 n,e=a.Reconcile(ctx);if e!=nil||n!=1{t.Fatalf("reversal %d %v",n,e)}
 n,e=a.Reconcile(ctx);if e!=nil||n!=0{t.Fatalf("replay reversal %d %v",n,e)}
 bal,e=a.SellerBalance(ctx,"did:plc:seller");if e!=nil||bal!=0{t.Fatalf("balance %d %v",bal,e)}
 mismatches,e:=base.ReconcileFake(ctx);if e!=nil||mismatches!=0{t.Fatalf("fake audit mismatch %d %v",mismatches,e)}
 var events int
 if e=db.DB.QueryRowContext(ctx,"SELECT COUNT(*) FROM fake_payment_dispute_audit WHERE entitlement_id=?",id).Scan(&events);e!=nil||events!=2{t.Fatalf("dispute audit %d %v",events,e)}
 a.Enabled=false
 if _,e=a.Reconcile(ctx);e==nil{t.Fatal("flag off allowed reconciliation")}
 a.Enabled=true;a.FeeBPS=10001
 if _,e=a.Reconcile(ctx);e==nil{t.Fatal("bad fee rate accepted")}
}
