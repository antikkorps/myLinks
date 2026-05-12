package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"

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
	folderRepo := repository.NewFolderRepository(pool)
	tagRepo := repository.NewTagRepository(pool)

	linkService := service.NewLinkService(pool, linkRepo, folderRepo, tagRepo)
	linkHandler := handler.NewLinkHandler(linkService)

	folderService := service.NewFolderService(pool, folderRepo)
	folderHandler := handler.NewFolderHandler(folderService)

	tagService := service.NewTagService(tagRepo, linkRepo)
	tagHandler := handler.NewTagHandler(tagService)

	refreshRepo := repository.NewRefreshTokenRepository(pool)
	authService := service.NewAuthService(pool, refreshRepo, cfg.JWTSecret, 15*time.Minute, 30*24*time.Hour)
	authHandler := handler.NewAuthHandler(authService, handler.CookieOptions{
		Secure:   cfg.CookieSecure,
		SameSite: "Lax",
	})

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.WebOrigin},
		AllowCredentials: true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type"},
	}))
	app.Get("/health", handler.Health(pool))
	app.Post("/auth/register", authHandler.Register)
	app.Post("/auth/login", authHandler.Login)
	app.Post("/auth/refresh", authHandler.Refresh)
	app.Post("/auth/logout", authHandler.Logout)
	app.Get("/auth/me", middleware.RequireAuth(cfg.JWTSecret), authHandler.Me)

	link := app.Group("/links", middleware.RequireAuth(cfg.JWTSecret))
	link.Get("/", linkHandler.List)
	link.Post("/", linkHandler.Create)
	link.Get("/:id", linkHandler.Get)
	link.Patch("/:id", linkHandler.Update)
	link.Delete("/:id", linkHandler.Delete)
	link.Get("/:id/tags", tagHandler.ListForLink)
	link.Post("/:id/tags/:tagId", tagHandler.Attach)
	link.Delete("/:id/tags/:tagId", tagHandler.Detach)

	folder := app.Group("/folders", middleware.RequireAuth(cfg.JWTSecret))
	folder.Get("/", folderHandler.List)
	folder.Post("/", folderHandler.Create)
	folder.Get("/:id", folderHandler.Get)
	folder.Put("/:id", folderHandler.Rename)
	folder.Delete("/:id", folderHandler.Delete)

	tag := app.Group("/tags", middleware.RequireAuth(cfg.JWTSecret))
	tag.Get("/", tagHandler.List)
	tag.Post("/", tagHandler.Create)
	tag.Get("/:id", tagHandler.Get)
	tag.Put("/:id", tagHandler.Rename)
	tag.Delete("/:id", tagHandler.Delete)

	log.Fatal(app.Listen(":" + cfg.APIPort))
}
