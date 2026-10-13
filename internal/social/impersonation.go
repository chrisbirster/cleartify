package social

import (
 "context"
 "github.com/chrisbirster/cleartify/internal/policy"
 "database/sql"
 "errors"
 "strings"
 "time"
)

var ErrImpersonationReview=errors.New("impersonation review denied")
const DefaultReportThreshold=policy.ImpersonationDistinctReportHoldThreshold

// DEFAULT (Oct 13 2026), owner can change. ASCII-only conservative
// candidate screening; NOT an AT Protocol handle/PDS attestation.
var reservedHandles=map[string]bool{
 "admin":true,"support":true,"security":true,"official":true,
 "cleartify":true,"moderator":true,"staff":true,
}
func handleSkeleton(handle string)string{
 h:=strings.ToLower(strings.TrimSpace(handle))
 var b strings.Builder
 for _,c:=range h{
  switch c {case '0':c='o';case '1':c='i';case '3':c='e';case '5':c='s';case '-','_','.':continue}
  b.WriteRune(c)
 }
 return b.String()
}
// ScreenHandle rejects direct or ASCII-leetspeak near matches to reserved
// platform identities. False positives require manual owner review.
func ScreenHandle(handle string)(string,bool){
 h:=strings.ToLower(strings.TrimSpace(handle))
 if h==""||len(h)>253||!handleRx.MatchString(h){return "invalid",false}
 for _,label:=range strings.Split(h,"."){
  skeleton:=handleSkeleton(label)
  if reservedHandles[label]{return "reserved",false}
  if reservedHandles[skeleton]{return "lookalike",false}
 }
 return "",true
}
type IdentityLabel struct{
 Verification string `json:"verification"`
 OnHold bool `json:"onHold"`
 HoldReason string `json:"holdReason,omitempty"`
}
// Impersonation is explicitly OFF until future reviewed integration.
// It reads local claimed state, never independently verifies ATProto DIDs.
type Impersonation struct{DB *sql.DB;Enabled bool;Now func()time.Time;ReportThreshold int}
func(s Impersonation)clock()string{if s.Now!=nil{return s.Now().UTC().Format(time.RFC3339Nano)};return time.Now().UTC().Format(time.RFC3339Nano)}
func(s Impersonation)threshold()int{if s.ReportThreshold>0&&s.ReportThreshold<=10{return s.ReportThreshold};return DefaultReportThreshold}
func(s Impersonation)Label(ctx context.Context,did string)(IdentityLabel,error){
 label:=IdentityLabel{Verification:"unverified"}
 if !s.Enabled||s.DB==nil{return label,nil}
 var claimed int;var kind string
 err:=s.DB.QueryRowContext(ctx,`SELECT claimed,kind FROM social_profiles WHERE did=?`,did).Scan(&claimed,&kind)
 if errors.Is(err,sql.ErrNoRows){return label,ErrNotFound};if err!=nil{return label,err}
 if claimed==1&&kind=="person"{label.Verification="locally_claimed_not_independently_verified"}
 var reason string
 err=s.DB.QueryRowContext(ctx,`SELECT reason FROM social_impersonation_holds WHERE subject_did=? AND released_at IS NULL`,did).Scan(&reason)
 if err==nil{label.OnHold=true;label.HoldReason=reason;label.Verification="unverified"}
 if err!=nil&&!errors.Is(err,sql.ErrNoRows){return label,err}
 return label,nil
}
// EvaluateRecords applies a reversible local hold only for reserved/lookalike
// handles or at least N DISTINCT reporters alleging impersonation.
// Holding does not change account ownership or publish external moderation.
func(s Impersonation)EvaluateRecords(ctx context.Context,did string)(IdentityLabel,error){
 if !s.Enabled||s.DB==nil{return IdentityLabel{},ErrImpersonationReview}
 var handle string
 err:=s.DB.QueryRowContext(ctx,`SELECT handle FROM social_profiles WHERE did=?`,did).Scan(&handle)
 if errors.Is(err,sql.ErrNoRows){return IdentityLabel{},ErrNotFound};if err!=nil{return IdentityLabel{},err}
 reason,allowed:=ScreenHandle(handle)
 if allowed {
  var reports int
  err=s.DB.QueryRowContext(ctx,`SELECT count(DISTINCT reporter_did) FROM social_reports WHERE subject_did=? AND reason='impersonation'`,did).Scan(&reports)
  if err!=nil{return IdentityLabel{},err}
  if reports>=s.threshold(){reason="reports"}
 }
 if reason!="" {
  _,err=s.DB.ExecContext(ctx,`INSERT INTO social_impersonation_holds(subject_did,reason,created_at,released_at) VALUES(?,?,?,NULL) ON CONFLICT(subject_did) DO UPDATE SET reason=excluded.reason,created_at=excluded.created_at,released_at=NULL`,did,reason,s.clock())
  if err!=nil{return IdentityLabel{},err}
 }
 return s.Label(ctx,did)
}
// ReleaseHold requires a trusted moderator caller in a future authenticated API.
// This method intentionally is not exposed through any API route.
func(s Impersonation)ReleaseHold(ctx context.Context,did string)error{
 if !s.Enabled||s.DB==nil{return ErrImpersonationReview}
 res,err:=s.DB.ExecContext(ctx,`UPDATE social_impersonation_holds SET released_at=? WHERE subject_did=? AND released_at IS NULL`,s.clock(),did)
 if err!=nil{return err};n,err:=res.RowsAffected();if err!=nil||n!=1{return ErrImpersonationReview}
 return nil
}
