package app

import (
	"io/fs"
	"net/http"

	"cleartify/internal/auth"
	"cleartify/internal/config"
	"cleartify/internal/server"
	"cleartify/internal/store"
)

type Application struct {
	Config     config.Config
	Store      *store.Store
	Auth       auth.Service
	HTTPServer *http.Server
	Handler    http.Handler
}

var newWorkOSService = func(cfg config.Config, sessionStore auth.SessionStore) (auth.Service, error) {
	return auth.NewWorkOSAuthService(
		cfg.WorkOSAPIKey,
		cfg.WorkOSClientID,
		cfg.WorkOSRedirectURI,
		cfg.AppPublicURL,
		cfg.AppMarketingURL,
		cfg.WorkOSCookiePassword,
		sessionStore,
		nil,
	)
}

var newHTTPServer = server.New

func New(cfg config.Config, frontend fs.FS) (*Application, error) {
	dataStore, err := store.Bootstrap(cfg.DatabasePath, cfg.SeedDemo)
	if err != nil {
		return nil, err
	}

	authService := auth.Service(auth.NewUnavailableAuthService(cfg.AppPublicURL, cfg.AppMarketingURL))
	if cfg.DemoAuthEnabled() {
		authService = auth.NewDemoModeAuthService(cfg.AppPublicURL, cfg.AppMarketingURL, cfg.WorkOSCookiePassword, dataStore)
	} else if cfg.WorkOSConfigured() {
		authService, err = newWorkOSService(cfg, dataStore)
		if err != nil {
			_ = dataStore.Close()
			return nil, err
		}
	}

	httpServer, err := newHTTPServer(cfg, dataStore, authService, frontend)
	if err != nil {
		_ = dataStore.Close()
		return nil, err
	}

	handler := httpServer.Handler()
	return &Application{
		Config:  cfg,
		Store:   dataStore,
		Auth:    authService,
		Handler: handler,
		HTTPServer: &http.Server{
			Addr:    cfg.Addr(),
			Handler: handler,
		},
	}, nil
}

func (a *Application) ListenAndServe() error {
	return a.HTTPServer.ListenAndServe()
}

func (a *Application) Close() error {
	_ = a.Store.Close()
	return a.HTTPServer.Close()
}
