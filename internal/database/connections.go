package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
)

// Each pooled connection is initialized independently, including connections
// opened after idle eviction or a driver reconnect. The source pool owns the
// Turso local/sync connection; the outer pool exposes only checked connections.
type checkedConnector struct{ source *sql.DB }

func (c checkedConnector) Driver() driver.Driver { return c.source.Driver() }
func (c checkedConnector) Connect(ctx context.Context) (driver.Conn, error) {
	conn, err := c.source.Conn(ctx)
	if err != nil {
		return nil, err
	}
	fail := func(e error) (driver.Conn, error) { conn.Close(); return nil, e }
	if _, err = conn.ExecContext(ctx, "PRAGMA foreign_keys=ON"); err != nil {
		return fail(err)
	}
	var enabled int
	if err = conn.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&enabled); err != nil {
		return fail(err)
	}
	if enabled != 1 {
		return fail(fmt.Errorf("foreign-key enforcement unavailable"))
	}
	return &checkedConnection{conn}, nil
}

type checkedConnection struct{ conn *sql.Conn }

func (c *checkedConnection) Close() error { return c.conn.Close() }
func (c *checkedConnection) Prepare(query string) (driver.Stmt, error) {
	s, e := c.conn.PrepareContext(context.Background(), query)
	return checkedStatement{s}, e
}
func (c *checkedConnection) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}
func (c *checkedConnection) BeginTx(ctx context.Context, o driver.TxOptions) (driver.Tx, error) {
	return c.conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.IsolationLevel(o.Isolation), ReadOnly: o.ReadOnly})
}
func args(values []driver.NamedValue) []any {
	out := make([]any, len(values))
	for i, v := range values {
		if v.Name != "" {
			out[i] = sql.Named(v.Name, v.Value)
		} else {
			out[i] = v.Value
		}
	}
	return out
}
func (c *checkedConnection) ExecContext(ctx context.Context, q string, v []driver.NamedValue) (driver.Result, error) {
	return c.conn.ExecContext(ctx, q, args(v)...)
}
func (c *checkedConnection) QueryContext(ctx context.Context, q string, v []driver.NamedValue) (driver.Rows, error) {
	r, e := c.conn.QueryContext(ctx, q, args(v)...)
	if e != nil {
		return nil, e
	}
	return wrapRows(r)
}

type checkedStatement struct{ s *sql.Stmt }

func (s checkedStatement) Close() error {
	if s.s == nil {
		return nil
	}
	return s.s.Close()
}
func (s checkedStatement) NumInput() int { return -1 }
func stmtArgs(v []driver.Value) []any {
	out := make([]any, len(v))
	for i, x := range v {
		out[i] = x
	}
	return out
}
func (s checkedStatement) Exec(v []driver.Value) (driver.Result, error) {
	return s.s.Exec(stmtArgs(v)...)
}
func (s checkedStatement) Query(v []driver.Value) (driver.Rows, error) {
	r, e := s.s.Query(stmtArgs(v)...)
	if e != nil {
		return nil, e
	}
	return wrapRows(r)
}

type checkedRows struct {
	rows    *sql.Rows
	columns []string
}

func wrapRows(r *sql.Rows) (driver.Rows, error) {
	cols, e := r.Columns()
	if e != nil {
		r.Close()
		return nil, e
	}
	return &checkedRows{r, cols}, nil
}
func (r *checkedRows) Columns() []string { return r.columns }
func (r *checkedRows) Close() error      { return r.rows.Close() }
func (r *checkedRows) Next(dest []driver.Value) error {
	if !r.rows.Next() {
		if e := r.rows.Err(); e != nil {
			return e
		}
		return io.EOF
	}
	values := make([]any, len(dest))
	ptrs := make([]any, len(dest))
	for i := range values {
		ptrs[i] = &values[i]
	}
	if e := r.rows.Scan(ptrs...); e != nil {
		return e
	}
	for i, v := range values {
		dest[i] = v
	}
	return nil
}
func enforceConnections(source *sql.DB) *sql.DB { return sql.OpenDB(checkedConnector{source}) }
