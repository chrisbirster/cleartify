package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	turso "turso.tech/database/tursogo"
)

type Store struct {
	DB     *sql.DB
	syncDB *turso.TursoSyncDb
}

type Config struct {
	Path       string
	RemoteURL  string
	AuthToken  string
	ClientName string
}

func Open(ctx context.Context, cfg Config) (*Store, error) {
	if cfg.Path == "" {
		cfg.Path = "./data/cleartify.db"
	}
	if cfg.ClientName == "" {
		cfg.ClientName = "cleartify"
	}

	if cfg.Path != ":memory:" {
		dir := filepath.Dir(cfg.Path)
		if dir != "." {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return nil, fmt.Errorf("create database directory: %w", err)
			}
		}
	}

	if cfg.RemoteURL != "" {
		if cfg.AuthToken == "" {
			return nil, errors.New("TURSO_AUTH_TOKEN is required when TURSO_REMOTE_URL is set")
		}
		bootstrap := true
		syncDB, err := turso.NewTursoSyncDb(ctx, turso.TursoSyncDbConfig{
			Path:             cfg.Path,
			RemoteUrl:        cfg.RemoteURL,
			AuthToken:        cfg.AuthToken,
			ClientName:       cfg.ClientName,
			BootstrapIfEmpty: &bootstrap,
			BusyTimeout:      5000,
		})
		if err != nil {
			return nil, fmt.Errorf("open Turso sync database: %w", err)
		}

		db, err := syncDB.Connect(ctx)
		if err != nil {
			return nil, fmt.Errorf("connect Turso sync database: %w", err)
		}
		if _, err := syncDB.Pull(ctx); err != nil {
			db.Close()
			return nil, fmt.Errorf("pull Turso database: %w", err)
		}
		return &Store{DB: db, syncDB: syncDB}, nil
	}

	db, err := sql.Open("turso", cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("open local Turso database: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &Store{DB: db}, nil
}

func (s *Store) Close() error {
	if s == nil || s.DB == nil {
		return nil
	}
	return s.DB.Close()
}

func (s *Store) Push(ctx context.Context) error {
	if s.syncDB == nil {
		return nil
	}
	if err := s.syncDB.Push(ctx); err != nil {
		return fmt.Errorf("push Turso database: %w", err)
	}
	return nil
}
