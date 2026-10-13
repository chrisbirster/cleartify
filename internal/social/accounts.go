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
type Accounts struct {Store *database.Store}
var ErrSocialIdentityRequired=errors.New("verified AT Protocol identity required")
type AccountResult struct {
 DID string `json:"did"`
 Claimed bool `json:"claimed"`
 Public bool `json:"public"`
}
func(s Accounts) EnsureAuthenticated(ctx context.Context,verifiedDID string)(AccountResult,error){
 if !strings.HasPrefix(verifiedDID,"did:")||len(verifiedDID)>2048{return AccountResult{},ErrSocialIdentityRequired}
 h:=sha256.Sum256([]byte(verifiedDID))
 handle:="acct-"+hex.EncodeToString(h[:12])
 now:=time.Now().UTC().Format(time.RFC3339Nano)
 _,err:=s.Store.DB.ExecContext(ctx,`INSERT INTO social_profiles(did,handle,display_name,kind,claimed,public_visible,updated_at)
 VALUES(?,?,'AT Protocol user','person',0,0,?)
 ON CONFLICT(did) DO NOTHING`,verifiedDID,handle,now)
 if err!=nil{return AccountResult{},err}
 if err=s.Store.Push(ctx);err!=nil{return AccountResult{},err}
 var claimed,public int
 err=s.Store.DB.QueryRowContext(ctx,`SELECT claimed,public_visible FROM social_profiles WHERE did=?`,verifiedDID).Scan(&claimed,&public)
 if errors.Is(err,sql.ErrNoRows){return AccountResult{},ErrSocialIdentityRequired}
 if err!=nil{return AccountResult{},err}
 return AccountResult{DID:verifiedDID,Claimed:claimed==1,Public:public==1},nil
}
