# ForgeFlow

Generate production-ready interactive UI prototypes using AI.

## Tech Stack

- Next.js 16 + React 19 + Tailwind CSS + shadcn/ui
- Go Fiber API
- PostgreSQL + Redis + NATS
- Turborepo + pnpm workspaces

## Prerequisites

- Node.js 22+
- pnpm 11.20.0+
- Go 1.26+
- Docker Desktop (for Postgres / Redis / NATS)

## Setup

```bash
pnpm install
cp .env.example .env
cp apps/web/.env.example apps/web/.env.local
docker compose up -d
```

## Develop

```bash
# Frontend (http://localhost:3000)
pnpm --filter web dev

# Backend (http://localhost:8080)
cd apps/api && go run ./cmd/server

# Or from the monorepo root (frontend only via turbo)
pnpm dev
```

## Useful endpoints

- `GET http://localhost:8080/health`
- `GET http://localhost:8080/api/v1/`

## Notes

- Always use `pnpm` in this repo (not `npm` / `npx`).
- The API starts even if Postgres/Redis are down, but `/health` reports `degraded` until Docker services are healthy.
