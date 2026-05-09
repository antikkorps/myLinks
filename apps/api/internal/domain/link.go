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

// LinkWithTags is the API response shape for a link enriched with its tags.
// Tags is always non-nil (empty slice when no tags) so JSON output is "tags": [].
type LinkWithTags struct {
	Link
	Tags []Tag `json:"tags"`
}
