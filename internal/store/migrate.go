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
var trancheBUsersOnboardingCompletedAtMigration = sqldata.MustMigration("004_tranche_b_users_onboarding_completed_at.sql")
var trancheBSessionsRefreshTokenMigration = sqldata.MustMigration("005_tranche_b_sessions_refresh_token.sql")
var trancheBSessionsWorkOSSessionIDMigration = sqldata.MustMigration("006_tranche_b_sessions_workos_session_id.sql")
var trancheBSessionsOrganizationIDMigration = sqldata.MustMigration("007_tranche_b_sessions_organization_id.sql")
var trancheBSessionsOrganizationNameMigration = sqldata.MustMigration("008_tranche_b_sessions_organization_name.sql")
var trancheBSessionsAccessTokenExpiresAtMigration = sqldata.MustMigration("009_tranche_b_sessions_access_token_expires_at.sql")
var liveWorkspaceWorkspaceSettingsMigration = sqldata.MustMigration("010_live_workspace_workspace_settings.sql")
var phase2ProjectsMigration = sqldata.MustMigration("011_phase2_projects.sql")

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
	if err := ensureNullableTextColumn(s.DB, "users", "onboarding_completed_at"); err != nil {
		return fmt.Errorf("ensure users.onboarding_completed_at: %w", err)
	}
	if err := ensureNullableTextColumn(s.DB, "sessions", "refresh_token"); err != nil {
		return fmt.Errorf("ensure sessions.refresh_token: %w", err)
	}
	if err := ensureNullableTextColumn(s.DB, "sessions", "workos_session_id"); err != nil {
		return fmt.Errorf("ensure sessions.workos_session_id: %w", err)
	}
	if err := ensureNullableTextColumn(s.DB, "sessions", "organization_id"); err != nil {
		return fmt.Errorf("ensure sessions.organization_id: %w", err)
	}
	if err := ensureNullableTextColumn(s.DB, "sessions", "organization_name"); err != nil {
		return fmt.Errorf("ensure sessions.organization_name: %w", err)
	}
	if err := ensureNullableTextColumn(s.DB, "sessions", "access_token_expires_at"); err != nil {
		return fmt.Errorf("ensure sessions.access_token_expires_at: %w", err)
	}
	if _, err := s.DB.Exec(liveWorkspaceWorkspaceSettingsMigration); err != nil {
		return fmt.Errorf("ensure workspace_settings table: %w", err)
	}
	if _, err := s.DB.Exec(phase2ProjectsMigration); err != nil {
		return fmt.Errorf("ensure projects table: %w", err)
	}
	if err := ensureNullableTextColumn(s.DB, "signals", "project_id"); err != nil {
		return fmt.Errorf("ensure signals.project_id: %w", err)
	}
	if err := ensureNullableTextColumn(s.DB, "clusters", "project_id"); err != nil {
		return fmt.Errorf("ensure clusters.project_id: %w", err)
	}
	if err := ensureNullableTextColumn(s.DB, "work_objects", "project_id"); err != nil {
		return fmt.Errorf("ensure work_objects.project_id: %w", err)
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
	if err := recordMigration(s.DB, "tranche_b_users_onboarding_completed_at", demoTimestamp.Format("2006-01-02T15:04:05Z07:00")); err != nil {
		return fmt.Errorf("record tranche b users migration: %w", err)
	}
	if err := recordMigration(s.DB, "tranche_b_sessions_refresh_token", demoTimestamp.Format("2006-01-02T15:04:05Z07:00")); err != nil {
		return fmt.Errorf("record tranche b sessions refresh token migration: %w", err)
	}
	if err := recordMigration(s.DB, "tranche_b_sessions_workos_session_id", demoTimestamp.Format("2006-01-02T15:04:05Z07:00")); err != nil {
		return fmt.Errorf("record tranche b sessions workos session id migration: %w", err)
	}
	if err := recordMigration(s.DB, "tranche_b_sessions_organization_id", demoTimestamp.Format("2006-01-02T15:04:05Z07:00")); err != nil {
		return fmt.Errorf("record tranche b sessions organization id migration: %w", err)
	}
	if err := recordMigration(s.DB, "tranche_b_sessions_organization_name", demoTimestamp.Format("2006-01-02T15:04:05Z07:00")); err != nil {
		return fmt.Errorf("record tranche b sessions organization name migration: %w", err)
	}
	if err := recordMigration(s.DB, "tranche_b_sessions_access_token_expires_at", demoTimestamp.Format("2006-01-02T15:04:05Z07:00")); err != nil {
		return fmt.Errorf("record tranche b sessions access token expiry migration: %w", err)
	}
	if err := recordMigration(s.DB, "live_workspace_workspace_settings", demoTimestamp.Format("2006-01-02T15:04:05Z07:00")); err != nil {
		return fmt.Errorf("record workspace settings migration: %w", err)
	}
	if err := recordMigration(s.DB, "phase2_projects", demoTimestamp.Format("2006-01-02T15:04:05Z07:00")); err != nil {
		return fmt.Errorf("record phase2 projects migration: %w", err)
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
	if table == "users" && column == "onboarding_completed_at" {
		statement = trancheBUsersOnboardingCompletedAtMigration
	}
	if table == "sessions" && column == "refresh_token" {
		statement = trancheBSessionsRefreshTokenMigration
	}
	if table == "sessions" && column == "workos_session_id" {
		statement = trancheBSessionsWorkOSSessionIDMigration
	}
	if table == "sessions" && column == "organization_id" {
		statement = trancheBSessionsOrganizationIDMigration
	}
	if table == "sessions" && column == "organization_name" {
		statement = trancheBSessionsOrganizationNameMigration
	}
	if table == "sessions" && column == "access_token_expires_at" {
		statement = trancheBSessionsAccessTokenExpiresAtMigration
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
