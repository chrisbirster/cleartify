package protocols

import (
 "context"
 "crypto/rand"
 "crypto/sha256"
 "database/sql"
 "encoding/hex"
 "encoding/json"
 "errors"
 "fmt"
 "strings"
 "time"
 "github.com/chrisbirster/cleartify/internal/database"
)
var ErrInstallUnavailable=errors.New("published protocol or valid permission manifest unavailable")
var ErrManifestChanged=errors.New("permission manifest digest mismatch")
var ErrInstallDecision=errors.New("install request is not pending or publication changed")
type InstallService struct {Store *database.Store; StrictPermissions bool}
type Manifest struct {
 PublicationID string `json:"publicationId"`
 Title string `json:"title"`
 Version int `json:"version"`
 Permissions []string `json:"permissions"`
 SpecSHA256 string `json:"specSha256"`
 ManifestSHA256 string `json:"manifestSha256"`
}
type installSnapshot struct{ID string;Manifest Manifest;Raw string}
func sha(raw []byte)string{digest:=sha256.Sum256(raw);return hex.EncodeToString(digest[:])}
func(s InstallService) manifest(ctx context.Context,id string)(installSnapshot,error){
 var x installSnapshot
 if id==""||len(id)>128{return x,ErrInstallUnavailable}
 var spec,title,slug string;var version int
 err:=s.Store.DB.QueryRowContext(ctx,`SELECT title,slug,version,spec_json FROM published_protocols WHERE id=? AND status='published'`,id).Scan(&title,&slug,&version,&spec)
 if errors.Is(err,sql.ErrNoRows){return x,ErrInstallUnavailable};if err!=nil{return x,err}
 if len(spec)>65536||version<=0||strings.TrimSpace(slug)==""{return x,ErrInstallUnavailable}
 if s.StrictPermissions && ValidatePreviewPermissions([]byte(spec))!=nil {return x,ErrInstallUnavailable}
 var raw struct{Permissions []string `json:"permissions"`}
 if err=json.Unmarshal([]byte(spec),&raw);err!=nil||raw.Permissions==nil||len(raw.Permissions)>32{return x,ErrInstallUnavailable}
 seen:=map[string]bool{}
 for _,p:=range raw.Permissions{if p==""||len(p)>100||seen[p]||strings.TrimSpace(p)!=p{return x,ErrInstallUnavailable};seen[p]=true}
 if raw.Permissions==nil{raw.Permissions=[]string{}}
 x.ID=id;x.Manifest=Manifest{PublicationID:id,Title:title,Version:version,Permissions:raw.Permissions,SpecSHA256:sha([]byte(spec))}
 // Canonical snapshot: fixed struct field order, including immutable spec digest.
 stable,err:=json.Marshal(struct{PublicationID string `json:"publicationId"`;Version int `json:"version"`;Permissions []string `json:"permissions"`;SpecSHA256 string `json:"specSha256"`}{id,version,raw.Permissions,x.Manifest.SpecSHA256});if err!=nil{return x,err}
 x.Manifest.ManifestSHA256=sha(stable)
 permissions,err:=json.Marshal(raw.Permissions);if err!=nil{return x,err}
 x.Raw=string(permissions)
 return x,nil
}
func(s InstallService) Preview(ctx context.Context,publicationID string)(Manifest,error){
 x,err:=s.manifest(ctx,publicationID);return x.Manifest,err
}
func(s InstallService) Request(ctx context.Context,actorDID,publicationID,ackDigest string)(string,error){
 if !strings.HasPrefix(actorDID,"did:")||len(actorDID)>2048{return "",ErrInstallUnavailable}
 x,err:=s.manifest(ctx,publicationID);if err!=nil{return "",err}
 if ackDigest!=x.Manifest.ManifestSHA256{return "",ErrManifestChanged}
 idBytes:=make([]byte,18);if _,err=rand.Read(idBytes);err!=nil{return "",err}
 id:=hex.EncodeToString(idBytes);now:=time.Now().UTC().Format(time.RFC3339Nano)
 _,err=s.Store.DB.ExecContext(ctx,`INSERT INTO protocol_install_requests(id,actor_did,publication_id,protocol_version,spec_sha256,manifest_json,manifest_sha256,acknowledged_at,status,requested_at)
 VALUES(?,?,?,?,?,?,?,?,'pending_approval',?)`,id,actorDID,publicationID,x.Manifest.Version,x.Manifest.SpecSHA256,x.Raw,x.Manifest.ManifestSHA256,now,now)
 if err!=nil{return "",err}
 if err=s.Store.Push(ctx);err!=nil{return "",err}
 return id,nil
}
func(s InstallService) Decide(ctx context.Context,id string,approve bool)error{
 if len(id)!=36{return ErrInstallDecision}
 var publicationID,storedSHA string;var version int
 err:=s.Store.DB.QueryRowContext(ctx,`SELECT publication_id,protocol_version,spec_sha256 FROM protocol_install_requests WHERE id=? AND status='pending_approval'`,id).Scan(&publicationID,&version,&storedSHA)
 if errors.Is(err,sql.ErrNoRows){return ErrInstallDecision};if err!=nil{return err}
 m,err:=s.manifest(ctx,publicationID);if errors.Is(err,ErrInstallUnavailable){return ErrInstallDecision};if err!=nil{return err}
 if m.Manifest.Version!=version||m.Manifest.SpecSHA256!=storedSHA{return ErrInstallDecision}
 status:="rejected";if approve{status="approved"}
 res,err:=s.Store.DB.ExecContext(ctx,`UPDATE protocol_install_requests SET status=?,decided_at=? WHERE id=? AND status='pending_approval' AND protocol_version=? AND spec_sha256=?`,status,time.Now().UTC().Format(time.RFC3339Nano),id,version,storedSHA)
 if err!=nil{return fmt.Errorf("decide install: %w",err)}
 n,err:=res.RowsAffected();if err!=nil{return err};if n!=1{return ErrInstallDecision}
 return s.Store.Push(ctx)
}
