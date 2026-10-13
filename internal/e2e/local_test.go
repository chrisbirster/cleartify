package e2e
import(
 "context"
 "errors"
 "testing"
 "github.com/chrisbirster/cleartify/internal/alerts"
 "github.com/chrisbirster/cleartify/internal/marketplace"
)
func TestStagedLocalSmoke(t *testing.T){
 ctx:=context.Background();sink:=alerts.NewLocal(4)
 h,err:=New(ctx,sink);if err!=nil{t.Fatal(err)};defer h.Close()
 if err=h.Stage(ctx,"migrate",func(ctx context.Context)error{
  var n int;return h.Store.DB.QueryRowContext(ctx,"SELECT count(*) FROM waitlist_entries").Scan(&n)
 });err!=nil{t.Fatal(err)}
 if err=h.Stage(ctx,"api",func(context.Context)error{if w:=h.Request("GET","/healthz");w.Code!=200{return errors.New("bad health")};return nil});err!=nil{t.Fatal(err)}
 if err=h.Stage(ctx,"fail-closed",func(context.Context)error{
  if w:=h.Request("POST","/api/me/fake-checkout");w.Code!=404{return errors.New("checkout unexpectedly enabled")}
  if w:=h.Request("POST","/api/me/agents/a/ownership/start");w.Code!=404{return errors.New("ownership unexpectedly enabled")}
  return nil
 });err!=nil{t.Fatal(err)}
 if err=h.Stage(ctx,"fake-payments",func(ctx context.Context)error{
  receipt,e:=(marketplace.FakeProvider{}).Charge(ctx,"local",100,"USD")
  if e!=nil||receipt.Provider!="fake"{return errors.New("fake provider failed")}
  if _,e=h.Fake.Purchase(ctx,"did:plc:test","missing");!errors.Is(e,marketplace.ErrDenied){return errors.New("disabled checkout accepted")}
  return nil
 });err!=nil{t.Fatal(err)}
 if err=h.Stage(ctx,"injected-failure",func(context.Context)error{return errors.New("synthetic")});err==nil{t.Fatal("stage failure swallowed")}
 events:=sink.Snapshot();if len(events)!=1||events[0].Key!="stage_failed"||events[0].Stage!="injected-failure"{t.Fatalf("alerts %+v",events)}
}
