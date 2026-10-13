package marketplace
import(
 "context"
 "testing"
 "github.com/chrisbirster/cleartify/internal/database"
)
func TestFakeReconciliationAndAccessGate(t *testing.T){
 ctx:=context.Background();store,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)};defer store.Close()
 if err=store.Migrate(ctx);err!=nil{t.Fatal(err)}
 s:=Service{DB:store.DB,Provider:FakeProvider{},Enabled:true}
 ok,err:=s.PublicationAccess(ctx,"did:plc:buyer","free");if err!=nil||!ok{t.Fatalf("free %v %v",ok,err)}
 _,err=store.DB.ExecContext(ctx,`INSERT INTO social_profiles(did,handle,display_name,kind,public_visible,claimed,updated_at) VALUES('did:plc:seller','seller.test','Seller','person',1,1,'now')`);if err!=nil{t.Fatal(err)}
 _,err=store.DB.ExecContext(ctx,`INSERT INTO published_protocols(id,creator_did,slug,version,title,summary,license,spec_json,status,created_at) VALUES('p','did:plc:seller','test',1,'Test','','MIT','{}','published','now')`);if err!=nil{t.Fatal(err)}
 _,err=store.DB.ExecContext(ctx,`INSERT INTO marketplace_listings(id,protocol_id,creator_did,title,pricing_kind,price_minor,currency,status,created_at) VALUES('l','p','did:plc:seller','Test listing','one_time',100,'USD','listed','now')`);if err!=nil{t.Fatal(err)}
 ok,err=s.PublicationAccess(ctx,"did:plc:buyer","p");if err!=nil||ok{t.Fatalf("unpaid %v %v",ok,err)}
 id,err:=s.Purchase(ctx,"did:plc:buyer","l");if err!=nil{t.Fatal(err)}
 again,err:=s.Purchase(ctx,"did:plc:buyer","l");if err!=nil||again!=id{t.Fatalf("idempotency %q %q %v",id,again,err)}
 ok,err=s.PublicationAccess(ctx,"did:plc:buyer","p");if err!=nil||!ok{t.Fatalf("paid %v %v",ok,err)}
 ok,err=s.PublicationAccess(ctx,"did:plc:other","p");if err!=nil||ok{t.Fatalf("cross-user %v %v",ok,err)}
 n,err:=s.ReconcileFake(ctx);if err!=nil||n!=0{t.Fatalf("reconcile purchase %d %v",n,err)}
 if err=s.Refund(ctx,"did:plc:buyer",id);err!=nil{t.Fatal(err)}
 ok,err=s.PublicationAccess(ctx,"did:plc:buyer","p");if err!=nil||ok{t.Fatalf("refunded %v %v",ok,err)}
 n,err=s.ReconcileFake(ctx);if err!=nil||n!=0{t.Fatalf("reconcile refund %d %v",n,err)}
 if _,err=s.Purchase(ctx,"did:plc:buyer","l");err!=ErrDenied{t.Fatalf("refunded repurchase unexpectedly allowed %v",err)}
 _,err=store.DB.ExecContext(ctx,`DELETE FROM marketplace_payment_audit WHERE entitlement_id=? AND event='fake_refund'`,id);if err!=nil{t.Fatal(err)}
 n,err=s.ReconcileFake(ctx);if err!=nil||n!=1{t.Fatalf("missing audit undetected %d %v",n,err)}
}
