package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chrisbirster/cleartify/internal/api"
	"github.com/chrisbirster/cleartify/internal/clear"
	"github.com/chrisbirster/cleartify/internal/beta"
	"github.com/chrisbirster/cleartify/internal/config"
	"github.com/chrisbirster/cleartify/internal/database"
	"github.com/chrisbirster/cleartify/internal/notifications"
	"github.com/chrisbirster/cleartify/internal/security"
	"github.com/chrisbirster/cleartify/internal/waitlist"
	"github.com/chrisbirster/cleartify/internal/protocols"
	"github.com/chrisbirster/cleartify/internal/social"
)

func main() {
	cfg := config.Load()

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

	tokens, err := security.NewTokens(cfg.AppSecret)
	if err != nil {
		slog.Error("configure app security", "error", err)
		os.Exit(1)
	}

	var sender notifications.Sender = notifications.DisabledSender{}
	if cfg.NotificationMode == "debug" {
		slog.Warn("notification debug mode enabled; beta secrets will be printed to logs")
		sender = notifications.DebugSender{}
	}

	waitlistService := waitlist.NewService(db)
	betaService := beta.NewService(db, tokens, sender, cfg.AppURL)
	handler := api.New(api.Options{
		Waitlist:   waitlistService,
		Beta:       betaService,
		AppOrigin:  cfg.AppOrigin,
		AdminToken: cfg.AdminToken,
		Clear: &clear.Service{DB:db.DB},
		Protocols: &protocols.Reader{DB:db.DB},
		Profiles: &social.Reader{DB:db.DB},
	})

	server := &http.Server{
		Addr:              ":" + cfg.Port,
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

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("http shutdown", "error", err)
	}
}
