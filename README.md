# SkillPass

Talent marketplace where jobseekers build structured career profiles, get AI-powered skill evaluations, and share their "skill passport" publicly. Verified companies discover candidates and post jobs.

## Tech Stack

| Layer | Choice |
|---|---|
| Runtime | Bun |
| Backend | Elysia 1.x |
| Frontend | React 19 + React Router v7 |
| Styling | Tailwind CSS v4 + DaisyUI 5 |
| Database | PostgreSQL |
| ORM | Drizzle |
| Auth | JWT (`@elysiajs/jwt`) |
| API Docs | Swagger (`@elysiajs/swagger`) |

## Prerequisites

- [Bun](https://bun.sh) >= 1.2
- [Docker](https://docker.com) (for PostgreSQL and full stack)
- Node.js 20+ (optional, for tools)

## Quick Start

```bash
# 1. Start PostgreSQL
docker compose up db -d

# 2. Run migrations and seed
bun run db:push
bun run db:seed

# 3. Start development servers (both server + web)
bun run dev
```

- **Web:** http://localhost:4200
- **API:** http://localhost:8800
- **Swagger docs:** http://localhost:8800/docs

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
| `bun run db:push` | Push Drizzle schema to DB |
| `bun run db:seed` | Seed industry categories |
| `bun run docker:up` | docker compose up --build |
| `bun run docker:down` | docker compose down |

## Project Structure

```
skillpass/
├── server/                — Elysia API backend
│   ├── src/
│   │   ├── index.ts       — Entry point, plugin registration
│   │   ├── db/            — Drizzle schema + DB client
│   │   ├── routes/        — Route handlers by domain
│   │   ├── middleware/    — Auth guards
│   │   └── lib/           — Utilities (password hashing)
│   └── Dockerfile
├── web/                   — React SPA frontend
│   ├── src/
│   │   ├── App.tsx        — Route definitions
│   │   ├── pages/         — Page components
│   │   ├── components/    — Reusable UI (layout, theme toggle)
│   │   ├── hooks/         — Auth context
│   │   └── lib/           — API client with JWT handling
│   └── Dockerfile
├── docker-compose.yml
└── docs/superpowers/
    ├── specs/             — Design documents (3 phases)
    └── plans/             — Implementation plans
```

## Phases

| Phase | Features |
|---|---|
| **1** (current) | Auth, profiles, company verification, candidate search, job postings, Skill Passport |
| **2** | AI evaluation & scoring, job applications, matching, Application Kanban |
| **3** | Company feedback, AI skill suggestions, Skill Gap Radar, Career Path, Company Rep Score |
