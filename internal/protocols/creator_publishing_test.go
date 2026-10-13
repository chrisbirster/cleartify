package protocols
import(
 "context"
 "encoding/json"
 "errors"
 "testing"
 "github.com/chrisbirster/cleartify/internal/database"
 "github.com/chrisbirster/cleartify/internal/security"
)
func TestCreatorPublishingLifecycle(t *testing.T){
 ctx:=context.Background();db,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)};defer db.Close()
 if err=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 tokens,err:=security.NewTokens("01234567890123456789012345678901");if err!=nil{t.Fatal(err)}
 service:=DraftService{DB:db.DB,Tokens:tokens}
 create:=func(slug,lic,spec string)string{
  t.Helper();id,e:=service.Create(ctx,"did:plc:author",DraftInput{Slug:slug,Version:1,Title:slug,License:lic,Spec:json.RawMessage(spec)})
  if e!=nil{t.Fatal(e)};return id
 }
 base:=create("base","CC-BY-4.0",`{"steps":[]}`)
 if err=service.PublishDraft(ctx,"did:plc:other",base,"private","original creator statement");!errors.Is(err,ErrPublishingConflict){t.Fatalf("cross-user: %v",err)}
 if err=service.PublishDraft(ctx,"did:plc:author",base,"public","original creator statement");!errors.Is(err,ErrPublishingConflict){t.Fatalf("skipped private: %v",err)}
 for _,status:=range []string{"private","unlisted","public"}{if err=service.PublishDraft(ctx,"did:plc:author",base,status,"original creator statement");err!=nil{t.Fatalf("%s: %v",status,err)}}
 if err=service.PublishDraft(ctx,"did:plc:author",base,"private","replay");!errors.Is(err,ErrPublishingConflict){t.Fatalf("replay: %v",err)}
 remix:=create("remix","CC-BY-4.0",`{"steps":[],"remixOf":"`+base+`","attribution":"Credit to author"}`)
 if err=service.PublishDraft(ctx,"did:plc:author",remix,"private","remix provenance reviewed");err!=nil{t.Fatal(err)}
 if err=service.PublishDraft(ctx,"did:plc:author",remix,"public","remix provenance reviewed");err!=nil{t.Fatal(err)}
 lineage,err:=(Takedowns{DB:db.DB}).Lineage(ctx,remix);if err!=nil||len(lineage)!=1||lineage[0]!=base{t.Fatalf("remix lineage %v %v",lineage,err)}
 discover,err:=(Reader{DB:db.DB}).Discover(ctx,20);if err!=nil{t.Fatal(err)}
 if len(discover)!=2{t.Fatalf("discover %+v",discover)}
 unlisted:=create("hidden","MIT",`{"steps":[]}`)
 if err=service.PublishDraft(ctx,"did:plc:author",unlisted,"private","original");err!=nil{t.Fatal(err)}
 if err=service.PublishDraft(ctx,"did:plc:author",unlisted,"unlisted","original");err!=nil{t.Fatal(err)}
 discover,err=(Reader{DB:db.DB}).Discover(ctx,20);if err!=nil{t.Fatal(err)}
 if len(discover)!=2{t.Fatalf("unlisted exposed: %+v",discover)}
 bad:=create("bad-remix","MIT",`{"steps":[],"remixOf":"`+base+`","attribution":"Credit to author"}`)
 if err=service.PublishDraft(ctx,"did:plc:author",bad,"private","provenance statement");!errors.Is(err,ErrPublishingInvalid){t.Fatalf("incompatible license %v",err)}
 var audits int;if err=db.DB.QueryRowContext(ctx,"SELECT count(*) FROM creator_publication_audit").Scan(&audits);err!=nil{t.Fatal(err)}
 if audits!=7{t.Fatalf("audit count=%d",audits)}
 takedown:=Takedowns{DB:db.DB}
 if _,err=takedown.Request(ctx,"did:plc:other",base,"rights_review");!errors.Is(err,ErrTakedownDenied){t.Fatalf("cross-user request %v",err)}
 req,err:=takedown.Request(ctx,"did:plc:author",base,"rights_review");if err!=nil{t.Fatal(err)}
 if _,err=takedown.Request(ctx,"did:plc:author",base,"rights_review");!errors.Is(err,ErrTakedownDenied){t.Fatalf("duplicate %v",err)}
 if err=takedown.Decide(ctx,req,"hidden");err!=nil{t.Fatal(err)}
 hidden,err:=(Reader{DB:db.DB}).Discover(ctx,20);if err!=nil||len(hidden)!=1{t.Fatalf("hidden discovery %d %v",len(hidden),err)}
 if err=takedown.Decide(ctx,req,"restored");err!=nil{t.Fatal(err)}
 visible,err:=(Reader{DB:db.DB}).Discover(ctx,20);if err!=nil||len(visible)!=2{t.Fatalf("restored discovery %d %v",len(visible),err)}
 if err=takedown.Decide(ctx,req,"restored");!errors.Is(err,ErrTakedownDenied){t.Fatalf("duplicate restoration %v",err)}
 var events int
 if err=db.DB.QueryRowContext(ctx,"SELECT count(*) FROM creator_takedown_audit WHERE request_id=?",req).Scan(&events);err!=nil||events!=3{t.Fatalf("audit %d %v",events,err)}
}
