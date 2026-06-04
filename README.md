# SkillPass

Talent marketplace where jobseekers build structured career profiles, get AI-powered skill evaluations, and share their "skill passport" publicly. Verified companies discover candidates and post jobs.

## Tech Stack

| Layer | Choice |
|---|---|
| Runtime | Go 1.24+ + Bun (web only) |
| Backend | Go + Gin + go-jet |
| Frontend | React 19 + React Router v7 |
| Styling | Tailwind CSS v4 + DaisyUI 5 |
| Database | PostgreSQL |
| Schema Mgmt | SQL DDL + go-jet codegen |
| Auth | JWT (golang-jwt) |

## Prerequisites

- [Go](https://go.dev) >= 1.24
- [Bun](https://bun.sh) >= 1.2 (for web tooling)
- [Docker](https://docker.com) (for PostgreSQL and full stack)
- [jet](https://github.com/go-jet/jet) CLI (optional, for codegen)

## Quick Start

```bash
# 1. Start PostgreSQL
docker compose up db -d

# 2. Run migrations and seed
bun run db:migrate
bun run db:seed

# 3. Start development servers (both server + web)
bun run dev
```

- **Web:** http://localhost:4200
- **API:** http://localhost:1234

## Full Docker Setup

```bash
# Build and start all services
bun run docker:up

# Stop
bun run docker:down
```

## Available Commands

| Command | Description |
|---|---|
| `bun run dev` | Start server + web concurrently |
| `bun run dev:server` | Start server only |
| `bun run dev:web` | Start web only |
| `bun run build` | Build web for production |
| `bun run db:migrate` | Run SQL migrations |
| `bun run db:seed` | Seed industry categories + admin user |
| `bun run db:generate` | Regenerate go-jet types from DB |
| `bun run docker:up` | docker compose up --build |
| `bun run docker:down` | docker compose down |

## Project Structure

```
skillpass/
├── server-go/              — Go/Gin API backend
│   ├── cmd/
│   │   ├── server/         — Entry point (main.go)
│   │   ├── migrate/        — SQL migration runner
│   │   └── seed/           — DB seeder (go-jet)
│   ├── internal/
│   │   ├── config/         — Env config
│   │   ├── db/             — pgx pool setup
│   │   ├── handlers/       — Route handlers by domain
│   │   ├── lib/            — Utilities (password hashing)
│   │   ├── middleware/     — Auth guards, role checks
│   │   └── gen/            — go-jet re-exports
│   ├── migrations/         — SQL DDL files
│   ├── .gen/               — go-jet generated types
│   └── Dockerfile
├── web/                    — React SPA frontend
│   ├── src/
│   │   ├── App.tsx         — Route definitions
│   │   ├── pages/          — Page components
│   │   ├── components/     — Reusable UI (layout, theme toggle)
│   │   ├── hooks/          — Auth context
│   │   └── lib/            — API client with JWT handling
│   └── Dockerfile
├── docker-compose.yml
└── docs/
    ├── specs/              — Design documents (3 phases)
    └── plans/              — Implementation plans
```
