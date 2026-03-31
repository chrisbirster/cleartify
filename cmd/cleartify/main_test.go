package main

import (
	"bytes"
	"errors"
	"net/http"
	"testing"

	"cleartify/internal/config"
)

type fakeApplication struct {
	listenErr error
	closed    bool
}

func (f *fakeApplication) ListenAndServe() error {
	return f.listenErr
}

func (f *fakeApplication) Close() error {
	f.closed = true
	return nil
}

func TestRunReturnsSuccessWhenApplicationStarts(t *testing.T) {
	var stderr bytes.Buffer
	originalBuilder := buildApplication
	t.Cleanup(func() { buildApplication = originalBuilder })

	fakeApp := &fakeApplication{listenErr: http.ErrServerClosed}
	buildApplication = func(cfg config.Config) (application, error) {
		if cfg.Port != "8123" {
			t.Fatalf("expected port 8123, got %q", cfg.Port)
		}
		return fakeApp, nil
	}

	code := run(func(key string) string {
		switch key {
		case "PORT":
			return "8123"
		default:
			return ""
		}
	}, &stderr)

	if code != 0 {
		t.Fatalf("expected run to succeed, got exit code %d with stderr %q", code, stderr.String())
	}
	if !fakeApp.closed {
		t.Fatal("expected application to be closed")
	}
}

func TestRunReturnsFailureOnInvalidConfig(t *testing.T) {
	var stderr bytes.Buffer

	code := run(func(key string) string {
		if key == "PORT" {
			return "not-a-port"
		}
		return ""
	}, &stderr)

	if code != 1 {
		t.Fatalf("expected run to fail, got exit code %d", code)
	}
	if stderr.Len() == 0 {
		t.Fatal("expected config error to be written to stderr")
	}
}

func TestRunReturnsFailureWhenBuildApplicationFails(t *testing.T) {
	var stderr bytes.Buffer
	originalBuilder := buildApplication
	t.Cleanup(func() { buildApplication = originalBuilder })

	buildApplication = func(config.Config) (application, error) {
		return nil, errors.New("build failed")
	}

	code := run(func(string) string { return "" }, &stderr)
	if code != 1 {
		t.Fatalf("expected run to fail, got exit code %d", code)
	}
	if stderr.Len() == 0 {
		t.Fatal("expected build error to be written to stderr")
	}
}

func TestRunReturnsFailureWhenListenAndServeFails(t *testing.T) {
	var stderr bytes.Buffer
	originalBuilder := buildApplication
	t.Cleanup(func() { buildApplication = originalBuilder })

	buildApplication = func(config.Config) (application, error) {
		return &fakeApplication{listenErr: errors.New("listen failed")}, nil
	}

	code := run(func(string) string { return "" }, &stderr)
	if code != 1 {
		t.Fatalf("expected run to fail, got exit code %d", code)
	}
	if stderr.Len() == 0 {
		t.Fatal("expected listen error to be written to stderr")
	}
}

func TestMainUsesExitFunction(t *testing.T) {
	originalBuilder := buildApplication
	originalExit := exitFn
	t.Cleanup(func() {
		buildApplication = originalBuilder
		exitFn = originalExit
	})

	buildApplication = func(cfg config.Config) (application, error) {
		return &fakeApplication{listenErr: http.ErrServerClosed}, nil
	}

	exitCode := -1
	exitFn = func(code int) {
		exitCode = code
	}

	main()

	if exitCode != 0 {
		t.Fatalf("expected main to exit 0, got %d", exitCode)
	}
}
