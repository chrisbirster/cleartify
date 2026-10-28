package agents
import(
 "context"
 "crypto/ed25519"
 "crypto/rand"
 "encoding/hex"
 "testing"
 "github.com/chrisbirster/cleartify/internal/database"
)
func TestOwnershipEd25519Ceremony(t *testing.T){
 ctx:=context.Background()
 db,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)};defer db.Close()
 if err=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 _,err=db.DB.ExecContext(ctx,`INSERT INTO social_profiles(did,handle,display_name,kind,claimed,public_visible,updated_at) VALUES('did:plc:owner','owner.example','Owner','person',1,1,'now')`);if err!=nil{t.Fatal(err)}
 reg,err:=(Service{DB:db.DB}).Register(ctx,"did:plc:owner","Bot","test");if err!=nil{t.Fatal(err)}
 svc:=Ownership{DB:db.DB,Push:db.Push,Verifier:Ed25519OwnershipVerifier{}}
 key,private,err:=ed25519.GenerateKey(rand.Reader);if err!=nil{t.Fatal(err)}
 if _,err=svc.Start(ctx,"did:plc:attacker",reg.ID,key);err==nil{t.Fatal("cross-owner challenge")}
 nonce,err:=svc.Start(ctx,"did:plc:owner",reg.ID,key);if err!=nil{t.Fatal(err)}
 if err=svc.Confirm(ctx,"did:plc:attacker",reg.ID,nonce,"x");err==nil{t.Fatal("cross-owner confirm")}
 if err=svc.Confirm(ctx,"did:plc:owner",reg.ID,nonce,"incorrect");err==nil{t.Fatal("invalid proof")}
 proof:=hex.EncodeToString(ed25519.Sign(private,OwnershipProofMessage(reg.ID,"did:plc:owner",nonce,key)))
 other,_,_:=ed25519.GenerateKey(rand.Reader);_ = other
 badProof:=hex.EncodeToString(ed25519.Sign(private,OwnershipProofMessage(reg.ID,"did:plc:attacker",nonce,key)))
 if err=svc.Confirm(ctx,"did:plc:owner",reg.ID,nonce,badProof);err==nil{t.Fatal("wrong domain owner accepted")}
 if err=svc.Confirm(ctx,"did:plc:owner",reg.ID,nonce,proof);err!=nil{t.Fatal(err)}
 if err=svc.Confirm(ctx,"did:plc:owner",reg.ID,nonce,proof);err==nil{t.Fatal("replay succeeded")}
 var audits int;if err=db.DB.QueryRowContext(ctx,"SELECT count(*) FROM agent_owner_audit").Scan(&audits);err!=nil||audits!=1{t.Fatalf("audit %d %v",audits,err)}
 entries,err:=Directory(ctx,db.DB,20);if err!=nil||len(entries)!=0{t.Fatal("proof automatically promoted agent",entries,err)}
 if err=svc.Promote(ctx,reg.ID);err!=nil{t.Fatal("reviewed activation",err)}
 entries,err=Directory(ctx,db.DB,20);if err!=nil||len(entries)!=1||entries[0].ID!=reg.ID{t.Fatal("promoted agent absent",entries,err)}
 if err=svc.Promote(ctx,reg.ID);err==nil{t.Fatal("activation replay")}
 adapter:=LocalTestAdapter{}
 if ok,err:=adapter.Health(ctx,reg.ID);err!=nil||!ok{t.Fatal("adapter health")}
 if _,err:=adapter.Challenge(ctx,reg.ID,nonce);err!=nil{t.Fatal(err)}
}
