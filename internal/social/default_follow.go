package social

import (
 "context"
 "crypto/sha256"
 "database/sql"
 "encoding/hex"
 "errors"
 "fmt"
 "strings"
 "time"

 "github.com/chrisbirster/cleartify/internal/database"
)

// DefaultFollow connects a locally authenticated AT Protocol social identity
// to an existing, claimed /@chris profile. No PDS or external network writes.
type DefaultFollow struct { Store *database.Store }

var ErrDefaultTargetUnavailable = errors.New("claimed /@chris profile is unavailable")
var ErrInvalidAuthenticatedDID = errors.New("invalid authenticated DID")

func placeholderHandle(did string) string {
 digest := sha256.Sum256([]byte(did))
 return "acct-" + hex.EncodeToString(digest[:12])
}

// OnVerifiedLogin is called only after Indigo has validated OAuth issuer,
// state and subject. Creates a private placeholder (not a public claim), then
// follows the local DID behind /@chris unless the user previously removed it.
func (s DefaultFollow) OnVerifiedLogin(ctx context.Context, did string) error {
 if !strings.HasPrefix(did,"did:")||len(did)>2048 {return ErrInvalidAuthenticatedDID}
 tx,err:=s.Store.DB.BeginTx(ctx,nil);if err!=nil{return err}
 defer tx.Rollback()
 now:=time.Now().UTC().Format(time.RFC3339Nano)
 _,err=tx.ExecContext(ctx,`INSERT INTO social_profiles(did,handle,display_name,kind,public_visible,claimed,updated_at)
 VALUES(?,?,'AT Protocol user','person',0,0,?) ON CONFLICT(did) DO NOTHING`,did,placeholderHandle(did),now)
 if err!=nil{return fmt.Errorf("create private social identity: %w",err)}
 var chrisDID string
 err=tx.QueryRowContext(ctx,`SELECT did FROM social_profiles WHERE handle='chris' AND kind='person' AND claimed=1 AND public_visible=1 LIMIT 1`).Scan(&chrisDID)
 if errors.Is(err,sql.ErrNoRows){if err:=tx.Commit();err!=nil{return err};return s.Store.Push(ctx)}
 if err!=nil{return err}
 if did!=chrisDID {
  _,err=tx.ExecContext(ctx,`INSERT INTO social_follows(follower_did,subject_did,created_at)
  SELECT ?,?,? WHERE NOT EXISTS (SELECT 1 FROM social_default_follow_optouts WHERE follower_did=? AND subject_did=?)
  ON CONFLICT(follower_did,subject_did) DO NOTHING`,did,chrisDID,now,did,chrisDID)
  if err!=nil{return fmt.Errorf("default follow: %w",err)}
 }
 if err:=tx.Commit();err!=nil{return err}
 return s.Store.Push(ctx)
}

// RemoveChrisFollow is an explicit, durable user opt-out. It never relies on a
// caller-supplied follower DID; the authenticated API supplies that identity.
func (s DefaultFollow) RemoveChrisFollow(ctx context.Context, authenticatedDID string) error {
 if !strings.HasPrefix(authenticatedDID,"did:")||len(authenticatedDID)>2048{return ErrInvalidAuthenticatedDID}
 tx,err:=s.Store.DB.BeginTx(ctx,nil);if err!=nil{return err}
 defer tx.Rollback()
 var target string
 err=tx.QueryRowContext(ctx,`SELECT did FROM social_profiles WHERE handle='chris' AND kind='person' AND claimed=1 AND public_visible=1 LIMIT 1`).Scan(&target)
 if errors.Is(err,sql.ErrNoRows){return ErrDefaultTargetUnavailable}
 if err!=nil{return err}
 now:=time.Now().UTC().Format(time.RFC3339Nano)
 if _,err=tx.ExecContext(ctx,`INSERT INTO social_default_follow_optouts(follower_did,subject_did,opted_out_at) VALUES(?,?,?)
 ON CONFLICT(follower_did,subject_did) DO UPDATE SET opted_out_at=excluded.opted_out_at`,authenticatedDID,target,now);err!=nil{return err}
 if _,err=tx.ExecContext(ctx,`DELETE FROM social_follows WHERE follower_did=? AND subject_did=?`,authenticatedDID,target);err!=nil{return err}
 if err:=tx.Commit();err!=nil{return err}
 return s.Store.Push(ctx)
}
