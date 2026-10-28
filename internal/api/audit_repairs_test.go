package api

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"github.com/bluesky-social/indigo/atproto/auth/oauth"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/chrisbirster/cleartify/internal/agents"
	"github.com/chrisbirster/cleartify/internal/atproto"
	"github.com/chrisbirster/cleartify/internal/database"
	"github.com/chrisbirster/cleartify/internal/interop"
	"github.com/chrisbirster/cleartify/internal/protocols"
	"github.com/chrisbirster/cleartify/internal/social"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLocalAgentWorkerRoutingAndOperatorAuthorization(t *testing.T) {
	ctx := context.Background()
	db, e := database.Open(ctx, database.Config{Path: ":memory:"})
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	if e = db.Migrate(ctx); e != nil {
		t.Fatal(e)
	}
	worker := &interop.AuthenticatedWorker{Enabled: true, DB: db.DB, Transport: interop.NewLocalTransport(true, db.DB, db.Push), Outbox: interop.Outbox{Enabled: true, DB: db.DB, Sink: interop.DurableInbox{DB: db.DB}, Lease: time.Second, Backoff: time.Second, MaxAttempts: 3}}
	for _, enabled := range []bool{false, true} {
		h := New(Options{AgentTransportEnabled: enabled, AgentWorker: worker, AgentOwnershipEnabled: true, AgentOwnership: &agents.Ownership{DB: db.DB, Push: db.Push}, AdminToken: "operator-test", AppOrigin: "http://localhost:3000"})
		for _, token := range []string{"", "operator-test"} {
			req := httptest.NewRequest("POST", "/api/admin/agents/local-worker/tick", nil)
			req.Header.Set("Origin", "http://localhost:3000")
			if token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			want := 404
			if enabled {
				want = 200
				if token == "" {
					want = 401
				}
			}
			if rec.Code != want {
				t.Fatalf("worker enabled=%v token=%q: %d %s", enabled, token, rec.Code, rec.Body.String())
			}
		}
		req := httptest.NewRequest("POST", "/api/me/agents/claimed/messages", strings.NewReader(`{"senderAgentId":"forged"}`))
		req.Header.Set("Origin", "http://localhost:3000")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if enabled && rec.Code < 400 {
			t.Fatal("unauthenticated sender accepted")
		}
		req = httptest.NewRequest("POST", "/api/admin/agents/claimed/activate", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		rec = httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != 401 {
			t.Fatal("unreviewed agent activated", rec.Code)
		}
	}
}
func TestHiddenPublicationCannotStartFakeExecution(t *testing.T) {
	ctx := context.Background()
	db, e := database.Open(ctx, database.Config{Path: ":memory:"})
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	if e = db.Migrate(ctx); e != nil {
		t.Fatal(e)
	}
	for _, q := range []string{
		`INSERT INTO users(id,display_name,primary_email,created_at,updated_at) VALUES('u','User','u@example.test','now','now')`,
		`INSERT INTO identity_links(id,clear_user_id,social_did,status,consent_version,consented_at,authn_clear_at,authn_social_at,linked_at) VALUES('link','u','did:web:reader.example.test','active','test','now','now','now','now')`,
		`INSERT INTO published_protocols(id,creator_did,slug,version,title,summary,license,spec_json,status,created_at) VALUES('p','did:plc:creator','demo',1,'Demo','','MIT','{"permissions":[],"steps":[{"operation":"local.set","key":"a","value":"b"}]}','published','now')`,
		`INSERT INTO creator_publications(protocol_id,visibility,provenance_note,approved_at,updated_at) VALUES('p','public','original','now','now')`,
	} {
		if _, e = db.DB.ExecContext(ctx, q); e != nil {
			t.Fatal(e)
		}
	}
	flow := protocols.InstallService{Store: db}
	m, e := flow.Preview(ctx, "p")
	if e != nil {
		t.Fatal(e)
	}
	request, e := flow.Request(ctx, "did:web:reader.example.test", "p", m.ManifestSHA256)
	if e != nil {
		t.Fatal(e)
	}
	if e = flow.Decide(ctx, request, true); e != nil {
		t.Fatal(e)
	}
	takedowns := protocols.Takedowns{DB: db.DB}
	id, e := takedowns.Request(ctx, "did:plc:creator", "p", "rights_review")
	if e != nil {
		t.Fatal(e)
	}
	if e = takedowns.Decide(ctx, id, "hidden"); e != nil {
		t.Fatal(e)
	}
	oauthHandler, cookie := auditSocialSession(t, db, "did:web:reader.example.test")
	server := &Server{fakeInstallRunDB: db.DB, fakeInstallRunPush: db.Push, clearPrincipalResolver: func(*http.Request) (string, error) { return "u", nil }, socialAuth: oauthHandler}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/me/fake-install-runs", strings.NewReader(`{"installRequestId":"`+request+`"}`))
	req.AddCookie(cookie)
	server.startFakeInstallRun(rec, req)
	if rec.Code != 409 {
		t.Fatalf("hidden execution: %d %s", rec.Code, rec.Body.String())
	}
	var n int
	if e = db.DB.QueryRowContext(ctx, `SELECT count(*) FROM clear_fake_install_runs`).Scan(&n); e != nil || n != 0 {
		t.Fatal("hidden run had side effects", n, e)
	}
	if e = takedowns.Decide(ctx, id, "restored"); e != nil {
		t.Fatal(e)
	}
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/me/fake-install-runs", strings.NewReader(`{"installRequestId":"`+request+`"}`))
	req.AddCookie(cookie)
	server.startFakeInstallRun(rec, req)
	if rec.Code != 201 {
		t.Fatalf("restored execution: %d %s", rec.Code, rec.Body.String())
	}
	// A pinned spec remains a real digest, not a slug/version alias.
	digest := sha256.Sum256([]byte(`{"permissions":[],"steps":[{"operation":"local.set","key":"a","value":"b"}]}`))
	if m.SpecSHA256 != hex.EncodeToString(digest[:]) {
		t.Fatal("lost spec pin")
	}
}

func auditSocialSession(t *testing.T, db *database.Store, did string) (*atproto.OAuthHandler, *http.Cookie) {
	t.Helper()
	ctx := context.Background()
	key := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{9}, 32))
	handler, e := atproto.NewOAuthHandler(db, "https://api.example.test", "https://app.example.test", key)
	if e != nil {
		t.Fatal(e)
	}
	parsed, e := syntax.ParseDID(did)
	if e != nil {
		t.Fatal(e)
	}
	raw, e := json.Marshal(oauth.ClientSessionData{AccountDID: parsed, SessionID: "audit-session"})
	if e != nil {
		t.Fatal(e)
	}
	box, e := atproto.NewSecretBox(key)
	if e != nil {
		t.Fatal(e)
	}
	encrypted, e := box.Seal(raw, "session:"+did+":audit-session")
	if e != nil {
		t.Fatal(e)
	}
	if _, e = db.DB.ExecContext(ctx, `INSERT INTO atproto_oauth_sessions(account_did,session_id,payload_ciphertext,created_at,updated_at) VALUES(?,'audit-session',?,'now','now')`, did, encrypted); e != nil {
		t.Fatal(e)
	}
	token := strings.Repeat("audit-cookie-", 4)
	digest := sha256.Sum256([]byte(token))
	if _, e = db.DB.ExecContext(ctx, `INSERT INTO atproto_browser_sessions(token_digest,account_did,oauth_session_id,expires_at,created_at) VALUES(?,?,'audit-session',?,'now')`, hex.EncodeToString(digest[:]), did, time.Now().UTC().Add(time.Hour).Format(time.RFC3339Nano)); e != nil {
		t.Fatal(e)
	}
	return handler, &http.Cookie{Name: "cleartify_atproto_session", Value: token}
}
func TestProfileSetupRequiresRealLocalOAuthSessionAndKeepsOtherProfilesPrivate(t *testing.T) {
	ctx := context.Background()
	db, e := database.Open(ctx, database.Config{Path: ":memory:"})
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	if e = db.Migrate(ctx); e != nil {
		t.Fatal(e)
	}
	auth, cookie := auditSocialSession(t, db, "did:web:reader.example.test")
	accounts := &social.Accounts{Store: db}
	h := New(Options{OAuth: auth, SocialAccountRegistrationEnabled: true, SocialAccounts: accounts, Profiles: &social.Reader{DB: db.DB}, AppOrigin: "http://localhost:3000"})
	call := func(body, origin string, withCookie bool) *httptest.ResponseRecorder {
		r := httptest.NewRequest("PUT", "/api/me/social-profile", strings.NewReader(body))
		r.Header.Set("Origin", origin)
		if withCookie {
			r.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	body := `{"displayName":"Reader","bio":"My profile","public":true}`
	if w := call(body, "http://localhost:3000", false); w.Code != 401 {
		t.Fatal("unauthenticated setup", w.Code)
	}
	if w := call(body, "https://evil.example", true); w.Code != 403 {
		t.Fatal("cross-origin setup", w.Code)
	}
	if w := call(`{"did":"did:web:other.example","displayName":"Other","public":true}`, "http://localhost:3000", true); w.Code != 400 {
		t.Fatal("caller chose DID", w.Code)
	}
	if w := call(body, "http://localhost:3000", true); w.Code != 200 {
		t.Fatal("setup failed", w.Code, w.Body.String())
	}
	own, e := accounts.Own(ctx, "did:web:reader.example.test")
	if e != nil {
		t.Fatal(e)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/api/profiles/"+own["handle"].(string), nil))
	if w.Code != 200 {
		t.Fatal("setup not publicly reachable", w.Code, w.Body.String())
	}
	if w := call(`{"displayName":"Reader","bio":"Private","public":false}`, "http://localhost:3000", true); w.Code != 200 {
		t.Fatal(w.Code)
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/api/profiles/"+own["handle"].(string), nil))
	if w.Code != 404 {
		t.Fatal("private profile exposed", w.Code)
	}
}
