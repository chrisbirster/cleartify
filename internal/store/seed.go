package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	sqldata "cleartify/data"
)

var demoTimestamp = time.Date(2026, 3, 1, 9, 0, 0, 0, time.UTC)
var jsonMarshal = json.Marshal
var beginTx = func(db *sql.DB, ctx context.Context) (*sql.Tx, error) {
	return db.BeginTx(ctx, nil)
}

var seedInsertUserStatement = sqldata.MustQuery("seed_insert_user.sql")
var seedInsertSessionStatement = sqldata.MustQuery("seed_insert_session.sql")
var seedInsertSignalStatement = sqldata.MustQuery("seed_insert_signal.sql")
var seedInsertClusterStatement = sqldata.MustQuery("seed_insert_cluster.sql")
var seedInsertClusterSignalStatement = sqldata.MustQuery("seed_insert_cluster_signal.sql")
var seedInsertWorkObjectStatement = sqldata.MustQuery("seed_insert_work_object.sql")
var seedInsertDecisionStatement = sqldata.MustQuery("seed_insert_decision.sql")
var seedInsertArtifactStatement = sqldata.MustQuery("seed_insert_artifact.sql")

func (s *Store) SeedDemo() error {
	ctx := context.Background()
	count, err := countRows(ctx, s.DB, "signals")
	if err != nil {
		return fmt.Errorf("count signals: %w", err)
	}
	if count > 0 {
		return nil
	}

	tx, err := beginTx(s.DB, ctx)
	if err != nil {
		return fmt.Errorf("begin seed tx: %w", err)
	}
	defer tx.Rollback()

	if err := seedUser(tx); err != nil {
		return err
	}
	if err := seedSignals(tx); err != nil {
		return err
	}
	if err := seedClusters(tx); err != nil {
		return err
	}
	if err := seedWorkObjects(tx); err != nil {
		return err
	}

	return tx.Commit()
}

func seedUser(tx *sql.Tx) error {
	_, err := tx.Exec(
		seedInsertUserStatement,
		"usr_demo",
		"demo@cleartify.com",
		"Demo User",
		demoTimestamp.Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("seed user: %w", err)
	}

	_, err = tx.Exec(
		seedInsertSessionStatement,
		"ses_demo",
		"usr_demo",
		demoTimestamp.Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("seed session: %w", err)
	}

	return nil
}

func seedSignals(tx *sql.Tx) error {
	type signalSeed struct {
		ID         string
		SourceType string
		SourceID   string
		Title      string
		Body       string
		URL        string
		Metadata   map[string]any
		CreatedAt  time.Time
	}

	signals := []signalSeed{
		{"sig_mobile_feedback", "feedback", "fb-100", "Mobile performance issues reported", "Dashboard feels slow on mobile during chart rendering.", "https://example.com/feedback/100", map[string]any{"priority": "high"}, demoTimestamp},
		{"sig_design_thread", "slack", "sl-222", "Design review thread from #product", "Team discussed a lighter onboarding sequence for first-run users.", "https://example.com/slack/222", map[string]any{"channel": "#product"}, demoTimestamp.Add(5 * time.Minute)},
		{"sig_pr_context", "github", "pr-234", "PR #234 needs context from original spec", "Code review is missing the original rationale for loading-state behavior.", "https://example.com/pr/234", map[string]any{"repo": "cleartify/web"}, demoTimestamp.Add(10 * time.Minute)},
		{"sig_comp_notes", "note", "nt-45", "Competitor analysis notes from team sync", "Notes highlight how competitor teams connect prototypes to shipped changes.", "", map[string]any{"author": "Demo User"}, demoTimestamp.Add(15 * time.Minute)},
		{"sig_proto_attach", "prototype", "pt-17", "Prototype attached for onboarding refresh", "Five mobile onboarding screens were attached for implementation planning.", "https://example.com/prototype/17", map[string]any{"screens": 5.0}, demoTimestamp.Add(20 * time.Minute)},
		{"sig_release_post", "release", "rl-9", "Release retro flagged missing product memory", "Recent release notes lost the decision trail behind a copy change.", "", map[string]any{"release": "2026.03.1"}, demoTimestamp.Add(25 * time.Minute)},
	}

	for _, signal := range signals {
		metadataJSON, err := jsonMarshal(signal.Metadata)
		if err != nil {
			return fmt.Errorf("marshal signal metadata: %w", err)
		}

		if _, err := tx.Exec(
			seedInsertSignalStatement,
			signal.ID,
			signal.SourceType,
			signal.SourceID,
			signal.Title,
			signal.Body,
			signal.URL,
			string(metadataJSON),
			signal.CreatedAt.Format(time.RFC3339),
		); err != nil {
			return fmt.Errorf("seed signal %s: %w", signal.ID, err)
		}
	}

	return nil
}

