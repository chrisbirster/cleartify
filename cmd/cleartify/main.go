package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"cleartify"
	"cleartify/internal/app"
	"cleartify/internal/config"
)

type application interface {
	ListenAndServe() error
	Close() error
}

var (
	buildApplication = func(cfg config.Config) (application, error) {
		return app.New(cfg, cleartify.Frontend())
	}
	exitFn = os.Exit
)

func main() {
	exitFn(run(os.Getenv, os.Stderr))
}

func run(getenv func(string) string, stderr io.Writer) int {
	cfg, err := config.Load(getenv)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "load config: %v\n", err)
		return 1
	}

	application, err := buildApplication(cfg)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "build app: %v\n", err)
		return 1
	}
	defer application.Close()

	_, _ = fmt.Fprintf(stderr, "Cleartify server listening on http://localhost:%s\n", cfg.Port)

	if err := application.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		_, _ = fmt.Fprintf(stderr, "listen and serve: %v\n", err)
		return 1
	}

	return 0
}
