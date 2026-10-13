package protocols

import (
 "context"
 "database/sql"
 "encoding/json"
 "errors"
 "regexp"
 "strings"
 "time"
 "github.com/chrisbirster/cleartify/internal/security"
)
var ErrInvalidDraft=errors.New("invalid protocol draft")
var ErrDraftConflict=errors.New("protocol draft version already exists")
var draftSlug=regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
type DraftInput struct{
 Slug string `json:"slug"`
 Version int `json:"version"`
 Title string `json:"title"`
 Summary string `json:"summary"`
 License string `json:"license"`
 Spec json.RawMessage `json:"spec"`
}
type DraftService struct{DB *sql.DB;Tokens *security.Tokens}
func(s DraftService) Create(ctx context.Context,verifiedDID string,d DraftInput)(string,error){
 if !strings.HasPrefix(verifiedDID,"did:")||len(verifiedDID)>256||len(d.Slug)>100||!draftSlug.MatchString(d.Slug)||d.Version<1||d.Version>1000000||strings.TrimSpace(d.Title)==""||len(d.Title)>160||len(d.Summary)>1000||len(d.License)>100||strings.TrimSpace(d.License)==""||len(d.Spec)==0||len(d.Spec)>32768||!json.Valid(d.Spec){return "",ErrInvalidDraft}
 var obj map[string]json.RawMessage
 if err:=json.Unmarshal(d.Spec,&obj);err!=nil||obj==nil{return "",ErrInvalidDraft}
 if err:=ValidateDraftSpec(d.Spec,d.License);err!=nil{return "",ErrInvalidDraft}
 id,err:=s.Tokens.RandomURLToken(18);if err!=nil{return "",err}
 _,err=s.DB.ExecContext(ctx,`INSERT INTO published_protocols(id,creator_did,slug,version,title,summary,license,spec_json,status,created_at) VALUES(?,?,?,?,?,?,?,?,'draft',?)`,id,verifiedDID,d.Slug,d.Version,strings.TrimSpace(d.Title),strings.TrimSpace(d.Summary),strings.TrimSpace(d.License),string(d.Spec),time.Now().UTC().Format(time.RFC3339Nano))
 if err!=nil{return "",ErrDraftConflict}
 return id,nil
}
