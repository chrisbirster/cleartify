package protocols

import (
 "errors"
 "testing"
)
func TestPreviewPermissionAllowlist(t *testing.T){
 for _,tc:=range []struct{name,spec string;safe bool}{
  {"local preview",`{"permissions":["local.read","local.write.preview"],"steps":[]}`,true},
  {"empty permission list",`{"permissions":[]}`,true},
  {"no manifest",`{"steps":[]}`,false},
  {"network",`{"permissions":["network.fetch"]}`,false},
  {"payment",`{"permissions":["payment.charge"]}`,false},
  {"message send",`{"permissions":["message.send"]}`,false},
  {"duplicate",`{"permissions":["local.read","local.read"]}`,false},
  {"unknown",`{"permissions":["executor.host"]}`,false},
 }{
  t.Run(tc.name,func(t *testing.T){err:=ValidatePreviewPermissions([]byte(tc.spec));if (err==nil)!=tc.safe{t.Fatalf("validation error %v, expected safe=%v",err,tc.safe)};if !tc.safe && !errors.Is(err,ErrUnsafePermissions){t.Fatalf("wrong error: %v",err)}})
 }
}
