package social

import (
 "context"
 "database/sql"
 "errors"
 "regexp"
 "strings"
)
var ErrNotFound=errors.New("public profile not found")
var ErrInvalidHandle=errors.New("invalid handle")
var handleRx=regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?(?:\.[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)*$`)

type Profile struct{
 DID string `json:"did"`
 Handle string `json:"handle"`
 DisplayName string `json:"displayName"`
 Bio string `json:"bio"`
 AvatarURL string `json:"avatarUrl"`
 Kind string `json:"kind"`
 Claimed bool `json:"claimed"`
}
type Reader struct {DB *sql.DB}
func (r Reader) PublicByHandle(ctx context.Context, handle string)(Profile,error){
 handle=strings.ToLower(strings.TrimSpace(handle))
 if len(handle)>253||!handleRx.MatchString(handle){return Profile{},ErrInvalidHandle}
 var p Profile;var claimed int
 err:=r.DB.QueryRowContext(ctx,`SELECT did,handle,display_name,bio,avatar_url,kind,claimed FROM social_profiles WHERE handle=? AND public_visible=1 LIMIT 1`,handle).Scan(&p.DID,&p.Handle,&p.DisplayName,&p.Bio,&p.AvatarURL,&p.Kind,&claimed)
 if errors.Is(err,sql.ErrNoRows){return Profile{},ErrNotFound}
 if err!=nil{return Profile{},err}
 p.Claimed=claimed==1
 return p,nil
}
