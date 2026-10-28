package protocols

import (
 "bytes"
 "encoding/json"
 "errors"
 "regexp"
 "strings"
 "time"
)

var ErrDraftSpec=errors.New("invalid draft spec or attribution")
var attributionID=regexp.MustCompile(`^[a-zA-Z0-9_-][a-zA-Z0-9._:-]{0,127}$`)
// ValidateDraftSpec is local draft validation, NOT permission to publish or
// execute arbitrary steps. Unknown executable fields are rejected.
func ValidateDraftSpec(raw []byte, license string) error {
 return validateDraftSpec(raw,license,false)
}
// ValidateDraftSpecWithGuides permits inert readable steps while keeping the legacy path strict.
func ValidateDraftSpecWithGuides(raw []byte,license string)error{return validateDraftSpec(raw,license,true)}
func validateDraftSpec(raw []byte,license string,allowGuides bool)error{
 var body struct {
  Steps []json.RawMessage `json:"steps"`
  Permissions []string `json:"permissions"`
  RemixOf string `json:"remixOf"`
  Attribution string `json:"attribution"`
 }
 dec:=json.NewDecoder(bytes.NewReader(raw))
 dec.DisallowUnknownFields()
 if err:=dec.Decode(&body);err!=nil{return ErrDraftSpec}
 if body.Steps==nil||len(body.Steps)>32{return ErrDraftSpec}
 seen:=map[string]bool{}
 if len(body.Permissions)>16{return ErrDraftSpec}
 for _,p:=range body.Permissions{
  if strings.TrimSpace(p)!=p||!localPreviewPermissions[p]||seen[p]{return ErrDraftSpec}
  seen[p]=true
 }
 for _,rawStep:=range body.Steps {
  var step struct{Operation string `json:"operation"`;Key string `json:"key"`;Value string `json:"value"`;Text string `json:"text"`;DueAt string `json:"dueAt"`}
  d:=json.NewDecoder(bytes.NewReader(rawStep));d.DisallowUnknownFields()
  if err:=d.Decode(&step);err!=nil{return ErrDraftSpec}
  if step.Operation=="guide" && allowGuides {
   if step.DueAt!="" {if _,err:=time.Parse(time.RFC3339,step.DueAt);err!=nil{return ErrDraftSpec}}
   if step.Key!=""||step.Value!=""||len(step.Text)<1||len(step.Text)>4096||strings.TrimSpace(step.Text)==""{return ErrDraftSpec}
   continue
  }
  if step.Text!=""||step.DueAt!=""||step.Operation!="local.set"&&step.Operation!="local.assert"{return ErrDraftSpec}
  if len(step.Key)==0||len(step.Key)>64||len(step.Value)>4096{return ErrDraftSpec}
  for _,c:=range step.Key {if !(c>='a'&&c<='z'||c>='A'&&c<='Z'||c>='0'&&c<='9'||c=='_'||c=='-'){return ErrDraftSpec}}
 }
 if body.RemixOf!="" {
  if !attributionID.MatchString(body.RemixOf)||len(strings.TrimSpace(body.Attribution))<3||len(body.Attribution)>500||strings.EqualFold(strings.TrimSpace(license),"all-rights-reserved"){return ErrDraftSpec}
 } else if body.Attribution!="" {return ErrDraftSpec}
 return nil
}
