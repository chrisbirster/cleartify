package api

import(
 "bytes"
 "context"
 "encoding/base64"
 "net/http"
 "net/http/httptest"
 "testing"

 "github.com/chrisbirster/cleartify/internal/atproto"
 "github.com/chrisbirster/cleartify/internal/database"
 "github.com/chrisbirster/cleartify/internal/social"
)
func TestProfileAndFollowEndpointsFailClosed(t *testing.T){
 ctx:=context.Background()
 db,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)}
 defer db.Close()
 if err=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 origin:="https://cleartify.example"
 req:=func(method,path,from string)*http.Request{r:=httptest.NewRequest(method,path,bytes.NewReader([]byte(`{"handle":"chris"}`)));r.Header.Set("Origin",from);return r}
 off:=New(Options{AppOrigin:origin,SocialActions:&social.Actions{Store:db}})
 for _,path:=range []string{"/api/me/profile/claim","/api/me/following/chris"}{
  w:=httptest.NewRecorder();off.ServeHTTP(w,req("POST",path,origin));if w.Code!=404{t.Fatalf("flag-off route %s exposed: %d",path,w.Code)}
 }
 key:=base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{9},32))
 oauth,err:=atproto.NewOAuthHandler(db,"https://api.cleartify.example",origin,key);if err!=nil{t.Fatal(err)}
 on:=New(Options{AppOrigin:origin,OAuth:oauth,SocialActions:&social.Actions{Store:db},SocialProfileActionsEnabled:true})
 for _,m:=range []string{http.MethodPost,http.MethodPut,http.MethodDelete}{
  path:="/api/me/following/chris";if m==http.MethodPost{path="/api/me/profile/claim"}
  w:=httptest.NewRecorder();on.ServeHTTP(w,req(m,path,"https://evil.example"));if w.Code!=403{t.Fatalf("bad origin %s got %d",m,w.Code)}
  w=httptest.NewRecorder();on.ServeHTTP(w,req(m,path,origin));if w.Code!=401{t.Fatalf("missing auth %s got %d",m,w.Code)}
 }
 var n int;if err=db.DB.QueryRowContext(ctx,`SELECT COUNT(*) FROM social_follows`).Scan(&n);err!=nil||n!=0{t.Fatalf("unauthenticated mutation: %d %v",n,err)}
}
