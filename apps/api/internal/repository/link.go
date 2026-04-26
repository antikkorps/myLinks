package repository

import (
	"context"

	"github.com/antikkorps/myLinks/apps/api/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LinkRepository struct {
	pool *pgxpool.Pool
}

func NewLinkRepository(pool *pgxpool.Pool) *LinkRepository {
	return &LinkRepository{pool: pool}
}

func (r *LinkRepository) ListAll(ctx context.Context) ([]domain.Link, error) {
	const query = `
		SELECT id, user_id, folder_id, url, title, description, image, created_at, updated_at
		FROM links
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query)
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
	rows, err := r.pool.Query(ctx, query, link.UserID, link.FolderID, link.URL, link.Title, link.Description, link.Image)
	if err != nil {
		return domain.Link{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Link])
}
