package api

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"log/slog"
	"regexp"
	"net/http"
	"strings"
	"time"

	"github.com/chrisbirster/cleartify/internal/beta"
	"github.com/chrisbirster/cleartify/internal/agents"
	"github.com/chrisbirster/cleartify/internal/clear"
	"github.com/chrisbirster/cleartify/internal/atproto"
	"github.com/chrisbirster/cleartify/internal/notifications"
	"github.com/chrisbirster/cleartify/internal/protocols"
	"github.com/chrisbirster/cleartify/internal/social"
	"github.com/chrisbirster/cleartify/internal/waitlist"
)

var referralCodePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,40}$`)

type Options struct {
	Waitlist   *waitlist.Service
	Beta       *beta.Service
	AppOrigin  string
	AdminToken string
	AgentRegistrationEnabled bool
	Agents *agents.Service
	Clear *clear.Service
	OAuth *atproto.OAuthHandler
	Protocols *protocols.Reader
	Profiles *social.Reader
	SocialAccountRegistrationEnabled bool
	SocialAccounts *social.Accounts
	SocialProfileActionsEnabled bool
	SocialActions *social.Actions
	DefaultFollowEnabled bool
	DefaultFollow *social.DefaultFollow
}

type Server struct {
	waitlist   *waitlist.Service
	beta       *beta.Service
	appOrigin  string
	adminToken string
	agents *agents.Service
	clear *clear.Service
	protocols *protocols.Reader
	profiles *social.Reader
	socialAccounts *social.Accounts
	socialActions *social.Actions
	socialAuth *atproto.OAuthHandler
	defaultFollow *social.DefaultFollow
	oauth *atproto.OAuthHandler
	signupLimiter *signupLimiter
}

func New(options Options) http.Handler {
	s := &Server{
		waitlist:   options.Waitlist,
		beta:       options.Beta,
		appOrigin:  strings.TrimRight(options.AppOrigin, "/"),
		adminToken: options.AdminToken,
		agents: options.Agents,
		clear: options.Clear,
		protocols: options.Protocols,
		profiles: options.Profiles,
		socialAccounts: options.SocialAccounts,
		socialActions: options.SocialActions,
		socialAuth: options.OAuth,
		defaultFollow: options.DefaultFollow,
		oauth: options.OAuth,
		signupLimiter: newSignupLimiter(),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	if options.SocialProfileActionsEnabled && options.OAuth!=nil && options.SocialActions!=nil {
		mux.HandleFunc("POST /api/me/profile/claim",s.claimOwnProfile)
		mux.HandleFunc("PUT /api/me/following/{handle}",s.updateOwnFollow)
		mux.HandleFunc("DELETE /api/me/following/{handle}",s.updateOwnFollow)
	}
	if options.SocialAccountRegistrationEnabled && options.OAuth!=nil && options.SocialAccounts!=nil {mux.HandleFunc("POST /api/social/accounts",s.registerSocialAccount)}
	if options.DefaultFollowEnabled && options.OAuth!=nil && options.DefaultFollow!=nil {mux.HandleFunc("DELETE /api/me/following/chris",s.removeDefaultFollow)}
	mux.HandleFunc("POST /api/admin/clear/run-requests",s.createRunRequest)
	mux.HandleFunc("POST /api/admin/clear/run-requests/{id}/decision",s.decideRunRequest)
	mux.HandleFunc("GET /api/protocols",s.discoverProtocols)
	mux.HandleFunc("GET /api/profiles/{handle}",s.getPublicProfile)
	mux.HandleFunc("POST /api/waitlist", s.joinWaitlist)
	mux.HandleFunc("POST /api/admin/invites", s.createBetaInvite)
	mux.HandleFunc("POST /api/beta/start", s.startBetaActivation)
	mux.HandleFunc("POST /api/beta/resend", s.resendBetaCode)
	mux.HandleFunc("POST /api/beta/verify", s.verifyBetaCode)

	if options.OAuth!=nil {options.OAuth.Register(mux)}
	if options.AgentRegistrationEnabled {mux.HandleFunc("POST /api/admin/agents",s.registerAgent)}
	return s.middleware(mux)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"service": "cleartify-api",
	})
}

func (s *Server) joinWaitlist(w http.ResponseWriter, r *http.Request) {
	if !s.limitSignup(w,r) { return }
	r.Body = http.MaxBytesReader(w, r.Body, 32<<10)

	var req struct {
		Name         string `json:"name"`
		Email        string `json:"email"`
		Phone        string `json:"phone"`
		EmailUpdates bool   `json:"emailUpdates"`
		ReferralCode string `json:"referralCode"`
		AcceptPolicies bool `json:"acceptPolicies"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	if req.ReferralCode != "" && !referralCodePattern.MatchString(req.ReferralCode) {
		writeJSON(w,http.StatusBadRequest,map[string]any{"error":"validation_error","field":"referralCode"})
		return
	}
	source := "landing"
	if req.ReferralCode != "" {source = "ref:" + req.ReferralCode}
	_, err := s.waitlist.Join(r.Context(), waitlist.Signup{
		Name:         req.Name,
		Email:        req.Email,
		Phone:        req.Phone,
		EmailUpdates: req.EmailUpdates,
		Source:       source,
		AcceptPolicies: req.AcceptPolicies,
	})
	if err != nil {
		var validation *waitlist.ValidationError
		switch {
		case errors.As(err, &validation):
			writeJSON(w, http.StatusBadRequest, map[string]any{
				"error":   "validation_error",
				"field":   validation.Field,
				"message": validation.Message,
			})
		case errors.Is(err, waitlist.ErrPhoneInUse):
			writeJSON(w, http.StatusConflict, map[string]any{
				"error":   "phone_in_use",
				"message": "That mobile number is already on the waitlist with another email.",
			})
		default:
			slog.Error("waitlist signup failed", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]any{
				"error": "internal_error",
			})
		}
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"joined": true})
}

