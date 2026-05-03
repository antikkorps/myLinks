package handler

import (
	"errors"
	"log"

	"github.com/antikkorps/myLinks/apps/api/internal/service"
	"github.com/gofiber/fiber/v3"
)

func (h *AuthHandler) Refresh(c fiber.Ctx) error {
	cookie := c.Cookies("refresh_token")
	if cookie == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "refresh token not found"})
	}

	result, err := h.auth.Refresh(c.Context(), cookie)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
		}
		log.Printf("refresh error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    result.AccessToken,
		Expires:  result.AccessExpiresAt,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
		Path:     "/",
	})
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    result.RefreshToken,
		Expires:  result.RefreshExpiresAt,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
		Path:     "/auth",
	})

	return c.JSON(result.User)
}
