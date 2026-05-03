package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/antikkorps/myLinks/apps/api/internal/domain"
	"github.com/antikkorps/myLinks/apps/api/internal/repository"
)

var ErrEmailAlreadyExists = errors.New("email already exists")
var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthService struct {
	pool      *pgxpool.Pool
	jwtSecret []byte
	tokenTTL  time.Duration
}

type LoginResult struct {
	User        domain.User
	AccessToken string
	ExpiresAt   time.Time
}

func NewAuthService(pool *pgxpool.Pool, jwtSecret string, tokenTTL time.Duration) *AuthService {
	return &AuthService{
		pool:      pool,
		jwtSecret: []byte(jwtSecret),
		tokenTTL:  tokenTTL,
	}
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

func (s *AuthService) generateToken(userID uuid.UUID) (string, time.Time,
	error) {
	expiresAt := time.Now().Add(s.tokenTTL)
	claims := jwt.MapClaims{
		"sub": userID.String(),
		"iat": time.Now().Unix(),
		"exp": expiresAt.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtSecret)
	return signed, expiresAt, err
}

func (s *AuthService) Login(ctx context.Context, email, password string) (LoginResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	accountRepo := repository.NewAccountRepository(s.pool)
	account, err := accountRepo.GetByProviderAndIdentifier(ctx, "password", email)
	if errors.Is(err, pgx.ErrNoRows) {
		return LoginResult{}, ErrInvalidCredentials
	}
	if err != nil {
		return LoginResult{}, err
	}
	if account.PasswordHash == nil {
		return LoginResult{}, ErrInvalidCredentials
	}

	match, err := argon2id.ComparePasswordAndHash(password, *account.PasswordHash)
	if err != nil {
		return LoginResult{}, err
	}
	if !match {
		return LoginResult{}, ErrInvalidCredentials
	}

	userRepo := repository.NewUserRepository(s.pool)
	user, err := userRepo.GetByID(ctx, account.UserID)
	if errors.Is(err, pgx.ErrNoRows) {
		return LoginResult{}, ErrInvalidCredentials
	}
	if err != nil {
		return LoginResult{}, err
	}

	token, expiresAt, err := s.generateToken(user.ID)
	if err != nil {
		return LoginResult{}, err
	}

	return LoginResult{
		User:        user,
		AccessToken: token,
		ExpiresAt:   expiresAt,
	}, nil
}
