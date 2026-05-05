package handler

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/antikkorps/myLinks/apps/api/internal/service"
)

type FolderHandler struct {
	svc *service.FolderService
}

func NewFolderHandler(svc *service.FolderService) *FolderHandler {
	return &FolderHandler{svc: svc}
}

type folderNameRequest struct {
	Name string `json:"name"`
}

func currentUserID(c fiber.Ctx) (uuid.UUID, bool) {
	id, ok := c.Locals("user_id").(uuid.UUID)
	return id, ok
}

func parseFolderID(c fiber.Ctx) (uuid.UUID, error) {
	return uuid.Parse(c.Params("id"))
}

// Create handles POST /folders
func (h *FolderHandler) Create(c fiber.Ctx) error {
	userID, ok := currentUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	req := new(folderNameRequest)
	if err := c.Bind().Body(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	folder, err := h.svc.Create(c.Context(), userID, req.Name)
	if errors.Is(err, service.ErrInvalidName) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name is required (1-100 chars)"})
	}
	if err != nil {
		log.Printf("create folder failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.Status(fiber.StatusCreated).JSON(folder)
}

// List handles GET /folders
func (h *FolderHandler) List(c fiber.Ctx) error {
	userID, ok := currentUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	folders, err := h.svc.List(c.Context(), userID)
	if err != nil {
		log.Printf("list folders failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(folders)
}

// Get handles GET /folders/:id
func (h *FolderHandler) Get(c fiber.Ctx) error {
	userID, ok := currentUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	id, err := parseFolderID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid folder id"})
	}
	folder, err := h.svc.Get(c.Context(), id, userID)
	if errors.Is(err, service.ErrFolderNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "folder not found"})
	}
	if err != nil {
		log.Printf("get folder failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(folder)
}

// Rename handles PUT /folders/:id
func (h *FolderHandler) Rename(c fiber.Ctx) error {
	userID, ok := currentUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	id, err := parseFolderID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid folder id"})
	}
	req := new(folderNameRequest)
	if err := c.Bind().Body(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	folder, err := h.svc.Rename(c.Context(), id, userID, req.Name)
	if errors.Is(err, service.ErrInvalidName) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name is required (1-100 chars)"})
	}
	if errors.Is(err, service.ErrFolderNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "folder not found"})
	}
	if err != nil {
		log.Printf("rename folder failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(folder)
}

// Delete handles DELETE /folders/:id
func (h *FolderHandler) Delete(c fiber.Ctx) error {
	userID, ok := currentUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	id, err := parseFolderID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid folder id"})
	}
	err = h.svc.Delete(c.Context(), id, userID)
	if errors.Is(err, service.ErrFolderNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "folder not found"})
	}
	if err != nil {
		log.Printf("delete folder failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
