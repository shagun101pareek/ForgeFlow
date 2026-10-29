# ForgeFlow

Generate interactive UI prototypes from a prompt: structured UI spec → React/CSS → live Sandpack preview.

## Tech stack

- **Web:** Next.js + React + Tailwind + Sandpack
- **API:** Go Fiber
- **Database:** PostgreSQL (required)
- **Optional in Compose:** Redis and NATS are started for health checks / future workers; the app does not depend on them for generate → preview today
- **CI:** GitHub Actions runs web typecheck/lint and Go vet/tests on PRs to `main`

## Prerequisites

- Node.js 22+
- pnpm 11.20.0+
- Go 1.26+
- Docker Desktop (Postgres on port 5433)

## Setup

```bash
pnpm install
cp .env.example .env
cp apps/web/.env.example apps/web/.env.local
docker compose up -d
```

Apply migrations once (from `apps/api`):

```bash
cd apps/api && go run ./cmd/migrate up
```

## Develop

```bash
# API — http://localhost:8080
cd apps/api && go run ./cmd/server

# Web — http://localhost:3000
pnpm --filter web dev
```

Useful checks: `GET http://localhost:8080/health`, `GET http://localhost:8080/api/v1/`.

## Generation providers

| `GENERATION_PROVIDER` | Behavior |
| --- | --- |
| `mock` (default) | Deterministic layouts from the prompt (landing, dashboard, portfolio, pricing). No API key. |
| `openai` | Uses `OPENAI_API_KEY` for the UI spec, then the same React codegen. |

Set these in the root `.env` before starting the API.

## 5-minute demo

1. Start Docker, migrate, API, and web (above).
2. Open http://localhost:3000 → **Sign up** → create a project.
3. On the empty canvas, pick a layout card (**Landing**, **Dashboard**, **Portfolio**, or **Pricing**). Generation starts and the preview fills the canvas.
4. Edit the prompt (try adding “testimonials”, “FAQ”, or “stats”), click **Generate**, then open an older version from the sidebar or **Discard** one.
5. Optional: open **Account → Settings** and save a GitHub personal access token (`repo` scope) for this browser.
6. **Actions → Download** for a ZIP, or **Publish** (uses the saved token when present).

## Notes

- Always use `pnpm` in this repo (not `npm` / `npx`).
- The API starts even if Postgres/Redis are down; `/health` reports `degraded` until Postgres is healthy.
- Preview runs client-side in Sandpack; there is no separate preview microservice.
- Settings stores the GitHub token in `localStorage` only on this device.
