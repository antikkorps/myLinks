package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v3"

	"github.com/antikkorps/myLinks/apps/api/internal/config"
	"github.com/antikkorps/myLinks/apps/api/internal/database"
	"github.com/antikkorps/myLinks/apps/api/internal/handler"
	"github.com/antikkorps/myLinks/apps/api/internal/repository"
)

func main() {
	cfg, err := config.Load()
      if err != nil { log.Fatalf("config: %v", err) }                   
   
      ctx := context.Background()                                       
      pool, err := database.Connect(ctx, cfg.DatabaseURL)
      if err != nil { log.Fatalf("database: %v", err) }                 
      defer pool.Close()

	  linkRepo := repository.NewLinkRepository(pool)
	  linkHandler := handler.NewLinkHandler(linkRepo)
                                                                        
      app := fiber.New()                                                
      app.Get("/health", handler.Health(pool))
      app.Get("/links", linkHandler.List)
                                                                        
      log.Fatal(app.Listen(":" + cfg.APIPort))

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Get("/health", func(c fiber.Ctx) error {
		pool.Ping(c.Context())
		if err := pool.Ping(c.Context()); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"status":"down", "error": err.Error()})
		}
		return c.JSON(fiber.Map{"status":"ok"})
	})

	log.Fatal(app.Listen(":8000"))
}
