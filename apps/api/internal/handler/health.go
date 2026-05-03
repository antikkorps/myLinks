package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Health(pool *pgxpool.Pool) fiber.Handler {
	return func(c fiber.Ctx) error {
		if err := pool.Ping(c.Context()); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"status": "down", "error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "ok"})
	}
}
