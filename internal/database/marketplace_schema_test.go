package database

import ("context";"testing")
func TestMarketplacePricingConstraints(t *testing.T){
 ctx:=context.Background()
 db,err:=Open(ctx,Config{Path:":memory:"});if err!=nil{t.Fatal(err)}
 defer db.Close()
 if err=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 _,err=db.DB.ExecContext(ctx,`INSERT INTO published_protocols(id,creator_did,slug,version,title,summary,license,spec_json,status,created_at) VALUES('p','did:plc:ewvi7nxzyoun6zhxrhs64oiz','weekly',1,'Weekly','summary','MIT','{}','published','now')`);if err!=nil{t.Fatal(err)}
 _,err=db.DB.ExecContext(ctx,`INSERT INTO marketplace_listings(id,protocol_id,creator_did,title,currency,price_minor,pricing_kind,created_at) VALUES('l','p','did:plc:ewvi7nxzyoun6zhxrhs64oiz','Weekly','USD',299,'one_time','now')`);if err!=nil{t.Fatal(err)}
 for _,price:=range []int{-1,1000001}{
  _,err=db.DB.ExecContext(ctx,`INSERT INTO marketplace_listings(id,protocol_id,creator_did,title,currency,price_minor,pricing_kind,created_at) VALUES(?, 'p','did:plc:ewvi7nxzyoun6zhxrhs64oiz','Weekly','USD',?,'one_time','now')`,"invalid"+string(rune(price+100)),price)
  if err==nil{t.Errorf("invalid price %d allowed",price)}
 }
}
