package api

import(
 "bytes"
 "encoding/base64"
 "net/http"
 "net/http/httptest"
 "testing"

 "github.com/chrisbirster/cleartify/internal/atproto"
 "github.com/chrisbirster/cleartify/internal/database"
 "github.com/chrisbirster/cleartify/internal/social"
 "context"
)
func TestSocialRegistrationFailClosedWithoutVerifiedOAuth(t *testing.T){
 ctx:=context.Background()
 db,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)}
 defer db.Close()
 if err=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 app:="https://cleartify.com"
 disabled:=New(Options{AppOrigin:app,SocialAccounts:&social.Accounts{Store:db}})
 req:=httptest.NewRequest(http.MethodPost,"/api/social/accounts",bytes.NewReader([]byte(`{}`)))
 req.Header.Set("Origin",app)
 rec:=httptest.NewRecorder();disabled.ServeHTTP(rec,req)
 if rec.Code!=http.StatusNotFound{t.Fatalf("default-off endpoint exposed: %d",rec.Code)}
 key:=base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{7},32))
 oauth,err:=atproto.NewOAuthHandler(db,"https://api.cleartify.com",app,key);if err!=nil{t.Fatal(err)}
 enabled:=New(Options{AppOrigin:app,SocialAccounts:&social.Accounts{Store:db},OAuth:oauth,SocialAccountRegistrationEnabled:true})
 rec=httptest.NewRecorder();enabled.ServeHTTP(rec,req)
 if rec.Code!=http.StatusUnauthorized{t.Fatalf("unauthenticated creation accepted: %d %s",rec.Code,rec.Body.String())}
 var count int
 if err=db.DB.QueryRowContext(ctx,"SELECT COUNT(*) FROM social_profiles").Scan(&count);err!=nil||count!=0{t.Fatalf("social profile persisted without OAuth: %d %v",count,err)}
 req=httptest.NewRequest(http.MethodPost,"/api/social/accounts",nil);req.Header.Set("Origin","https://evil.example")
 rec=httptest.NewRecorder();enabled.ServeHTTP(rec,req)
 if rec.Code!=http.StatusForbidden{t.Fatalf("cross-site request accepted: %d",rec.Code)}
}