func (s *Server) createBetaInvite(w http.ResponseWriter, r *http.Request) {
	if s.adminToken == "" {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"error": "admin_not_configured",
		})
		return
	}
	if !secureBearerMatch(r.Header.Get("Authorization"), s.adminToken) {
		writeJSON(w, http.StatusUnauthorized, map[string]any{
			"error": "unauthorized",
		})
		return
	}

	var req struct {
		Email string `json:"email"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	result, err := s.beta.Invite(r.Context(), req.Email)
	switch {
	case err == nil:
		writeJSON(w, http.StatusAccepted, map[string]any{
			"invited":  true,
			"inviteId": result.InviteID,
		})
	case errors.Is(err, beta.ErrInviteNotFound):
		writeJSON(w, http.StatusNotFound, map[string]any{
			"error": "waitlist_entry_not_found",
		})
	case errors.Is(err, notifications.ErrUnavailable):
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"error": "notification_delivery_unavailable",
		})
	default:
		slog.Error("create beta invite failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error": "internal_error",
		})
	}
}

func (s *Server) startBetaActivation(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	result, err := s.beta.Start(r.Context(), req.Token)
	if err != nil {
		writeBetaError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{
		"activationToken": result.ActivationToken,
		"phoneHint":       result.PhoneHint,
		"expiresIn":       result.ExpiresIn,
	})
}

func (s *Server) resendBetaCode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ActivationToken string `json:"activationToken"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	result, err := s.beta.Resend(r.Context(), req.ActivationToken)
	if err != nil {
		writeBetaError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{
		"activationToken": result.ActivationToken,
		"phoneHint":       result.PhoneHint,
		"expiresIn":       result.ExpiresIn,
	})
}

