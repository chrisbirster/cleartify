package followups

import (
 "context"
 "crypto/rand"
 "database/sql"
 "encoding/hex"
 "errors"
 "strings"
 "time"
 "github.com/chrisbirster/cleartify/internal/notifications"
)

var ErrInvalid = errors.New("invalid follow-up")
var ErrNotFound = errors.New("follow-up not found")
type Service struct{
 DB *sql.DB
 Sender notifications.Sender
 Now func()time.Time
}
func(s Service)clock()time.Time{if s.Now!=nil{return s.Now().UTC()};return time.Now().UTC()}
func(s Service)Ready(ctx context.Context)error{
 // Explicitly initialized only when feature flag is enabled.
 for _,ddl:=range []string{
 `CREATE TABLE IF NOT EXISTS clear_followups (
   id TEXT PRIMARY KEY,user_id TEXT NOT NULL,text TEXT NOT NULL,due_at TEXT NOT NULL,
   status TEXT NOT NULL CHECK(status IN ('open','done')),created_at TEXT NOT NULL,done_at TEXT
 )`,
 `CREATE TABLE IF NOT EXISTS clear_followup_nudges (
   followup_id TEXT NOT NULL,due_at TEXT NOT NULL,created_at TEXT NOT NULL,
   delivered_at TEXT,PRIMARY KEY(followup_id,due_at)
 )`,
 }{
  if _,err:=s.DB.ExecContext(ctx,ddl);err!=nil{return err}
 }
 return nil
}
type Followup struct{
 ID string `json:"id"`
 Text string `json:"text"`
 DueAt string `json:"dueAt"`
 Status string `json:"status"`
}
func(s Service)Create(ctx context.Context,user,text string,due time.Time)(Followup,error){
 text=strings.TrimSpace(text)
 if user==""||len(text)<1||len(text)>4096||due.IsZero(){return Followup{},ErrInvalid}
 var b [18]byte;if _,err:=rand.Read(b[:]);err!=nil{return Followup{},err}
 f:=Followup{ID:hex.EncodeToString(b[:]),Text:text,DueAt:due.UTC().Format(time.RFC3339Nano),Status:"open"}
 _,err:=s.DB.ExecContext(ctx,`INSERT INTO clear_followups(id,user_id,text,due_at,status,created_at)
 SELECT ?,?,?,?,'open',? WHERE EXISTS(SELECT 1 FROM users WHERE id=? AND status='active')`,f.ID,user,f.Text,f.DueAt,s.clock().Format(time.RFC3339Nano),user)
 if err!=nil{return Followup{},err}
 return f,nil
}
func(s Service)Done(ctx context.Context,user,id string)error{
 res,err:=s.DB.ExecContext(ctx,`UPDATE clear_followups SET status='done',done_at=? WHERE id=? AND user_id=? AND status='open'`,s.clock().Format(time.RFC3339Nano),id,user)
 if err!=nil{return err}
 n,err:=res.RowsAffected();if err!=nil{return err};if n==0{return ErrNotFound};return nil
}
func(s Service)List(ctx context.Context,user string)([]Followup,error){
 rows,err:=s.DB.QueryContext(ctx,`SELECT id,text,due_at,status FROM clear_followups WHERE user_id=? ORDER BY created_at,id`,user)
 if err!=nil{return nil,err};defer rows.Close()
 out:=[]Followup{}
 for rows.Next(){var f Followup;if err=rows.Scan(&f.ID,&f.Text,&f.DueAt,&f.Status);err!=nil{return nil,err};out=append(out,f)}
 return out,rows.Err()
}
// Tick atomically records a unique nudge in a durable in-app outbox.
// Returning a nudge is the test/in-app delivery path. An optional Sender is
// used only when the caller explicitly requests SMS delivery elsewhere.
func(s Service)Tick(ctx context.Context)(int,error){
 now:=s.clock().Format(time.RFC3339Nano)
 res,err:=s.DB.ExecContext(ctx,`INSERT OR IGNORE INTO clear_followup_nudges(followup_id,due_at,created_at)
 SELECT id,due_at,? FROM clear_followups WHERE status='open' AND due_at<=?`,now,now)
 if err!=nil{return 0,err};n,err:=res.RowsAffected();return int(n),err
}
type Nudge struct{
 FollowupID string `json:"followupId"`
 Text string `json:"text"`
 DueAt string `json:"dueAt"`
}
func(s Service)Inbox(ctx context.Context,user string)([]Nudge,error){
 rows,err:=s.DB.QueryContext(ctx,`SELECT n.followup_id,f.text,n.due_at FROM clear_followup_nudges n
 JOIN clear_followups f ON f.id=n.followup_id WHERE f.user_id=? AND f.status='open' ORDER BY n.created_at,n.followup_id`,user)
 if err!=nil{return nil,err};defer rows.Close()
 out:=[]Nudge{}
 for rows.Next(){var n Nudge;if err=rows.Scan(&n.FollowupID,&n.Text,&n.DueAt);err!=nil{return nil,err};out=append(out,n)}
 return out,rows.Err()
}
