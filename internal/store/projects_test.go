package store

import (
	"context"
	"testing"
)

func TestListProjectsReturnsSeededSampleProject(t *testing.T) {
	t.Parallel()

	store := newTestStore(t, true)

	projects, err := store.ListProjects(context.Background())
	if err != nil {
		t.Fatalf("expected projects, got error: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 seeded project, got %d", len(projects))
	}
	if projects[0].ID != sampleProjectID {
		t.Fatalf("expected seeded project id %q, got %q", sampleProjectID, projects[0].ID)
	}

	signals, err := store.ListSignals(context.Background())
	if err != nil {
		t.Fatalf("expected signals, got error: %v", err)
	}
	for _, signal := range signals {
		if signal.ProjectID != sampleProjectID {
			t.Fatalf("expected signal %q to be scoped to %q, got %q", signal.ID, sampleProjectID, signal.ProjectID)
		}
	}

	clusters, err := store.ListClusters(context.Background())
	if err != nil {
		t.Fatalf("expected clusters, got error: %v", err)
	}
	for _, cluster := range clusters {
		if cluster.ProjectID != sampleProjectID {
			t.Fatalf("expected cluster %q to be scoped to %q, got %q", cluster.ID, sampleProjectID, cluster.ProjectID)
		}
	}

	workObjects, err := store.ListWorkObjects(context.Background())
	if err != nil {
		t.Fatalf("expected work objects, got error: %v", err)
	}
	for _, workObject := range workObjects {
		if workObject.ProjectID != sampleProjectID {
			t.Fatalf("expected work object %q to be scoped to %q, got %q", workObject.ID, sampleProjectID, workObject.ProjectID)
		}
	}
}

func TestCreateAndUpdateProject(t *testing.T) {
	t.Parallel()

	store := newTestStore(t, false)

	project, err := store.CreateProject(context.Background(), "wrk_live", "Onboarding refresh", "Ship the calmer first-run path.", "")
	if err != nil {
		t.Fatalf("expected project create to succeed: %v", err)
	}
	if project.Status != "active" {
		t.Fatalf("expected default status active, got %q", project.Status)
	}
	if project.Slug != "onboarding-refresh" {
		t.Fatalf("expected slug to be normalized, got %q", project.Slug)
	}

	updatedProject, err := store.UpdateProject(context.Background(), project.ID, "Onboarding refresh v2", "Narrow the scope to setup success.", "paused")
	if err != nil {
		t.Fatalf("expected project update to succeed: %v", err)
	}
	if updatedProject.Status != "paused" {
		t.Fatalf("expected updated status paused, got %q", updatedProject.Status)
	}
	if updatedProject.Name != "Onboarding refresh v2" {
		t.Fatalf("expected updated name, got %q", updatedProject.Name)
	}
}
