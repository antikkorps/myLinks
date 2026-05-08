package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/antikkorps/myLinks/apps/api/internal/domain"
)

type TagRepository struct {
	db DBTX
}

func NewTagRepository(db DBTX) *TagRepository {
	return &TagRepository{db: db}
}

func (r *TagRepository) Create(ctx context.Context, t domain.Tag) (domain.Tag, error) {
	const query = `
		INSERT INTO tags (user_id, name)
		VALUES ($1, $2)
		RETURNING id, user_id, name, created_at, updated_at
	`
	rows, err := r.db.Query(ctx, query, t.UserID, t.Name)
	if err != nil {
		return domain.Tag{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Tag])
}

func (r *TagRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Tag, error) {
	const query = `
		SELECT id, user_id, name, created_at, updated_at
		FROM tags
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[domain.Tag])
}

// GetByIDForUser returns the tag only if it belongs to userID.
// Ownership is enforced in the WHERE clause: a tag owned by another user
// is indistinguishable from a non-existent tag (pgx.ErrNoRows).
func (r *TagRepository) GetByIDForUser(ctx context.Context, id, userID uuid.UUID) (domain.Tag, error) {
	const query = `
		SELECT id, user_id, name, created_at, updated_at
		FROM tags
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`
	rows, err := r.db.Query(ctx, query, id, userID)
	if err != nil {
		return domain.Tag{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Tag])
}

// UpdateName renames a tag owned by userID. Returns pgx.ErrNoRows if
// no row matched (not found or not owned).
func (r *TagRepository) UpdateName(ctx context.Context, id, userID uuid.UUID, name string) (domain.Tag, error) {
	const query = `
		UPDATE tags
		SET name = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		RETURNING id, user_id, name, created_at, updated_at
	`
	rows, err := r.db.Query(ctx, query, id, userID, name)
	if err != nil {
		return domain.Tag{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Tag])
}

// SoftDelete marks the tag as deleted. Returns pgx.ErrNoRows if no row
// matched (not found, not owned, or already deleted).
func (r *TagRepository) SoftDelete(ctx context.Context, id, userID uuid.UUID) error {
	const query = `
		UPDATE tags
		SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`
	result, err := r.db.Exec(ctx, query, id, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// Get or create a tag by name. Returns the existing tag if it already exists (not deleted).
func (r *TagRepository) GetOrCreateByName(ctx context.Context, name string) (domain.Tag, error) {
	const query = `
		INSERT INTO tags (user_id, name)
		VALUES ($1, $2)
		ON CONFLICT (user_id, name) WHERE deleted_at IS NULL
		DO UPDATE SET updated_at = tags.updated_at
		RETURNING id, user_id, name, created_at, updated_at
	`
	rows, err := r.db.Query(ctx, query, name)
	if err != nil {
		return domain.Tag{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Tag])
}

// AttachToLink attaches a tag to a link. Idempotent: re-attaching an
// already-attached tag is a no-op (handled by ON CONFLICT DO NOTHING).
// Ownership of both link and tag must be verified by the caller.
func (r *TagRepository) AttachToLink(ctx context.Context, tagID, linkID uuid.UUID) error {
	const query = `
		INSERT INTO link_tags (tag_id, link_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, tagID, linkID)
	return err
}

// DetachFromLink removes the link between a tag and a link. Idempotent:
// detaching a tag that is not attached is a no-op (no error returned).
// Ownership must be verified by the caller.
func (r *TagRepository) DetachFromLink(ctx context.Context, tagID, linkID uuid.UUID) error {
	const query = `
		DELETE FROM link_tags
		WHERE tag_id = $1 AND link_id = $2
	`
	_, err := r.db.Exec(ctx, query, tagID, linkID)
	return err
}

// ListByLink returns the tags attached to a link, sorted by name.
// Soft-deleted tags are excluded even if their link_tags rows still exist
// (soft-deleting a tag does not cascade-clean the pivot table).
// Ownership of the link must be verified by the caller.
func (r *TagRepository) ListByLink(ctx context.Context, linkID uuid.UUID) ([]domain.Tag, error) {
	const query = `
		SELECT t.id, t.user_id, t.name, t.created_at, t.updated_at
		FROM tags t
		INNER JOIN link_tags lt ON t.id = lt.tag_id
		WHERE lt.link_id = $1 AND t.deleted_at IS NULL
		ORDER BY t.name ASC
	`
	rows, err := r.db.Query(ctx, query, linkID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[domain.Tag])
}
