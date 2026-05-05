package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/antikkorps/myLinks/apps/api/internal/domain"
	"github.com/antikkorps/myLinks/apps/api/internal/repository"
)

var (
	ErrFolderNotFound = errors.New("folder not found")
	ErrInvalidName    = errors.New("invalid folder name")
)

const maxFolderNameLen = 100

type FolderService struct {
	pool       *pgxpool.Pool
	folderRepo *repository.FolderRepository
}

func NewFolderService(pool *pgxpool.Pool, folderRepo *repository.FolderRepository) *FolderService {
	return &FolderService{pool: pool, folderRepo: folderRepo}
}

func sanitizeName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" {
		return "", ErrInvalidName
	}
	if len(name) > maxFolderNameLen {
		return "", ErrInvalidName
	}
	return name, nil
}

func (s *FolderService) Create(ctx context.Context, userID uuid.UUID, rawName string) (domain.Folder, error) {
	name, err := sanitizeName(rawName)
	if err != nil {
		return domain.Folder{}, err
	}
	return s.folderRepo.Create(ctx, domain.Folder{UserID: userID, Name: name})
}

func (s *FolderService) List(ctx context.Context, userID uuid.UUID) ([]domain.Folder, error) {
	return s.folderRepo.ListByUserID(ctx, userID)
}

func (s *FolderService) Get(ctx context.Context, id, userID uuid.UUID) (domain.Folder, error) {
	folder, err := s.folderRepo.GetByIDForUser(ctx, id, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Folder{}, ErrFolderNotFound
	}
	return folder, err
}

func (s *FolderService) Rename(ctx context.Context, id, userID uuid.UUID, rawName string) (domain.Folder, error) {
	name, err := sanitizeName(rawName)
	if err != nil {
		return domain.Folder{}, err
	}
	folder, err := s.folderRepo.UpdateName(ctx, id, userID, name)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Folder{}, ErrFolderNotFound
	}
	return folder, err
}

// Delete soft-deletes the folder and detaches its links (folder_id -> NULL)
// atomically. Both writes succeed or none do.
func (s *FolderService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	txRepo := repository.NewFolderRepository(tx)

	if err := txRepo.SoftDelete(ctx, id, userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrFolderNotFound
		}
		return err
	}
	if err := txRepo.DetachLinksFromFolder(ctx, id, userID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