func (s *Server) verifyBetaCode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ActivationToken string `json:"activationToken"`
		Code            string `json:"code"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	_, err := s.beta.Verify(r.Context(), req.ActivationToken, req.Code)
	if err != nil {
		writeBetaError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"activated": true,
	})
}

func writeBetaError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, beta.ErrInviteNotFound), errors.Is(err, beta.ErrActivationNotFound):
		writeJSON(w, http.StatusNotFound, map[string]any{
			"error":   "activation_not_found",
			"message": "This beta activation link is not valid.",
		})
	case errors.Is(err, beta.ErrInviteExpired), errors.Is(err, beta.ErrActivationExpired):
		writeJSON(w, http.StatusGone, map[string]any{
			"error":   "activation_expired",
			"message": "This beta activation has expired.",
		})
	case errors.Is(err, beta.ErrInviteUsed):
		writeJSON(w, http.StatusConflict, map[string]any{
			"error":   "invite_used",
			"message": "This invitation link has already been used.",
		})
	case errors.Is(err, beta.ErrAlreadyActivated):
		writeJSON(w, http.StatusConflict, map[string]any{
			"error":   "already_activated",
			"message": "This beta invitation is already activated.",
		})
	case errors.Is(err, beta.ErrInvalidCode):
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error":   "invalid_code",
			"message": "That code did not match.",
		})
	case errors.Is(err, beta.ErrLocked):
		writeJSON(w, http.StatusTooManyRequests, map[string]any{
			"error":   "activation_locked",
			"message": "Too many incorrect attempts. This activation is locked.",
		})
	case errors.Is(err, beta.ErrResendTooSoon):
		writeJSON(w, http.StatusTooManyRequests, map[string]any{
			"error":   "resend_too_soon",
			"message": "Wait a minute before requesting another code.",
		})
	case errors.Is(err, beta.ErrPhoneAlreadyActive):
		writeJSON(w, http.StatusConflict, map[string]any{
			"error":   "phone_already_active",
			"message": "That mobile number already belongs to an active Cleartify account.",
		})
	case errors.Is(err, notifications.ErrUnavailable):
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"error": "notification_delivery_unavailable",
		})
	default:
		slog.Error("beta activation failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error": "internal_error",
		})
	}
}

func (s *Server) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		if s.appOrigin != "" {
			origin := r.Header.Get("Origin")
			if origin == s.appOrigin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
			}
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		started := time.Now()
		next.ServeHTTP(w, r)
		slog.Debug("request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(started))
	})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 32<<10)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "invalid_request",
		})
		return false
	}
	return true
}

func secureBearerMatch(header, expected string) bool {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return false
	}
	got := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	return len(got) == len(expected) && subtle.ConstantTimeCompare([]byte(got), []byte(expected)) == 1
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Error("write response", "error", err)
	}
}

func (s *Server) getPublicProfile(w http.ResponseWriter,r *http.Request){
 if s.profiles==nil {writeJSON(w,http.StatusServiceUnavailable,map[string]string{"error":"not_configured"});return}
 p,err:=s.profiles.PublicByHandle(r.Context(),r.PathValue("handle"))
 if err!=nil {
  if errors.Is(err,social.ErrInvalidHandle){writeJSON(w,http.StatusBadRequest,map[string]string{"error":"invalid_handle"});return}
  if errors.Is(err,social.ErrNotFound){writeJSON(w,http.StatusNotFound,map[string]string{"error":"profile_not_found"});return}
  slog.Error("public profile lookup", "error", err)
  writeJSON(w,http.StatusInternalServerError,map[string]string{"error":"internal_error"});return
 }
 writeJSON(w,http.StatusOK,p)
}

func (s *Server) discoverProtocols(w http.ResponseWriter,r *http.Request){
 if s.protocols==nil {writeJSON(w,http.StatusServiceUnavailable,map[string]string{"error":"not_configured"});return}
 result,err:=s.protocols.Discover(r.Context(),20)
 if err!=nil{slog.Error("protocol discovery", "error",err);writeJSON(w,http.StatusInternalServerError,map[string]string{"error":"internal_error"});return}
 writeJSON(w,http.StatusOK,map[string]any{"protocols":result})
}

func(s *Server) authorizedClear(w http.ResponseWriter,r *http.Request) bool {
 if s.clear==nil || s.adminToken=="" {writeJSON(w,http.StatusServiceUnavailable,map[string]string{"error":"not_configured"});return false}
 if !secureBearerMatch(r.Header.Get("Authorization"),s.adminToken){writeJSON(w,http.StatusUnauthorized,map[string]string{"error":"unauthorized"});return false}
 return true
}
func(s *Server) createRunRequest(w http.ResponseWriter,r *http.Request){
 if !s.authorizedClear(w,r){return}
 var req struct{UserID string `json:"userId"`;PublicationID string `json:"publicationId"`}
 if !decodeJSON(w,r,&req){return}
 id,err:=s.clear.Request(r.Context(),req.UserID,req.PublicationID)
 if err!=nil {if errors.Is(err,clear.ErrNotPublished){writeJSON(w,http.StatusBadRequest,map[string]string{"error":"not_installed_or_published"});return};slog.Error("run request", "error",err);writeJSON(w,http.StatusInternalServerError,map[string]string{"error":"internal_error"});return}
 writeJSON(w,http.StatusAccepted,map[string]any{"id":id,"status":"pending_approval","executed":false})
}
func(s *Server) decideRunRequest(w http.ResponseWriter,r *http.Request){
 if !s.authorizedClear(w,r){return}
 var req struct{UserID string `json:"userId"`;Approve bool `json:"approve"`}
 if !decodeJSON(w,r,&req){return}
 if err:=s.clear.Decide(r.Context(),req.UserID,r.PathValue("id"),req.Approve);err!=nil{
 if errors.Is(err,clear.ErrNotPending){writeJSON(w,http.StatusConflict,map[string]string{"error":"not_pending"});return}
 writeJSON(w,http.StatusInternalServerError,map[string]string{"error":"internal_error"});return}
 status:="rejected";if req.Approve{status="approved"}
 writeJSON(w,http.StatusOK,map[string]any{"status":status,"executed":false})
}

func(s *Server) registerAgent(w http.ResponseWriter,r *http.Request){
 if s.agents==nil||s.adminToken=="" {writeJSON(w,http.StatusServiceUnavailable,map[string]string{"error":"not_configured"});return}
 if !secureBearerMatch(r.Header.Get("Authorization"),s.adminToken) {writeJSON(w,http.StatusUnauthorized,map[string]string{"error":"unauthorized"});return}
 var req struct{OwnerDID string `json:"ownerDid"`; DisplayName string `json:"displayName"`; Description string `json:"description"`}
 if !decodeJSON(w,r,&req){return}
 reg,err:=s.agents.Register(r.Context(),req.OwnerDID,req.DisplayName,req.Description)
 if err!=nil {
  if errors.Is(err,agents.ErrInvalidRegistration){writeJSON(w,http.StatusBadRequest,map[string]string{"error":"invalid_registration"});return}
  slog.Error("agent registration failed","error",err)
  writeJSON(w,http.StatusInternalServerError,map[string]string{"error":"internal_error"});return
 }
 writeJSON(w,http.StatusCreated,reg)
}
