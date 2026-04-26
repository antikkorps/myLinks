# myLinks

Application de gestion de liens (type Linkwarden simplifié) — projet pédagogique.

## Stack

- **API** — Go + Fiber (`apps/api`)
- **Scraper** — Rust + Axum (`apps/scraper`)
- **Web** — Nuxt 3 (`apps/web`)
- **DB** — PostgreSQL 16 (Docker)

## Structure

```
apps/
  api/         # Backend Go (Fiber)
  scraper/     # Microservice Rust (Axum)
  web/         # Frontend Nuxt 3
packages/
  shared/      # Types partagés
infrastructure/
  docker-compose.yml
migrations/    # Migrations SQL (golang-migrate)
```

## Démarrage rapide

```bash
cp .env.example .env
make db-up          # Démarre Postgres
make migrate-up     # Applique les migrations
make db-psql        # Ouvre psql pour inspecter
```

## Commandes utiles

| Commande | Description |
|---|---|
| `make db-up` | Démarre Postgres en arrière-plan |
| `make db-down` | Stoppe Postgres |
| `make db-logs` | Logs Postgres |
| `make db-psql` | Shell psql interactif |
| `make migrate-up` | Applique toutes les migrations |
| `make migrate-down` | Rollback la dernière migration |
| `make migrate-status` | Affiche la version courante |
| `make migrate-create NAME=xxx` | Crée une nouvelle paire up/down |
