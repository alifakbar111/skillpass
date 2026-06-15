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

- [Go](https://go.dev) >= 1.26
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
- **Storybook:** http://localhost:6006

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
