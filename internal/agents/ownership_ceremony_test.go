package agents
import(
 "context"
 "crypto/sha256"
 "encoding/hex"
 "testing"
 "github.com/chrisbirster/cleartify/internal/database"
)
func TestOwnershipFakeCeremony(t *testing.T){
 ctx:=context.Background()
 db,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)};defer db.Close()
 if err=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 _,err=db.DB.ExecContext(ctx,`INSERT INTO social_profiles(did,handle,display_name,kind,claimed,public_visible,updated_at) VALUES('did:plc:owner','owner.example','Owner','person',1,1,'now')`);if err!=nil{t.Fatal(err)}
 reg,err:=(Service{DB:db.DB}).Register(ctx,"did:plc:owner","Bot","test");if err!=nil{t.Fatal(err)}
 svc:=Ownership{DB:db.DB,Push:db.Push,Verifier:LocalFakeVerifier{}}
 key:=make([]byte,32);for i:=range key{key[i]=byte(i)}
 if _,err=svc.Start(ctx,"did:plc:attacker",reg.ID,key);err==nil{t.Fatal("cross-owner challenge")}
 nonce,err:=svc.Start(ctx,"did:plc:owner",reg.ID,key);if err!=nil{t.Fatal(err)}
 if err=svc.Confirm(ctx,"did:plc:attacker",reg.ID,nonce,"x");err==nil{t.Fatal("cross-owner confirm")}
 if err=svc.Confirm(ctx,"did:plc:owner",reg.ID,nonce,"incorrect");err==nil{t.Fatal("invalid proof")}
 proofBytes:=sha256.Sum256(append(append([]byte(nil),key...),[]byte(nonce)...));proof:=hex.EncodeToString(proofBytes[:])
 if err=svc.Confirm(ctx,"did:plc:owner",reg.ID,nonce,proof);err!=nil{t.Fatal(err)}
 if err=svc.Confirm(ctx,"did:plc:owner",reg.ID,nonce,proof);err==nil{t.Fatal("replay succeeded")}
 var audits int;if err=db.DB.QueryRowContext(ctx,"SELECT count(*) FROM agent_owner_audit").Scan(&audits);err!=nil||audits!=1{t.Fatalf("audit %d %v",audits,err)}
 adapter:=LocalTestAdapter{}
 if ok,err:=adapter.Health(ctx,reg.ID);err!=nil||!ok{t.Fatal("adapter health")}
 if _,err:=adapter.Challenge(ctx,reg.ID,nonce);err!=nil{t.Fatal(err)}
}
