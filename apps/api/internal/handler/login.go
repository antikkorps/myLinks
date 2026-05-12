package handler

import (
	"errors"
	"log"
	"net/mail"

	"github.com/antikkorps/myLinks/apps/api/internal/service"
	"github.com/gofiber/fiber/v3"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	req := new(LoginRequest)
	if err := c.Bind().Body(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if _, err := mail.ParseAddress(req.Email); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid email"})
	}

	result, err := h.auth.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
		}
		log.Printf("login error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    result.AccessToken,
		Expires:  result.AccessExpiresAt,
		HTTPOnly: true,
		Secure:   h.opts.Secure,
		SameSite: h.opts.SameSite,
		Path:     "/",
	})
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    result.RefreshToken,
		Expires:  result.RefreshExpiresAt,
		HTTPOnly: true,
		Secure:   h.opts.Secure,
		SameSite: h.opts.SameSite,
		Path:     "/auth",
	})

	return c.JSON(result.User)
}
