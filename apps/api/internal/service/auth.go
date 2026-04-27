package service

import (
	"context"
	"errors"
	"strings"

	"github.com/alexedwards/argon2id"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/antikkorps/myLinks/apps/api/internal/domain"
	"github.com/antikkorps/myLinks/apps/api/internal/repository"
)

var ErrEmailAlreadyExists = errors.New("email already exists")

type AuthService struct {
	pool *pgxpool.Pool
}

func NewAuthService(pool *pgxpool.Pool) *AuthService {
	return &AuthService{pool: pool}
}

type RegisterInput struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

func (s *AuthService) Register(ctx context.Context, in RegisterInput) (domain.User, error) {
	hash, err := argon2id.CreateHash(in.Password, argon2id.DefaultParams)
	if err != nil {
		return domain.User{}, err
	}

	email := strings.ToLower(strings.TrimSpace(in.Email))

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.User{}, err
	}
	defer tx.Rollback(ctx)

	userRepo := repository.NewUserRepository(tx)
	accountRepo := repository.NewAccountRepository(tx)

	user, err := userRepo.Create(ctx, domain.User{
		Email:     email,
		FirstName: in.FirstName,
		LastName:  in.LastName,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return domain.User{}, ErrEmailAlreadyExists
		}
		return domain.User{}, err
	}

	if _, err := accountRepo.Create(ctx, domain.Account{
		UserID:         user.ID,
		Provider:       "password",
		ProviderUserID: email,
		PasswordHash:   &hash,
	}); err != nil {
		return domain.User{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
