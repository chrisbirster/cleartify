package social

import(
 "context"
 "testing"
 "github.com/chrisbirster/cleartify/internal/database"
)
func TestPrivateSocialAccountRequiresVerifiedDID(t *testing.T){
 ctx:=context.Background()
 db,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)}
 defer db.Close()
 if err=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 accounts:=Accounts{Store:db}
 if _,err=accounts.EnsureAuthenticated(ctx,"");err!=ErrSocialIdentityRequired{t.Fatalf("empty DID allowed: %v",err)}
 if _,err=accounts.EnsureAuthenticated(ctx,"email@example.com");err!=ErrSocialIdentityRequired{t.Fatalf("non DID allowed: %v",err)}
 did:="did:plc:aaaaaaaaaaaaaaaaaaaaaaaa"
 v,err:=accounts.EnsureAuthenticated(ctx,did);if err!=nil{t.Fatal(err)}
 if v.DID!=did||v.Claimed||v.Public{t.Fatalf("public social account without verification: %+v",v)}
 _,err=accounts.EnsureAuthenticated(ctx,did);if err!=nil{t.Fatal(err)}
 var n int
 if err=db.DB.QueryRowContext(ctx,`SELECT count(*) FROM social_profiles WHERE did=? AND claimed=0 AND public_visible=0`,did).Scan(&n);err!=nil||n!=1{t.Fatalf("not idempotent: %d %v",n,err)}
}
