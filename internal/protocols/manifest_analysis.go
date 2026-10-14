package protocols
import("bytes";"encoding/json";"sort")

// PermissionDiff is an offline review artifact; it never grants permissions.
type PermissionDiff struct{Added []string `json:"added"`;Removed []string `json:"removed"`;Changed bool `json:"changed"`}
func AnalyzeManifest(raw []byte)error{
 if ValidatePreviewPermissions(raw)!=nil{return ErrUnsafePermissions}
 var spec struct{Permissions []string `json:"permissions"`;Steps []json.RawMessage `json:"steps"`}
 d:=json.NewDecoder(bytes.NewReader(raw));d.DisallowUnknownFields()
 if e:=d.Decode(&spec);e!=nil{return ErrUnsafePermissions}
 if len(spec.Steps)>32{return ErrUnsafePermissions}
 for _,step:=range spec.Steps{
  var v struct{Operation string `json:"operation"`;Key string `json:"key"`;Value string `json:"value"`}
  parser:=json.NewDecoder(bytes.NewReader(step));parser.DisallowUnknownFields()
  if e:=parser.Decode(&v);e!=nil{return ErrUnsafePermissions}
  if v.Operation!="local.set"&&v.Operation!="local.assert"{return ErrUnsafePermissions}
  if len(v.Key)<1||len(v.Key)>64||len(v.Value)>4096{return ErrUnsafePermissions}
  for _,r:=range v.Key{if !(r>='a'&&r<='z'||r>='A'&&r<='Z'||r>='0'&&r<='9'||r=='_'||r=='-'){return ErrUnsafePermissions}}
 }
 return nil
}
func DiffPermissions(before,after []string)PermissionDiff{
 a,b:=map[string]bool{},map[string]bool{}
 for _,x:=range before{a[x]=true};for _,x:=range after{b[x]=true}
 diff:=PermissionDiff{Added:[]string{},Removed:[]string{}}
 for x:=range b{if !a[x]{diff.Added=append(diff.Added,x)}}
 for x:=range a{if !b[x]{diff.Removed=append(diff.Removed,x)}}
 sort.Strings(diff.Added);sort.Strings(diff.Removed)
 diff.Changed=len(diff.Added)+len(diff.Removed)>0
 return diff
}
