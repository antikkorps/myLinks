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
	ErrLinkURLRequired    = errors.New("url is required")
	ErrLinkFolderNotFound = errors.New("folder not found")
)

type LinkService struct {
	pool       *pgxpool.Pool
	linkRepo   *repository.LinkRepository
	folderRepo *repository.FolderRepository
	tagRepo    *repository.TagRepository
}

func NewLinkService(
	pool *pgxpool.Pool,
	linkRepo *repository.LinkRepository,
	folderRepo *repository.FolderRepository,
	tagRepo *repository.TagRepository,
) *LinkService {
	return &LinkService{pool: pool, linkRepo: linkRepo, folderRepo: folderRepo, tagRepo: tagRepo}
}

// CreateLinkInput is the service-layer input for link creation.
// TagNames are raw strings; they are normalized + GetOrCreate'd inside the TX.
type CreateLinkInput struct {
	URL         string
	Title       *string
	Description *string
	Image       *string
	FolderID    *uuid.UUID
	TagNames    []string
}

// UpdateLinkInput is the service-layer patch payload.
//
//	URL/Title/Description/Image: nil = no change. For Title/Description/Image,
//	  pass *("") to clear (stored as NULL).
//	FolderID + ClearFolder: ClearFolder=true detaches; otherwise FolderID nil = no change,
//	  non-nil = set to that folder.
//	Tags: nil = no change; non-nil (even empty) = replace full set.
type UpdateLinkInput struct {
	URL         *string
	Title       *string
	Description *string
	Image       *string
	FolderID    *uuid.UUID
	ClearFolder bool
	Tags        *[]string
}

func (s *LinkService) Create(ctx context.Context, userID uuid.UUID, in CreateLinkInput) (domain.LinkWithTags, error) {
	if strings.TrimSpace(in.URL) == "" {
		return domain.LinkWithTags{}, ErrLinkURLRequired
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.LinkWithTags{}, err
	}
	defer tx.Rollback(ctx)

	linkTxRepo := repository.NewLinkRepository(tx)
	folderTxRepo := repository.NewFolderRepository(tx)
	tagTxRepo := repository.NewTagRepository(tx)

	if in.FolderID != nil {
		if _, err := folderTxRepo.GetByIDForUser(ctx, *in.FolderID, userID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.LinkWithTags{}, ErrLinkFolderNotFound
			}
			return domain.LinkWithTags{}, err
		}
	}

	link, err := linkTxRepo.Create(ctx, domain.Link{
		UserID:      userID,
		FolderID:    in.FolderID,
		URL:         in.URL,
		Title:       in.Title,
		Description: in.Description,
		Image:       in.Image,
	})
	if err != nil {
		return domain.LinkWithTags{}, err
	}

	tags, err := resolveAndAttachTags(ctx, tagTxRepo, userID, link.ID, in.TagNames)
	if err != nil {
		return domain.LinkWithTags{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.LinkWithTags{}, err
	}
	return domain.LinkWithTags{Link: link, Tags: tags}, nil
}

func (s *LinkService) Get(ctx context.Context, userID, id uuid.UUID) (domain.LinkWithTags, error) {
	link, err := s.linkRepo.GetByIDForUser(ctx, id, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.LinkWithTags{}, ErrLinkNotFound
	}
	if err != nil {
		return domain.LinkWithTags{}, err
	}
	tags, err := s.tagRepo.ListByLink(ctx, link.ID)
	if err != nil {
		return domain.LinkWithTags{}, err
	}
	if tags == nil {
		tags = []domain.Tag{}
	}
	return domain.LinkWithTags{Link: link, Tags: tags}, nil
}

func (s *LinkService) List(ctx context.Context, userID uuid.UUID) ([]domain.LinkWithTags, error) {
	links, err := s.linkRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(links) == 0 {
		return []domain.LinkWithTags{}, nil
	}
	ids := make([]uuid.UUID, len(links))
	for i, l := range links {
		ids[i] = l.ID
	}
	tagsByLink, err := s.tagRepo.ListByLinkIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make([]domain.LinkWithTags, len(links))
	for i, l := range links {
		tags := tagsByLink[l.ID]
		if tags == nil {
			tags = []domain.Tag{}
		}
		out[i] = domain.LinkWithTags{Link: l, Tags: tags}
	}
	return out, nil
}

