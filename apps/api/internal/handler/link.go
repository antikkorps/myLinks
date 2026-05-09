package handler

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/antikkorps/myLinks/apps/api/internal/service"
)

type LinkHandler struct {
	svc *service.LinkService
}

func NewLinkHandler(svc *service.LinkService) *LinkHandler {
	return &LinkHandler{svc: svc}
}

type createLinkRequest struct {
	URL         string     `json:"url"`
	Title       *string    `json:"title,omitempty"`
	Description *string    `json:"description,omitempty"`
	Image       *string    `json:"image,omitempty"`
	FolderID    *uuid.UUID `json:"folder_id,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
}

type updateLinkRequest struct {
	URL         *string    `json:"url,omitempty"`
	Title       *string    `json:"title,omitempty"`
	Description *string    `json:"description,omitempty"`
	Image       *string    `json:"image,omitempty"`
	FolderID    *uuid.UUID `json:"folder_id,omitempty"`
	ClearFolder bool       `json:"clear_folder,omitempty"`
	Tags        *[]string  `json:"tags,omitempty"`
}

// List handles GET /links
func (h *LinkHandler) List(c fiber.Ctx) error {
	userID, ok := currentUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	links, err := h.svc.List(c.Context(), userID)
	if err != nil {
		log.Printf("list links failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(links)
}

// Get handles GET /links/:id
func (h *LinkHandler) Get(c fiber.Ctx) error {
	userID, ok := currentUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid link id"})
	}
	link, err := h.svc.Get(c.Context(), userID, id)
	if errors.Is(err, service.ErrLinkNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "link not found"})
	}
	if err != nil {
		log.Printf("get link failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(link)
}

// Create handles POST /links
func (h *LinkHandler) Create(c fiber.Ctx) error {
	userID, ok := currentUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	req := new(createLinkRequest)
	if err := c.Bind().Body(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	created, err := h.svc.Create(c.Context(), userID, service.CreateLinkInput{
		URL:         req.URL,
		Title:       req.Title,
		Description: req.Description,
		Image:       req.Image,
		FolderID:    req.FolderID,
		TagNames:    req.Tags,
	})
	if errors.Is(err, service.ErrLinkURLRequired) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "url is required"})
	}
	if errors.Is(err, service.ErrLinkFolderNotFound) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "folder not found"})
	}
	if err != nil {
		log.Printf("create link failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.Status(fiber.StatusCreated).JSON(created)
}

// Update handles PATCH /links/:id
func (h *LinkHandler) Update(c fiber.Ctx) error {
	userID, ok := currentUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid link id"})
	}
	req := new(updateLinkRequest)
	if err := c.Bind().Body(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	updated, err := h.svc.Update(c.Context(), userID, id, service.UpdateLinkInput{
		URL:         req.URL,
		Title:       req.Title,
		Description: req.Description,
		Image:       req.Image,
		FolderID:    req.FolderID,
		ClearFolder: req.ClearFolder,
		Tags:        req.Tags,
	})
	if errors.Is(err, service.ErrLinkNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "link not found"})
	}
	if errors.Is(err, service.ErrLinkURLRequired) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "url cannot be empty"})
	}
	if errors.Is(err, service.ErrLinkFolderNotFound) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "folder not found"})
	}
	if err != nil {
		log.Printf("update link failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(updated)
}

// Delete handles DELETE /links/:id
func (h *LinkHandler) Delete(c fiber.Ctx) error {
	userID, ok := currentUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid link id"})
	}
	err = h.svc.Delete(c.Context(), userID, id)
	if errors.Is(err, service.ErrLinkNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "link not found"})
	}
	if err != nil {
		log.Printf("delete link failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
