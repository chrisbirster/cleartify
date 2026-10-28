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

var ErrInvalid=errors.New("invalid follow-up")
var ErrNotFound=errors.New("follow-up not found")
var ErrRateLimit=errors.New("follow-up limit reached")
const MaxCreatedPerDay=20
const MaxNudgesPerDay=30
const MaxDeliveryAttempts=5
type Service struct{
 DB *sql.DB
 Sender notifications.Sender
 Now func()time.Time
 Channel string // "in_app" (default) or "sms"; no implicit SMS
}
func(s Service)clock()time.Time{if s.Now!=nil{return s.Now().UTC()};return time.Now().UTC()}
func(s Service)NowUTC()time.Time{return s.clock()}
func(s Service)channel()string{if s.Channel=="sms"{return "sms"};return "in_app"}
func(s Service)ChannelName()string{return s.channel()}
func(s Service)Ready(ctx context.Context)error{
 for _,ddl:=range []string{
 `CREATE TABLE IF NOT EXISTS clear_followups (
 id TEXT PRIMARY KEY,user_id TEXT NOT NULL,text TEXT NOT NULL,due_at TEXT NOT NULL,
 status TEXT NOT NULL CHECK(status IN ('open','done')),created_at TEXT NOT NULL,done_at TEXT)`,
 `CREATE TABLE IF NOT EXISTS clear_followup_nudges (
 followup_id TEXT NOT NULL,due_at TEXT NOT NULL,created_at TEXT NOT NULL,
 delivered_at TEXT,PRIMARY KEY(followup_id,due_at))`,
 `CREATE TABLE IF NOT EXISTS clear_followup_delivery (
 followup_id TEXT PRIMARY KEY,channel TEXT NOT NULL,status TEXT NOT NULL,
 attempts INTEGER NOT NULL DEFAULT 0,next_attempt_at TEXT NOT NULL,
 delivered_at TEXT,lease_until TEXT,last_error TEXT)`,
 `CREATE INDEX IF NOT EXISTS idx_clear_followups_user_date ON clear_followups(user_id,created_at)`,
 }{if _,err:=s.DB.ExecContext(ctx,ddl);err!=nil{return err}}
 return nil
}
type Followup struct{
 ID string `json:"id"`
 Text string `json:"text"`
 DueAt string `json:"dueAt"`
 Status string `json:"status"`
}
func(s Service)Create(ctx context.Context,user,text string,due time.Time)(Followup,error){
 text=strings.TrimSpace(text);now:=s.clock()
 if user==""||len(text)<1||len(text)>4096||due.IsZero()||due.After(now.AddDate(2,0,0)){return Followup{},ErrInvalid}
 tx,err:=s.DB.BeginTx(ctx,nil);if err!=nil{return Followup{},err};defer tx.Rollback()
 var active int
 if err=tx.QueryRowContext(ctx,"SELECT COUNT(*) FROM users WHERE id=? AND status='active'",user).Scan(&active);err!=nil{return Followup{},err}
 if active==0{return Followup{},ErrInvalid}
 var count int
 if err=tx.QueryRowContext(ctx,"SELECT COUNT(*) FROM clear_followups WHERE user_id=? AND created_at>=?",user,now.Add(-24*time.Hour).Format(time.RFC3339Nano)).Scan(&count);err!=nil{return Followup{},err}
 if count>=MaxCreatedPerDay{return Followup{},ErrRateLimit}
 var b [18]byte;if _,err=rand.Read(b[:]);err!=nil{return Followup{},err}
 f:=Followup{ID:hex.EncodeToString(b[:]),Text:text,DueAt:due.UTC().Format(time.RFC3339Nano),Status:"open"}
 _,err=tx.ExecContext(ctx,`INSERT INTO clear_followups(id,user_id,text,due_at,status,created_at) VALUES(?,?,?,?,'open',?)`,f.ID,user,f.Text,f.DueAt,now.Format(time.RFC3339Nano))
 if err!=nil{return Followup{},err}
 if err=tx.Commit();err!=nil{return Followup{},err};return f,nil
}
func(s Service)Done(ctx context.Context,user,id string)error{
 res,err:=s.DB.ExecContext(ctx,`UPDATE clear_followups SET status='done',done_at=? WHERE id=? AND user_id=? AND status='open'`,s.clock().Format(time.RFC3339Nano),id,user)
 if err!=nil{return err};n,err:=res.RowsAffected();if err!=nil{return err};if n==0{return ErrNotFound};return nil
}
func(s Service)List(ctx context.Context,user string)([]Followup,error){
 rows,err:=s.DB.QueryContext(ctx,`SELECT id,text,due_at,status FROM clear_followups WHERE user_id=? ORDER BY created_at,id`,user)
 if err!=nil{return nil,err};defer rows.Close()
 out:=[]Followup{}
 for rows.Next(){var f Followup;if err=rows.Scan(&f.ID,&f.Text,&f.DueAt,&f.Status);err!=nil{return nil,err};out=append(out,f)}
 return out,rows.Err()
}
func(s Service)Tick(ctx context.Context)(int,error){
 now:=s.clock().Format(time.RFC3339Nano)
 // Unique per (followup,due time), persistent across worker restarts.
 var before int
 if err:=s.DB.QueryRowContext(ctx,"SELECT COUNT(*) FROM clear_followup_nudges").Scan(&before);err!=nil{return 0,err}
 _,err:=s.DB.ExecContext(ctx,`INSERT OR IGNORE INTO clear_followup_nudges(followup_id,due_at,created_at)
 SELECT eligible.id,eligible.due_at,?
 FROM (
  SELECT f.id,f.user_id,f.due_at,
   ROW_NUMBER() OVER (PARTITION BY f.user_id ORDER BY f.due_at,f.id) AS position
  FROM clear_followups f
  WHERE f.status='open' AND f.due_at<=?
   AND NOT EXISTS(SELECT 1 FROM clear_followup_nudges existing WHERE existing.followup_id=f.id)
 ) eligible
 WHERE eligible.position <= ? - (
  SELECT COUNT(*) FROM clear_followup_nudges past JOIN clear_followups owner ON owner.id=past.followup_id
  WHERE owner.user_id=eligible.user_id AND past.created_at>=?
 )`,now,now,MaxNudgesPerDay,s.clock().Add(-24*time.Hour).Format(time.RFC3339Nano))
 if err!=nil{return 0,err}
 var after int
 if err:=s.DB.QueryRowContext(ctx,"SELECT COUNT(*) FROM clear_followup_nudges").Scan(&after);err!=nil{return 0,err}
 n:=after-before
 _,err=s.DB.ExecContext(ctx,`INSERT OR IGNORE INTO clear_followup_delivery(followup_id,channel,status,next_attempt_at)
 SELECT n.followup_id,?,'pending',? FROM clear_followup_nudges n`,s.channel(),now)
 return n,err
}
type Nudge struct{
 FollowupID string `json:"followupId"`
 Text string `json:"text"`
 DueAt string `json:"dueAt"`
 Channel string `json:"channel"`
 DeliveryStatus string `json:"deliveryStatus"`
}
func(s Service)Inbox(ctx context.Context,user string)([]Nudge,error){
 rows,err:=s.DB.QueryContext(ctx,`SELECT n.followup_id,f.text,n.due_at,COALESCE(d.channel,'in_app'),COALESCE(d.status,'pending') FROM clear_followup_nudges n
 JOIN clear_followups f ON f.id=n.followup_id LEFT JOIN clear_followup_delivery d ON d.followup_id=f.id
 WHERE f.user_id=? AND f.status='open' ORDER BY n.created_at,n.followup_id`,user)
 if err!=nil{return nil,err};defer rows.Close()
 out:=[]Nudge{}
 for rows.Next(){var n Nudge;if err=rows.Scan(&n.FollowupID,&n.Text,&n.DueAt,&n.Channel,&n.DeliveryStatus);err!=nil{return nil,err};out=append(out,n)}
 return out,rows.Err()
}
// DeliverDue claims a nudge with a persistent lease. A crash after external
// SendSMS succeeds but before the DB commit can cause duplicate SMS; external
// exactly-once delivery cannot be guaranteed without provider idempotency.
func(s Service)DeliverDue(ctx context.Context)(int,error){
 now:=s.clock()
 rows,err:=s.DB.QueryContext(ctx,`SELECT d.followup_id FROM clear_followup_delivery d JOIN clear_followups f ON f.id=d.followup_id
 WHERE f.status='open' AND d.status NOT IN ('delivered','dead') AND d.next_attempt_at<=?
 AND (d.lease_until IS NULL OR d.lease_until<=?) ORDER BY d.next_attempt_at LIMIT 30`,now.Format(time.RFC3339Nano),now.Format(time.RFC3339Nano))
 if err!=nil{return 0,err}
 ids:=[]string{};for rows.Next(){var id string;if err=rows.Scan(&id);err!=nil{rows.Close();return 0,err};ids=append(ids,id)}
 if err=rows.Err();err!=nil{rows.Close();return 0,err};rows.Close()
 delivered:=0
 for _,id:=range ids{
  lease:=now.Add(time.Minute).Format(time.RFC3339Nano)
  res,e:=s.DB.ExecContext(ctx,`UPDATE clear_followup_delivery SET status='leased',lease_until=? WHERE followup_id=? AND status NOT IN ('delivered','dead') AND next_attempt_at<=? AND (lease_until IS NULL OR lease_until<=?) AND EXISTS (SELECT 1 FROM clear_followups WHERE id=? AND status='open')`,lease,id,now.Format(time.RFC3339Nano),now.Format(time.RFC3339Nano),id)
  if e!=nil{return delivered,e};n,e:=res.RowsAffected();if e!=nil{return delivered,e};if n!=1{continue}
  var user,text,channel string;var attempts int
  e=s.DB.QueryRowContext(ctx,`SELECT f.user_id,f.text,d.channel,d.attempts FROM clear_followups f JOIN clear_followup_delivery d ON d.followup_id=f.id WHERE f.id=? AND f.status='open'`,id).Scan(&user,&text,&channel,&attempts)
  if e!=nil{continue}
  if channel=="in_app"{
   // In-app inbox exists durably, so its delivery acknowledgment is transactional.
   _,e=s.DB.ExecContext(ctx,`UPDATE clear_followup_delivery SET status='delivered',delivered_at=?,lease_until=NULL WHERE followup_id=? AND lease_until=? AND EXISTS(SELECT 1 FROM clear_followups WHERE id=? AND status='open')`,now.Format(time.RFC3339Nano),id,lease,id)
  }else if channel=="sms"{
   var phone string
   e=s.DB.QueryRowContext(ctx,`SELECT phone_e164 FROM phone_identities WHERE user_id=? AND verified_at IS NOT NULL LIMIT 1`,user).Scan(&phone)
   if e==nil&&s.Sender!=nil{e=s.Sender.SendSMS(ctx,notifications.SMS{To:phone,Text:"Cleartify follow-up: "+text})}else if e==nil{e=notifications.ErrUnavailable}
   if e==nil{_,e=s.DB.ExecContext(ctx,`UPDATE clear_followup_delivery SET status='delivered',delivered_at=?,lease_until=NULL WHERE followup_id=? AND lease_until=?`,now.Format(time.RFC3339Nano),id,lease)}
  }else{e=ErrInvalid}
  if e!=nil{
   delay:=time.Minute*time.Duration(1<<min(attempts,6))
   _,updateErr:=s.DB.ExecContext(ctx,`UPDATE clear_followup_delivery SET status=CASE WHEN attempts+1>=? THEN 'dead' ELSE 'retry' END,attempts=attempts+1,next_attempt_at=?,lease_until=NULL,last_error='delivery_failed' WHERE followup_id=? AND lease_until=?`,MaxDeliveryAttempts,now.Add(delay).Format(time.RFC3339Nano),id,lease)
   if updateErr!=nil{return delivered,updateErr}
   continue
  }
  delivered++
 }
 return delivered,nil
}
// Run handles ticks on injected clock signals. Production can supply a ticker;
// tests inject a channel and cancel the context for graceful shutdown.
func(s Service)Run(ctx context.Context,ticks <-chan time.Time)error{
 for{
  select{
  case <-ctx.Done():return nil
  case _,ok:=<-ticks:
   if !ok{return nil}
   if _,err:=s.Tick(ctx);err!=nil{if ctx.Err()!=nil{return nil};return err}
   if _,err:=s.DeliverDue(ctx);err!=nil{if ctx.Err()!=nil{return nil};return err}
  }
 }
}
