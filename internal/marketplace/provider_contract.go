package marketplace

import (
 "context"
 "crypto/hmac"
 "crypto/sha256"
 "encoding/hex"
 "errors"
 "strings"
 "time"
)

// PaymentAdapter is the provider-neutral contract for a future review. The
// runtime remains FakeProvider-only and rejects other Provider implementations.
type PaymentAdapter interface {
 Charge(context.Context,string,int64,string)(Receipt,error)
 Refund(context.Context,string)(Receipt,error)
}
var _ PaymentAdapter = FakeProvider{}

// SignatureVerifier is an offline verification boundary. It takes a complete
// raw payload and never fetches provider resources or interprets commands.
type SignatureVerifier interface {
 Sign([]byte,string)(string,error)
 Verify([]byte,string,string)error
}
var ErrWebhookSignature=errors.New("invalid fake webhook signature")
type FakeWebhookSigner struct{TestKey []byte}
func(s FakeWebhookSigner)mac(payload []byte,eventID string)([]byte,error){
 if len(s.TestKey)<16||len(s.TestKey)>128||eventID==""||len(eventID)>128||len(payload)>64*1024{return nil,ErrWebhookSignature}
 h:=hmac.New(sha256.New,s.TestKey)
 h.Write([]byte("cleartify-fake-webhook-v1\n"))
 h.Write([]byte(eventID));h.Write([]byte("\n"));h.Write(payload)
 return h.Sum(nil),nil
}
func(s FakeWebhookSigner)Sign(payload []byte,eventID string)(string,error){
 mac,e:=s.mac(payload,eventID);if e!=nil{return "",e}
 return hex.EncodeToString(mac),nil
}
func(s FakeWebhookSigner)Verify(payload []byte,eventID,signature string)error{
 want,e:=s.mac(payload,eventID);if e!=nil{return e}
 got,e:=hex.DecodeString(signature)
 if e!=nil||!hmac.Equal(got,want){return ErrWebhookSignature}
 return nil
}
type FakeWebhookEvent struct{ID string;Kind string;ReceiptID string;OccurredAt time.Time}
func ValidateFakeWebhookEvent(e FakeWebhookEvent)error{
 if e.ID==""||len(e.ID)>128||!strings.HasPrefix(e.ReceiptID,"fake-")||e.OccurredAt.IsZero(){return ErrWebhookSignature}
 switch e.Kind{case "fake.purchase","fake.refund","fake.dispute":return nil}
 return ErrWebhookSignature
}
// SellerComplianceState is a placeholder only. It carries no bank account,
// taxpayer identifier, legal classification, or external onboarding token.
type SellerComplianceState struct{SellerDID string;PayoutState string;TaxDataState string}
