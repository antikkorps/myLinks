# Charge les variables de .env si présent
ifneq (,$(wildcard .env))
	include .env
	export
endif

DATABASE_URL ?= postgres://mylinks:mylinks_dev@localhost:5434/mylinks?sslmode=disable
MIGRATIONS_DIR := migrations

.PHONY: db-up db-down db-logs db-psql migrate-up migrate-down migrate-status migrate-create api-run api-build web-install web-dev web-build

## --- Docker (Postgres) ---
db-up:
	cd infrastructure && docker compose up -d

db-down:
	cd infrastructure && docker compose down

db-logs:
	cd infrastructure && docker compose logs -f postgres

db-psql:
	docker exec -it mylinks_postgres psql -U mylinks -d mylinks

## --- Migrations ---
migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1

migrate-status:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" version

# Usage: make migrate-create NAME=add_users_table
migrate-create:
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)

## --- API (Go) ---
api-run:
	cd apps/api && go run ./cmd/api

api-build:
	cd apps/api && go build -o bin/api ./cmd/api

## --- Web (Nuxt) ---
web-install:
	cd apps/web && pnpm install

web-dev:
	cd apps/web && pnpm dev

web-build:
	cd apps/web && pnpm build
