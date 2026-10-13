package api
import(
 "crypto/hmac"
 "crypto/sha256"
 "database/sql"
 "encoding/hex"
 "net"
 "net/http"
 "strconv"
 "time"
)
// allowDBSignup limits by peer IP from RemoteAddr, ignoring client-supplied
// forwarding headers. Each 15-minute bucket is persisted transactionally.
// A misconfigured enabled limiter denies signups rather than silently bypassing.
func(s *Server) allowDBSignup(w http.ResponseWriter,r *http.Request)bool{
 if !s.dbAbuseEnabled{return true}
 if s.dbAbuseDB==nil||len(s.dbAbuseSecret)<32{writeJSON(w,503,map[string]string{"error":"signup_guard_unavailable"});return false}
 ip,_,err:=net.SplitHostPort(r.RemoteAddr)
 if err!=nil||net.ParseIP(ip)==nil{writeJSON(w,503,map[string]string{"error":"signup_guard_unavailable"});return false}
 mac:=hmac.New(sha256.New,[]byte(s.dbAbuseSecret))
 mac.Write([]byte("signup-v1:"+ip))
 key:=hex.EncodeToString(mac.Sum(nil))
 bucket:=time.Now().UTC().Unix()/900
 tx,err:=s.dbAbuseDB.BeginTx(r.Context(),nil)
 if err!=nil{writeJSON(w,503,map[string]string{"error":"signup_guard_unavailable"});return false}
 defer tx.Rollback()
 _,err=tx.ExecContext(r.Context(),`INSERT INTO signup_abuse_windows(identity_hash,bucket,attempts) VALUES(?,?,1) ON CONFLICT(identity_hash,bucket) DO UPDATE SET attempts=attempts+1`,key,bucket)
 if err!=nil{writeJSON(w,503,map[string]string{"error":"signup_guard_unavailable"});return false}
 var attempts int
 err=tx.QueryRowContext(r.Context(),`SELECT attempts FROM signup_abuse_windows WHERE identity_hash=? AND bucket=?`,key,bucket).Scan(&attempts)
 if err!=nil&&err!=sql.ErrNoRows{writeJSON(w,503,map[string]string{"error":"signup_guard_unavailable"});return false}
 if err!=nil||tx.Commit()!=nil{writeJSON(w,503,map[string]string{"error":"signup_guard_unavailable"});return false}
 if attempts>5{
  w.Header().Set("Retry-After",strconv.FormatInt(900-time.Now().UTC().Unix()%900,10))
  writeJSON(w,429,map[string]string{"error":"signup_rate_limited"});return false
 }
 return true
}
