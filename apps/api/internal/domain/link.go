package domain

import (
	"time"

	"github.com/google/uuid"
)

type Link struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	FolderID    *uuid.UUID `json:"folder_id" db:"folder_id"`
	URL         string     `json:"url" db:"url"`
	Title       *string    `json:"title" db:"title"`
	Description *string    `json:"description" db:"description"`
	Image       *string    `json:"image" db:"image"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	// DeletedAt intentionally omitted from API responses
}
