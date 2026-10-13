package interop
import(
 "bytes"
 "context"
 "crypto/ed25519"
 "crypto/rand"
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "net"
 "net/http"
 "net/http/httptest"
 "testing"
 "time"
 "github.com/chrisbirster/cleartify/internal/agents"
 "github.com/chrisbirster/cleartify/internal/database"
)
func TestLoopbackSignedReplayAndConsent(t *testing.T){
 ctx:=context.Background();store,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)};defer store.Close()
 if err=store.Migrate(ctx);err!=nil{t.Fatal(err)}
 now:=time.Date(2026,10,13,12,0,0,0,time.UTC)
 for _,u:=range []string{"alice","bob"}{_,err=store.DB.ExecContext(ctx,`INSERT INTO social_profiles(did,handle,display_name,kind,public_visible,claimed,updated_at) VALUES(?,?,?,'person',1,1,'now')`,"did:plc:"+u,u+".test",u);if err!=nil{t.Fatal(err)}}
 a,err:=(agents.Service{DB:store.DB}).Register(ctx,"did:plc:alice","A","");if err!=nil{t.Fatal(err)}
 b,err:=(agents.Service{DB:store.DB}).Register(ctx,"did:plc:bob","B","");if err!=nil{t.Fatal(err)}
 grant:=agents.GrantService{DB:store.DB,Now:func()time.Time{return now}}
 ap,priv,_:=ed25519.GenerateKey(rand.Reader);bp,bpriv,_:=ed25519.GenerateKey(rand.Reader)
 if err=grant.Bind(ctx,"did:plc:alice",a.ID,ap,ed25519.Sign(priv,agents.OwnerProofMessage(a.ID,"did:plc:alice")));err!=nil{t.Fatal(err)}
 if err=grant.Bind(ctx,"did:plc:bob",b.ID,bp,ed25519.Sign(bpriv,agents.OwnerProofMessage(b.ID,"did:plc:bob")));err!=nil{t.Fatal(err)}
 for _,id:=range []string{a.ID,b.ID}{if _,err=store.DB.ExecContext(ctx,"UPDATE agent_registrations SET status='verified' WHERE id=?",id);err!=nil{t.Fatal(err)}}
 if _,err=store.DB.ExecContext(ctx,`INSERT INTO agent_interop_keys(key_id,agent_id,algorithm,public_key,created_at) VALUES('key',?,'Ed25519',?,'now')`,a.ID,[]byte(ap));err!=nil{t.Fatal(err)}
 if _,err=grant.Grant(ctx,"did:plc:alice",a.ID,b.ID,"message.preview",time.Hour);err!=nil{t.Fatal(err)}
 if err=(LocalService{Store:store,Now:func()time.Time{return now}}).Consent(ctx,"did:plc:bob",b.ID,a.ID,time.Hour);err!=nil{t.Fatal(err)}
 transport:=NewLocalTransport(true,store.DB,store.Push);transport.now=func()time.Time{return now}
 makeEnv:=func(id string)Envelope{d:=sha256.Sum256([]byte("test"));e:=Envelope{MessageID:id,SenderAgentID:a.ID,RecipientAgentID:b.ID,KeyID:"key",PayloadDigest:hex.EncodeToString(d[:]),IssuedAt:now.Format(time.RFC3339),ExpiresAt:now.Add(time.Minute).Format(time.RFC3339)};raw,_:=e.canonical();hash:=sha256.Sum256(raw);e.Signature=ed25519.Sign(priv,hash[:]);return e}
 handler:=LoopbackHandler(true,transport)
 send:=func(id,remote string)int{raw,_:=json.Marshal(makeEnv(id));r:=httptest.NewRequest("POST","/agent-envelope",bytes.NewReader(raw));r.RemoteAddr=remote;r.Header.Set("Content-Type","application/json");w:=httptest.NewRecorder();handler.ServeHTTP(w,r);return w.Code}
 if n:=send("id1","203.0.113.1:9000");n!=403{t.Fatalf("nonloopback %d",n)}
 if n:=send("id1","127.0.0.1:9000");n!=202{t.Fatalf("accepted %d",n)}
 if n:=send("id1","127.0.0.1:9000");n!=409{t.Fatalf("replay %d",n)}
 if err=transport.RevokeConsent(ctx,"did:plc:bob",b.ID,a.ID);err!=nil{t.Fatal(err)}
 if n:=send("id2","127.0.0.1:9000");n!=403{t.Fatalf("revoked %d",n)}
 if n:=send("id3","192.168.1.5:9000");n!=403{t.Fatalf("LAN %d",n)}
 if n:=send("id3","invalid");n!=403{t.Fatalf("malformed peer %d",n)}
 if n:=send("id3","127.0.0.1:9000");n!=403{t.Fatalf("revoked localhost %d",n)}
 if got:=LoopbackHandler(false,transport);got==nil{t.Fatal("nil disabled handler")}
 if _,err=ListenLoopback(false);err!=ErrTransportDisabled{t.Fatalf("flag-off listener %v",err)}
 ln,err:=ListenLoopback(true);if err!=nil{t.Fatal(err)};defer ln.Close()
 host,_,err:=net.SplitHostPort(ln.Addr().String());if err!=nil||host!="127.0.0.1"{t.Fatalf("unsafe bind %s",ln.Addr())}
 client:=&http.Client{Timeout:time.Second}
 srv:=&http.Server{Handler:handler};go srv.Serve(ln);defer srv.Close()
 // Real TCP loopback request verifies network path independently of httptest.
 raw,_:=json.Marshal(makeEnv("id4"));req,err:=http.NewRequest("POST","http://"+ln.Addr().String()+"/agent-envelope",bytes.NewReader(raw));if err!=nil{t.Fatal(err)};req.Header.Set("Content-Type","application/json")
 resp,err:=client.Do(req);if err!=nil{t.Fatal(err)};resp.Body.Close();if resp.StatusCode!=403{t.Fatalf("expected revoked network denial got %d",resp.StatusCode)}
}
