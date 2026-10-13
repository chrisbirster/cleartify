package identitylink

import (
 "context"
 "errors"
 "testing"
 "time"
 "github.com/chrisbirster/cleartify/internal/database"
)
func TestCommitSingleUseAndUniqueness(t *testing.T){
 ctx:=context.Background()
 db,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)};defer db.Close()
 if err=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 now:=time.Date(2026,10,13,12,0,0,0,time.UTC)
 stamp:=func(t time.Time)string{return t.UTC().Format(time.RFC3339Nano)}
 for _,id:=range []string{"one","two"}{
  if _,err=db.DB.ExecContext(ctx,"INSERT INTO users(id,display_name,primary_email,status,created_at,updated_at) VALUES(?,?,?,'active',?,?)",id,id,id+"@example.com",stamp(now),stamp(now));err!=nil{t.Fatal(err)}
 }
 mk:=func(nonce,user,did string,exp time.Time){
  t.Helper()
  if _,err:=db.DB.ExecContext(ctx,"INSERT INTO identity_link_challenges(nonce_hash,clear_user_id,social_did,state,consent_version,expires_at) VALUES(?,?,?,'pending','v1',?)",nonce,user,did,stamp(exp));err!=nil{t.Fatal(err)}
 }
 mk("nonce1","one","did:plc:alice",now.Add(time.Minute))
 svc:=Store{DB:db.DB}
 request:=func(nonce,user,did,id,event string)CommitRequest{return CommitRequest{ID:id,NonceHash:nonce,ClearUserID:user,SocialDID:did,ConsentVersion:"v1",ClearAuthenticatedAt:now,SocialAuthenticatedAt:now,Now:now,EventID:event}}
 if err:=svc.Commit(ctx,request("nonce1","two","did:plc:alice","bad","bad"));!errors.Is(err,ErrChallenge){t.Fatalf("forged principal allowed: %v",err)}
 if err:=svc.Commit(ctx,request("nonce1","one","did:plc:alice","l1","e1"));err!=nil{t.Fatal(err)}
 if err:=svc.Commit(ctx,request("nonce1","one","did:plc:alice","l2","e2"));!errors.Is(err,ErrChallenge){t.Fatalf("replay: %v",err)}
 mk("nonce2","two","did:plc:alice",now.Add(time.Minute))
 if err:=svc.Commit(ctx,request("nonce2","two","did:plc:alice","l3","e3"));!errors.Is(err,ErrConflict){t.Fatalf("duplicate did: %v",err)}
 mk("nonce3","one","did:plc:bob",now.Add(time.Minute))
 if err:=svc.Commit(ctx,request("nonce3","one","did:plc:bob","l4","e4"));!errors.Is(err,ErrConflict){t.Fatalf("duplicate user: %v",err)}
 mk("nonce4","two","did:plc:charlie",now.Add(-time.Minute))
 if err:=svc.Commit(ctx,request("nonce4","two","did:plc:charlie","l5","e5"));!errors.Is(err,ErrChallenge){t.Fatalf("expired: %v",err)}
 var links,events int
 if err:=db.DB.QueryRowContext(ctx,"SELECT count(*) FROM identity_links WHERE status='active'").Scan(&links);err!=nil{t.Fatal(err)}
 if err:=db.DB.QueryRowContext(ctx,"SELECT count(*) FROM identity_link_events").Scan(&events);err!=nil{t.Fatal(err)}
 if links!=1||events!=1{t.Fatalf("partial writes: links=%d events=%d",links,events)}
}
