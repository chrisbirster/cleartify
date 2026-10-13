package protocols

import (
 "encoding/json"
 "errors"
 "strings"
)
var ErrUnsafePermissions=errors.New("protocol permissions exceed conservative local-only policy")

// DEFAULT (Oct 13 2026), owner can change.
// A deliberately tiny preview-only permission taxonomy, never authorization
// for remote network calls, payment, message sends or device side effects.
var localPreviewPermissions=map[string]bool{
 "local.read":true,
 "local.write.preview":true,
 "notify.preview":true,
}
func ValidatePreviewPermissions(spec []byte)error{
 if len(spec)==0||len(spec)>65536{return ErrUnsafePermissions}
 var body struct{Permissions []string `json:"permissions"`;Steps []json.RawMessage `json:"steps"`}
 if err:=json.Unmarshal(spec,&body);err!=nil||body.Permissions==nil||len(body.Permissions)>16{return ErrUnsafePermissions}
 seen:=make(map[string]bool)
 for _,p:=range body.Permissions{
  if strings.TrimSpace(p)!=p||!localPreviewPermissions[p]||seen[p]{return ErrUnsafePermissions}
  seen[p]=true
 }
 if len(body.Steps)>100{return ErrUnsafePermissions}
 return nil
}
