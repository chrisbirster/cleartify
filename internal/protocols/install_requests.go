package protocols

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chrisbirster/cleartify/internal/database"
	"strings"
	"time"
)

var ErrInstallUnavailable = errors.New("published protocol or valid permission manifest unavailable")
var ErrManifestChanged = errors.New("permission manifest digest mismatch")
var ErrInstallDecision = errors.New("install request is not pending or publication changed")

type InstallService struct {
	Store                 *database.Store
	StrictPermissions     bool
	PaidEnforcement       bool
	PermissionDiffEnabled bool
	HoldVisibility        bool
}
type Manifest struct {
	PublicationID     string          `json:"publicationId"`
	Title             string          `json:"title"`
	Version           int             `json:"version"`
	Permissions       []string        `json:"permissions"`
	SpecSHA256        string          `json:"specSha256"`
	ManifestSHA256    string          `json:"manifestSha256"`
	PermissionChanges *PermissionDiff `json:"permissionChanges,omitempty"`
}
type installSnapshot struct {
	ID       string
	Manifest Manifest
	Raw      string
}

func sha(raw []byte) string { digest := sha256.Sum256(raw); return hex.EncodeToString(digest[:]) }

type manifestQuerier interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func (s InstallService) manifest(ctx context.Context, id string) (installSnapshot, error) {
	return s.manifestFrom(ctx, s.Store.DB, id)
}
func (s InstallService) manifestFrom(ctx context.Context, query manifestQuerier, id string) (installSnapshot, error) {
	var x installSnapshot
	if id == "" || len(id) > 128 {
		return x, ErrInstallUnavailable
	}
	var spec, title, slug string
	var version int
	access := AccessiblePublicationSQL
	if s.HoldVisibility {
		access += " AND " + PublicationHoldSQL
	}
	err := query.QueryRowContext(ctx, `SELECT title,slug,version,spec_json FROM published_protocols WHERE id=? AND `+access, id).Scan(&title, &slug, &version, &spec)
	if errors.Is(err, sql.ErrNoRows) {
		return x, ErrInstallUnavailable
	}
	if err != nil {
		return x, err
	}
	if len(spec) > 65536 || version <= 0 || strings.TrimSpace(slug) == "" {
		return x, ErrInstallUnavailable
	}
	normalized, err := normalizePermissionSpec([]byte(spec))
	if err != nil {
		return x, ErrInstallUnavailable
	}
	if s.PermissionDiffEnabled && AnalyzeManifest(normalized) != nil {
		return x, ErrInstallUnavailable
	}
	if s.StrictPermissions && ValidatePreviewPermissions(normalized) != nil {
		return x, ErrInstallUnavailable
	}
	var raw struct {
		Permissions []string `json:"permissions"`
	}
	if err = json.Unmarshal(normalized, &raw); err != nil || raw.Permissions == nil || len(raw.Permissions) > 32 {
		return x, ErrInstallUnavailable
	}
	seen := map[string]bool{}
	for _, p := range raw.Permissions {
		if p == "" || len(p) > 100 || seen[p] || strings.TrimSpace(p) != p {
			return x, ErrInstallUnavailable
		}
		seen[p] = true
	}
	if raw.Permissions == nil {
		raw.Permissions = []string{}
	}
	x.ID = id
	x.Manifest = Manifest{PublicationID: id, Title: title, Version: version, Permissions: raw.Permissions, SpecSHA256: sha([]byte(spec))}
	// Canonical snapshot: fixed struct field order, including immutable spec digest.
	stable, err := json.Marshal(struct {
		PublicationID string   `json:"publicationId"`
		Version       int      `json:"version"`
		Permissions   []string `json:"permissions"`
		SpecSHA256    string   `json:"specSha256"`
	}{id, version, raw.Permissions, x.Manifest.SpecSHA256})
	if err != nil {
		return x, err
	}
	x.Manifest.ManifestSHA256 = sha(stable)
	if s.PermissionDiffEnabled {
		var previous string
		prevErr := query.QueryRowContext(ctx, `SELECT spec_json FROM published_protocols WHERE creator_did=(SELECT creator_did FROM published_protocols WHERE id=?) AND slug=? AND version<? ORDER BY version DESC LIMIT 1`, id, slug, version).Scan(&previous)
		if prevErr != nil && !errors.Is(prevErr, sql.ErrNoRows) {
			return x, prevErr
		}
		if prevErr == nil {
			var old struct {
				Permissions []string `json:"permissions"`
			}
			normalizedPrevious, parseErr := normalizePermissionSpec([]byte(previous))
			if parseErr != nil || json.Unmarshal(normalizedPrevious, &old) != nil || old.Permissions == nil {
				return x, ErrInstallUnavailable
			}
			diff := DiffPermissions(old.Permissions, raw.Permissions)
			x.Manifest.PermissionChanges = &diff
		}
	}
	permissions, err := json.Marshal(raw.Permissions)
	if err != nil {
		return x, err
	}
	x.Raw = string(permissions)
	return x, nil
}
func (s InstallService) Preview(ctx context.Context, publicationID string) (Manifest, error) {
	if s.PaidEnforcement {
		var count int
		if err := s.Store.DB.QueryRowContext(ctx, "SELECT count(*) FROM marketplace_listings WHERE protocol_id=? AND pricing_kind='one_time' AND price_minor>0", publicationID).Scan(&count); err != nil {
			return Manifest{}, err
		}
		if count > 0 {
			return Manifest{}, ErrInstallUnavailable
		}
	}
	x, err := s.manifest(ctx, publicationID)
	return x.Manifest, err
}
func (s InstallService) Request(ctx context.Context, actorDID, publicationID, ackDigest string) (string, error) {
	if !strings.HasPrefix(actorDID, "did:") || len(actorDID) > 2048 {
		return "", ErrInstallUnavailable
	}
	tx, err := s.Store.DB.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	x, err := s.manifestFrom(ctx, tx, publicationID)
	if err != nil {
		return "", err
	}
	blocked, err := installBlocked(ctx, tx, actorDID, publicationID)
	if err != nil {
		return "", err
	}
	if blocked {
		return "", ErrInstallUnavailable
	}
	if ackDigest != x.Manifest.ManifestSHA256 {
		return "", ErrManifestChanged
	}
	idBytes := make([]byte, 18)
	if _, err = rand.Read(idBytes); err != nil {
		return "", err
	}
	id := hex.EncodeToString(idBytes)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err = tx.ExecContext(ctx, `INSERT INTO protocol_install_requests(id,actor_did,publication_id,protocol_version,spec_sha256,manifest_json,manifest_sha256,acknowledged_at,status,requested_at)
 VALUES(?,?,?,?,?,?,?,?,'pending_approval',?)`, id, actorDID, publicationID, x.Manifest.Version, x.Manifest.SpecSHA256, x.Raw, x.Manifest.ManifestSHA256, now, now)
	if err != nil {
		return "", err
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	if err = s.Store.Push(ctx); err != nil {
		return "", err
	}
	return id, nil
}
func (s InstallService) Decide(ctx context.Context, id string, approve bool) error {
	if len(id) != 36 {
		return ErrInstallDecision
	}
	tx, err := s.Store.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var publicationID, storedSHA, actorDID string
	var version int
	err = tx.QueryRowContext(ctx, `SELECT publication_id,protocol_version,spec_sha256,actor_did FROM protocol_install_requests WHERE id=? AND status='pending_approval'`, id).Scan(&publicationID, &version, &storedSHA, &actorDID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrInstallDecision
	}
	if err != nil {
		return err
	}
	blocked, err := installBlocked(ctx, tx, actorDID, publicationID)
	if err != nil {
		return err
	}
	if approve && blocked {
		return ErrInstallDecision
	}
	m, err := s.manifestFrom(ctx, tx, publicationID)
	if errors.Is(err, ErrInstallUnavailable) {
		return ErrInstallDecision
	}
	if err != nil {
		return err
	}
	if m.Manifest.Version != version || m.Manifest.SpecSHA256 != storedSHA {
		return ErrInstallDecision
	}
	status := "rejected"
	if approve {
		status = "approved"
	}
	res, err := tx.ExecContext(ctx, `UPDATE protocol_install_requests SET status=?,decided_at=? WHERE id=? AND status='pending_approval' AND protocol_version=? AND spec_sha256=?`, status, time.Now().UTC().Format(time.RFC3339Nano), id, version, storedSHA)
	if err != nil {
		return fmt.Errorf("decide install: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return ErrInstallDecision
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return s.Store.Push(ctx)
}

// Authoring treats omitted/null permissions as empty. Normalize only for
// manifest validation/serialization; the pinned digest still covers raw spec.
func normalizePermissionSpec(raw []byte) ([]byte, error) {
	var body map[string]json.RawMessage
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, err
	}
	if body == nil {
		return nil, ErrInstallUnavailable
	}
	if value, ok := body["permissions"]; !ok || string(value) == "null" {
		body["permissions"] = json.RawMessage(`[]`)
	}
	return json.Marshal(body)
}

func installBlocked(ctx context.Context, q manifestQuerier, actor, publication string) (bool, error) {
	var n int
	err := q.QueryRowContext(ctx, `SELECT count(*) FROM social_blocks b JOIN published_protocols p ON p.id=?
 WHERE (b.actor_did=? AND b.subject_did=p.creator_did) OR (b.actor_did=p.creator_did AND b.subject_did=?)`, publication, actor, actor).Scan(&n)
	return n > 0, err
}
