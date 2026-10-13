package social

import (
 "context"
 "database/sql"
 "errors"
 "fmt"
 "strings"
 "time"

 "github.com/chrisbirster/cleartify/internal/database"
)

var ErrOwnership = errors.New("profile DID does not belong to authenticated identity")
var ErrNotClaimable = errors.New("handle has not been independently linked to this DID")
var ErrInvalidFollow = errors.New("invalid follow")

type Actions struct { Store *database.Store }

// ClaimExisting refuses any reassignment of an unverified handle. The locally
// verified OAuth subject must match the profile's immutable DID, and a
// previously linked non-placeholder handle must already belong to that DID.
// A real PDS/DID handle-resolution check is still needed before launch.
func(a Actions) ClaimExisting(ctx context.Context,actorDID,handle string) error {
 handle=strings.ToLower(strings.TrimSpace(handle))
 if !strings.HasPrefix(actorDID,"did:") || len(actorDID)>2048 {return ErrOwnership}
 if len(handle)>253 || !handleRx.MatchString(handle) || strings.HasPrefix(handle,"acct-"){return ErrNotClaimable}
 var ownerDID,kind string
 err:=a.Store.DB.QueryRowContext(ctx,`SELECT did,kind FROM social_profiles WHERE handle=? LIMIT 1`,handle).Scan(&ownerDID,&kind)
 if errors.Is(err,sql.ErrNoRows){return ErrNotClaimable}
 if err!=nil{return err}
 if ownerDID!=actorDID || kind!="person"{return ErrOwnership}
 res,err:=a.Store.DB.ExecContext(ctx,`UPDATE social_profiles SET claimed=1,public_visible=1,updated_at=? WHERE did=? AND handle=? AND kind='person'`,time.Now().UTC().Format(time.RFC3339Nano),actorDID,handle)
 if err!=nil{return err}
 n,err:=res.RowsAffected();if err!=nil{return err};if n!=1{return ErrOwnership}
 return a.Store.Push(ctx)
}

// SetFollow updates only the locally stored graph, never the PDS graph.
func(a Actions) SetFollow(ctx context.Context,actorDID,targetHandle string,follow bool)error {
 targetHandle=strings.ToLower(strings.TrimSpace(targetHandle))
 if !strings.HasPrefix(actorDID,"did:") || len(actorDID)>2048 || len(targetHandle)>253 || !handleRx.MatchString(targetHandle){return ErrInvalidFollow}
 tx,err:=a.Store.DB.BeginTx(ctx,nil);if err!=nil{return err}
 defer tx.Rollback()
 var actor int
 err=tx.QueryRowContext(ctx,`SELECT 1 FROM social_profiles WHERE did=? AND kind='person'`,actorDID).Scan(&actor)
 if errors.Is(err,sql.ErrNoRows){return ErrOwnership};if err!=nil{return err}
 var target,handle string
 err=tx.QueryRowContext(ctx,`SELECT did,handle FROM social_profiles WHERE handle=? AND public_visible=1 AND claimed=1 LIMIT 1`,targetHandle).Scan(&target,&handle)
 if errors.Is(err,sql.ErrNoRows){return ErrNotFound};if err!=nil{return err}
 if target==actorDID{return ErrInvalidFollow}
 now:=time.Now().UTC().Format(time.RFC3339Nano)
 if follow {
  // An explicit follow resets a prior opt-out, including for /@chris.
  if _,err=tx.ExecContext(ctx,`DELETE FROM social_default_follow_optouts WHERE follower_did=? AND subject_did=?`,actorDID,target);err!=nil{return err}
  if _,err=tx.ExecContext(ctx,`INSERT INTO social_follows(follower_did,subject_did,created_at) VALUES(?,?,?) ON CONFLICT(follower_did,subject_did) DO NOTHING`,actorDID,target,now);err!=nil{return err}
 } else {
  if handle=="chris" {
   if _,err=tx.ExecContext(ctx,`INSERT INTO social_default_follow_optouts(follower_did,subject_did,opted_out_at) VALUES(?,?,?) ON CONFLICT(follower_did,subject_did) DO UPDATE SET opted_out_at=excluded.opted_out_at`,actorDID,target,now);err!=nil{return fmt.Errorf("persist default-follow opt-out: %w",err)}
  }
  if _,err=tx.ExecContext(ctx,`DELETE FROM social_follows WHERE follower_did=? AND subject_did=?`,actorDID,target);err!=nil{return err}
 }
 if err=tx.Commit();err!=nil{return err}
 return a.Store.Push(ctx)
}
