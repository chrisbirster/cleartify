package protocols

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var ErrPublishingInvalid = errors.New("publication rejected")
var ErrPublishingConflict = errors.New("publication transition conflict")

// PublishDraft transitions immutable drafts through private, unlisted and public
// with one transaction. Visibility is authoritative in creator_publications.
func (s DraftService) PublishDraft(ctx context.Context, actor, id, target, provenance string) error {
	if s.DB == nil || !strings.HasPrefix(actor, "did:") || len(actor) > 256 || id == "" || len(id) > 256 || len(provenance) > 500 || strings.TrimSpace(provenance) == "" {
		return ErrPublishingInvalid
	}
	if target != "private" && target != "unlisted" && target != "public" {
		return ErrPublishingInvalid
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var owner, status, license, spec string
	err = tx.QueryRowContext(ctx, "SELECT creator_did,status,license,spec_json FROM published_protocols WHERE id=?", id).Scan(&owner, &status, &license, &spec)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrPublishingConflict
	}
	if err != nil {
		return err
	}
	if owner != actor {
		return ErrPublishingConflict
	}
	// Existing draft validation is repeated against immutable data at publication time.
	var validationErr error
	if s.GuideStepsEnabled {
		validationErr = ValidateDraftSpecWithGuides([]byte(spec), license)
	} else {
		validationErr = ValidateDraftSpec([]byte(spec), license)
	}
	if validationErr != nil {
		return ErrPublishingInvalid
	}
	if license != "CC0-1.0" && license != "CC-BY-4.0" && license != "MIT" && license != "Apache-2.0" {
		return ErrPublishingInvalid
	}
	var input struct {
		RemixOf     string `json:"remixOf"`
		Attribution string `json:"attribution"`
	}
	if err = json.Unmarshal([]byte(spec), &input); err != nil {
		return ErrPublishingInvalid
	}
	if input.RemixOf != "" {
		if input.RemixOf == id {
			return ErrPublishingInvalid
		}
		var sourceLicense, sourceStatus, sourceVisibility string
		err = tx.QueryRowContext(ctx, `SELECT p.license,p.status,c.visibility FROM published_protocols p JOIN creator_publications c ON c.protocol_id=p.id WHERE p.id=?`, input.RemixOf).Scan(&sourceLicense, &sourceStatus, &sourceVisibility)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPublishingInvalid
		}
		if err != nil {
			return err
		}
		if sourceStatus != "published" || sourceVisibility == "private" || len(strings.TrimSpace(input.Attribution)) < 3 {
			return ErrPublishingInvalid
		}
		// Conservative compatibility; legal validity and provenance still need owner review.
		if sourceLicense != "CC0-1.0" && !(sourceLicense == "CC-BY-4.0" && license == "CC-BY-4.0") {
			return ErrPublishingInvalid
		}
	}
	var previous string
	err = tx.QueryRowContext(ctx, "SELECT visibility FROM creator_publications WHERE protocol_id=?", id).Scan(&previous)
	if errors.Is(err, sql.ErrNoRows) {
		previous = "draft"
	} else if err != nil {
		return err
	}
	// No backward changes to immutable visibility in this first workflow.
	valid := (previous == "draft" && target == "private") || (previous == "private" && (target == "unlisted" || target == "public")) || (previous == "unlisted" && target == "public")
	if !valid && previous != target {
		return ErrPublishingConflict
	}
	var activeHold int
	if err = tx.QueryRowContext(ctx, "SELECT count(*) FROM creator_takedown_requests WHERE protocol_id=? AND status='hidden'", id).Scan(&activeHold); err != nil {
		return err
	}
	if activeHold > 0 {
		return ErrPublishingConflict
	}
	if previous == target {
		if err = tx.Commit(); err != nil {
			return err
		}
		if s.Push != nil {
			return s.Push(ctx)
		}
		return nil
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if previous == "draft" {
		_, err = tx.ExecContext(ctx, `INSERT INTO creator_publications(protocol_id,visibility,original_source_id,provenance_note,approved_at,updated_at) VALUES(?,?,?,?,?,?)`, id, target, input.RemixOf, provenance, now, now)
	} else {
		_, err = tx.ExecContext(ctx, `UPDATE creator_publications SET visibility=?,updated_at=? WHERE protocol_id=? AND visibility=?`, target, now, id, previous)
	}
	if err != nil {
		return err
	}
	if previous == "draft" && input.RemixOf != "" {
		_, err = tx.ExecContext(ctx, `INSERT INTO creator_remix_lineage(protocol_id,parent_protocol_id,attribution,recorded_at) VALUES(?,?,?,?)`, id, input.RemixOf, input.Attribution, now)
		if err != nil {
			return err
		}
	}
	if target != "private" {
		result, e := tx.ExecContext(ctx, `UPDATE published_protocols SET status='published' WHERE id=? AND status IN ('draft','published')`, id)
		if e != nil {
			return e
		}
		n, e := result.RowsAffected()
		if e != nil || n != 1 {
			return ErrPublishingConflict
		}
	}
	b := make([]byte, 16)
	if _, err = rand.Read(b); err != nil {
		return err
	}
	auditID := hex.EncodeToString(b)
	_, err = tx.ExecContext(ctx, `INSERT INTO creator_publication_audit(id,protocol_id,actor_did,from_visibility,to_visibility,recorded_at) VALUES(?,?,?,?,?,?)`, auditID, id, actor, previous, target, now)
	if err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	if s.Push != nil {
		return s.Push(ctx)
	}
	return nil
}
