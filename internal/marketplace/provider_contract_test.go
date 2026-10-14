package marketplace

import (
 "context"
 "errors"
 "testing"
 "time"
 "github.com/chrisbirster/cleartify/internal/database"
)

func providerContract(t *testing.T, p PaymentAdapter) {
 t.Helper()
 ctx:=context.Background()
 r,e:=p.Charge(ctx,"contract-1",1299,"USD")
 if e!=nil||r.ID!="fake-contract-1"||r.Provider!="fake"{t.Fatalf("charge: %+v %v",r,e)}
 r2,e:=p.Charge(ctx,"contract-1",1299,"USD")
 if e!=nil||r2!=r{t.Fatalf("replay: %+v %v",r2,e)}
 if _,e=p.Charge(ctx,"contract-2",-1,"USD");!errors.Is(e,ErrPayment){t.Fatal("negative amount accepted")}
 if _,e=p.Charge(ctx,"contract-2",9,"EUR");!errors.Is(e,ErrPayment){t.Fatal("unsupported currency accepted")}
 refunded,e:=p.Refund(ctx,r.ID)
 if e!=nil||refunded!=r{t.Fatalf("refund: %+v %v",refunded,e)}
 if _,e=p.Refund(ctx,"external-receipt");!errors.Is(e,ErrPayment){t.Fatal("nonfake receipt accepted")}
}
func TestFakeProviderContract(t *testing.T){providerContract(t,FakeProvider{})}
func TestFakeWebhookSigningContract(t *testing.T){
 signer:=FakeWebhookSigner{TestKey:[]byte("fake-test-only-signing-key-32bytes")}
 payload:=[]byte(`{"event":"fake.refund","id":"evt-1"}`)
 sig,e:=signer.Sign(payload,"evt-1");if e!=nil{t.Fatal(e)}
 if e=signer.Verify(payload,"evt-1",sig);e!=nil{t.Fatal(e)}
 for _,attempt:=range []struct{body []byte;id,signature string}{
 {[]byte(`{"event":"fake.refund","id":"evt-2"}`),"evt-1",sig},
 {payload,"evt-2",sig},
 {payload,"evt-1","abc"},
 }{
  if e=signer.Verify(attempt.body,attempt.id,attempt.signature);!errors.Is(e,ErrWebhookSignature){t.Fatalf("modified webhook accepted: %v",e)}
 }
 if e=ValidateFakeWebhookEvent(FakeWebhookEvent{ID:"evt-1",Kind:"fake.refund",ReceiptID:"fake-1",OccurredAt:time.Now()});e!=nil{t.Fatal(e)}
 if e=ValidateFakeWebhookEvent(FakeWebhookEvent{ID:"evt-1",Kind:"payment.execute",ReceiptID:"fake-1",OccurredAt:time.Now()});e==nil{t.Fatal("unknown webhook event accepted")}
}
func TestFakeKeyedPurchaseRefundIdempotency(t *testing.T){
 ctx:=context.Background()
 d,e:=database.Open(ctx,database.Config{Path:":memory:"});if e!=nil{t.Fatal(e)};defer d.Close()
 if e=d.Migrate(ctx);e!=nil{t.Fatal(e)}
 d.DB.SetMaxOpenConns(1)
 for _,q:=range []string{
  `INSERT INTO social_profiles(did,handle,display_name,kind,public_visible,claimed,updated_at) VALUES('did:plc:s','s.test','S','person',1,1,'now')`,
  `INSERT INTO published_protocols(id,creator_did,slug,version,title,summary,license,spec_json,status,created_at) VALUES('p','did:plc:s','p',1,'P','','MIT','{}','published','now')`,
  `INSERT INTO marketplace_listings(id,protocol_id,creator_did,title,pricing_kind,price_minor,currency,status,created_at) VALUES('l','p','did:plc:s','P','one_time',500,'USD','listed','now')`,
 }{
  if _,e=d.DB.ExecContext(ctx,q);e!=nil{t.Fatal(e)}
 }
 k:=KeyedOperations{DB:d.DB}
 if _,e=k.Purchase(ctx,"did:plc:b","purchase-001","l");!errors.Is(e,ErrDenied){t.Fatalf("off flag: %v",e)}
 k.Enabled=true
 id,e:=k.Purchase(ctx,"did:plc:b","purchase-001","l");if e!=nil{t.Fatal(e)}
 again,e:=k.Purchase(ctx,"did:plc:b","purchase-001","l");if e!=nil||again!=id{t.Fatalf("purchase replay: %s %v",again,e)}
 if _,e=k.Purchase(ctx,"did:plc:b","purchase-001","other-listing");!errors.Is(e,ErrIdempotencyConflict){t.Fatalf("key reuse allowed: %v",e)}
 if e=k.Refund(ctx,"did:plc:b","refund-001",id);e!=nil{t.Fatal(e)}
 if e=k.Refund(ctx,"did:plc:b","refund-001",id);e!=nil{t.Fatalf("refund replay: %v",e)}
 if e=k.Refund(ctx,"did:plc:b","purchase-001",id);!errors.Is(e,ErrIdempotencyConflict){t.Fatalf("cross-operation key reuse: %v",e)}
 var purchases,refunds int
 if e=d.DB.QueryRowContext(ctx,"SELECT COUNT(*) FROM marketplace_payment_audit WHERE event='fake_purchase'").Scan(&purchases);e!=nil{t.Fatal(e)}
 if e=d.DB.QueryRowContext(ctx,"SELECT COUNT(*) FROM marketplace_payment_audit WHERE event='fake_refund'").Scan(&refunds);e!=nil{t.Fatal(e)}
 if purchases!=1||refunds!=1{t.Fatalf("unexpected audit counts %d %d",purchases,refunds)}
}
