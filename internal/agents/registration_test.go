package agents
import(
 "context"
 "testing"
 "github.com/chrisbirster/cleartify/internal/database"
)
func TestRegisterUnverifiedAgentOnlyForClaimedOwner(t *testing.T){
 ctx:=context.Background()
 db,err:=database.Open(ctx,database.Config{Path:":memory:"});if err!=nil{t.Fatal(err)}
 defer db.Close()
 if err=db.Migrate(ctx);err!=nil{t.Fatal(err)}
 s:=Service{DB:db.DB}
 if _,err:=s.Register(ctx,"did:plc:ewvi7nxzyoun6zhxrhs64oiz","Agent","bio");err!=ErrInvalidRegistration{t.Fatalf("unclaimed accepted: %v",err)}
 _,err=db.DB.ExecContext(ctx,`INSERT INTO social_profiles(did,handle,display_name,kind,public_visible,claimed,updated_at) VALUES('did:plc:ewvi7nxzyoun6zhxrhs64oiz','creator.example','Creator','person',1,1,'now')`);if err!=nil{t.Fatal(err)}
 reg,err:=s.Register(ctx,"did:plc:ewvi7nxzyoun6zhxrhs64oiz","Demo Agent","local only")
 if err!=nil{t.Fatal(err)}
 if reg.Status!="unverified"||reg.ID==""{t.Fatalf("unexpected registration: %+v",reg)}
 var status string
 if err:=db.DB.QueryRowContext(ctx,`SELECT status FROM agent_registrations WHERE id=?`,reg.ID).Scan(&status);err!=nil||status!="unverified"{t.Fatalf("status: %s %v",status,err)}
}
