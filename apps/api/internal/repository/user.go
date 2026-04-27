package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/antikkorps/myLinks/apps/api/internal/domain"
)

type UserRepository struct {
	db DBTX
}

func NewUserRepository(db DBTX) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	const query = `
		INSERT INTO users (email, first_name, last_name)
		VALUES ($1, $2, $3)
		RETURNING id, email, first_name, last_name, created_at, updated_at, deleted_at
	`
	rows, err := r.db.Query(ctx, query, user.Email, user.FirstName, user.LastName)
	if err != nil {
		return domain.User{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.User])
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	const query = `
		SELECT id, email, first_name, last_name, created_at, updated_at, deleted_at
		FROM users
		WHERE LOWER(email) = LOWER($1) AND deleted_at IS NULL
	`
	rows, err := r.db.Query(ctx, query, email)
	if err != nil {
		return domain.User{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.User])
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	const query = `
		SELECT id, email, first_name, last_name, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`
	rows, err := r.db.Query(ctx, query, id)
	if err != nil {
		return domain.User{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.User])
}
