package handler

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/antikkorps/myLinks/apps/api/internal/service"
)

type TagHandler struct {
	svc *service.TagService
}

func NewTagHandler(svc *service.TagService) *TagHandler {
	return &TagHandler{svc: svc}
}

type tagNameRequest struct {
	Name string `json:"name"`
}

func parseTagID(c fiber.Ctx) (uuid.UUID, error) {
	return uuid.Parse(c.Params("tagId"))
}

func parseLinkID(c fiber.Ctx) (uuid.UUID, error) {
	return uuid.Parse(c.Params("id"))
}

// Create handles POST /tags
func (h *TagHandler) Create(c fiber.Ctx) error {
	userID, ok := currentUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	req := new(tagNameRequest)
	if err := c.Bind().Body(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	tag, err := h.svc.CreateTag(c.Context(), userID, req.Name)
	if errors.Is(err, service.ErrTagNameInvalid) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "tag name is invalid"})
	}
	if err != nil {
		log.Printf("create tag failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.Status(fiber.StatusCreated).JSON(tag)
}

// List handles GET /tags
func (h *TagHandler) List(c fiber.Ctx) error {
	userID, ok := currentUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	tags, err := h.svc.ListTags(c.Context(), userID)
	if err != nil {
		log.Printf("list tags failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(tags)
}

// Get handles GET /tags/:id  (uses :id route param like folders)
func (h *TagHandler) Get(c fiber.Ctx) error {
	userID, ok := currentUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid tag id"})
	}
	tag, err := h.svc.GetTagByID(c.Context(), userID, id)
	if errors.Is(err, service.ErrTagNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "tag not found"})
	}
	if err != nil {
		log.Printf("get tag failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(tag)
}

// Rename handles PUT /tags/:id
func (h *TagHandler) Rename(c fiber.Ctx) error {
	userID, ok := currentUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid tag id"})
	}
	req := new(tagNameRequest)
	if err := c.Bind().Body(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	tag, err := h.svc.RenameTag(c.Context(), userID, id, req.Name)
	if errors.Is(err, service.ErrTagNameInvalid) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "tag name is invalid"})
	}
	if errors.Is(err, service.ErrTagNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "tag not found"})
	}
	if err != nil {
		log.Printf("rename tag failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(tag)
}

// Delete handles DELETE /tags/:id
func (h *TagHandler) Delete(c fiber.Ctx) error {
	userID, ok := currentUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid tag id"})
	}
	err = h.svc.DeleteTag(c.Context(), userID, id)
	if errors.Is(err, service.ErrTagNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "tag not found"})
	}
	if err != nil {
		log.Printf("delete tag failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ListForLink handles GET /links/:id/tags
func (h *TagHandler) ListForLink(c fiber.Ctx) error {
	userID, ok := currentUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	linkID, err := parseLinkID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid link id"})
	}
	tags, err := h.svc.ListTagsForLink(c.Context(), userID, linkID)
	if errors.Is(err, service.ErrLinkNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "link not found"})
	}
	if err != nil {
		log.Printf("list tags for link failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(tags)
}

// Attach handles POST /links/:id/tags/:tagId
func (h *TagHandler) Attach(c fiber.Ctx) error {
	userID, ok := currentUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	linkID, err := parseLinkID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid link id"})
	}
	tagID, err := parseTagID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid tag id"})
	}
	err = h.svc.AttachTagToLink(c.Context(), userID, tagID, linkID)
	if errors.Is(err, service.ErrLinkNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "link not found"})
	}
	if errors.Is(err, service.ErrTagNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "tag not found"})
	}
	if err != nil {
		log.Printf("attach tag failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// Detach handles DELETE /links/:id/tags/:tagId
func (h *TagHandler) Detach(c fiber.Ctx) error {
	userID, ok := currentUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	linkID, err := parseLinkID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid link id"})
	}
	tagID, err := parseTagID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid tag id"})
	}
	err = h.svc.DetachTagFromLink(c.Context(), userID, tagID, linkID)
	if errors.Is(err, service.ErrLinkNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "link not found"})
	}
	if errors.Is(err, service.ErrTagNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "tag not found"})
	}
	if err != nil {
		log.Printf("detach tag failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
