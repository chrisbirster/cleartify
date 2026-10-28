package protocols

import(
 "context"
 "encoding/json"
 "errors"
 "testing"
 "github.com/chrisbirster/cleartify/internal/database"
 "github.com/chrisbirster/cleartify/internal/security"
)
func TestPrivateCreatorDrafts(t *testing.T){
 ctx:=context.Background()
 db,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)};defer db.Close()
 if err:=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 tok,err:=security.NewTokens("01234567890123456789012345678901");if err!=nil{t.Fatal(err)}
 svc:=DraftService{DB:db.DB,Tokens:tok}
 in:=DraftInput{Slug:"original-protocol",Version:1,Title:"Original",Summary:"Private draft",License:"all-rights-reserved",Spec:json.RawMessage(`{"steps":[]}`)}
 if _,err=svc.Create(ctx,"did:plc:alice",in);err!=nil{t.Fatal(err)}
 if _,err=svc.Create(ctx,"did:plc:alice",in);err!=nil{t.Fatalf("identical retry: %v",err)}
 in.Title="Conflicting title"
 if _,err=svc.Create(ctx,"did:plc:alice",in);!errors.Is(err,ErrDraftConflict){t.Fatalf("conflicting revision: %v",err)}
 in.Title="Original"
 if _,err=svc.Create(ctx,"did:plc:bob",in);err!=nil{t.Fatal(err)}
 in.Version=2
 if _,err=svc.Create(ctx,"did:plc:alice",in);err!=nil{t.Fatal(err)}
 in.Spec=json.RawMessage(`[1,2]`)
 if _,err=svc.Create(ctx,"did:plc:alice",in);!errors.Is(err,ErrInvalidDraft){t.Fatalf("array spec: %v",err)}
 var published int
 if err:=db.DB.QueryRowContext(ctx,"SELECT count(*) FROM published_protocols WHERE status='published'").Scan(&published);err!=nil{t.Fatal(err)}
 if published!=0{t.Fatal("draft leaked into public discovery")}
}
