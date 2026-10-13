package api

import (
 "context"
 "crypto/rand"
 "crypto/sha256"
 "encoding/hex"
 "database/sql"
 "errors"
 "net/http"
 "strings"
 "time"

 "github.com/chrisbirster/cleartify/internal/identitylink"
)

const linkConsentVersion = "internal-m2-v1"
const linkCookie = "cleartify_link_challenge"
const linkTTL = 10*time.Minute

// SocialDIDVerifier is supplied by Indigo OAuth in production and fake in tests.
// It must never trust caller-provided identity fields.
type SocialDIDVerifier interface { SocialSessionDID(*http.Request)(string,bool) }

func challengeToken()(string,string,error){
 b:=make([]byte,32);if _,err:=rand.Read(b);err!=nil{return "","",err}
 token:=hex.EncodeToString(b);return token,challengeHash(token),nil
}
func challengeHash(token string)string{sum:=sha256.Sum256([]byte(token));return hex.EncodeToString(sum[:])}
func(s *Server) linkCookieSet(w http.ResponseWriter,r *http.Request,value string,seconds int){
 http.SetCookie(w,&http.Cookie{Name:linkCookie,Value:value,Path:"/api/me/identity-link",HttpOnly:true,Secure:strings.HasPrefix(s.appOrigin,"https://"),SameSite:http.SameSiteStrictMode,MaxAge:seconds})
}
func(s *Server) linkPrincipals(w http.ResponseWriter,r *http.Request)(string,string,bool){
 if s.clearSessions==nil||s.linkSocial==nil||s.identityLinks==nil {writeJSON(w,503,map[string]string{"error":"link_unavailable"});return "","",false}
 c,err:=r.Cookie("clear_session");if err!=nil{writeJSON(w,401,map[string]string{"error":"unauthenticated"});return "","",false}
 // Require a recently issued Clear session (after phone verification), not
 // merely a still-active 30-day sliding session.
 user,err:=s.clearSessions.AuthenticateFresh(r.Context(),c.Value,linkTTL)
 if err!=nil{writeJSON(w,401,map[string]string{"error":"clear_reauthentication_required"});return "","",false}
 did,ok:=s.linkSocial.SocialSessionDID(r)
 if !ok||!strings.HasPrefix(did,"did:"){writeJSON(w,401,map[string]string{"error":"social_reauthentication_required"});return "","",false}
 return user,did,true
}
func(s *Server) linkStart(w http.ResponseWriter,r *http.Request){
 user,did,ok:=s.linkPrincipals(w,r);if !ok{return}
 raw,digest,err:=challengeToken();if err!=nil{writeJSON(w,503,map[string]string{"error":"link_unavailable"});return}
 now:=time.Now().UTC()
 _,err=s.identityLinks.DB.ExecContext(r.Context(),`INSERT INTO identity_link_challenges(nonce_hash,clear_user_id,social_did,state,consent_version,expires_at) VALUES(?,?,?,'pending',?,?)`,digest,user,did,linkConsentVersion,now.Add(linkTTL).Format(time.RFC3339Nano))
 if err!=nil{writeJSON(w,503,map[string]string{"error":"link_unavailable"});return}
 if err=s.linkPush(r);err!=nil{writeJSON(w,503,map[string]string{"error":"link_unavailable"});return}
 s.linkCookieSet(w,r,raw,int(linkTTL.Seconds()))
 w.Header().Set("Cache-Control","no-store")
 writeJSON(w,201,map[string]any{"pending":true,"expiresInSeconds":int(linkTTL.Seconds()),"consentVersion":linkConsentVersion})
}
func(s *Server) linkChallenge(w http.ResponseWriter,r *http.Request)(string,string,string,bool){
 user,did,ok:=s.linkPrincipals(w,r);if !ok{return "","","",false}
 c,err:=r.Cookie(linkCookie);if err!=nil||len(c.Value)!=64 {writeJSON(w,401,map[string]string{"error":"challenge_missing"});return "","","",false}
 var u,d,state,expiry,consent string
 err=s.identityLinks.DB.QueryRowContext(r.Context(),`SELECT clear_user_id,social_did,state,expires_at,consent_version FROM identity_link_challenges WHERE nonce_hash=?`,challengeHash(c.Value)).Scan(&u,&d,&state,&expiry,&consent)
 deadline,parseErr:=time.Parse(time.RFC3339Nano,expiry)
 if err!=nil||parseErr!=nil||state!="pending"||u!=user||d!=did||consent!=linkConsentVersion||!time.Now().UTC().Before(deadline){
  writeJSON(w,409,map[string]string{"error":"challenge_invalid"});return "","","",false
 }
 return user,did,challengeHash(c.Value),true
}
func(s *Server) linkPreview(w http.ResponseWriter,r *http.Request){
 user,did,_,ok:=s.linkChallenge(w,r);if !ok{return}
 w.Header().Set("Cache-Control","no-store")
 writeJSON(w,200,map[string]any{"clearUserId":user,"socialDID":did,"consentVersion":linkConsentVersion,"privateDataPublished":false,"linkGrantsClearAccess":false,"requiresExplicitConsent":true})
}
func(s *Server) linkConfirm(w http.ResponseWriter,r *http.Request){
 user,did,nonce,ok:=s.linkChallenge(w,r);if !ok{return}
 var input struct{Consent bool `json:"consent"`;ConsentVersion string `json:"consentVersion"`}
 if !decodeJSON(w,r,&input){return}
 if !input.Consent||input.ConsentVersion!=linkConsentVersion{writeJSON(w,400,map[string]string{"error":"consent_required"});return}
 id,_,err:=challengeToken();if err!=nil{writeJSON(w,503,map[string]string{"error":"link_unavailable"});return}
 event,_,err:=challengeToken();if err!=nil{writeJSON(w,503,map[string]string{"error":"link_unavailable"});return}
 now:=time.Now().UTC()
 // Fresh principals were independently revalidated above; the DB store
 // atomically consumes the challenge and enforces one-to-one ownership.
 err= (identitylink.Store{DB:s.identityLinks.DB}).Commit(r.Context(),identitylink.CommitRequest{
  ID:id,NonceHash:nonce,ClearUserID:user,SocialDID:did,ConsentVersion:linkConsentVersion,
  ClearAuthenticatedAt:now,SocialAuthenticatedAt:now,Now:now,EventID:event,
 })
 if errors.Is(err,identitylink.ErrConflict){writeJSON(w,409,map[string]string{"error":"link_conflict"});return}
 if err!=nil{writeJSON(w,409,map[string]string{"error":"link_unavailable"});return}
 if err=s.linkPush(r);err!=nil{writeJSON(w,503,map[string]string{"error":"link_unavailable"});return}
 s.linkCookieSet(w,r,"",-1)
 w.Header().Set("Cache-Control","no-store")
 writeJSON(w,200,map[string]any{"linked":true,"privateDataPublished":false})
}
func(s *Server) linkUnlink(w http.ResponseWriter,r *http.Request){
 user,did,ok:=s.linkPrincipals(w,r);if !ok{return}
 var input struct{Confirm bool `json:"confirm"`}
 if !decodeJSON(w,r,&input){return}
 if !input.Confirm{writeJSON(w,400,map[string]string{"error":"confirmation_required"});return}
 event,_,err:=challengeToken();if err!=nil{writeJSON(w,503,map[string]string{"error":"link_unavailable"});return}
 tx,err:=s.identityLinks.DB.BeginTx(r.Context(),nil);if err!=nil{writeJSON(w,503,map[string]string{"error":"link_unavailable"});return};defer tx.Rollback()
 now:=time.Now().UTC().Format(time.RFC3339Nano)
 res,err:=tx.ExecContext(r.Context(),`UPDATE identity_links SET status='revoked',revoked_at=? WHERE clear_user_id=? AND social_did=? AND status='active'`,now,user,did)
 if err!=nil{writeJSON(w,503,map[string]string{"error":"link_unavailable"});return}
 n,err:=res.RowsAffected();if err!=nil||n!=1{writeJSON(w,409,map[string]string{"error":"link_not_active"});return}
 _,err=tx.ExecContext(r.Context(),`UPDATE identity_link_challenges SET state='cancelled' WHERE clear_user_id=? AND state='pending'`,user)
 if err==nil{_,err=tx.ExecContext(r.Context(),`INSERT INTO identity_link_events(id,clear_user_id,social_did,action,recorded_at) VALUES(?,?,?,'unlinked',?)`,event,user,did,now)}
 if err!=nil||tx.Commit()!=nil{writeJSON(w,503,map[string]string{"error":"link_unavailable"});return}
 if err=s.linkPush(r);err!=nil{writeJSON(w,503,map[string]string{"error":"link_unavailable"});return}
 s.linkCookieSet(w,r,"",-1)
 w.Header().Set("Cache-Control","no-store")
 writeJSON(w,200,map[string]any{"linked":false})
}
func(s *Server) linkPush(r *http.Request)error{if s.identityLinks.Push==nil{return nil};return s.identityLinks.Push(r.Context())}

// LinkPersistence exposes storage and its durable push without granting API callers SQL access.
type LinkPersistence struct{DB *sql.DB;Push func(context.Context)error}
