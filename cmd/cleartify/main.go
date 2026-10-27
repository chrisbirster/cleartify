package main

import (
	"context"
 "fmt"
 "log"
 "strings"
	"errors"
	"log/slog"
	"net/http"
 "net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chrisbirster/cleartify/internal/api"
 "github.com/chrisbirster/cleartify/internal/operations"
 "github.com/chrisbirster/cleartify/internal/alpha"
	"github.com/chrisbirster/cleartify/internal/agents"
 "github.com/chrisbirster/cleartify/internal/interop"
	"github.com/chrisbirster/cleartify/internal/clear"
	"github.com/chrisbirster/cleartify/internal/atproto"
	"github.com/chrisbirster/cleartify/internal/beta"
 "github.com/chrisbirster/cleartify/internal/clearsession"
	"github.com/chrisbirster/cleartify/internal/config"
	"github.com/chrisbirster/cleartify/internal/database"
	"github.com/chrisbirster/cleartify/internal/notifications"
	"github.com/chrisbirster/cleartify/internal/security"
	"github.com/chrisbirster/cleartify/internal/waitlist"
	"github.com/chrisbirster/cleartify/internal/protocols"
 "github.com/chrisbirster/cleartify/internal/marketplace"
	"github.com/chrisbirster/cleartify/internal/social"
)

