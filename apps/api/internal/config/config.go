package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL  string
	APIPort      string
	JWTSecret    string
	CookieSecure bool
	WebOrigin    string
}

func Load() (*Config, error) {
	// Load environment variables from .env file
	if err := godotenv.Load("../../.env"); err != nil {
		fmt.Printf("No .env file found: %v\n", err)
	}

	cfg := &Config{
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		APIPort:      os.Getenv("API_PORT"),
		JWTSecret:    os.Getenv("JWT_SECRET"),
		WebOrigin:    os.Getenv("WEB_ORIGIN"),
		CookieSecure: true,
	}

	if v := os.Getenv("COOKIE_SECURE"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return nil, fmt.Errorf("COOKIE_SECURE invalid: %w", err)
		}
		cfg.CookieSecure = b
	}

	if cfg.WebOrigin == "" {
		cfg.WebOrigin = "http://localhost:3000"
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.APIPort == "" {
		cfg.APIPort = "8000" // default port
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}
