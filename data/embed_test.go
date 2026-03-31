package sqldata

import (
	"strings"
	"testing"
)

func TestQueryLoadsEmbeddedSQL(t *testing.T) {
	t.Parallel()

	query, err := Query("list_signals.sql")
	if err != nil {
		t.Fatalf("expected embedded query: %v", err)
	}
	if !strings.Contains(query, "FROM signals") {
		t.Fatalf("expected signals query, got %q", query)
	}
}

func TestMigrationLoadsEmbeddedSQL(t *testing.T) {
	t.Parallel()

	migration, err := Migration("000_phase0_init.sql")
	if err != nil {
		t.Fatalf("expected embedded migration: %v", err)
	}
	if !strings.Contains(migration, "CREATE TABLE IF NOT EXISTS work_objects") {
		t.Fatalf("expected work_objects migration, got %q", migration)
	}
}

func TestQueryReturnsHelpfulErrorForMissingSQL(t *testing.T) {
	t.Parallel()

	_, err := Query("missing.sql")
	if err == nil {
		t.Fatal("expected missing query error")
	}
	if !strings.Contains(err.Error(), "read embedded sql query/missing.sql") {
		t.Fatalf("expected helpful embedded query error, got %q", err.Error())
	}
}

func TestMigrationReturnsHelpfulErrorForMissingSQL(t *testing.T) {
	t.Parallel()

	_, err := Migration("missing.sql")
	if err == nil {
		t.Fatal("expected missing migration error")
	}
	if !strings.Contains(err.Error(), "read embedded sql migration/missing.sql") {
		t.Fatalf("expected helpful embedded migration error, got %q", err.Error())
	}
}

func TestMustHelpersPanicForMissingSQL(t *testing.T) {
	t.Parallel()

	assertPanics := func(name string, fn func()) {
		t.Helper()
		defer func() {
			if recover() == nil {
				t.Fatalf("expected %s to panic", name)
			}
		}()
		fn()
	}

	assertPanics("MustQuery", func() {
		MustQuery("missing.sql")
	})
	assertPanics("MustMigration", func() {
		MustMigration("missing.sql")
	})
}

func TestMustHelpersReturnEmbeddedSQL(t *testing.T) {
	t.Parallel()

	if query := MustQuery("list_clusters.sql"); !strings.Contains(query, "FROM clusters") {
		t.Fatalf("expected embedded clusters query, got %q", query)
	}

	if migration := MustMigration("001_phase1_work_objects_cluster_id.sql"); !strings.Contains(migration, "ALTER TABLE work_objects ADD COLUMN cluster_id TEXT") {
		t.Fatalf("expected embedded phase 1 migration, got %q", migration)
	}
}

func TestRenderSubstitutesTemplateValues(t *testing.T) {
	t.Parallel()

	rendered := Render("SELECT * FROM {{table}} WHERE id = '{{id}}';", map[string]string{
		"table": "signals",
		"id":    "sig_1",
	})
	if rendered != "SELECT * FROM signals WHERE id = 'sig_1';" {
		t.Fatalf("expected rendered template, got %q", rendered)
	}
}
