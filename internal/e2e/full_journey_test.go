package e2e

import (
 "context"
 "encoding/json"
 "errors"
 "fmt"
 "testing"
 "time"

 "github.com/chrisbirster/cleartify/internal/alerts"
 "github.com/chrisbirster/cleartify/internal/alpha"
 "github.com/chrisbirster/cleartify/internal/clearsession"
 "github.com/chrisbirster/cleartify/internal/identitylink"
 "github.com/chrisbirster/cleartify/internal/marketplace"
 "github.com/chrisbirster/cleartify/internal/protocols"
 "github.com/chrisbirster/cleartify/internal/runner"
 "github.com/chrisbirster/cleartify/internal/security"
 "github.com/chrisbirster/cleartify/internal/waitlist"
)

// Local journey composed from the same production persistence and service
// primitives, replacing providers with in-memory/local fakes. Each stage is
// independently named to produce a useful local-only alert on failure.
func TestProductionShapedLocalJourney(t *testing.T){
 ctx:=context.Background();sink:=alerts.NewLocal(8)
 h,err:=New(ctx,sink);if err!=nil{t.Fatal(err)};defer h.Close()
 run:=func(name string,fn func(context.Context)error){t.Helper();if err:=h.Stage(ctx,name,fn);err!=nil{t.Fatal(err)}}
 did:="did:plc:localbuyer";creator:="did:plc:localcreator"
 var userToken,publication,installID,entitlement string
 tokens,err:=security.NewTokens("local-only-test-secret-0123456789012345");if err!=nil{t.Fatal(err)}
 sessions:=&clearsession.Service{Store:h.Store,Tokens:tokens}
 gate:=alpha.Gate{DB:h.Store.DB}
 draft:=protocols.DraftService{DB:h.Store.DB,Tokens:tokens}
 installs:=protocols.InstallService{Store:h.Store}
 fake:=marketplace.Service{DB:h.Store.DB,Provider:marketplace.FakeProvider{},Enabled:true}
 run("signup",func(ctx context.Context)error{
  _,err:=waitlist.NewService(h.Store).Join(ctx,waitlist.Signup{Name:"Local Buyer",Email:"localbuyer@example.test",Phone:"+17175550123",AcceptPolicies:true,Source:"e2e"});if err!=nil{return err}
  _,err=h.Store.DB.ExecContext(ctx,`INSERT INTO users(id,display_name,primary_email,status,created_at,updated_at) VALUES('u1','Local Buyer','localbuyer@example.test','active','now','now')`);return err
 })
 run("session-alpha",func(ctx context.Context)error{
  _,err:=h.Store.DB.ExecContext(ctx,`INSERT INTO beta_invites(id,waitlist_entry_id,status,email_token_hash,email_token_expires_at,activated_user_id,activated_at,created_at,updated_at) SELECT 'i1',id,'activated','localhash','later','u1','now','now','now' FROM waitlist_entries WHERE email_normalized='localbuyer@example.test'`);if err!=nil{return err}
  if err=gate.Admit(ctx,"u1");err!=nil{return err}
  userToken,err=sessions.Issue(ctx,"u1");if err!=nil{return err}
  id,e:=sessions.Authenticate(ctx,userToken);if e!=nil{return e};if id!="u1"{return errors.New("incorrect Clear principal")}
  return nil
 })
 run("linked-identity-fake",func(ctx context.Context)error{
  now:=time.Now().UTC();nonce:="local-fake-challenge"
  _,err:=h.Store.DB.ExecContext(ctx,`INSERT INTO identity_link_challenges(nonce_hash,clear_user_id,social_did,state,consent_version,expires_at) VALUES(?,'u1',?,'pending','local-test-consent',?)`,nonce,did,now.Add(time.Minute).Format(time.RFC3339Nano));if err!=nil{return err}
  if err=(identitylink.Store{DB:h.Store.DB}).Commit(ctx,identitylink.CommitRequest{ID:"link1",NonceHash:nonce,ClearUserID:"u1",SocialDID:did,ConsentVersion:"local-test-consent",ClearAuthenticatedAt:now,SocialAuthenticatedAt:now,Now:now,EventID:"link-event1"});err!=nil{return err}
  var n int;err=h.Store.DB.QueryRowContext(ctx,`SELECT count(*) FROM identity_links WHERE clear_user_id='u1' AND social_did=? AND status='active'`,did).Scan(&n)
  if err!=nil{return err};if n!=1{return errors.New("identity link not active")}
  return nil
 })
 run("creator-publish",func(ctx context.Context)error{
  var e error
  publication,e=draft.Create(ctx,creator,protocols.DraftInput{Slug:"local-e2e",Version:1,Title:"Local E2E",License:"MIT",Spec:json.RawMessage(`{"permissions":[],"steps":[{"operation":"local.set","key":"ok","value":"yes"},{"operation":"local.assert","key":"ok","value":"yes"}]}`)})
  if e!=nil{return e}
  if e=draft.PublishDraft(ctx,creator,publication,"private","local original");e!=nil{return e}
  return draft.PublishDraft(ctx,creator,publication,"public","local original")
 })
 run("paid-gate-fake-checkout",func(ctx context.Context)error{
  _,err:=h.Store.DB.ExecContext(ctx,`INSERT INTO marketplace_listings(id,protocol_id,creator_did,title,pricing_kind,price_minor,currency,status,created_at) VALUES('local-listing',?,?,'Local Test','one_time',100,'USD','listed','now')`,publication,creator);if err!=nil{return err}
  allowed,e:=fake.PublicationAccess(ctx,did,publication);if e!=nil{return e};if allowed{return errors.New("unpaid access")}
  entitlement,e=fake.Purchase(ctx,did,"local-listing");if e!=nil{return e}
  repeat,e:=fake.Purchase(ctx,did,"local-listing");if e!=nil||repeat!=entitlement{return errors.New("fake purchase not idempotent")}
  allowed,e=fake.PublicationAccess(ctx,did,publication);if e!=nil{return e};if !allowed{return errors.New("paid fake entitlement missing")}
  n,e:=fake.ReconcileFake(ctx);if e!=nil{return e};if n!=0{return errors.New("fake audit reconciliation mismatch")}
  return nil
 })
 run("install-approval-to-local-run",func(ctx context.Context)error{
  manifest,e:=installs.Preview(ctx,publication);if e!=nil{return e}
  installID,e=installs.Request(ctx,did,publication,manifest.ManifestSHA256);if e!=nil{return e}
  if e=installs.Decide(ctx,installID,true);e!=nil{return e}
  var status string;e=h.Store.DB.QueryRowContext(ctx,"SELECT status FROM protocol_install_requests WHERE id=?",installID).Scan(&status);if e!=nil{return e}
  if status!="approved"{return errors.New("install not approved")}
  local:=runner.New(true)
  plan:=runner.Plan{Steps:[]runner.Step{{Operation:"local.set",Key:"ok",Value:"yes"},{Operation:"local.assert",Key:"ok",Value:"yes"}}}
  snapshot,e:=local.Start(plan);if e!=nil{return e}
  finished,e:=local.Run(ctx,snapshot.ID);if e!=nil{return e}
  if finished.Status!="completed"||finished.Values["ok"]!="yes"{return errors.New("local execution did not converge")}
  // Durable fake-run result + audit, never a real executor.
  payload,e:=json.Marshal(finished);if e!=nil{return e}
  _,e=h.Store.DB.ExecContext(ctx,`INSERT INTO clear_fake_install_runs(id,install_request_id,user_id,actor_did,publication_id,manifest_sha256,status,result_json,created_at,finished_at) VALUES('local-run',?,'u1',?,?,?,'completed',?,'now','now')`,installID,did,publication,manifest.ManifestSHA256,string(payload));if e!=nil{return e}
  _,e=h.Store.DB.ExecContext(ctx,`INSERT INTO clear_fake_install_audit(id,run_id,action,recorded_at) VALUES('local-run-audit','local-run','completed','now')`);return e
 })
 run("moderation-hold-and-release",func(ctx context.Context)error{
  _,e:=h.Store.DB.ExecContext(ctx,`INSERT INTO social_profiles(did,handle,display_name,kind,public_visible,claimed,updated_at) VALUES(?,'creator.test','Creator','person',1,1,'now')`,creator);if e!=nil{return e}
  _,e=h.Store.DB.ExecContext(ctx,`INSERT INTO social_impersonation_holds(subject_did,reason,created_at) VALUES(?,'lookalike','now')`,creator);if e!=nil{return e}
  reader:=protocols.Reader{DB:h.Store.DB,HoldVisibility:true}
  items,e:=reader.Discover(ctx,10);if e!=nil{return e};if len(items)!=0{return errors.New("held creator exposed")}
  _,e=h.Store.DB.ExecContext(ctx,`UPDATE social_impersonation_holds SET released_at='now' WHERE subject_did=?`,creator);if e!=nil{return e}
  items,e=reader.Discover(ctx,10);if e!=nil{return e};if len(items)!=1{return errors.New("released creator not visible")}
  return nil
 })
 run("fake-refund-and-audit",func(ctx context.Context)error{
  if e:=fake.Refund(ctx,did,entitlement);e!=nil{return e}
  allowed,e:=fake.PublicationAccess(ctx,did,publication);if e!=nil{return e};if allowed{return errors.New("refunded access remained active")}
  n,e:=fake.ReconcileFake(ctx);if e!=nil{return e};if n!=0{return fmt.Errorf("%d mismatched fake payment events",n)}
  return nil
 })
 run("local-http-and-noop-alert",func(ctx context.Context)error{
  if got:=h.Request("GET","/healthz").Code;got!=200{return fmt.Errorf("health %d",got)}
  if got:=h.Request("POST","/api/me/fake-checkout").Code;got!=404{return errors.New("disabled production route active")}
  return alerts.Noop{}.Emit(ctx,alerts.Event{Key:"journey_ok",Stage:"local",At:h.Clock()})
 })
 if len(sink.Snapshot())!=0{t.Fatalf("unexpected failures: %+v",sink.Snapshot())}
}
