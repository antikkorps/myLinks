package domain

import (
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID             uuid.UUID `json:"-" db:"id"`
	UserID         uuid.UUID `json:"-" db:"user_id"`
	Provider       string    `json:"-" db:"provider"`
	ProviderUserID string    `json:"-" db:"provider_user_id"`
	PasswordHash   *string   `json:"-" db:"password_hash"`
	CreatedAt      time.Time `json:"-" db:"created_at"`
	UpdatedAt      time.Time `json:"-" db:"updated_at"`
}
