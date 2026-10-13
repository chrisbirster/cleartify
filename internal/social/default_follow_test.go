package social

import(
 "context"
 "database/sql"
 "testing"

 "github.com/chrisbirster/cleartify/internal/database"
)

func TestAutoFollowChrisAndPersistentUserRemoval(t *testing.T){
 ctx:=context.Background()
 store,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)}
 defer store.Close()
 if err:=store.Migrate(ctx);err!=nil{t.Fatal(err)}
 chris:="did:plc:ewvi7nxzyoun6zhxrhs64oiz"
 alice:="did:plc:aaaaaaaaaaaaaaaaaaaaaaaa"
 bob:="did:plc:bbbbbbbbbbbbbbbbbbbbbbbb"
 _,err=store.DB.ExecContext(ctx,`INSERT INTO social_profiles(did,handle,display_name,kind,claimed,public_visible,updated_at) VALUES(?,'chris','Chris','person',1,1,'now')`,chris)
 if err!=nil{t.Fatal(err)}
 s:=DefaultFollow{Store:store}
 if err=s.OnVerifiedLogin(ctx,alice);err!=nil{t.Fatal(err)}
 var count int
 if err=store.DB.QueryRowContext(ctx,`SELECT count(*) FROM social_follows WHERE follower_did=? AND subject_did=?`,alice,chris).Scan(&count);err!=nil||count!=1{t.Fatalf("missing default follow %d %v",count,err)}
 var vis,claimed int
 if err=store.DB.QueryRowContext(ctx,`SELECT public_visible,claimed FROM social_profiles WHERE did=?`,alice).Scan(&vis,&claimed);err!=nil||vis!=0||claimed!=0{t.Fatalf("placeholder incorrectly public/claimed: %d %d %v",vis,claimed,err)}
 if err=s.OnVerifiedLogin(ctx,alice);err!=nil{t.Fatal(err)}
 if err=s.RemoveChrisFollow(ctx,alice);err!=nil{t.Fatal(err)}
 if err=s.OnVerifiedLogin(ctx,alice);err!=nil{t.Fatal(err)}
 if err=store.DB.QueryRowContext(ctx,`SELECT count(*) FROM social_follows WHERE follower_did=?`,alice).Scan(&count);err!=nil||count!=0{t.Fatalf("opt-out not respected: %d %v",count,err)}
 if err=s.OnVerifiedLogin(ctx,bob);err!=nil{t.Fatal(err)}
 if err=store.DB.QueryRowContext(ctx,`SELECT count(*) FROM social_follows WHERE follower_did=? AND subject_did=?`,bob,chris).Scan(&count);err!=nil||count!=1{t.Fatalf("new account not followed: %d %v",count,err)}
}
func TestAutoFollowRequiresClaimedChrisAndNoOutboundCalls(t *testing.T){
 ctx:=context.Background()
 store,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)}
 defer store.Close()
 if err:=store.Migrate(ctx);err!=nil{t.Fatal(err)}
 s:=DefaultFollow{Store:store}
 did:="did:plc:cccccccccccccccccccccccc"
 if err=s.OnVerifiedLogin(ctx,did);err!=nil{t.Fatal(err)}
 var n int
 if err=store.DB.QueryRowContext(ctx,`SELECT COUNT(*) FROM social_follows`).Scan(&n);err!=nil||n!=0{t.Fatalf("unexpected follow before claimed target: %d %v",n,err)}
 if err=s.RemoveChrisFollow(ctx,did);err!=ErrDefaultTargetUnavailable{t.Fatalf("missing target removal %v",err)}
 if err=s.OnVerifiedLogin(ctx,"not-a-did");err!=ErrInvalidAuthenticatedDID{t.Fatalf("invalid DID accepted: %v",err)}
 if err=store.DB.QueryRowContext(ctx,`SELECT COUNT(*) FROM social_default_follow_optouts`).Scan(&n);err!=nil||n!=0{t.Fatalf("unexpected optouts %d %v",n,err)}
 _=sql.ErrNoRows
}
