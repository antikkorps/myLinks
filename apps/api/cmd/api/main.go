package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/antikkorps/myLinks/apps/api/internal/config"
	"github.com/antikkorps/myLinks/apps/api/internal/database"
	"github.com/antikkorps/myLinks/apps/api/internal/handler"
	"github.com/antikkorps/myLinks/apps/api/internal/middleware"
	"github.com/antikkorps/myLinks/apps/api/internal/repository"
	"github.com/antikkorps/myLinks/apps/api/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()
	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	linkRepo := repository.NewLinkRepository(pool)
	linkHandler := handler.NewLinkHandler(linkRepo)

	refreshRepo := repository.NewRefreshTokenRepository(pool)
	authService := service.NewAuthService(pool, refreshRepo, cfg.JWTSecret, 15*time.Minute, 30*24*time.Hour)
	authHandler := handler.NewAuthHandler(authService)

	app := fiber.New()
	app.Get("/health", handler.Health(pool))
	app.Post("/auth/register", authHandler.Register)
	app.Post("/auth/login", authHandler.Login)
	app.Post("/auth/refresh", authHandler.Refresh)

	link := app.Group("/links", middleware.RequireAuth(cfg.JWTSecret))
	link.Get("/", linkHandler.List)
	link.Post("/", linkHandler.Create)

	log.Fatal(app.Listen(":" + cfg.APIPort))
}
