package e2e

import(
 "context"
 "errors"
 "fmt"
 "net/http"
 "net/http/httptest"
 "time"

 "github.com/chrisbirster/cleartify/internal/alerts"
 "github.com/chrisbirster/cleartify/internal/api"
 "github.com/chrisbirster/cleartify/internal/database"
 "github.com/chrisbirster/cleartify/internal/marketplace"
)

// Harness is strictly local: in-memory DB and httptest transport, no listener,
// OAuth provider, hosted database, outbound calls, secrets or real checkout.
type Harness struct {
 Store *database.Store
 HTTP http.Handler
 Fake marketplace.Service
 Alerts alerts.Sink
 Clock func()time.Time
}
func New(ctx context.Context,sink alerts.Sink)(*Harness,error){
 if sink==nil{sink=alerts.Noop{}}
 store,err:=database.Open(ctx,database.Config{Path:":memory:"})
 if err!=nil{return nil,err}
 if err=store.Migrate(ctx);err!=nil{store.Close();return nil,err}
 return &Harness{Store:store,HTTP:api.New(api.Options{}),Fake:marketplace.Service{DB:store.DB,Provider:marketplace.FakeProvider{}},Alerts:sink,Clock:func()time.Time{return time.Date(2026,10,13,0,0,0,0,time.UTC)}},nil
}
func(h *Harness)Close()error{return h.Store.Close()}
func(h *Harness)Stage(ctx context.Context,stage string,fn func(context.Context)error)error{
 if err:=ctx.Err();err!=nil{return err}
 if stage==""||fn==nil{return errors.New("invalid E2E stage")}
 if err:=fn(ctx);err!=nil{
  if emitErr:=h.Alerts.Emit(ctx,alerts.Event{Key:"stage_failed",Stage:stage,At:h.Clock()});emitErr!=nil{return fmt.Errorf("stage %s: %w (alert: %v)",stage,err,emitErr)}
  return fmt.Errorf("stage %s: %w",stage,err)
 }
 return nil
}
func(h *Harness)Request(method,path string)*httptest.ResponseRecorder{
 w:=httptest.NewRecorder()
 h.HTTP.ServeHTTP(w,httptest.NewRequest(method,path,nil))
 return w
}
