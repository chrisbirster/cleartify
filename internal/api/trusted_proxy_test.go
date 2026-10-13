package api
import(
 "context"
 "net/http/httptest"
 "testing"
 "time"
 "github.com/chrisbirster/cleartify/internal/database"
)
func TestProxyChainTrustAndSpoofing(t *testing.T){
 nets,err:=ParseTrustedProxyCIDRs("127.0.0.0/8,10.0.0.0/8");if err!=nil{t.Fatal(err)}
 s:=&Server{dbAbuseProxyEnabled:true,dbAbuseTrustedProxyCIDRs:nets}
 check:=func(remote,xff,want string){t.Helper();r:=httptest.NewRequest("POST","/api/waitlist",nil);r.RemoteAddr=remote;r.Header.Set("X-Forwarded-For",xff);got,err:=s.signupPeer(r);if err!=nil||got!=want{t.Fatalf("%s xff=%s got=%s err=%v want=%s",remote,xff,got,err,want)}}
 check("192.0.2.2:1234","198.51.100.99","192.0.2.2")
 check("127.0.0.1:1234","198.51.100.77","198.51.100.77")
 check("127.0.0.1:1234","203.0.113.99, 10.0.0.2","203.0.113.99")
 check("127.0.0.1:1234","6.6.6.6,198.51.100.77","198.51.100.77")
 r:=httptest.NewRequest("POST","/api/waitlist",nil);r.RemoteAddr="127.0.0.1:1234";r.Header.Set("X-Forwarded-For","not-an-ip")
 if _,err=s.signupPeer(r);err==nil{t.Fatal("malformed forwarding not rejected")}
 s.dbAbuseProxyEnabled=false
 if got,err:=s.signupPeer(r);err!=nil||got!="127.0.0.1"{t.Fatalf("default direct peer %s %v",got,err)}
 if _,err:=ParseTrustedProxyCIDRs("0.0.0.0/0,wat");err==nil{t.Fatal("invalid config accepted")}
}
func TestCounterCleanupRetainsCurrentBucket(t *testing.T){
 ctx:=context.Background();store,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)};defer store.Close()
 if err=store.Migrate(ctx);err!=nil{t.Fatal(err)}
 now:=time.Now().UTC().Unix()/900
 for _,v:=range []int64{now-300,now-1,now}{
  if _,err=store.DB.ExecContext(ctx,"INSERT INTO signup_abuse_windows(identity_hash,bucket,attempts) VALUES(?,?,1)","x",v);err!=nil{t.Fatal(err)}
 }
 s:=&Server{dbAbuseEnabled:true,dbAbuseCleanupEnabled:true,dbAbuseDB:store.DB,dbAbuseSecret:"01234567890123456789012345678901"}
 r:=httptest.NewRequest("POST","/api/waitlist",nil);r.RemoteAddr="192.0.2.5:1234"
 if !s.allowDBSignup(httptest.NewRecorder(),r){t.Fatal("signup denied")}
 var old,active int
 if err=store.DB.QueryRowContext(ctx,"SELECT count(*) FROM signup_abuse_windows WHERE bucket < ?",now-192).Scan(&old);err!=nil||old!=0{t.Fatalf("old buckets %d %v",old,err)}
 if err=store.DB.QueryRowContext(ctx,"SELECT count(*) FROM signup_abuse_windows WHERE bucket=?",now).Scan(&active);err!=nil||active!=2{t.Fatalf("active buckets %d %v",active,err)}
}
