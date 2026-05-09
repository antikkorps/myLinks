package repository

import (
	"context"

	"github.com/antikkorps/myLinks/apps/api/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type LinkRepository struct {
	db DBTX
}

func NewLinkRepository(db DBTX) *LinkRepository {
	return &LinkRepository{db: db}
}

func (r *LinkRepository) ListAll(ctx context.Context) ([]domain.Link, error) {
	const query = `
		SELECT id, user_id, folder_id, url, title, description, image, created_at, updated_at
		FROM links
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[domain.Link])
}

func (r *LinkRepository) Create(ctx context.Context, link domain.Link) (domain.Link, error) {
	const query = `
		INSERT INTO links (user_id, folder_id, url, title, description, image)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, folder_id, url, title, description, image, created_at, updated_at
	`
	rows, err := r.db.Query(ctx, query, link.UserID, link.FolderID, link.URL, link.Title, link.Description, link.Image)
	if err != nil {
		return domain.Link{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Link])
}

// GetByIDForUser returns the link only if it belongs to userID.
// Ownership is enforced in the WHERE clause: a link owned by another user
// is indistinguishable from a non-existent link (pgx.ErrNoRows).
func (r *LinkRepository) GetByIDForUser(ctx context.Context, id, userID uuid.UUID) (domain.Link, error) {
	const query = `
		SELECT id, user_id, folder_id, url, title, description, image, created_at, updated_at
		FROM links
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`
	rows, err := r.db.Query(ctx, query, id, userID)
	if err != nil {
		return domain.Link{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Link])
}

func (r *LinkRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Link, error) {
	const query = `
		SELECT id, user_id, folder_id, url, title, description, image, created_at, updated_at
		FROM links
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[domain.Link])
}

// Update writes the mutable fields of l (folder_id, url, title, description, image).
// Returns pgx.ErrNoRows if no row matched (not found or not owned).
func (r *LinkRepository) Update(ctx context.Context, id, userID uuid.UUID, l domain.Link) (domain.Link, error) {
	const query = `
		UPDATE links
		SET folder_id = $3, url = $4, title = $5, description = $6, image = $7,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		RETURNING id, user_id, folder_id, url, title, description, image, created_at, updated_at
	`
	rows, err := r.db.Query(ctx, query, id, userID, l.FolderID, l.URL, l.Title, l.Description, l.Image)
	if err != nil {
		return domain.Link{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Link])
}

// SoftDelete marks the link as deleted. Returns pgx.ErrNoRows if no row matched.
func (r *LinkRepository) SoftDelete(ctx context.Context, id, userID uuid.UUID) error {
	const query = `
		UPDATE links
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
