package social

import (
 "context"
 "errors"
 "testing"
 "github.com/chrisbirster/cleartify/internal/database"
)

func TestPublicProfileByHandleUsesStableDID(t *testing.T){
 ctx:=context.Background()
 store,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)}
 defer store.Close()
 if err:=store.Migrate(ctx);err!=nil{t.Fatal(err)}
 _,err=store.DB.ExecContext(ctx,`INSERT INTO social_profiles(did,handle,display_name,bio,kind,public_visible,claimed,updated_at) VALUES
 ('did:plc:ewvi7nxzyoun6zhxrhs64oiz','chris.example','Chris','Hi','person',1,1,'2026-10-12'),
 ('did:plc:aaaaaaaaaaaaaaaaaaaaaaaa','private.example','Private','hidden','agent',0,0,'2026-10-12')`)
 if err!=nil{t.Fatal(err)}
 reader:=Reader{DB:store.DB}
 p,err:=reader.PublicByHandle(ctx,"CHRIS.example")
 if err!=nil{t.Fatal(err)}
 if p.DID!="did:plc:ewvi7nxzyoun6zhxrhs64oiz" || !p.Claimed {t.Fatalf("%+v",p)}
 if _,err:=reader.PublicByHandle(ctx,"private.example");!errors.Is(err,ErrNotFound){t.Fatalf("private leaked: %v",err)}
 if _,err:=reader.PublicByHandle(ctx,"../../bad");!errors.Is(err,ErrInvalidHandle){t.Fatalf("invalid accepted: %v",err)}
 _,err=store.DB.ExecContext(ctx,`UPDATE social_profiles SET handle='new.example' WHERE did='did:plc:ewvi7nxzyoun6zhxrhs64oiz'`)
 if err!=nil{t.Fatal(err)}
 renamed,err:=reader.PublicByHandle(ctx,"new.example");if err!=nil{t.Fatal(err)}
 if renamed.DID!=p.DID{t.Fatal("changing handle changes identity")}
}
