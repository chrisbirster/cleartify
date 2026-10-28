package social

import (
	"context"
	"errors"
	"github.com/chrisbirster/cleartify/internal/database"
	"github.com/chrisbirster/cleartify/internal/security"
	"testing"
)

func TestOrdinaryProfileSetupBlocksAndHiddenUnfollow(t *testing.T) {
	ctx := context.Background()
	db, e := database.Open(ctx, database.Config{Path: ":memory:"})
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	if e = db.Migrate(ctx); e != nil {
		t.Fatal(e)
	}
	accounts := Accounts{Store: db}
	a := "did:plc:alice"
	b := "did:plc:bob"
	for _, did := range []string{a, b} {
		if _, e = accounts.Setup(ctx, did, ProfileInput{DisplayName: did, Bio: "Local profile", Public: true}); e != nil {
			t.Fatal(e)
		}
	}
	own, e := accounts.Own(ctx, b)
	if e != nil {
		t.Fatal(e)
	}
	handle := own["handle"].(string)
	if _, e = (Reader{DB: db.DB}).PublicByHandle(ctx, handle); e != nil {
		t.Fatal("new account cannot become public", e)
	}
	actions := Actions{Store: db}
	if e = actions.SetFollow(ctx, a, handle, true); e != nil {
		t.Fatal(e)
	}
	tokens, _ := security.NewTokens("social-audit-regression-secret-000000")
	mod := Moderation{DB: db.DB, Tokens: tokens, Push: db.Push}
	if e = mod.Block(ctx, b, a, true); e != nil {
		t.Fatal(e)
	}
	var n int
	if e = db.DB.QueryRowContext(ctx, `SELECT count(*) FROM social_follows`).Scan(&n); e != nil || n != 0 {
		t.Fatal("block kept follow graph", n, e)
	}
	if e = actions.SetFollow(ctx, a, handle, true); !errors.Is(e, ErrInvalidFollow) {
		t.Fatal("blocked follow allowed", e)
	}
	if e = mod.Block(ctx, b, a, false); e != nil {
		t.Fatal(e)
	}
	if e = actions.SetFollow(ctx, a, handle, true); e != nil {
		t.Fatal(e)
	}
	if _, e = accounts.Setup(ctx, b, ProfileInput{DisplayName: "Bob", Public: false}); e != nil {
		t.Fatal(e)
	}
	if e = actions.SetFollow(ctx, a, handle, false); e != nil {
		t.Fatal("hidden profile cannot be unfollowed", e)
	}
	if e = actions.SetFollow(ctx, a, handle, true); !errors.Is(e, ErrNotFound) {
		t.Fatal("hidden profile followed", e)
	}
	// A different authenticated DID cannot claim an existing reserved profile.
	if _, e = db.DB.ExecContext(ctx, `INSERT INTO social_profiles(did,handle,display_name,kind,claimed,updated_at) VALUES('did:plc:reserved','reserved','Reserved','person',0,'now')`); e != nil {
		t.Fatal(e)
	}
	if _, e = accounts.Setup(ctx, "did:plc:reserved", ProfileInput{DisplayName: "Hijack", Public: true}); e == nil {
		t.Fatal("reserved profile bypassed claim ceremony")
	}
}