func (s *LinkService) Update(ctx context.Context, userID, id uuid.UUID, in UpdateLinkInput) (domain.LinkWithTags, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.LinkWithTags{}, err
	}
	defer tx.Rollback(ctx)

	linkTxRepo := repository.NewLinkRepository(tx)
	folderTxRepo := repository.NewFolderRepository(tx)
	tagTxRepo := repository.NewTagRepository(tx)

	current, err := linkTxRepo.GetByIDForUser(ctx, id, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.LinkWithTags{}, ErrLinkNotFound
	}
	if err != nil {
		return domain.LinkWithTags{}, err
	}

	if in.URL != nil {
		if strings.TrimSpace(*in.URL) == "" {
			return domain.LinkWithTags{}, ErrLinkURLRequired
		}
		current.URL = *in.URL
	}
	if in.Title != nil {
		current.Title = nilIfEmpty(in.Title)
	}
	if in.Description != nil {
		current.Description = nilIfEmpty(in.Description)
	}
	if in.Image != nil {
		current.Image = nilIfEmpty(in.Image)
	}
	switch {
	case in.ClearFolder:
		current.FolderID = nil
	case in.FolderID != nil:
		if _, err := folderTxRepo.GetByIDForUser(ctx, *in.FolderID, userID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.LinkWithTags{}, ErrLinkFolderNotFound
			}
			return domain.LinkWithTags{}, err
		}
		current.FolderID = in.FolderID
	}

	updated, err := linkTxRepo.Update(ctx, id, userID, current)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.LinkWithTags{}, ErrLinkNotFound
	}
	if err != nil {
		return domain.LinkWithTags{}, err
	}

	var tags []domain.Tag
	if in.Tags != nil {
		if err := tagTxRepo.DetachAllFromLink(ctx, updated.ID); err != nil {
			return domain.LinkWithTags{}, err
		}
		tags, err = resolveAndAttachTags(ctx, tagTxRepo, userID, updated.ID, *in.Tags)
		if err != nil {
			return domain.LinkWithTags{}, err
		}
	} else {
		tags, err = tagTxRepo.ListByLink(ctx, updated.ID)
		if err != nil {
			return domain.LinkWithTags{}, err
		}
	}
	if tags == nil {
		tags = []domain.Tag{}
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.LinkWithTags{}, err
	}
	return domain.LinkWithTags{Link: updated, Tags: tags}, nil
}

func (s *LinkService) Delete(ctx context.Context, userID, id uuid.UUID) error {
	err := s.linkRepo.SoftDelete(ctx, id, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrLinkNotFound
	}
	return err
}

// resolveAndAttachTags normalizes raw names, dedupes, GetOrCreates each tag for
// the user, then attaches them to the link. Returns the resolved tags sorted
// stably by attach order. Empty input returns an empty slice without error.
func resolveAndAttachTags(
	ctx context.Context,
	tagRepo *repository.TagRepository,
	userID, linkID uuid.UUID,
	rawNames []string,
) ([]domain.Tag, error) {
	if len(rawNames) == 0 {
		return []domain.Tag{}, nil
	}
	seen := make(map[string]struct{}, len(rawNames))
	tags := make([]domain.Tag, 0, len(rawNames))
	for _, raw := range rawNames {
		name, err := normalizeTagName(raw)
		if err != nil {
			// Silently skip invalid names rather than failing the whole request.
			// Alternative: surface ErrTagNameInvalid. V1 choice = lenient.
			continue
		}
		if _, dup := seen[name]; dup {
			continue
		}
		seen[name] = struct{}{}

		tag, err := tagRepo.GetOrCreateByName(ctx, userID, name)
		if err != nil {
			return nil, err
		}
		if err := tagRepo.AttachToLink(ctx, tag.ID, linkID); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func nilIfEmpty(s *string) *string {
	if s == nil || *s == "" {
		return nil
	}
	return s
}
