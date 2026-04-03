package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	sqldata "cleartify/data"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	DB *sql.DB
}

var sqliteDriverName = "sqlite3"

var sqlitePragmas = []string{
	"PRAGMA journal_mode = WAL;",
	"PRAGMA foreign_keys = ON;",
	"PRAGMA busy_timeout = 5000;",
}

var countRowsTemplate = sqldata.MustQuery("count_rows.sql")

func Open(path string) (*Store, error) {
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create sqlite dir: %w", err)
		}
	}

	db, err := sql.Open(sqliteDriverName, path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	for _, pragma := range sqlitePragmas {
		if _, err := db.Exec(pragma); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("apply sqlite pragma: %w", err)
		}
	}

	return &Store{DB: db}, nil
}

func Bootstrap(path string, seedDemo bool) (*Store, error) {
	store, err := Open(path)
	if err != nil {
		return nil, err
	}

	if err := store.Migrate(); err != nil {
		_ = store.Close()
		return nil, err
	}

	if seedDemo {
		if err := store.SeedDemo(); err != nil {
			_ = store.Close()
			return nil, err
		}
	}

	return store, nil
}

func (s *Store) Close() error {
	if s == nil || s.DB == nil {
		return nil
	}
	return s.DB.Close()
}

func countRows(ctx context.Context, db *sql.DB, table string) (int, error) {
	var count int
	statement := sqldata.Render(countRowsTemplate, map[string]string{"table": table})
	if err := db.QueryRowContext(ctx, statement).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func decodeJSONMap(raw sql.NullString) (map[string]any, error) {
	if !raw.Valid || raw.String == "" {
		return nil, nil
	}

	var value map[string]any
	if err := json.Unmarshal([]byte(raw.String), &value); err != nil {
		return nil, err
	}
	return value, nil
}

func encodeJSONMap(value map[string]any) (any, error) {
	if len(value) == 0 {
		return nil, nil
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return string(encoded), nil
}

func decodeTime(raw string) (time.Time, error) {
	return time.Parse(time.RFC3339, raw)
}

func decodeNullableTime(raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}

	parsed, err := decodeTime(raw)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}
