package app

import (
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"

	"cleartify/internal/auth"
	"cleartify/internal/config"
	"cleartify/internal/server"
)

func testFrontendFS() fs.FS {
	return fstest.MapFS{
		"index.html":    {Data: []byte("<html>app</html>")},
		"assets/app.js": {Data: []byte("console.log('app')")},
	}
}

func TestNewUsesDemoAuthByDefault(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		Port:            "0",
		DatabasePath:    t.TempDir() + "/cleartify.db",
		AppMarketingURL: "http://localhost:4173",
		SeedDemo:        true,
	}

	application, err := New(cfg, testFrontendFS())
	if err != nil {
		t.Fatalf("expected app to build: %v", err)
	}
	t.Cleanup(func() { _ = application.Close() })

	if _, ok := application.Auth.(*auth.DemoAuthService); !ok {
		t.Fatalf("expected demo auth service, got %T", application.Auth)
	}
	if application.Handler == nil {
		t.Fatal("expected handler to be initialized")
	}
}

func TestNewUsesWorkOSWhenConfigured(t *testing.T) {
	t.Parallel()

	originalFactory := newWorkOSService
	t.Cleanup(func() { newWorkOSService = originalFactory })

	newWorkOSService = func(config.Config) (auth.Service, error) {
		return auth.NewDemoAuthService(), nil
	}

	cfg := config.Config{
		Port:                 "0",
		DatabasePath:         t.TempDir() + "/cleartify.db",
		AppMarketingURL:      "http://localhost:4173",
		WorkOSAPIKey:         "api",
		WorkOSClientID:       "client",
		WorkOSCookiePassword: "cookie",
	}

	application, err := New(cfg, testFrontendFS())
	if err != nil {
		t.Fatalf("expected configured app to build: %v", err)
	}
	t.Cleanup(func() { _ = application.Close() })

	if _, ok := application.Auth.(*auth.DemoAuthService); !ok {
		t.Fatalf("expected stubbed workos service, got %T", application.Auth)
	}
}

func TestListenAndServeAndClose(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		Port:            "0",
		DatabasePath:    t.TempDir() + "/cleartify.db",
		AppMarketingURL: "http://localhost:4173",
	}

	application, err := New(cfg, testFrontendFS())
	if err != nil {
		t.Fatalf("expected app to build: %v", err)
	}
	t.Cleanup(func() { _ = application.Close() })

	go func() {
		time.Sleep(25 * time.Millisecond)
		_ = application.Close()
	}()

	if err := application.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		t.Fatalf("expected clean shutdown, got %v", err)
	}
}

func TestNewReturnsBootstrapError(t *testing.T) {
	t.Parallel()

	parent := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(parent, []byte("x"), 0o644); err != nil {
		t.Fatalf("expected file fixture: %v", err)
	}

	cfg := config.Config{
		Port:            "0",
		DatabasePath:    filepath.Join(parent, "cleartify.db"),
		AppMarketingURL: "http://localhost:4173",
	}

	if _, err := New(cfg, testFrontendFS()); err == nil {
		t.Fatal("expected bootstrap error")
	}
}

func TestNewReturnsWorkOSError(t *testing.T) {
	originalFactory := newWorkOSService
	t.Cleanup(func() { newWorkOSService = originalFactory })

	newWorkOSService = func(config.Config) (auth.Service, error) {
		return nil, errors.New("workos failed")
	}

	cfg := config.Config{
		Port:                 "0",
		DatabasePath:         filepath.Join(t.TempDir(), "cleartify.db"),
		AppMarketingURL:      "http://localhost:4173",
		WorkOSAPIKey:         "api",
		WorkOSClientID:       "client",
		WorkOSCookiePassword: "cookie",
	}

	if _, err := New(cfg, testFrontendFS()); err == nil {
		t.Fatal("expected workos factory error")
	}
}

func TestNewReturnsServerConstructionError(t *testing.T) {
	originalServerFactory := newHTTPServer
	t.Cleanup(func() { newHTTPServer = originalServerFactory })

	newHTTPServer = func(config.Config, server.DataStore, auth.Service, fs.FS) (*server.Server, error) {
		return nil, errors.New("server failed")
	}

	cfg := config.Config{
		Port:            "0",
		DatabasePath:    filepath.Join(t.TempDir(), "cleartify.db"),
		AppMarketingURL: "http://localhost:4173",
	}

	if _, err := New(cfg, testFrontendFS()); err == nil {
		t.Fatal("expected server construction error")
	}
}
