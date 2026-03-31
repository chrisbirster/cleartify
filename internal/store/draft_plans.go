package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	sqldata "cleartify/data"
	"cleartify/internal/domain"
)

var draftPlanNow = time.Now
var lookupWorkObjectDraftPlanStatement = sqldata.MustQuery("lookup_work_object_draft_plan.sql")
var upsertWorkObjectDraftPlanStatement = sqldata.MustQuery("upsert_work_object_draft_plan.sql")

var ErrInvalidDraftPlanFields = errors.New("invalid draft plan fields")
var ErrDraftPlanNotFound = errors.New("draft plan not found")

func (s *Store) LookupDraftPlanForWorkObject(ctx context.Context, workObjectID string) (*domain.DraftPlan, error) {
	row := s.DB.QueryRowContext(ctx, lookupWorkObjectDraftPlanStatement, workObjectID)

	var draftPlan domain.DraftPlan
	var rawSteps string
	var rawEdgeCases string
	var rawOpenQuestions string
	var createdAt string
	var updatedAt string
	if err := row.Scan(
		&draftPlan.WorkObjectID,
		&draftPlan.Overview,
		&rawSteps,
		&rawEdgeCases,
		&rawOpenQuestions,
		&createdAt,
		&updatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("lookup draft plan: %w", err)
	}

	steps, err := decodeStringSlice(rawSteps)
	if err != nil {
		return nil, fmt.Errorf("decode implementation steps: %w", err)
	}
	edgeCases, err := decodeStringSlice(rawEdgeCases)
	if err != nil {
		return nil, fmt.Errorf("decode edge cases: %w", err)
	}
	openQuestions, err := decodeStringSlice(rawOpenQuestions)
	if err != nil {
		return nil, fmt.Errorf("decode open questions: %w", err)
	}
	draftPlan.CreatedAt, err = decodeTime(createdAt)
	if err != nil {
		return nil, fmt.Errorf("decode draft plan created_at: %w", err)
	}
	draftPlan.UpdatedAt, err = decodeTime(updatedAt)
	if err != nil {
		return nil, fmt.Errorf("decode draft plan updated_at: %w", err)
	}
	draftPlan.ImplementationSteps = steps
	draftPlan.EdgeCases = edgeCases
	draftPlan.OpenQuestions = openQuestions

	return &draftPlan, nil
}

func (s *Store) GenerateDraftPlanForWorkObject(ctx context.Context, workObjectID string) (domain.DraftPlan, error) {
	workObject, err := s.lookupWorkObject(ctx, workObjectID)
	if err != nil {
		return domain.DraftPlan{}, err
	}

	var linkedCluster *domain.Cluster
	if workObject.ClusterID != "" {
		cluster, err := s.lookupClusterForWorkObject(ctx, workObject.ClusterID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return domain.DraftPlan{}, err
		}
		if err == nil {
			linkedCluster = &cluster
		}
	}

	existingPlan, err := s.LookupDraftPlanForWorkObject(ctx, workObjectID)
	if err != nil {
		return domain.DraftPlan{}, err
	}

	timestamp := draftPlanNow().UTC().Truncate(time.Second)
	createdAt := timestamp
	if existingPlan != nil {
		createdAt = existingPlan.CreatedAt
	}

	draftPlan := buildDraftPlan(workObject, linkedCluster, createdAt, timestamp)
	rawSteps, err := encodeStringSlice(draftPlan.ImplementationSteps)
	if err != nil {
		return domain.DraftPlan{}, fmt.Errorf("encode implementation steps: %w", err)
	}
	rawEdgeCases, err := encodeStringSlice(draftPlan.EdgeCases)
	if err != nil {
		return domain.DraftPlan{}, fmt.Errorf("encode edge cases: %w", err)
	}
	rawOpenQuestions, err := encodeStringSlice(draftPlan.OpenQuestions)
	if err != nil {
		return domain.DraftPlan{}, fmt.Errorf("encode open questions: %w", err)
	}

	if err := s.persistDraftPlan(ctx, draftPlan, rawSteps, rawEdgeCases, rawOpenQuestions); err != nil {
		return domain.DraftPlan{}, err
	}

	return draftPlan, nil
}

func (s *Store) UpdateDraftPlanForWorkObject(
	ctx context.Context,
	workObjectID string,
	overview string,
	implementationSteps []string,
	edgeCases []string,
	openQuestions []string,
) (domain.DraftPlan, error) {
	if _, err := s.lookupWorkObject(ctx, workObjectID); err != nil {
		return domain.DraftPlan{}, err
	}

	existingPlan, err := s.LookupDraftPlanForWorkObject(ctx, workObjectID)
	if err != nil {
		return domain.DraftPlan{}, err
	}
	if existingPlan == nil {
		return domain.DraftPlan{}, ErrDraftPlanNotFound
	}

	normalizedOverview := strings.TrimSpace(overview)
	normalizedSteps := normalizeDraftPlanItems(implementationSteps)
	normalizedEdgeCases := normalizeDraftPlanItems(edgeCases)
	normalizedOpenQuestions := normalizeDraftPlanItems(openQuestions)
	if normalizedOverview == "" || len(normalizedSteps) == 0 || len(normalizedEdgeCases) == 0 || len(normalizedOpenQuestions) == 0 {
		return domain.DraftPlan{}, ErrInvalidDraftPlanFields
	}

	updatedAt := draftPlanNow().UTC().Truncate(time.Second)
	draftPlan := domain.DraftPlan{
		WorkObjectID:        workObjectID,
		Overview:            normalizedOverview,
		ImplementationSteps: normalizedSteps,
		EdgeCases:           normalizedEdgeCases,
		OpenQuestions:       normalizedOpenQuestions,
		CreatedAt:           existingPlan.CreatedAt,
		UpdatedAt:           updatedAt,
	}

	rawSteps, err := encodeStringSlice(draftPlan.ImplementationSteps)
	if err != nil {
		return domain.DraftPlan{}, fmt.Errorf("encode implementation steps: %w", err)
	}
	rawEdgeCases, err := encodeStringSlice(draftPlan.EdgeCases)
	if err != nil {
		return domain.DraftPlan{}, fmt.Errorf("encode edge cases: %w", err)
	}
	rawOpenQuestions, err := encodeStringSlice(draftPlan.OpenQuestions)
	if err != nil {
		return domain.DraftPlan{}, fmt.Errorf("encode open questions: %w", err)
	}

	if err := s.persistDraftPlan(ctx, draftPlan, rawSteps, rawEdgeCases, rawOpenQuestions); err != nil {
		return domain.DraftPlan{}, err
	}

	return draftPlan, nil
}

