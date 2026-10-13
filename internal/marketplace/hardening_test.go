package marketplace
import("context";"sync";"testing";"github.com/chrisbirster/cleartify/internal/database")
func TestConcurrentFakePurchaseAndLedgerIdempotency(t *testing.T){
 ctx:=context.Background();d,e:=database.Open(ctx,database.Config{Path:":memory:"});if e!=nil{t.Fatal(e)};defer d.Close();d.DB.SetMaxOpenConns(1)
 if e=d.Migrate(ctx);e!=nil{t.Fatal(e)}
 for _,q:=range []string{
 `INSERT INTO social_profiles(did,handle,display_name,kind,public_visible,claimed,updated_at) VALUES('did:plc:s','s.test','S','person',1,1,'now')`,
 `INSERT INTO published_protocols(id,creator_did,slug,version,title,summary,license,spec_json,status,created_at) VALUES('p','did:plc:s','s',1,'S','','MIT','{}','published','now')`,
 `INSERT INTO marketplace_listings(id,protocol_id,creator_did,title,pricing_kind,price_minor,currency,status,created_at) VALUES('l','p','did:plc:s','S','one_time',100,'USD','listed','now')`,
 }{if _,e=d.DB.ExecContext(ctx,q);e!=nil{t.Fatal(e)}}
 svc:=Service{DB:d.DB,Provider:FakeProvider{},Enabled:true}
 const n=12
 ids:=make([]string,n);errs:=make([]error,n)
 var wg sync.WaitGroup
 for i:=0;i<n;i++{wg.Add(1);go func(i int){defer wg.Done();ids[i],errs[i]=svc.Purchase(ctx,"did:plc:b","l")}(i)};wg.Wait()
 for i:=range ids{if errs[i]!=nil||ids[i]==""||ids[i]!=ids[0]{t.Fatalf("purchase %d: %q %v",i,ids[i],errs[i])}}
 a:=FakeAccounting{DB:d.DB,Enabled:true,FeeBPS:500}
 counts:=make([]int,n)
 for i:=0;i<n;i++{wg.Add(1);go func(i int){defer wg.Done();counts[i],errs[i]=a.RunReconciliationJob(ctx)}(i)};wg.Wait()
 sum:=0;for i:=range counts{if errs[i]!=nil{t.Fatalf("reconcile %d: %v",i,errs[i])};sum+=counts[i]}
 if sum!=1{t.Fatalf("double booking %d",sum)}
 bal,e:=a.SellerBalance(ctx,"did:plc:s");if e!=nil||bal!=95{t.Fatalf("balance %d %v",bal,e)}
}
