package repository

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/antikkorps/myLinks/apps/api/internal/domain"
)

type AccountRepository struct {
	db DBTX
}

func NewAccountRepository(db DBTX) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(ctx context.Context, account domain.Account) (domain.Account, error) {
	const query = `
		INSERT INTO accounts (user_id, provider, provider_user_id, password_hash)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, provider, provider_user_id, password_hash, created_at, updated_at
	`
	rows, err := r.db.Query(ctx, query, account.UserID, account.Provider, account.ProviderUserID, account.PasswordHash)
	if err != nil {
		return domain.Account{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Account])
}

func (r *AccountRepository) GetByProviderAndIdentifier(ctx context.Context, provider, identifier string) (domain.Account, error) {
	const query = `
		SELECT id, user_id, provider, provider_user_id, password_hash, created_at, updated_at
		FROM accounts
		WHERE provider = $1 AND provider_user_id = $2
	`
	rows, err := r.db.Query(ctx, query, provider, identifier)
	if err != nil {
		return domain.Account{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Account])
}
