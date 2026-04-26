package handler

import (
	"log"

	"github.com/antikkorps/myLinks/apps/api/internal/repository"
	"github.com/gofiber/fiber/v3"
)

type LinkHandler struct {
	repo *repository.LinkRepository
}

func NewLinkHandler(repo *repository.LinkRepository) *LinkHandler {
	return &LinkHandler{repo: repo}
}

// List handles GET /links
func (h *LinkHandler) List(c fiber.Ctx) error {
	links, err := h.repo.ListAll(c.Context())
	if err != nil {
		log.Printf("Error fetching links: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch links"})
	}
	return c.JSON(links)
}