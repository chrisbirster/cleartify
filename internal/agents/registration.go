package agents

import (
 "context"
 "crypto/rand"
 "database/sql"
 "encoding/hex"
 "errors"
 "strings"
 "time"
)
var ErrInvalidRegistration=errors.New("invalid or unverified agent owner")
type Service struct{DB *sql.DB}
type Registration struct {
 ID string `json:"id"`
 OwnerDID string `json:"ownerDid"`
 DisplayName string `json:"displayName"`
 Description string `json:"description"`
 Status string `json:"status"`
}
func (s Service) Register(ctx context.Context,ownerDID,name,description string)(Registration,error){
 name=strings.TrimSpace(name);description=strings.TrimSpace(description)
 if len(name)<1||len(name)>80||len(description)>500||!strings.HasPrefix(ownerDID,"did:")||len(ownerDID)>2048{return Registration{},ErrInvalidRegistration}
 var isClaimed int
 err:=s.DB.QueryRowContext(ctx,`SELECT claimed FROM social_profiles WHERE did=? AND kind='person' AND public_visible=1`,ownerDID).Scan(&isClaimed)
 if errors.Is(err,sql.ErrNoRows)||isClaimed!=1{return Registration{},ErrInvalidRegistration};if err!=nil{return Registration{},err}
 token:=make([]byte,16);if _,err=rand.Read(token);err!=nil{return Registration{},err}
 reg:=Registration{ID:hex.EncodeToString(token),OwnerDID:ownerDID,DisplayName:name,Description:description,Status:"unverified"}
 _,err=s.DB.ExecContext(ctx,`INSERT INTO agent_registrations(id,owner_did,display_name,description,status,created_at) VALUES(?,?,?,?,'unverified',?)`,reg.ID,reg.OwnerDID,reg.DisplayName,reg.Description,time.Now().UTC().Format(time.RFC3339Nano))
 return reg,err
}
