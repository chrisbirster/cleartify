package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	sqldata "cleartify/data"

	"cleartify/internal/auth"
	"cleartify/internal/domain"
)

var upsertAuthUserStatement = sqldata.MustQuery("upsert_auth_user.sql")
var lookupAuthSessionStatement = sqldata.MustQuery("lookup_auth_session.sql")
var upsertAuthSessionStatement = sqldata.MustQuery("upsert_auth_session.sql")
var deleteAuthSessionStatement = sqldata.MustQuery("delete_auth_session.sql")
var markUserOnboardingCompleteStatement = sqldata.MustQuery("mark_user_onboarding_complete.sql")
var lookupWorkspaceSettingsStatement = sqldata.MustQuery("lookup_workspace_settings.sql")
var upsertWorkspaceSettingsStatement = sqldata.MustQuery("upsert_workspace_settings.sql")
var authSessionNow = time.Now

func (s *Store) UpsertAuthUser(ctx context.Context, user domain.User) error {
	createdAt := authSessionNow().UTC().Truncate(time.Second).Format(time.RFC3339)
	if _, err := s.DB.ExecContext(
		ctx,
		upsertAuthUserStatement,
		strings.TrimSpace(user.ID),
		strings.TrimSpace(user.Email),
		strings.TrimSpace(user.DisplayName),
		createdAt,
	); err != nil {
		return fmt.Errorf("upsert auth user: %w", err)
	}
	return nil
}

func (s *Store) LookupAuthSession(ctx context.Context, sessionID string) (auth.SessionRecord, error) {
	row := s.DB.QueryRowContext(ctx, lookupAuthSessionStatement, strings.TrimSpace(sessionID))

	var (
		record                auth.SessionRecord
		userID                string
		userEmail             string
		userDisplayName       string
		onboardingCompletedAt string
		refreshToken          string
		workOSSessionID       string
		organizationID        string
		organizationName      string
		accessTokenExpiresAt  string
	)

	if err := row.Scan(
		&record.ID,
		&userID,
		&userEmail,
		&userDisplayName,
		&onboardingCompletedAt,
		&refreshToken,
		&workOSSessionID,
		&organizationID,
		&organizationName,
		&accessTokenExpiresAt,
	); err != nil {
		return auth.SessionRecord{}, err
	}

	record.User = domain.User{
		ID:          userID,
		Email:       userEmail,
		DisplayName: userDisplayName,
	}
	record.RefreshToken = refreshToken
	record.WorkOSSessionID = workOSSessionID
	record.OrganizationID = organizationID

	if organizationID != "" || organizationName != "" {
		record.Organization = &domain.Organization{
			ID:   organizationID,
			Name: organizationName,
			Slug: slugify(storeFirstNonEmpty(organizationName, organizationID)),
		}
		record.Workspace = &domain.Workspace{
			ID:   storeFirstNonEmpty(organizationID, "workspace_"+userID),
			Name: storeFirstNonEmpty(organizationName, "Current workspace"),
			Slug: slugify(storeFirstNonEmpty(organizationName, organizationID, "current-workspace")),
		}
	} else {
		record.Workspace = &domain.Workspace{
			ID:   "workspace_" + userID,
			Name: userDisplayName + "'s workspace",
			Slug: slugify(userDisplayName + " workspace"),
		}
	}
	if record.Workspace != nil {
		workspaceMode, sampleInitializedAt, liveInitializedAt, err := s.lookupWorkspaceSettings(ctx, record.Workspace.ID)
		if err != nil {
			return auth.SessionRecord{}, err
		}
		record.Workspace.Mode = workspaceMode
		record.Workspace.SampleInitializedAt = sampleInitializedAt
		record.Workspace.LiveInitializedAt = liveInitializedAt
	}
	record.NeedsOnboarding = onboardingCompletedAt == "" || record.Workspace == nil || strings.TrimSpace(record.Workspace.Mode) == ""

	if accessTokenExpiresAt != "" {
		expiresAt, err := decodeTime(accessTokenExpiresAt)
		if err != nil {
			return auth.SessionRecord{}, fmt.Errorf("decode session expiry: %w", err)
		}
		record.AccessTokenExpiresAt = &expiresAt
	}

	return record, nil
}

