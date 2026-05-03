package domain

import (
	"net/netip"
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID         uuid.UUID   `db:"id" json:"-"`
	UserID     uuid.UUID   `db:"user_id" json:"-"`
	TokenHash  string      `db:"token_hash" json:"-"`
	ExpiresAt  time.Time   `db:"expires_at" json:"-"`
	RevokedAt  *time.Time  `db:"revoked_at" json:"-"`
	UserAgent  *string     `db:"user_agent" json:"-"`
	IP         *netip.Addr `db:"ip" json:"-"`
	LastUsedAt *time.Time  `db:"last_used_at" json:"-"`
	CreatedAt  time.Time   `db:"created_at" json:"-"`
}