func main() {
 if len(os.Args)==2&&os.Args[1]=="config-preflight"{
  env:=map[string]string{}
  for _,entry:=range os.Environ(){
   parts:=strings.SplitN(entry,"=",2)
   if len(parts)==2{env[parts[0]]=parts[1]}
  }
  failures:=operations.Preflight(env)
  for _,f:=range failures{fmt.Fprintf(os.Stderr,"config-preflight: %s: %s\n",f.Name,f.Problem)}
  if len(failures)!=0{os.Exit(1)}
  fmt.Println("config-preflight: valid")
  return
 }
 cfg := config.Load()
 hardeningEnabled:=strings.EqualFold(strings.TrimSpace(os.Getenv("OPERATIONS_HARDENING_ENABLED")),"true")
 if hardeningEnabled {
  slog.SetDefault(slog.New(operations.RedactingHandler{Next:slog.NewJSONHandler(os.Stdout,nil)}))
  // Debug sender prints beta secrets: never run it in hardened mode.
  if cfg.NotificationMode=="debug"{log.Print("NOTIFICATION_MODE: unsafe_debug_mode");os.Exit(1)}
 }


	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.Open(ctx, database.Config{
		Path:       cfg.DatabasePath,
		RemoteURL:  cfg.TursoRemoteURL,
		AuthToken:  cfg.TursoAuthToken,
		ClientName: "cleartify-api",
	})
	if err != nil {
		slog.Error("open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Migrate(ctx); err != nil {
		slog.Error("migrate database", "error", err)
		os.Exit(1)
	}

	// Recovery is a startup gate: never serve requests if an interrupted
 // fake run cannot be reconciled and its audit durably persisted.
 if cfg.FakeRunRecoveryEnabled && cfg.FakeInstallRunEnabled {
  recovered, recoveryErr := api.RecoverReservedFakeRuns(ctx, db.DB, db.Push)
  if recoveryErr != nil {slog.Error("fake run recovery failed closed","error",recoveryErr);os.Exit(1)}
  if recovered>0 {slog.Warn("interrupted fake runs marked failed; never automatically re-executed","count",recovered)}
 }
 tokens, err := security.NewTokens(cfg.AppSecret)
	if err != nil {
		slog.Error("configure app security", "error", err)
		os.Exit(1)
	}

	var sender notifications.Sender = notifications.DisabledSender{}
	if cfg.NotificationMode == "debug" {
  if !(strings.HasPrefix(cfg.AppOrigin,"http://localhost:")||strings.HasPrefix(cfg.AppOrigin,"http://127.0.0.1:")) {slog.Error("debug notifications require loopback local app origin");os.Exit(1)}
		slog.Warn("notification debug mode enabled; beta secrets will be printed to logs")
		sender = notifications.DebugSender{}
	}

	waitlistService := waitlist.NewService(db)
 waitlistService.ConsentKeysEnabled=cfg.ConsentKeysEnabled
	betaService := beta.NewService(db, tokens, sender, cfg.AppURL)
	var oauthHandler *atproto.OAuthHandler
	if cfg.AtprotoOAuthEnabled {
		var err error
		oauthHandler,err=atproto.NewOAuthHandler(db,cfg.AtprotoPublicOrigin,cfg.AppOrigin,cfg.AtprotoSessionKey)
		if err!=nil {slog.Error("AT Protocol OAuth configuration invalid", "error", err);os.Exit(1)}
	}
	var defaultFollow *social.DefaultFollow
 if cfg.SocialDefaultFollowEnabled && cfg.AtprotoOAuthEnabled && oauthHandler!=nil {
  defaultFollow=&social.DefaultFollow{Store:db}
  oauthHandler.SetOnVerifiedLogin(defaultFollow.OnVerifiedLogin)
 }
 var clearSessions *clearsession.Service
 if cfg.ClearSessionsEnabled {clearSessions=&clearsession.Service{Store:db,Tokens:tokens}}
 var returningLogin *clearsession.Login
 if cfg.ReturningLoginEnabled && clearSessions!=nil {returningLogin=&clearsession.Login{Sessions:clearSessions,Sender:sender};if err=returningLogin.Ready(ctx);err!=nil{slog.Error("returning login storage failed", "error",err);os.Exit(1)}}
 // Local transport has no HTTP listener, outbound network, or production delivery.
 _ = interop.NewLocalTransport(cfg.AgentTransportEnabled, db.DB, db.Push)
 var trustedProxyNetworks []*net.IPNet
 if cfg.DBAbuseEnabled && cfg.DBAbuseProxyEnabled {
  networks,parseErr:=api.ParseTrustedProxyCIDRs(cfg.DBAbuseTrustedProxyCIDRs)
  if parseErr!=nil {slog.Error("invalid trusted proxy CIDRs","error",parseErr);os.Exit(1)}
  trustedProxyNetworks=networks
 }
 handler := api.New(api.Options{
  DBAbuseEnabled:cfg.DBAbuseEnabled,
  DBAbuseProxyEnabled:cfg.DBAbuseEnabled && cfg.DBAbuseProxyEnabled,
  DBAbuseCleanupEnabled:cfg.DBAbuseEnabled && cfg.DBAbuseCleanupEnabled,
  DBAbuseTrustedProxyCIDRs:trustedProxyNetworks,
  DBAbuseDB:db.DB,
  DBAbuseSecret:cfg.AppSecret,
  IdentityRecoveryEnabled:cfg.IdentityRecoveryEnabled && cfg.IdentityLinkingEnabled && cfg.AtprotoOAuthEnabled && cfg.ClearSessionsEnabled,
  IdentityLinkingEnabled:cfg.IdentityLinkingEnabled && cfg.AtprotoOAuthEnabled && cfg.ClearSessionsEnabled,
  IdentityLinks:&api.LinkPersistence{DB:db.DB,Push:db.Push},
  LinkSocial:oauthHandler,
  ModerationBroaderEnabled:cfg.ModerationBroaderEnabled,
  ModerationEnforcementEnabled:cfg.ModerationEnforcementEnabled,
  ModerationRuntimeEnabled:cfg.ModerationRuntimeEnabled,
  ModerationDB:db.DB,
  AgentOwnershipEnabled:cfg.AgentOwnershipEnabled && cfg.AtprotoOAuthEnabled,
  AgentOwnership:&agents.Ownership{DB:db.DB,Push:db.Push,Verifier:agents.Ed25519OwnershipVerifier{}},
  AgentDirectoryEnabled:cfg.AgentDirectoryEnabled,
  AgentDirectoryDB:db.DB,
  SocialModerationEnabled:cfg.SocialModerationEnabled && cfg.AtprotoOAuthEnabled,
  Moderation:&social.Moderation{DB:db.DB,Tokens:tokens},
  PaidEntitlementEnforcementEnabled:cfg.PaidEntitlementEnforcementEnabled && cfg.FakeCheckoutEnabled && cfg.AtprotoOAuthEnabled,
  FakeCheckoutEnabled:cfg.FakeCheckoutEnabled && cfg.AtprotoOAuthEnabled,
  FakeCheckout:&marketplace.Service{DB:db.DB,Provider:marketplace.FakeProvider{},Enabled:cfg.FakeCheckoutEnabled && cfg.AtprotoOAuthEnabled},
  FakeInstallRunEnabled:cfg.FakeInstallRunEnabled && cfg.PrivateAlphaRoutesEnabled && cfg.ClearSessionsEnabled && cfg.AtprotoOAuthEnabled,
  FakeInstallRunDB:db.DB,
  FakeInstallRunPush:db.Push,
  CreatorTakedownsEnabled:cfg.CreatorTakedownsEnabled && cfg.AtprotoOAuthEnabled,
  CreatorTakedowns:&protocols.Takedowns{DB:db.DB},
  CreatorPublishingEnabled:cfg.CreatorPublishingEnabled && cfg.CreatorDraftsEnabled && cfg.AtprotoOAuthEnabled,
  CreatorDraftsEnabled:cfg.CreatorDraftsEnabled && cfg.AtprotoOAuthEnabled,
  CreatorDrafts:&protocols.DraftService{DB:db.DB,Tokens:tokens,GuideStepsEnabled:cfg.GuideStepsEnabled},
  PrivateAlphaRoutesEnabled:cfg.PrivateAlphaRoutesEnabled,
  AlphaEligibility:alpha.Gate{DB:db.DB},
  ClearPrincipalResolver:api.ClearSessionPrincipal(clearSessions),
  ClearSessionsEnabled:cfg.ClearSessionsEnabled,
  ClearSessions:clearSessions,
  ReturningLoginEnabled:cfg.ReturningLoginEnabled,
  ReturningLogin:returningLogin,
		Waitlist:   waitlistService,
		Beta:       betaService,
		AppOrigin:  cfg.AppOrigin,
		AdminToken: cfg.AdminToken,
		AgentRegistrationEnabled:cfg.AgentRegistrationEnabled,
		Agents:&agents.Service{DB:db.DB},
		Clear: &clear.Service{DB:db.DB},
		OAuth: oauthHandler,
		Protocols: &protocols.Reader{DB:db.DB,HoldVisibility:cfg.ModerationBroaderEnabled && cfg.ModerationRuntimeEnabled,PaidEnforcement:cfg.PaidEntitlementEnforcementEnabled && cfg.FakeCheckoutEnabled},
		Profiles: &social.Reader{DB:db.DB},
		ProtocolInstallFlowEnabled: cfg.ProtocolInstallFlowEnabled && cfg.AtprotoOAuthEnabled,
		InstallFlow: &protocols.InstallService{Store:db,StrictPermissions:cfg.StrictProtocolPermissionsEnabled,PermissionDiffEnabled:cfg.ManifestPermissionDiffEnabled,PaidEnforcement:cfg.PaidEntitlementEnforcementEnabled && cfg.FakeCheckoutEnabled},
		DefaultFollowEnabled: cfg.SocialDefaultFollowEnabled && cfg.AtprotoOAuthEnabled,
		DefaultFollow: defaultFollow,
		SocialAccountRegistrationEnabled: cfg.SocialAccountRegistrationEnabled && cfg.AtprotoOAuthEnabled,
		SocialAccounts: &social.Accounts{Store:db},
		SocialProfileActionsEnabled: cfg.SocialProfileActionsEnabled && cfg.AtprotoOAuthEnabled,
		SocialActions: &social.Actions{Store:db},
	})

	var readiness *operations.Readiness
 if hardeningEnabled{
  readiness=&operations.Readiness{DB:db.DB}
  handler=operations.WithReadiness(handler,readiness)
 }
 server := &http.Server{
		Addr:              func()string{if cfg.NotificationMode=="debug"{return "127.0.0.1:"+cfg.Port};return ":"+cfg.Port}(),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("cleartify api listening", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
	case err := <-errCh:
		slog.Error("http server failed", "error", err)
		stop()
	}

	if readiness!=nil{readiness.Drain()}
 shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("http shutdown", "error", err)
	}
}
