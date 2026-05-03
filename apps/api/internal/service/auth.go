package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
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
	pool        *pgxpool.Pool
	refreshRepo *repository.RefreshTokenRepository
	jwtSecret   []byte
	accessTTL   time.Duration
	refreshTTL  time.Duration
}

type LoginResult struct {
	User             domain.User
	AccessToken      string
	AccessExpiresAt  time.Time
	RefreshToken     string
	RefreshExpiresAt time.Time
}

func NewAuthService(pool *pgxpool.Pool, refreshRepo *repository.RefreshTokenRepository, jwtSecret string, accessTTL, refreshTTL time.Duration) *AuthService {
	return &AuthService{
		pool:        pool,
		refreshRepo: refreshRepo,
		jwtSecret:   []byte(jwtSecret),
		accessTTL:   accessTTL,
		refreshTTL:  refreshTTL,
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

func (s *AuthService) generateRefreshToken() (raw, hash string, expiresAt time.Time, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", time.Time{}, err
	}
	raw = base64.RawURLEncoding.EncodeToString(b)
	sum := sha256.Sum256([]byte(raw))
	hash = hex.EncodeToString(sum[:])
	expiresAt = time.Now().Add(s.refreshTTL)
	return raw, hash, expiresAt, nil
}

func (s *AuthService) generateToken(userID uuid.UUID) (string, time.Time,
	error) {
	expiresAt := time.Now().Add(s.accessTTL)
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

	refreshRaw, refreshHash, refreshExp, err := s.generateRefreshToken()
	if err != nil {
		return LoginResult{}, err
	}
	if _, err := s.refreshRepo.Create(ctx, domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: refreshHash,
		ExpiresAt: refreshExp,
	}); err != nil {
		return LoginResult{}, err
	}

	return LoginResult{
		User:             user,
		AccessToken:      token,
		AccessExpiresAt:  expiresAt,
		RefreshToken:     refreshRaw,
		RefreshExpiresAt: refreshExp,
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshTokenRaw string) (LoginResult, error) {
	sum := sha256.Sum256([]byte(refreshTokenRaw))
	hash := hex.EncodeToString(sum[:])

	rt, err := s.refreshRepo.GetByTokenHash(ctx, hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return LoginResult{}, ErrInvalidCredentials
	}
	if err != nil {
		return LoginResult{}, err
	}

	if rt.RevokedAt != nil {
		// Reuse detection: a revoked token is being replayed.
		// Revoke every refresh of this user as a precaution.
		_ = s.refreshRepo.RevokeAllForUser(ctx, rt.UserID)
		return LoginResult{}, ErrInvalidCredentials
	}

	if time.Now().After(rt.ExpiresAt) {
		return LoginResult{}, ErrInvalidCredentials
	}

	userRepo := repository.NewUserRepository(s.pool)
	user, err := userRepo.GetByID(ctx, rt.UserID)
	if errors.Is(err, pgx.ErrNoRows) {
		return LoginResult{}, ErrInvalidCredentials
	}
	if err != nil {
		return LoginResult{}, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return LoginResult{}, err
	}
	defer tx.Rollback(ctx)

	txRepo := repository.NewRefreshTokenRepository(tx)
	if err := txRepo.Revoke(ctx, rt.ID); err != nil {
		return LoginResult{}, err
	}

	newRaw, newHash, newExp, err := s.generateRefreshToken()
	if err != nil {
		return LoginResult{}, err
	}
	if _, err := txRepo.Create(ctx, domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: newHash,
		ExpiresAt: newExp,
	}); err != nil {
		return LoginResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return LoginResult{}, err
	}

	accessToken, accessExp, err := s.generateToken(user.ID)
	if err != nil {
		return LoginResult{}, err
	}

	return LoginResult{
		User:             user,
		AccessToken:      accessToken,
		AccessExpiresAt:  accessExp,
		RefreshToken:     newRaw,
		RefreshExpiresAt: newExp,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshTokenRaw string) error {
	if refreshTokenRaw == "" {
		return nil
	}
	sum := sha256.Sum256([]byte(refreshTokenRaw))
	hash := hex.EncodeToString(sum[:])

	rt, err := s.refreshRepo.GetByTokenHash(ctx, hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	return s.refreshRepo.Revoke(ctx, rt.ID)
}
