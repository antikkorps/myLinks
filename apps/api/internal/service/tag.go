package service

import (
	"context"
	"errors"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/antikkorps/myLinks/apps/api/internal/domain"
	"github.com/antikkorps/myLinks/apps/api/internal/repository"
)

var (
	ErrTagNotFound    = errors.New("tag not found")
	ErrTagNameInvalid = errors.New("tag name is invalid")
	ErrLinkNotFound   = errors.New("link not found")
)

type TagService struct {
	tagRepo  *repository.TagRepository
	linkRepo *repository.LinkRepository
}

func NewTagService(tagRepo *repository.TagRepository, linkRepo *repository.LinkRepository) *TagService {
	return &TagService{tagRepo: tagRepo, linkRepo: linkRepo}
}

// normalizeTagName applies the normalization pipeline:
//  1. trim spaces
//  2. NFD + drop combining marks (so "é" → "e")
//  3. lowercase
//  4. keep only [a-z0-9_], drop everything else
//  5. prefix with "#"
//
// Returns ErrTagNameInvalid if the result is empty (or just "#").
func normalizeTagName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", ErrTagNameInvalid
	}
	// strip the leading # if the user already typed one — we re-add it at the end
	name = strings.TrimPrefix(name, "#")

	// NFD decomposes "é" into "e" + combining acute, then we drop the marks.
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	decomposed, _, err := transform.String(t, name)
	if err != nil {
		return "", ErrTagNameInvalid
	}

	decomposed = strings.ToLower(decomposed)

	var b strings.Builder
	b.Grow(len(decomposed) + 1)
	for _, r := range decomposed {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "", ErrTagNameInvalid
	}
	return "#" + b.String(), nil
}

func (s *TagService) CreateTag(ctx context.Context, userID uuid.UUID, rawName string) (domain.Tag, error) {
	name, err := normalizeTagName(rawName)
	if err != nil {
		return domain.Tag{}, err
	}
	// GetOrCreateByName makes CreateTag idempotent: posting twice returns the
	// same tag instead of erroring on the unique index. UX-friendly default.
	return s.tagRepo.GetOrCreateByName(ctx, userID, name)
}

func (s *TagService) ListTags(ctx context.Context, userID uuid.UUID) ([]domain.Tag, error) {
	return s.tagRepo.ListByUserID(ctx, userID)
}

func (s *TagService) GetTagByID(ctx context.Context, userID, tagID uuid.UUID) (domain.Tag, error) {
	tag, err := s.tagRepo.GetByIDForUser(ctx, tagID, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Tag{}, ErrTagNotFound
	}
	return tag, err
}

func (s *TagService) RenameTag(ctx context.Context, userID, tagID uuid.UUID, rawName string) (domain.Tag, error) {
	name, err := normalizeTagName(rawName)
	if err != nil {
		return domain.Tag{}, err
	}
	tag, err := s.tagRepo.UpdateName(ctx, tagID, userID, name)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Tag{}, ErrTagNotFound
	}
	return tag, err
}

func (s *TagService) DeleteTag(ctx context.Context, userID, tagID uuid.UUID) error {
	err := s.tagRepo.SoftDelete(ctx, tagID, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrTagNotFound
	}
	return err
}

// AttachTagToLink ensures both the link and the tag belong to userID before
// inserting into the pivot. The repo pivot methods don't take userID by design
// — ownership is enforced here.
func (s *TagService) AttachTagToLink(ctx context.Context, userID, tagID, linkID uuid.UUID) error {
	if _, err := s.linkRepo.GetByIDForUser(ctx, linkID, userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrLinkNotFound
		}
		return err
	}
	if _, err := s.tagRepo.GetByIDForUser(ctx, tagID, userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrTagNotFound
		}
		return err
	}
	return s.tagRepo.AttachToLink(ctx, tagID, linkID)
}

func (s *TagService) DetachTagFromLink(ctx context.Context, userID, tagID, linkID uuid.UUID) error {
	if _, err := s.linkRepo.GetByIDForUser(ctx, linkID, userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrLinkNotFound
		}
		return err
	}
	if _, err := s.tagRepo.GetByIDForUser(ctx, tagID, userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrTagNotFound
		}
		return err
	}
	return s.tagRepo.DetachFromLink(ctx, tagID, linkID)
}

// ListTagsForLink returns the tags attached to a link. Verifies link ownership first.
func (s *TagService) ListTagsForLink(ctx context.Context, userID, linkID uuid.UUID) ([]domain.Tag, error) {
	if _, err := s.linkRepo.GetByIDForUser(ctx, linkID, userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrLinkNotFound
		}
		return nil, err
	}
	return s.tagRepo.ListByLink(ctx, linkID)
}
