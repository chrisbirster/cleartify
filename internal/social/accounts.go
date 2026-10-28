package social

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/chrisbirster/cleartify/internal/database"
)

// Accounts creates a private, unclaimed social identity only for a DID
// sourced by the caller from an authenticated Indigo OAuth browser session.
type Accounts struct{ Store *database.Store }

var ErrSocialIdentityRequired = errors.New("verified AT Protocol identity required")

type AccountResult struct {
	DID     string `json:"did"`
	Claimed bool   `json:"claimed"`
	Public  bool   `json:"public"`
}

func (s Accounts) EnsureAuthenticated(ctx context.Context, verifiedDID string) (AccountResult, error) {
	if !strings.HasPrefix(verifiedDID, "did:") || len(verifiedDID) > 2048 {
		return AccountResult{}, ErrSocialIdentityRequired
	}
	h := sha256.Sum256([]byte(verifiedDID))
	handle := "acct-" + hex.EncodeToString(h[:12])
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := s.Store.DB.ExecContext(ctx, `INSERT INTO social_profiles(did,handle,display_name,kind,claimed,public_visible,updated_at)
 VALUES(?,?,'AT Protocol user','person',0,0,?)
 ON CONFLICT(did) DO NOTHING`, verifiedDID, handle, now)
	if err != nil {
		return AccountResult{}, err
	}
	if err = s.Store.Push(ctx); err != nil {
		return AccountResult{}, err
	}
	var claimed, public int
	err = s.Store.DB.QueryRowContext(ctx, `SELECT claimed,public_visible FROM social_profiles WHERE did=?`, verifiedDID).Scan(&claimed, &public)
	if errors.Is(err, sql.ErrNoRows) {
		return AccountResult{}, ErrSocialIdentityRequired
	}
	if err != nil {
		return AccountResult{}, err
	}
	return AccountResult{DID: verifiedDID, Claimed: claimed == 1, Public: public == 1}, nil
}

// Setup edits only the authenticated DID's local profile. Its generated local
// handle is retained; this does not claim or transfer any reserved handle.
type ProfileInput struct {
	DisplayName string `json:"displayName"`
	Bio         string `json:"bio"`
	Public      bool   `json:"public"`
}

func (s Accounts) Setup(ctx context.Context, did string, input ProfileInput) (AccountResult, error) {
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	input.Bio = strings.TrimSpace(input.Bio)
	if input.DisplayName == "" || len(input.DisplayName) > 160 || len(input.Bio) > 2000 {
		return AccountResult{}, ErrSocialIdentityRequired
	}
	if _, err := s.EnsureAuthenticated(ctx, did); err != nil {
		return AccountResult{}, err
	}
	// Reserved profiles keep the existing explicit claim ceremony.
	res, err := s.Store.DB.ExecContext(ctx, `UPDATE social_profiles SET display_name=?,bio=?,claimed=1,public_visible=?,updated_at=? WHERE did=? AND kind='person' AND (claimed=1 OR handle LIKE 'acct-%')`, input.DisplayName, input.Bio, input.Public, time.Now().UTC().Format(time.RFC3339Nano), did)
	if err != nil {
		return AccountResult{}, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return AccountResult{}, err
	}
	if n != 1 {
		return AccountResult{}, ErrSocialIdentityRequired
	}
	if err = s.Store.Push(ctx); err != nil {
		return AccountResult{}, err
	}
	return AccountResult{DID: did, Claimed: true, Public: input.Public}, nil
}

func(s Accounts)Own(ctx context.Context,did string)(map[string]any,error){
 if !strings.HasPrefix(did,"did:"){return nil,ErrSocialIdentityRequired}
 var handle,name,bio string;var visible,claimed bool
 err:=s.Store.DB.QueryRowContext(ctx,`SELECT handle,display_name,bio,public_visible,claimed FROM social_profiles WHERE did=? AND kind='person'`,did).Scan(&handle,&name,&bio,&visible,&claimed)
 if err!=nil{return nil,err};return map[string]any{"did":did,"handle":handle,"displayName":name,"bio":bio,"public":visible,"claimed":claimed},nil
}
