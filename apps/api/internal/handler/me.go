package handler

import (
	"log"

	"github.com/gofiber/fiber/v3"
)

func (h *AuthHandler) Me(c fiber.Ctx) error {
	userID, ok := currentUserID(c)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	user, err := h.auth.GetUser(c.Context(), userID)
	if err != nil {
		log.Printf("me: get user: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(user)
}
