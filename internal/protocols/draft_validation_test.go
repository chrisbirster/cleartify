package protocols

import(
 "encoding/json"
 "errors"
 "testing"
)
func TestDraftValidationRejectsExecutableOrInvalidRemix(t *testing.T){
 cases:=[]struct{spec,license string;ok bool}{
  {`{"steps":[]}`,"all-rights-reserved",true},
  {`{"steps":[{"operation":"local.set","key":"score","value":"1"}],"permissions":["local.write.preview"]}`,"CC0-1.0",true},
  {`{"steps":[{"operation":"http.get","key":"url"}]}`,"CC0-1.0",false},
  {`{"steps":[{"operation":"local.set","key":"score","value":"1","shell":"echo"}]}`,"CC0-1.0",false},
  {`{"steps":[],"permissions":["network.send"]}`,"CC0-1.0",false},
  {`{"steps":[],"remixOf":"original-1"}`,"CC0-1.0",false},
  {`{"steps":[],"remixOf":"original-1","attribution":"Original source: original-1"}`,"all-rights-reserved",false},
  {`{"steps":[],"remixOf":"original-1","attribution":"Original source: original-1"}`,"CC-BY-4.0",true},
  {`{"steps":[],"attribution":"Unlinked claim"}`,"CC0-1.0",false},
  {`{"steps":null}`,"CC0-1.0",false},
  {`{"steps":[],"executable":"rm -rf"}`,"CC0-1.0",false},
 }
 for _,tc:=range cases {if err:=ValidateDraftSpec(json.RawMessage(tc.spec),tc.license);(err==nil)!=tc.ok{t.Errorf("spec %s license %s: %v",tc.spec,tc.license,err)}}
 if err:=ValidateDraftSpec([]byte(`[]`),"CC0-1.0");!errors.Is(err,ErrDraftSpec){t.Fatal(err)}
}