func seedClusters(tx *sql.Tx) error {
	type clusterSeed struct {
		ID        string
		Title     string
		Summary   string
		Severity  string
		Status    string
		SignalIDs []string
	}

	clusters := []clusterSeed{
		{
			ID:        "clu_mobile_perf",
			Title:     "Mobile performance improvements",
			Summary:   "Signals point to chart-heavy flows feeling slow on mobile devices.",
			Severity:  "high",
			Status:    "active",
			SignalIDs: []string{"sig_mobile_feedback", "sig_pr_context"},
		},
		{
			ID:        "clu_onboarding_refresh",
			Title:     "Onboarding refresh",
			Summary:   "Design review and prototype signals suggest a lighter onboarding flow.",
			Severity:  "medium",
			Status:    "active",
			SignalIDs: []string{"sig_design_thread", "sig_proto_attach"},
		},
		{
			ID:        "clu_product_memory",
			Title:     "Product memory and release traceability",
			Summary:   "Retros and notes highlight missing rationale in review and release flows.",
			Severity:  "medium",
			Status:    "active",
			SignalIDs: []string{"sig_comp_notes", "sig_release_post"},
		},
	}

	for _, cluster := range clusters {
		if _, err := tx.Exec(
			seedInsertClusterStatement,
			cluster.ID,
			cluster.Title,
			cluster.Summary,
			cluster.Severity,
			cluster.Status,
			demoTimestamp.Format(time.RFC3339),
			demoTimestamp.Add(30*time.Minute).Format(time.RFC3339),
		); err != nil {
			return fmt.Errorf("seed cluster %s: %w", cluster.ID, err)
		}

		for _, signalID := range cluster.SignalIDs {
			if _, err := tx.Exec(
				seedInsertClusterSignalStatement,
				cluster.ID,
				signalID,
			); err != nil {
				return fmt.Errorf("seed cluster signal %s/%s: %w", cluster.ID, signalID, err)
			}
		}
	}

	return nil
}

func seedWorkObjects(tx *sql.Tx) error {
	type workObjectSeed struct {
		ClusterID      string
		ID             string
		Title          string
		Summary        string
		Goal           string
		Status         string
		Priority       string
		OwnerType      string
		OwnerID        string
		PrototypeURL   string
		PrototypeNote  string
		ReleaseReadyAt string
	}

	workObjects := []workObjectSeed{
		{
			ClusterID:     "clu_mobile_perf",
			ID:            "wo_mobile_perf",
			Title:         "Speed up mobile dashboard rendering",
			Summary:       "Improve dashboard responsiveness on small screens without losing analytical depth.",
			Goal:          "Reduce perceived latency for mobile dashboard visits.",
			Status:        "planning",
			Priority:      "high",
			OwnerType:     "user",
			OwnerID:       "usr_demo",
			PrototypeURL:  "https://example.com/prototype/mobile-perf",
			PrototypeNote: "",
		},
		{
			ClusterID:     "clu_onboarding_refresh",
			ID:            "wo_onboarding_refresh",
			Title:         "Prototype-to-Plan onboarding refresh",
			Summary:       "Translate the new onboarding prototype into a reviewed implementation plan.",
			Goal:          "Ship a cleaner first-run experience grounded in the new prototype set.",
			Status:        "in_review",
			Priority:      "medium",
			OwnerType:     "user",
			OwnerID:       "usr_demo",
			PrototypeURL:  "https://example.com/prototype/onboarding-refresh",
			PrototypeNote: "Five approved onboarding screens anchor the first Prototype-to-Plan pass for this Work Object.",
		},
	}

	for _, workObject := range workObjects {
		if _, err := tx.Exec(
			seedInsertWorkObjectStatement,
			workObject.ID,
			workObject.ClusterID,
			workObject.Title,
			workObject.Summary,
			workObject.Goal,
			workObject.Status,
			workObject.Priority,
			workObject.OwnerType,
			workObject.OwnerID,
			workObject.PrototypeURL,
			workObject.PrototypeNote,
			workObject.ReleaseReadyAt,
			demoTimestamp.Format(time.RFC3339),
			demoTimestamp.Add(45*time.Minute).Format(time.RFC3339),
		); err != nil {
			return fmt.Errorf("seed work object %s: %w", workObject.ID, err)
		}
	}

	decisions := []struct {
		ID, WorkObjectID, Title, Summary, Rationale string
	}{
		{"dec_mobile_virtualize", "wo_mobile_perf", "Virtualize heavy dashboard sections", "Prioritize virtualization before a full chart rewrite.", "It improves mobile responsiveness quickly and reduces scope risk."},
		{"dec_onboarding_scope", "wo_onboarding_refresh", "Keep the onboarding refresh to five key screens", "Limit the first pass to the prototype set already reviewed.", "This keeps the plan aligned with approved design context."},
	}
	for _, decision := range decisions {
		if _, err := tx.Exec(
			seedInsertDecisionStatement,
			decision.ID,
			decision.WorkObjectID,
			decision.Title,
			decision.Summary,
			decision.Rationale,
			demoTimestamp.Add(50*time.Minute).Format(time.RFC3339),
		); err != nil {
			return fmt.Errorf("seed decision %s: %w", decision.ID, err)
		}
	}

	artifacts := []struct {
		ID, WorkObjectID, ArtifactType, Title, Body string
		Metadata                                    map[string]any
	}{
		{"art_mobile_plan", "wo_mobile_perf", "plan", "Mobile optimization implementation plan", "Draft plan generated from mobile performance context.", map[string]any{"source": "prototype-to-plan"}},
		{"art_onboarding_notes", "wo_onboarding_refresh", "review_notes", "Onboarding review checklist", "Review notes for the five-screen onboarding refresh.", map[string]any{"reviewer": "Demo User"}},
	}
	for _, artifact := range artifacts {
		metadataJSON, err := jsonMarshal(artifact.Metadata)
		if err != nil {
			return fmt.Errorf("marshal artifact metadata: %w", err)
		}
		if _, err := tx.Exec(
			seedInsertArtifactStatement,
			artifact.ID,
			artifact.WorkObjectID,
			artifact.ArtifactType,
			artifact.Title,
			artifact.Body,
			string(metadataJSON),
			demoTimestamp.Add(55*time.Minute).Format(time.RFC3339),
		); err != nil {
			return fmt.Errorf("seed artifact %s: %w", artifact.ID, err)
		}
	}

	return nil
}
