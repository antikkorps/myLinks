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
