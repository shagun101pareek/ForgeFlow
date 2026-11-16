# ForgeFlow

Generate interactive UI prototypes from a prompt: structured UI spec → React/CSS → live Sandpack preview.

## Tech stack

- **Web:** Next.js + React + Tailwind + Sandpack
- **API:** Go Fiber
- **Database:** PostgreSQL (required)
- **Optional:** Redis (Compose profile `redis` only; health ping, unused by app logic)
- **CI:** GitHub Actions runs web typecheck/lint/build and Go vet/tests on PRs to `main`

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

That starts **Postgres only**. Optional Redis:

```bash
docker compose --profile redis up -d
```

Apply migrations once (paths resolve from the API module):

```bash
cd apps/api && go run ./cmd/migrate up
# optional: go run ./cmd/migrate version
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
- `/health` and `/ready` require Postgres. Redis is reported when present but optional.
- Preview runs client-side in Sandpack; there is no separate preview microservice.
- Settings stores the GitHub token in `localStorage` only on this device.
