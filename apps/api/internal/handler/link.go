package handler

import (
	"log"

	"github.com/google/uuid"

	"github.com/antikkorps/myLinks/apps/api/internal/domain"
	"github.com/antikkorps/myLinks/apps/api/internal/repository"
	"github.com/gofiber/fiber/v3"
)

type LinkHandler struct {
	repo *repository.LinkRepository
}

type createLinkRequest struct {
	URL         string  `json:"url"`
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Image       *string `json:"image,omitempty"`
}

func NewLinkHandler(repo *repository.LinkRepository) *LinkHandler {
	return &LinkHandler{repo: repo}
}

// List handles GET /links
func (h *LinkHandler) List(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	links, err := h.repo.ListByUserID(c.Context(), userID)
	if err != nil {
		log.Printf("Error fetching links: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch links"})
	}
	return c.JSON(links)
}

// Create handles POST /links
func (h *LinkHandler) Create(c fiber.Ctx) error {
	req := new(createLinkRequest)
	if err := c.Bind().Body(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.URL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "url is required"})
	}

	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	newLink := domain.Link{
		UserID:      userID,
		URL:         req.URL,
		Title:       req.Title,
		Description: req.Description,
		Image:       req.Image,
	}

	created, err := h.repo.Create(c.Context(), newLink)
	if err != nil {
		log.Printf("create link failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.Status(fiber.StatusCreated).JSON(created)
}
