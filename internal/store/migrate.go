package store

import (
	"database/sql"
	"fmt"

	sqldata "cleartify/data"
)

var insertMigrationStatement = sqldata.MustQuery("insert_migration.sql")
var pragmaTableInfoTemplate = sqldata.MustQuery("pragma_table_info.sql")
var addNullableTextColumnTemplate = sqldata.MustQuery("add_nullable_text_column.sql")
var phase1WorkObjectsClusterIDMigration = sqldata.MustMigration("001_phase1_work_objects_cluster_id.sql")
var phase3WorkObjectsPrototypeNoteMigration = sqldata.MustMigration("002_phase3_work_objects_prototype_note.sql")
var phase5WorkObjectsReleaseReadyAtMigration = sqldata.MustMigration("003_phase5_work_objects_release_ready_at.sql")

var recordMigration = func(db *sql.DB, name, appliedAt string) error {
	_, err := db.Exec(insertMigrationStatement, name, appliedAt)
	return err
}

var queryTableInfo = func(db *sql.DB, table string) (*sql.Rows, error) {
	return db.Query(sqldata.Render(pragmaTableInfoTemplate, map[string]string{"table": table}))
}

var phase0MigrationStatements = []string{
	sqldata.MustMigration("000_phase0_init.sql"),
}

func (s *Store) Migrate() error {
	for _, statement := range phase0MigrationStatements {
		if _, err := s.DB.Exec(statement); err != nil {
			return fmt.Errorf("run migration: %w", err)
		}
	}

	if err := ensureNullableTextColumn(s.DB, "work_objects", "cluster_id"); err != nil {
		return fmt.Errorf("ensure work_objects.cluster_id: %w", err)
	}
	if err := ensureNullableTextColumn(s.DB, "work_objects", "prototype_note"); err != nil {
		return fmt.Errorf("ensure work_objects.prototype_note: %w", err)
	}
	if err := ensureNullableTextColumn(s.DB, "work_objects", "release_ready_at"); err != nil {
		return fmt.Errorf("ensure work_objects.release_ready_at: %w", err)
	}

	if err := recordMigration(s.DB, "phase0_init", demoTimestamp.Format("2006-01-02T15:04:05Z07:00")); err != nil {
		return fmt.Errorf("record migration: %w", err)
	}
	if err := recordMigration(s.DB, "phase1_work_objects_cluster_id", demoTimestamp.Format("2006-01-02T15:04:05Z07:00")); err != nil {
		return fmt.Errorf("record phase1 migration: %w", err)
	}
	if err := recordMigration(s.DB, "phase3_work_objects_prototype_note", demoTimestamp.Format("2006-01-02T15:04:05Z07:00")); err != nil {
		return fmt.Errorf("record phase3 migration: %w", err)
	}
	if err := recordMigration(s.DB, "phase5_work_objects_release_ready_at", demoTimestamp.Format("2006-01-02T15:04:05Z07:00")); err != nil {
		return fmt.Errorf("record phase5 migration: %w", err)
	}

	return nil
}

func ensureNullableTextColumn(db *sql.DB, table, column string) error {
	hasColumn, err := tableHasColumn(db, table, column)
	if err != nil {
		return err
	}
	if hasColumn {
		return nil
	}

	statement := sqldata.Render(addNullableTextColumnTemplate, map[string]string{
		"table":  table,
		"column": column,
	})
	if table == "work_objects" && column == "cluster_id" {
		statement = phase1WorkObjectsClusterIDMigration
	}
	if table == "work_objects" && column == "prototype_note" {
		statement = phase3WorkObjectsPrototypeNoteMigration
	}
	if table == "work_objects" && column == "release_ready_at" {
		statement = phase5WorkObjectsReleaseReadyAtMigration
	}

	if _, err := db.Exec(statement); err != nil {
		return fmt.Errorf("alter table: %w", err)
	}
	return nil
}

func tableHasColumn(db *sql.DB, table, column string) (bool, error) {
	rows, err := queryTableInfo(db, table)
	if err != nil {
		return false, fmt.Errorf("query table info: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var columnType string
		var notNull int
		var defaultValue sql.NullString
		var primaryKey int
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return false, fmt.Errorf("scan table info: %w", err)
		}
		if name == column {
			return true, nil
		}
	}

	return false, rows.Err()
}
