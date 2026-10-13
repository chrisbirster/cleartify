package identitylink

import (
 "context"
 "database/sql"
 "errors"
 "time"
)

var ErrConflict=errors.New("identity link conflict")
var ErrChallenge=errors.New("identity link challenge invalid")

// Store is a persistence primitive only. Do not call Commit from an HTTP route
// without independent, fresh Clear and OAuth principal verification and consent.
type Store struct{DB *sql.DB}
type CommitRequest struct {
 ID string
 NonceHash string
 ClearUserID string
 SocialDID string
 ConsentVersion string
 ClearAuthenticatedAt time.Time
 SocialAuthenticatedAt time.Time
 Now time.Time
 EventID string
}
func(s Store) Commit(ctx context.Context,r CommitRequest)error{
 if r.ID==""||r.NonceHash==""||r.ClearUserID==""||r.SocialDID==""||r.ConsentVersion==""||r.EventID==""||r.ClearAuthenticatedAt.IsZero()||r.SocialAuthenticatedAt.IsZero()||r.Now.IsZero(){return ErrChallenge}
 now:=r.Now.UTC().Format(time.RFC3339Nano)
 tx,err:=s.DB.BeginTx(ctx,nil);if err!=nil{return err}
 defer tx.Rollback()
 var state,expires,consent,user,did string
 err=tx.QueryRowContext(ctx,"SELECT state,expires_at,consent_version,clear_user_id,social_did FROM identity_link_challenges WHERE nonce_hash=?",r.NonceHash).Scan(&state,&expires,&consent,&user,&did)
 if errors.Is(err,sql.ErrNoRows){return ErrChallenge};if err!=nil{return err}
 deadline,parseErr:=time.Parse(time.RFC3339Nano,expires)
 if parseErr!=nil||!r.Now.Before(deadline)||state!="pending"||consent!=r.ConsentVersion||user!=r.ClearUserID||did!=r.SocialDID{return ErrChallenge}
 result,err:=tx.ExecContext(ctx,"UPDATE identity_link_challenges SET state='consumed',consumed_at=? WHERE nonce_hash=? AND state='pending'",now,r.NonceHash)
 if err!=nil{return err}
 n,err:=result.RowsAffected();if err!=nil{return err};if n!=1{return ErrChallenge}
 _,err=tx.ExecContext(ctx,`INSERT INTO identity_links(id,clear_user_id,social_did,status,consent_version,consented_at,authn_clear_at,authn_social_at,linked_at) VALUES(?,?,?,'active',?,?,?,?,?)`,r.ID,r.ClearUserID,r.SocialDID,r.ConsentVersion,now,r.ClearAuthenticatedAt.UTC().Format(time.RFC3339Nano),r.SocialAuthenticatedAt.UTC().Format(time.RFC3339Nano),now)
 if err!=nil{return ErrConflict}
 _,err=tx.ExecContext(ctx,"INSERT INTO identity_link_events(id,clear_user_id,social_did,action,recorded_at) VALUES(?,?,?,'linked',?)",r.EventID,r.ClearUserID,r.SocialDID,now)
 if err!=nil{return err}
 return tx.Commit()
}
