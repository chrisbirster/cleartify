package database

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
)

func identifier(s string) string { return `"` + strings.ReplaceAll(s, `"`, `""`) + `"` }

// Explicitly inspect declared relations: the embedded engine's native
// foreign_key_check currently returns no violations for historical orphans.
// Never delete or silently repair these records during application startup.
func (s *Store) CheckForeignKeys(ctx context.Context) error {
	rows, e := s.DB.QueryContext(ctx, `SELECT name FROM sqlite_master WHERE type='table'`)
	if e != nil {
		return e
	}
	tables := []string{}
	for rows.Next() {
		var name string
		if e = rows.Scan(&name); e != nil {
			rows.Close()
			return e
		}
		tables = append(tables, name)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return e
	}
	type relation struct {
		parent                string
		childCols, parentCols []string
	}
	for _, table := range tables {
		refs := map[int]*relation{}
		rows, e = s.DB.QueryContext(ctx, "PRAGMA foreign_key_list("+identifier(table)+")")
		if e != nil {
			return e
		}
		for rows.Next() {
			var id, seq int
			var parent, from, update, del, match string
			var to sql.NullString
			if e = rows.Scan(&id, &seq, &parent, &from, &to, &update, &del, &match); e != nil {
				rows.Close()
				return e
			}
			if refs[id] == nil {
				refs[id] = &relation{parent: parent}
			}
			refs[id].childCols = append(refs[id].childCols, from)
			refs[id].parentCols = append(refs[id].parentCols, to.String)
		}
		e = rows.Err()
		rows.Close()
		if e != nil {
			return e
		}
		for _, ref := range refs {
			if len(ref.parentCols) > 0 && ref.parentCols[0] == "" {
				pk := map[int]string{}
				rows, e = s.DB.QueryContext(ctx, "PRAGMA table_info("+identifier(ref.parent)+")")
				if e != nil {
					return e
				}
				for rows.Next() {
					var cid, notnull, order int
					var name, kind string
					var defaultValue any
					if e = rows.Scan(&cid, &name, &kind, &notnull, &defaultValue, &order); e != nil {
						rows.Close()
						return e
					}
					if order > 0 {
						pk[order] = name
					}
				}
				e = rows.Err()
				rows.Close()
				if e != nil {
					return e
				}
				orders := []int{}
				for order := range pk {
					orders = append(orders, order)
				}
				sort.Ints(orders)
				ref.parentCols = nil
				for _, order := range orders {
					ref.parentCols = append(ref.parentCols, pk[order])
				}
			}
			if len(ref.parentCols) != len(ref.childCols) {
				return fmt.Errorf("invalid foreign key in %s", table)
			}
			present, joins := []string{}, []string{}
			for i, col := range ref.childCols {
				present = append(present, "c."+identifier(col)+" IS NOT NULL")
				joins = append(joins, "p."+identifier(ref.parentCols[i])+"=c."+identifier(col))
			}
			var count int
			query := "SELECT count(*) FROM " + identifier(table) + " c WHERE " + strings.Join(present, " AND ") + " AND NOT EXISTS(SELECT 1 FROM " + identifier(ref.parent) + " p WHERE " + strings.Join(joins, " AND ") + ")"
			if e = s.DB.QueryRowContext(ctx, query).Scan(&count); e != nil {
				return e
			}
			if count > 0 {
				return fmt.Errorf("foreign-key violations in %s; repair before startup", table)
			}
		}
	}
	return nil
}
