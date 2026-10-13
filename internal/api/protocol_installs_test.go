package api

import(
 "bytes"
 "context"
 "encoding/base64"
 "net/http/httptest"
 "testing"
 "github.com/chrisbirster/cleartify/internal/atproto"
 "github.com/chrisbirster/cleartify/internal/database"
 "github.com/chrisbirster/cleartify/internal/protocols"
)
func TestInstallEndpointsAreFlaggedAndRequireOAuth(t *testing.T){
 ctx:=context.Background()
 db,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)}
 defer db.Close()
 if err=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 origin:="https://app.example"
 flow:=&protocols.InstallService{Store:db}
 off:=New(Options{AppOrigin:origin,InstallFlow:flow})
 req:=httptest.NewRequest("POST","/api/me/protocol-install-requests",nil);req.Header.Set("Origin",origin)
 w:=httptest.NewRecorder();off.ServeHTTP(w,req);if w.Code!=404{t.Fatalf("off endpoint exposed %d",w.Code)}
 secret:=base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{5},32))
 auth,err:=atproto.NewOAuthHandler(db,"https://api.app.example",origin,secret);if err!=nil{t.Fatal(err)}
 on:=New(Options{AppOrigin:origin,OAuth:auth,InstallFlow:flow,ProtocolInstallFlowEnabled:true,AdminToken:"admin-test"})
 w=httptest.NewRecorder();on.ServeHTTP(w,req);if w.Code!=401{t.Fatalf("missing OAuth got %d",w.Code)}
 req=httptest.NewRequest("POST","/api/me/protocol-install-requests",nil);req.Header.Set("Origin","https://evil.example")
 w=httptest.NewRecorder();on.ServeHTTP(w,req);if w.Code!=403{t.Fatalf("bad origin got %d",w.Code)}
 req=httptest.NewRequest("POST","/api/admin/protocol-install-requests/invalid/decision",nil)
 w=httptest.NewRecorder();on.ServeHTTP(w,req);if w.Code!=401{t.Fatalf("missing admin got %d",w.Code)}
}
