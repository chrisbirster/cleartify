package social
import(
 "context"
 "errors"
 "testing"
 "github.com/chrisbirster/cleartify/internal/database"
)
func TestProfileClaimCannotStealHandleAndFollowIsReversible(t *testing.T){
 ctx:=context.Background()
 db,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)}
 defer db.Close()
 if err=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 alice:="did:plc:aaaaaaaaaaaaaaaaaaaaaaaa"
 bob:="did:plc:bbbbbbbbbbbbbbbbbbbbbbbb"
 chris:="did:plc:cccccccccccccccccccccccc"
 _,err=db.DB.ExecContext(ctx,`INSERT INTO social_profiles(did,handle,display_name,kind,claimed,public_visible,updated_at) VALUES
 (?,'alice.test','Alice','person',0,0,'now'),
 (?,'bob.test','Bob','person',1,1,'now'),
 (?,'chris','Chris','person',1,1,'now')`,alice,bob,chris)
 if err!=nil{t.Fatal(err)}
 a:=Actions{Store:db}
 if err=a.ClaimExisting(ctx,alice,"bob.test");!errors.Is(err,ErrOwnership){t.Fatalf("stole handle: %v",err)}
 if err=a.ClaimExisting(ctx,alice,"not-linked.test");!errors.Is(err,ErrNotClaimable){t.Fatalf("unlinked handle: %v",err)}
 if err=a.ClaimExisting(ctx,alice,"alice.test");err!=nil{t.Fatal(err)}
 var claimed,visible int
 if err=db.DB.QueryRowContext(ctx,`SELECT claimed,public_visible FROM social_profiles WHERE did=?`,alice).Scan(&claimed,&visible);err!=nil||claimed!=1||visible!=1{t.Fatalf("claim not persisted: %v",err)}
 if err=a.SetFollow(ctx,alice,"alice.test",true);!errors.Is(err,ErrInvalidFollow){t.Fatalf("self-follow: %v",err)}
 if err=a.SetFollow(ctx,alice,"bob.test",true);err!=nil{t.Fatal(err)}
 if err=a.SetFollow(ctx,alice,"bob.test",true);err!=nil{t.Fatal(err)}
 var n int
 if err=db.DB.QueryRowContext(ctx,`SELECT count(*) FROM social_follows WHERE follower_did=? AND subject_did=?`,alice,bob).Scan(&n);err!=nil||n!=1{t.Fatalf("duplicate follows %d %v",n,err)}
 if err=a.SetFollow(ctx,alice,"bob.test",false);err!=nil{t.Fatal(err)}
 if err=a.SetFollow(ctx,alice,"chris",false);err!=nil{t.Fatal(err)}
 if err=db.DB.QueryRowContext(ctx,`SELECT count(*) FROM social_default_follow_optouts WHERE follower_did=? AND subject_did=?`,alice,chris).Scan(&n);err!=nil||n!=1{t.Fatalf("optout missing %d %v",n,err)}
 if err=a.SetFollow(ctx,alice,"chris",true);err!=nil{t.Fatal(err)}
 if err=db.DB.QueryRowContext(ctx,`SELECT count(*) FROM social_default_follow_optouts WHERE follower_did=?`,alice).Scan(&n);err!=nil||n!=0{t.Fatalf("explicit re-follow didn't reset optout %d %v",n,err)}
}