func (s *Store) persistDraftPlan(ctx context.Context, draftPlan domain.DraftPlan, rawSteps string, rawEdgeCases string, rawOpenQuestions string) error {
	if _, err := s.DB.ExecContext(
		ctx,
		upsertWorkObjectDraftPlanStatement,
		draftPlan.WorkObjectID,
		draftPlan.Overview,
		rawSteps,
		rawEdgeCases,
		rawOpenQuestions,
		draftPlan.CreatedAt.Format(time.RFC3339),
		draftPlan.UpdatedAt.Format(time.RFC3339),
	); err != nil {
		return fmt.Errorf("upsert draft plan: %w", err)
	}
	return nil
}

func buildDraftPlan(workObject domain.WorkObject, linkedCluster *domain.Cluster, createdAt time.Time, updatedAt time.Time) domain.DraftPlan {
	clusterContext := "This draft plan stays grounded in the Work Object itself."
	if linkedCluster != nil {
		clusterContext = fmt.Sprintf(
			"Linked Cluster context from \"%s\" keeps the plan attached to its originating Signals and current %s severity.",
			linkedCluster.Title,
			linkedCluster.Severity,
		)
	}

	prototypeContext := "The current prototype note should be expanded before implementation starts."
	switch {
	case workObject.PrototypeURL != "" && workObject.PrototypeNote != "":
		prototypeContext = "The attached prototype link and note define the first implementation pass together."
	case workObject.PrototypeURL != "":
		prototypeContext = "The attached prototype link provides the primary visual reference for this first planning pass."
	case workObject.PrototypeNote != "":
		prototypeContext = "The attached prototype note provides the primary product reference for this first planning pass."
	}

	thirdStep := "Capture any missing context that would normally come from a linked Cluster before expanding the scope."
	thirdQuestion := "What additional Signal context should be attached before implementation moves beyond the draft?"
	thirdEdgeCase := "This Work Object has no linked Cluster, so surrounding Signal context may still be incomplete."
	if linkedCluster != nil {
		thirdStep = fmt.Sprintf(
			"Carry forward the linked Cluster context from \"%s\" so the implementation stays proportional to the current %s severity and summary.",
			linkedCluster.Title,
			linkedCluster.Severity,
		)
		thirdQuestion = fmt.Sprintf("Which Signals from \"%s\" should still shape acceptance criteria during implementation?", linkedCluster.Title)
		thirdEdgeCase = fmt.Sprintf("The linked Cluster is currently marked %s severity, so the implementation scope should stay proportional to that context.", linkedCluster.Severity)
	}

	fourthStep := "Use the attached prototype note as the review reference while the first implementation pass is clarified."
	firstQuestion := "Which concrete prototype states should be inferred from the attached note before execution starts?"
	firstEdgeCase := "The Work Object currently relies on a prototype note without a linked prototype URL."
	if workObject.PrototypeURL != "" {
		fourthStep = "Use the attached prototype link as the review reference while the first implementation pass is clarified."
		firstQuestion = "Which prototype screens or states must ship in the first implementation pass?"
		firstEdgeCase = "The linked prototype may drift if the reference changes after planning starts."
	}

	secondEdgeCase := "The prototype note is still empty, so rationale may need to be captured during planning."
	secondQuestion := "What product judgment should be written down alongside the prototype reference before execution begins?"
	if workObject.PrototypeNote != "" {
		secondEdgeCase = "The prototype note may still leave some interaction details or handoff expectations unresolved."
		secondQuestion = "What implementation details are still implied, but not stated explicitly, in the prototype note?"
	}

	return domain.DraftPlan{
		WorkObjectID: workObject.ID,
		Overview: fmt.Sprintf(
			"This draft plan turns \"%s\" into executable work using the current summary, goal, and prototype context. %s %s",
			workObject.Title,
			prototypeContext,
			clusterContext,
		),
		ImplementationSteps: []string{
			fmt.Sprintf("Review the Work Object goal for \"%s\" and keep the first pass aligned with: %s", workObject.Title, workObject.Goal),
			fmt.Sprintf("Translate the current summary into the smallest useful implementation slice: %s", workObject.Summary),
			thirdStep,
			fourthStep,
		},
		EdgeCases: []string{
			firstEdgeCase,
			secondEdgeCase,
			thirdEdgeCase,
		},
		OpenQuestions: []string{
			firstQuestion,
			secondQuestion,
			thirdQuestion,
		},
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

func encodeStringSlice(values []string) (string, error) {
	encoded, err := json.Marshal(values)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func decodeStringSlice(raw string) ([]string, error) {
	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return nil, err
	}
	return values, nil
}

func normalizeDraftPlanItems(items []string) []string {
	normalized := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		normalized = append(normalized, trimmed)
	}
	return normalized
}
