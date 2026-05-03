package handler

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
)

func (h *AuthHandler) Logout(c fiber.Ctx) error {
	cookie := c.Cookies("refresh_token")
	if err := h.auth.Logout(c.Context(), cookie); err != nil {
		log.Printf("logout error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	expired := time.Unix(0, 0)
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    "",
		Expires:  expired,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
		Path:     "/",
	})
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  expired,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
		Path:     "/auth",
	})

	return c.SendStatus(fiber.StatusNoContent)
}
