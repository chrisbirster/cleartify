package interop

import (
 "context"
 "crypto/ed25519"
 "crypto/rand"
 "crypto/sha256"
 "encoding/hex"
 "errors"
 "testing"
 "time"
 "github.com/chrisbirster/cleartify/internal/agents"
 "github.com/chrisbirster/cleartify/internal/database"
)
func TestTwoAgentLocalConsentRotationAndReplay(t *testing.T){
 ctx:=context.Background()
 store,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)};defer store.Close()
 if err=store.Migrate(ctx);err!=nil{t.Fatal(err)}
 now:=time.Date(2026,10,13,14,0,0,0,time.UTC)
 alice,bob:="did:plc:alice","did:plc:bob"
 for _,u:=range []struct{did,handle string}{{alice,"alice.local"},{bob,"bob.local"}}{
  if _,err=store.DB.ExecContext(ctx,`INSERT INTO social_profiles(did,handle,display_name,kind,public_visible,claimed,updated_at) VALUES(?,?,?,'person',1,1,'now')`,u.did,u.handle,u.handle);err!=nil{t.Fatal(err)}
 }
 a,err:=(agents.Service{DB:store.DB}).Register(ctx,alice,"A","");if err!=nil{t.Fatal(err)}
 b,err:=(agents.Service{DB:store.DB}).Register(ctx,bob,"B","");if err!=nil{t.Fatal(err)}
 ag:=agents.GrantService{DB:store.DB,Now:func()time.Time{return now}}
 aPub,aPriv,_:=ed25519.GenerateKey(rand.Reader);bPub,bPriv,_:=ed25519.GenerateKey(rand.Reader)
 if err=ag.Bind(ctx,alice,a.ID,aPub,ed25519.Sign(aPriv,agents.OwnerProofMessage(a.ID,alice)));err!=nil{t.Fatal(err)}
 if err=ag.Bind(ctx,bob,b.ID,bPub,ed25519.Sign(bPriv,agents.OwnerProofMessage(b.ID,bob)));err!=nil{t.Fatal(err)}
 for _,id:=range []string{a.ID,b.ID}{
  if _,err=store.DB.ExecContext(ctx,`UPDATE agent_registrations SET status='verified' WHERE id=?`,id);err!=nil{t.Fatal(err)}
 }
 if _,err=store.DB.ExecContext(ctx,`INSERT INTO agent_interop_keys(key_id,agent_id,algorithm,public_key,created_at) VALUES('old',?,'Ed25519',?,'now')`,a.ID,[]byte(aPub));err!=nil{t.Fatal(err)}
 svc:=LocalService{Store:store,Now:func()time.Time{return now}}
 env:=func(id,key string,private ed25519.PrivateKey)Envelope{
  dig:=sha256.Sum256([]byte("local-only"))
  e:=Envelope{MessageID:id,SenderAgentID:a.ID,RecipientAgentID:b.ID,KeyID:key,PayloadDigest:hex.EncodeToString(dig[:]),IssuedAt:now.Format(time.RFC3339),ExpiresAt:now.Add(5*time.Minute).Format(time.RFC3339)}
  canonical,_:=e.canonical();digest:=sha256.Sum256(canonical);e.Signature=ed25519.Sign(private,digest[:]);return e
 }
 if err=svc.VerifyAndRecordConsented(ctx,env("one","old",aPriv));!errors.Is(err,ErrConsent){t.Fatal("missing consent/grant accepted",err)}
 if _,err=ag.Grant(ctx,alice,a.ID,b.ID,"message.preview",time.Hour);err!=nil{t.Fatal(err)}
 if err=svc.VerifyAndRecordConsented(ctx,env("one","old",aPriv));!errors.Is(err,ErrConsent){t.Fatal("missing consent accepted",err)}
 if err=svc.Consent(ctx,alice,b.ID,a.ID,time.Hour);!errors.Is(err,ErrConsent){t.Fatal("wrong owner consent accepted",err)}
 if err=svc.Consent(ctx,bob,b.ID,a.ID,time.Hour);err!=nil{t.Fatal(err)}
 if err=svc.VerifyAndRecordConsented(ctx,env("one","old",aPriv));err!=nil{t.Fatal(err)}
 if err=svc.VerifyAndRecordConsented(ctx,env("one","old",aPriv));!errors.Is(err,ErrReplay){t.Fatal("replay accepted",err)}
 nextPub,nextPriv,_:=ed25519.GenerateKey(rand.Reader)
 proof:=ed25519.Sign(aPriv,[]byte("cleartify-rotate-v1\n"+a.ID+"\nold\nnew\n"+string(nextPub)))
 if err=svc.RotateKey(ctx,a.ID,"old","new",nextPub,proof);err!=nil{t.Fatal(err)}
 if err=svc.VerifyAndRecordConsented(ctx,env("two","old",aPriv));!errors.Is(err,ErrKeyUnavailable){t.Fatal("revoked key accepted",err)}
 if err=svc.VerifyAndRecordConsented(ctx,env("two","new",nextPriv));err!=nil{t.Fatal(err)}
 if err=svc.RevokeConsent(ctx,alice,b.ID,a.ID);!errors.Is(err,ErrConsent){t.Fatal("wrong-owner revoke succeeded",err)}
 if err=svc.RevokeConsent(ctx,bob,b.ID,a.ID);err!=nil{t.Fatal(err)}
 if err=svc.VerifyAndRecordConsented(ctx,env("three","new",nextPriv));!errors.Is(err,ErrConsent){t.Fatal("revoked consent accepted",err)}
}