func (s *Store) UpsertAuthSession(ctx context.Context, session auth.SessionRecord) error {
	expiresAt := ""
	if session.AccessTokenExpiresAt != nil {
		expiresAt = session.AccessTokenExpiresAt.UTC().Format(time.RFC3339)
	}

	if _, err := s.DB.ExecContext(
		ctx,
		upsertAuthSessionStatement,
		strings.TrimSpace(session.ID),
		strings.TrimSpace(session.User.ID),
		authSessionNow().UTC().Truncate(time.Second).Format(time.RFC3339),
		strings.TrimSpace(session.RefreshToken),
		strings.TrimSpace(session.WorkOSSessionID),
		strings.TrimSpace(session.OrganizationID),
		strings.TrimSpace(workspaceName(session)),
		expiresAt,
	); err != nil {
		return fmt.Errorf("upsert auth session: %w", err)
	}
	return nil
}

func (s *Store) DeleteAuthSession(ctx context.Context, sessionID string) error {
	if _, err := s.DB.ExecContext(ctx, deleteAuthSessionStatement, strings.TrimSpace(sessionID)); err != nil {
		return fmt.Errorf("delete auth session: %w", err)
	}
	return nil
}

func (s *Store) SetWorkspaceMode(ctx context.Context, workspaceID string, mode string, selectedAt time.Time) error {
	workspaceID = strings.TrimSpace(workspaceID)
	mode = strings.TrimSpace(mode)
	if workspaceID == "" || mode == "" {
		return fmt.Errorf("workspace settings require workspace id and mode")
	}

	timestamp := selectedAt.UTC().Truncate(time.Second).Format(time.RFC3339)
	sampleInitializedAt := ""
	liveInitializedAt := ""
	switch mode {
	case "sample":
		sampleInitializedAt = timestamp
	case "live":
		liveInitializedAt = timestamp
	}

	if _, err := s.DB.ExecContext(
		ctx,
		upsertWorkspaceSettingsStatement,
		workspaceID,
		mode,
		sampleInitializedAt,
		liveInitializedAt,
		timestamp,
		timestamp,
	); err != nil {
		return fmt.Errorf("set workspace mode: %w", err)
	}

	return nil
}

func (s *Store) EnsureSampleData() error {
	return s.SeedDemo()
}

func (s *Store) MarkUserOnboardingComplete(ctx context.Context, userID string, completedAt time.Time) error {
	result, err := s.DB.ExecContext(
		ctx,
		markUserOnboardingCompleteStatement,
		completedAt.UTC().Format(time.RFC3339),
		strings.TrimSpace(userID),
	)
	if err != nil {
		return fmt.Errorf("mark onboarding complete: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("mark onboarding complete rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s *Store) lookupWorkspaceSettings(ctx context.Context, workspaceID string) (string, *time.Time, *time.Time, error) {
	row := s.DB.QueryRowContext(ctx, lookupWorkspaceSettingsStatement, strings.TrimSpace(workspaceID))

	var (
		mode                string
		sampleInitializedAt string
		liveInitializedAt   string
	)
	if err := row.Scan(&mode, &sampleInitializedAt, &liveInitializedAt); err != nil {
		if err == sql.ErrNoRows {
			return "", nil, nil, nil
		}
		return "", nil, nil, fmt.Errorf("lookup workspace settings: %w", err)
	}

	decodedSampleInitializedAt, err := decodeNullableTime(sampleInitializedAt)
	if err != nil {
		return "", nil, nil, fmt.Errorf("decode sample initialized at: %w", err)
	}
	decodedLiveInitializedAt, err := decodeNullableTime(liveInitializedAt)
	if err != nil {
		return "", nil, nil, fmt.Errorf("decode live initialized at: %w", err)
	}

	return strings.TrimSpace(mode), decodedSampleInitializedAt, decodedLiveInitializedAt, nil
}

func workspaceName(session auth.SessionRecord) string {
	if session.Workspace != nil && strings.TrimSpace(session.Workspace.Name) != "" {
		return session.Workspace.Name
	}
	if session.Organization != nil && strings.TrimSpace(session.Organization.Name) != "" {
		return session.Organization.Name
	}
	return ""
}

func slugify(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return ""
	}

	var builder strings.Builder
	lastDash := false
	for _, char := range value {
		switch {
		case char >= 'a' && char <= 'z':
			builder.WriteRune(char)
			lastDash = false
		case char >= '0' && char <= '9':
			builder.WriteRune(char)
			lastDash = false
		default:
			if !lastDash {
				builder.WriteRune('-')
				lastDash = true
			}
		}
	}

	return strings.Trim(builder.String(), "-")
}

func storeFirstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
