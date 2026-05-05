package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/antikkorps/myLinks/apps/api/internal/domain"
)

type FolderRepository struct {
	db DBTX
}

func NewFolderRepository(db DBTX) *FolderRepository {
	return &FolderRepository{db: db}
}

func (r *FolderRepository) Create(ctx context.Context, f domain.Folder) (domain.Folder, error) {
	const query = `
		INSERT INTO folders (user_id, name)
		VALUES ($1, $2)
		RETURNING id, user_id, name, created_at, updated_at
	`
	rows, err := r.db.Query(ctx, query, f.UserID, f.Name)
	if err != nil {
		return domain.Folder{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Folder])
}

func (r *FolderRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Folder, error) {
	const query = `
		SELECT id, user_id, name, created_at, updated_at
		FROM folders
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[domain.Folder])
}

// GetByIDForUser returns the folder only if it belongs to userID.
// Ownership is enforced in the WHERE clause: a folder owned by another user
// is indistinguishable from a non-existent folder (pgx.ErrNoRows).
func (r *FolderRepository) GetByIDForUser(ctx context.Context, id, userID uuid.UUID) (domain.Folder, error) {
	const query = `
		SELECT id, user_id, name, created_at, updated_at
		FROM folders
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`
	rows, err := r.db.Query(ctx, query, id, userID)
	if err != nil {
		return domain.Folder{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Folder])
}

// UpdateName renames a folder owned by userID. Returns pgx.ErrNoRows if
// no row matched (not found or not owned).
func (r *FolderRepository) UpdateName(ctx context.Context, id, userID uuid.UUID, name string) (domain.Folder, error) {
	const query = `
		UPDATE folders
		SET name = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		RETURNING id, user_id, name, created_at, updated_at
	`
	rows, err := r.db.Query(ctx, query, id, userID, name)
	if err != nil {
		return domain.Folder{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Folder])
}

// SoftDelete marks the folder as deleted. Returns pgx.ErrNoRows if no row
// matched (not found, not owned, or already deleted).
func (r *FolderRepository) SoftDelete(ctx context.Context, id, userID uuid.UUID) error {
	const query = `
		UPDATE folders
		SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`
	tag, err := r.db.Exec(ctx, query, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// DetachLinksFromFolder sets links.folder_id = NULL for every link of userID
// currently in folderID. Used inside the delete-folder transaction so that
// links survive their folder's removal (option B1).
func (r *FolderRepository) DetachLinksFromFolder(ctx context.Context, folderID, userID uuid.UUID) error {
	const query = `
		UPDATE links
		SET folder_id = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE folder_id = $1 AND user_id = $2 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, folderID, userID)
	return err
}
