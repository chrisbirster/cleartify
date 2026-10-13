package protocols

import (
 "context"
 "database/sql"
 "errors"
 "fmt"
 "strings"
)

type Published struct {
 ID string `json:"id"`
 CreatorDID string `json:"creatorDid"`
 Slug string `json:"slug"`
 Version int `json:"version"`
 Title string `json:"title"`
 Summary string `json:"summary"`
 License string `json:"license"`
}
type Reader struct {DB *sql.DB;PaidEnforcement bool}
func (r Reader) Discover(ctx context.Context,limit int)([]Published,error){
 if limit<=0||limit>50{limit=20}
 query:=`SELECT id,creator_did,slug,version,title,summary,license FROM published_protocols WHERE status='published' AND NOT EXISTS (SELECT 1 FROM creator_publications c WHERE c.protocol_id=published_protocols.id AND c.visibility!='public') ORDER BY created_at DESC,id DESC LIMIT ?`
 if r.PaidEnforcement {query=`SELECT id,creator_did,slug,version,title,summary,license FROM published_protocols WHERE status='published' AND NOT EXISTS (SELECT 1 FROM creator_publications c WHERE c.protocol_id=published_protocols.id AND c.visibility!='public') AND NOT EXISTS (SELECT 1 FROM marketplace_listings l WHERE l.protocol_id=published_protocols.id AND l.status='listed' AND l.pricing_kind='one_time' AND l.price_minor>0) ORDER BY created_at DESC,id DESC LIMIT ?`}
 rows,err:=r.DB.QueryContext(ctx,query,limit)
 if err!=nil{return nil,fmt.Errorf("list protocols: %w",err)}
 defer rows.Close()
 out:=make([]Published,0)
 for rows.Next(){var p Published;if err:=rows.Scan(&p.ID,&p.CreatorDID,&p.Slug,&p.Version,&p.Title,&p.Summary,&p.License);err!=nil{return nil,err};out=append(out,p)}
 return out,rows.Err()
}
var ErrNotFound=errors.New("published protocol not found")
func(r Reader) Version(ctx context.Context,creatorDID,slug string,version int)(Published,error){
 var p Published
 if !strings.HasPrefix(creatorDID,"did:")||slug==""||len(slug)>100||version<1{return p,ErrNotFound}
 err:=r.DB.QueryRowContext(ctx,`SELECT id,creator_did,slug,version,title,summary,license FROM published_protocols WHERE creator_did=? AND slug=? AND version=? AND status='published' AND NOT EXISTS (SELECT 1 FROM creator_publications c WHERE c.protocol_id=published_protocols.id AND c.visibility='private')`,creatorDID,slug,version).Scan(&p.ID,&p.CreatorDID,&p.Slug,&p.Version,&p.Title,&p.Summary,&p.License)
 if errors.Is(err,sql.ErrNoRows){return Published{},ErrNotFound}
 return p,err
}
