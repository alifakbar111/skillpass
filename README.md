# SkillPass

Talent marketplace where jobseekers build structured career profiles, get AI-powered skill evaluations, and share their "skill passport" publicly. Verified companies discover candidates and post jobs.

## Tech Stack

| Layer | Choice |
|---|---|
| Runtime | Go 1.26+ (server), Bun (web tooling) |
| Backend | Go + Gin + go-jet |
| Frontend | React 19 + React Router v7 + TanStack Query v5 |
| Styling | Tailwind CSS v4 + DaisyUI 5 |
| Forms | react-hook-form + Zod |
| Database | PostgreSQL |
| Schema Mgmt | SQL DDL + go-jet codegen |
| Auth | JWT (golang-jwt) |
| Linting | Biome (replaces ESLint + Prettier) |
| Testing | vitest (web) · Go testing (server) |
| CI | GitHub Actions |

## Features

- **Jobseeker profiles** — structured career data, skills, experience, education
- **AI evaluation** — multi-LLM skill assessment with score badges and charts
- **Skill passport** — public profile sharing with QR code
- **Job matching** — AI-powered candidate-job matching with skills gap analysis
- **Applications** — kanban board for tracking application status
- **Company search** — discover candidates by skills, industry, location
- **Company profiles** — verification, job postings, applicant management
- **Feedback system** — structured feedback between jobseekers and companies
- **Company reviews** — reputation scoring and review history
- **Career paths** — AI-generated career progression suggestions
- **Analytics** — profile views, application metrics, engagement tracking
- **HRIS module** — employee management, branches, departments, positions, org chart, RBAC
- **Notifications** — real-time notification bell with unread counts
- **Webhooks** — company webhook integrations with SSRF protection
- **Accessibility** — WCAG 2.1 AA compliant (skip links, ARIA, focus management)

## Prerequisites

