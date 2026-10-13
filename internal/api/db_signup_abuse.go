package api
import(
 "crypto/hmac"
 "errors"
 "crypto/sha256"
 "database/sql"
 "encoding/hex"
 "net"
 "strings"
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
 ip,err:=s.signupPeer(r)
 if err!=nil{writeJSON(w,503,map[string]string{"error":"signup_guard_unavailable"});return false}
 mac:=hmac.New(sha256.New,[]byte(s.dbAbuseSecret))
 mac.Write([]byte("signup-v1:"+ip))
 key:=hex.EncodeToString(mac.Sum(nil))
 now:=time.Now().UTC()
 bucket:=now.Unix()/900
 tx,err:=s.dbAbuseDB.BeginTx(r.Context(),nil)
 if err!=nil{writeJSON(w,503,map[string]string{"error":"signup_guard_unavailable"});return false}
 defer tx.Rollback()
 if s.dbAbuseCleanupEnabled {
  // Conservative 48h retention, on the existing transactional write path.
  if _,err=tx.ExecContext(r.Context(),`DELETE FROM signup_abuse_windows WHERE bucket < ?`,bucket-192);err!=nil{writeJSON(w,503,map[string]string{"error":"signup_guard_unavailable"});return false}
 }
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

// signupPeer never trusts an unrecognized upstream. Trusted CIDRs must be
// configured explicitly; XFF is evaluated right-to-left, stopping at the
// first non-trusted hop so a client cannot spoof earlier headers.
func(s *Server) signupPeer(r *http.Request)(string,error){
 host,_,err:=net.SplitHostPort(r.RemoteAddr)
 if err!=nil{return "",err}
 peer:=net.ParseIP(host)
 if peer==nil{return "",ErrPeerInvalid}
 if !s.dbAbuseProxyEnabled {return peer.String(),nil}
 trusted:=func(ip net.IP)bool{for _,network:=range s.dbAbuseTrustedProxyCIDRs{if network.Contains(ip){return true}};return false}
 if !trusted(peer){return peer.String(),nil}
 forwarded:=r.Header.Values("X-Forwarded-For")
 if len(forwarded)!=1||len(forwarded[0])>1024{return "",ErrPeerInvalid}
 hops:=strings.Split(forwarded[0],",")
 if len(hops)==0||len(hops)>16{return "",ErrPeerInvalid}
 for i:=len(hops)-1;i>=0;i--{
  ip:=net.ParseIP(strings.TrimSpace(hops[i]))
  if ip==nil{return "",ErrPeerInvalid}
  peer=ip
  if !trusted(peer){return peer.String(),nil}
 }
 return peer.String(),nil
}
var ErrPeerInvalid=errors.New("invalid proxy peer chain")
