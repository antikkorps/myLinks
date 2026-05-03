package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/antikkorps/myLinks/apps/api/internal/domain"
)

type RefreshTokenRepository struct {
	db DBTX
}

func NewRefreshTokenRepository(db DBTX) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, token domain.RefreshToken) (domain.RefreshToken, error) {
	const query = `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at, user_agent, ip)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, token_hash, expires_at, revoked_at, user_agent, ip, last_used_at, created_at
	`
	rows, err := r.db.Query(ctx, query, token.UserID, token.TokenHash, token.ExpiresAt, token.UserAgent, token.IP)
	if err != nil {
		return domain.RefreshToken{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.RefreshToken])
}

func (r *RefreshTokenRepository) GetByTokenHash(ctx context.Context, hash string) (domain.RefreshToken, error) {
	const query = `
		SELECT id, user_id, token_hash, expires_at, revoked_at, user_agent, ip, last_used_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`
	rows, err := r.db.Query(ctx, query, hash)
	if err != nil {
		return domain.RefreshToken{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.RefreshToken])
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	const query = `
		UPDATE refresh_tokens SET revoked_at = NOW()
		WHERE id = $1 AND revoked_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	const query = `
		UPDATE refresh_tokens SET revoked_at = NOW()
		WHERE user_id = $1 AND revoked_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}