| Tool | Version | Install |
|---|---|---|
| [Go](https://go.dev) | >= 1.26 | `go version` |
| [Bun](https://bun.sh) | >= 1.2 | `bun --version` |
| [Docker](https://docker.com) | Any recent | `docker --version` |
| [jet CLI](https://github.com/go-jet/jet) | Latest | `go install github.com/go-jet/jet/v2/cmd/jet@latest` (optional — for codegen) |

> **Quick version check:**
> ```bash
> go version    # expect go1.26+
> bun --version # expect 1.2+
> docker --version
> ```

## Local Setup (Step by Step)

Follow these steps in order for a fresh local development environment.

### 1. Clone the Repository

```bash
git clone <repo-url> skillpass
cd skillpass
```

### 2. Install Dependencies

Install root-level tooling (Biome linter, concurrently, lefthook git hooks, openapi-typescript, etc.):

```bash
bun install
```

Then install frontend dependencies:

```bash
bun --cwd web install
```

### 3. Configure Environment Variables

Environment files are pre-configured for local dev. If starting fresh, copy the examples:

```bash
# Server config (JWT secret, DB connection, LLM keys, SMTP, etc.)
cp server-go/.env.example server-go/.env

# Frontend config (API path, proxy target)
cp web/.env.example web/.env
```

> **Important:** The default `.env` files already in the repo work out of the box.
> - Server uses `server-go/.env` — contains DB creds, JWT secret, LLM config, SMTP settings
> - Frontend uses `web/.env` — sets `VITE_API_BASE_PATH=/api/v1` and proxied API target to `localhost:1234`
> - A root `.env` also exists for reference (duplicates key server vars)
>
> For AI evaluation features, set your LLM API key in `server-go/.env`:
> ```
> LLM_PROVIDER=openai
> LLM_API_KEY=sk-...
> LLM_MODEL=gpt-4o-mini
> ```

### 4. Start PostgreSQL

```bash
docker compose up db -d
```

This starts a PostgreSQL 16 container on port `5432` with database `skillpass` and user `postgres` / password `postgres`.

Wait a few seconds for the container to become healthy:

```bash
docker compose ps  # should show "healthy" for db
```

### 5. Install Lefthook (Git Hooks)

```bash
bun run lefthook install
```

This installs:
- **pre-commit hook** — auto-formats staged files with Biome
- **pre-push hook** — runs Go + web tests, vulnerability checks, API drift detection

> **First-time tip:** Run with `--force` if hooks were previously installed:
> ```bash
> bun x lefthook install --force
> ```

### 6. Run Database Migrations

```bash
bun run db:migrate
```
### 7. Seed Initial Data

```bash
bun run db:seed
```
### 8. Generate go-jet Types (Optional)

```bash
bun run db:generate
```

Regenerates the Go type-safe query builder types from the live database schema into `server-go/.gen/`. Run this **after** any schema change.

> Requires the `jet` CLI: `go install github.com/go-jet/jet/v2/cmd/jet@latest`

### 9. Generate API Types

```bash
bun run api:generate
```

Regenerates:
- Swagger/OpenAPI spec → `server-go/docs/swagger.json`
- TypeScript API client types → `web/src/lib/generated/api.d.ts`

> Run this whenever server response structs change, and **before pushing** (the pre-push hook enforces it).

### 10. Start Development Servers

```bash
bun run dev
```

This starts both servers concurrently:

| Service | URL | Command |
|---|---|---|
| **Web (Vite dev server)** | http://localhost:4200 | `bun --cwd web dev` |
| **API (Go/Gin)** | http://localhost:1234 | `go -C server-go run ./cmd/server/` |
| **Storybook** | http://localhost:6006 | `bun --cwd web storybook` |

The Vite dev server proxies `/api/*` and `/uploads/*` requests to the Go backend at `http://localhost:1234`.

### Complete One-Liner (Fresh Start)

If you've already cloned the repo and installed dependencies:

```bash
docker compose up db -d && 
bun run db:migrate && 
bun run db:seed && 
bun run dev
```

### Docker Full Stack (Alternative)

Run everything in containers (no local tooling needed except Docker):

```bash
# Build and start all services (DB, server, web)
bun run docker:up

# Stop
bun run docker:down
```

This uses the `docker-compose.yml` which builds:
- **db** — PostgreSQL 16
- **server** — Go API (Dockerfile in `server-go/`)
- **web** — Nginx-served React SPA (Dockerfile in `web/`)

> For local development, the non-Docker approach is recommended (hot reload, faster iteration). Use `docker:up` for integration testing or demo deployments.

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
| `bun run api:generate` | Regenerate Swagger + TypeScript API types |
| `bun run api:check` | API drift check (pre-push gate) |
| `bun run lint` | Biome check |
| `bun run lint:fix` | Biome check + auto-fix |
| `bun run format` | Biome format (write) |
| `bun run format:check` | Biome format (read-only) |
| `bun --cwd web test` | Run web tests (vitest) |
| `bun run test:server` | Run Go tests |
| `bun --cwd web typecheck` | TypeScript type check |
| `bun --cwd web storybook` | Start Storybook (port 6006) |
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
│   │   ├── lib/            — Utilities (password hashing, LLM client)
│   │   ├── middleware/     — Auth guards, role checks, rate limiting
│   │   ├── gen/            — go-jet re-exports
│   │   ├── evaluation/     — AI evaluation handler + service
│   │   ├── application/    — Application tracking
│   │   ├── matching/       — Job-candidate matching
│   │   ├── career/         — Career path suggestions
│   │   ├── feedback/       — Feedback system
│   │   ├── companyreviews/ — Company reputation
│   │   ├── analytics/      — Profile view analytics
│   │   ├── notification/   — Notification system
│   │   ├── webhook/        — Webhook integrations
│   │   ├── email/          — Email templates + sender
│   │   ├── resume/         — PDF resume generation
│   │   ├── hris/           — HRIS module (employee, org)
│   │   ├── spdid/          — SP-DID records
│   │   ├── rbac/           — Role-based access control
│   │   ├── storage/        — File storage
│   │   ├── profileviews/   — Profile view tracking
│   │   ├── authtoken/      — Auth token management
│   │   └── testutil/       — Test helpers + factories
│   ├── migrations/         — SQL DDL files (000001-000017)
│   ├── .gen/               — go-jet generated types
│   ├── docs/               — Swagger spec
│   └── Dockerfile
├── web/                    — React SPA frontend
│   ├── src/
│   │   ├── App.tsx         — Route definitions (React Router v7)
│   │   ├── pages/          — Page components (folder per page)
│   │   │   ├── jobseeker/  — Evaluation, Applications, Matches, Feedback, Career, Analytics
│   │   │   ├── company/    — FeedbackHistory, Reputation
│   │   │   └── hris/       — Employees, Branches, Departments, Positions, OrgChart, Roles
│   │   ├── components/     — Reusable UI
│   │   │   ├── layout/     — Navbar, RootLayout, NotificationBell, VerifyEmailBanner
│   │   │   ├── ui/         — Form components, ErrorBoundary, ThemeToggle, LoadingFallback
│   │   │   ├── jobseeker/  — ApplicationKanban, AvatarUploader, EvaluationScoreBadge
│   │   │   ├── company/    — JobMatches
│   │   │   ├── hris/       — HRISLayout, HRISSidebar
│   │   │   ├── onboarding/ — ChecklistCard, CompanyOnboarding, JobseekerOnboarding
│   │   │   └── passport/   — SharePassport (QR code)
│   │   ├── hooks/          — useAuth, useIndustries, usePermissions
│   │   ├── lib/            — api.ts, api-types.ts, domain modules, schemas, generated types
│   │   └── stories/        — Storybook stories (tokens, components, patterns)
│   ├── .storybook/         — Storybook config
│   ├── vitest.config.ts    — Test config
│   └── Dockerfile
├── .agents/                — Agent definitions, rules, skills
├── .claude/                — Claude Code agents
├── .opencode/              — OpenCode agents
├── docker-compose.yml
├── lefthook.yml            — Git hooks
├── biome.json              — Linter config
├── design-tokens.json      — Design tokens
└── docs/                   — specs, plans, migration docs
```

## Architecture

### Backend (Go/Gin)
- Routes registered in `cmd/server/main.go` via `gin.Group("/api/v1/...")`
- JWT auth via `internal/middleware/auth.go` — Bearer token
- Role guards: `RequireRole("company")` + `RequireVerifiedCompany(pool)`
- Rate limiting on auth endpoints
- All responses use **camelCase** JSON field names
- Named response structs (never raw `gin.H` for success paths)
- Swagger docs auto-generated via `swag`

### Frontend (React SPA)
- TanStack Query v5 for server state management
- react-hook-form + Zod for form validation
- Lazy-loaded routes with `React.lazy` + `Suspense`
- `api()` wrapper for all authenticated requests
- Path alias `@/*` → `src/*`

### Database
- PostgreSQL with pgx pool
- 17 SQL migrations (000001-000017)
- go-jet codegen for type-safe queries
- DB-backed migration tracking (no file markers)

## Git Hooks (lefthook)
- **pre-commit:** `bun run format` (auto-fix code style, auto-stages)
- **pre-push:** Go tests, web tests, govulncheck, bun audit, API drift check, gen-types annotation check

## License

Private — All rights reserved.
