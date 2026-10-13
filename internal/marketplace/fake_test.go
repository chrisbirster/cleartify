package marketplace
import(
 "context"
 "errors"
 "testing"
 "github.com/chrisbirster/cleartify/internal/database"
)
func TestFakeProviderAndDisabledDefault(t *testing.T){
 p:=FakeProvider{}
 r,err:=p.Charge(context.Background(),"abc",200,"USD")
 if err!=nil||r.Provider!="fake"||r.ID!="fake-abc"{t.Fatalf("%+v %v",r,err)}
 if _,err:=p.Charge(context.Background(),"abc",-1,"USD");!errors.Is(err,ErrPayment){t.Fatal(err)}
 if _,err:=p.Charge(context.Background(),"abc",20,"EUR");!errors.Is(err,ErrPayment){t.Fatal(err)}
 if _,err:=p.Refund(context.Background(),"real-abc");!errors.Is(err,ErrPayment){t.Fatal(err)}
 ctx:=context.Background()
 db,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)};defer db.Close()
 if err:=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 s:=Service{DB:db.DB,Provider:p}
 if _,err:=s.Purchase(ctx,"did:plc:buyer","missing");!errors.Is(err,ErrDenied){t.Fatal(err)}
 if err:=s.AuthorizeSeller(ctx,"did:plc:seller","missing");!errors.Is(err,ErrDenied){t.Fatal(err)}
 if yes,err:=s.Entitled(ctx,"did:plc:buyer","missing");err!=nil||yes{t.Fatal(yes,err)}
 s.Enabled=true
 if _,err:=s.Purchase(ctx,"did:plc:buyer","missing");!errors.Is(err,ErrDenied){t.Fatal(err)}
 if err:=s.Refund(ctx,"did:plc:other","missing");!errors.Is(err,ErrDenied){t.Fatal(err)}
 if err:=s.AuthorizeSeller(ctx,"did:plc:other","missing");!errors.Is(err,ErrDenied){t.Fatal(err)}
}
func TestRejectNonFakeProvider(t *testing.T){
 s:=Service{Enabled:true,Provider:nil}
 if _,err:=s.provider();!errors.Is(err,ErrPayment){t.Fatal(err)}
}
