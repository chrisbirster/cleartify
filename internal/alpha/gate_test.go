package alpha

import (
 "context"
 "testing"
 "github.com/chrisbirster/cleartify/internal/database"
)
func TestAlphaMustBeActivatedAndRevocable(t *testing.T){
 ctx:=context.Background()
 db,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)}
 defer db.Close()
 if err=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 _,err=db.DB.ExecContext(ctx,`INSERT INTO users(id,display_name,primary_email,created_at,updated_at) VALUES('u1','Demo','a@b.co','now','now')`)
 if err!=nil{t.Fatal(err)}
 gate:=Gate{DB:db.DB}
 if err=gate.Admit(ctx,"u1");err!=ErrNotActivated{t.Fatalf("unverified admitted: %v",err)}
 _,err=db.DB.ExecContext(ctx,`INSERT INTO waitlist_entries(id,name,email,email_normalized,phone_e164,created_at,updated_at) VALUES('w','Demo','a@b.co','a@b.co','+17175551212','now','now')`);if err!=nil{t.Fatal(err)}
 _,err=db.DB.ExecContext(ctx,`INSERT INTO beta_invites(id,waitlist_entry_id,status,email_token_hash,email_token_expires_at,activated_user_id,activated_at,created_at,updated_at) VALUES('i','w','activated','hash','later','u1','now','now','now')`);if err!=nil{t.Fatal(err)}
 if err=gate.Admit(ctx,"u1");err!=nil{t.Fatal(err)}
 eligible,err:=gate.Eligible(ctx,"u1");if err!=nil||!eligible{t.Fatalf("admitted user ineligible: %v",err)}
 if err=gate.Revoke(ctx,"u1");err!=nil{t.Fatal(err)}
 eligible,err=gate.Eligible(ctx,"u1");if err!=nil||eligible{t.Fatalf("revoked user eligible: %v",err)}
}
