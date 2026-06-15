# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**SkillPass** is a talent marketplace where jobseekers build structured career profiles with AI-powered skill evaluations and share public "skill passports." Verified companies discover candidates and post jobs.

**Monorepo structure:**
```
skillpass/              — root orchestration
├── server-go/         — Go (Gin) API backend
├── web/               — React SPA frontend
├── .agents/           — Agent definitions, rules, skills
└── docs/              — specs and plans
```

## Tech Stack

| Layer | Technology |
|-------|------------|
| **Runtime** | Go 1.26+ (server) · Bun (web tooling) |
| **Backend** | Gin framework · pgx (database driver) · go-jet (ORM codegen) |
| **Frontend** | React 19 · React Router v7 · Vite 7 · TanStack Query v5 |
| **Forms** | react-hook-form + Zod validation |
| **Styling** | Tailwind CSS v4 · DaisyUI 5 (no tailwind.config files) |
| **Database** | PostgreSQL · 17 SQL DDL migrations |
| **Auth** | JWT (golang-jwt) · Bearer token in Authorization header |
| **Linting/Formatting** | Biome 2.5+ (single binary, replaces ESLint + Prettier) |
| **Testing** | vitest + @testing-library/react (web) · Go `testing` + `httptest` (server) |
| **CI** | GitHub Actions |
| **Icons** | lucide-react |
| **QR** | qrcode.react |

## Development Commands

**From root directory (all commands use `bun`):**

| Action | Command |
|--------|---------|
| **Start development** | `bun run dev` (server + web concurrently) |
| Server only | `bun run dev:server` |
| Web only | `bun run dev:web` |
| **Database** | |
| Run migrations | `bun run db:migrate` |
| Seed initial data | `bun run db:seed` |
| Regenerate go-jet types | `bun run db:generate` |
| Regenerate API types | `bun run api:generate` (after changing any response struct or swag annotation) |
| API drift check | `bun run api:check` |
| **Code quality** | |
| Lint (check) | `bun run lint` |
| Lint + auto-fix | `bun run lint:fix` |
| Format (write) | `bun run format` |
| Format check (read-only) | `bun run format:check` |
| **Testing** | |
| Web typecheck | `bun --cwd web typecheck` (tsc --noEmit) |
| Web tests | `bun --cwd web test` (vitest) |
| Go tests | `bun run test:server` (go test -p 1) |
| **Build/Deploy** | |
| Build web for production | `bun run build` |
| Docker full stack up | `bun run docker:up` |
| Docker full stack down | `bun run docker:down` |
| **Storybook** | |
| Start Storybook | `bun --cwd web storybook` (port 6006) |

**Local dev startup (non-Docker):**
```bash
docker compose up db -d                # Start PostgreSQL container
bun run db:migrate                     # Run SQL migrations
bun run db:seed                        # Seed initial data (admin + industries)
bun run dev                            # Start server + web concurrently
```

**Dev URLs:**
- Web: http://localhost:4200
- API: http://localhost:1234
- Storybook: http://localhost:6006
- Vite (in web) proxies `/api/*` and `/uploads` → `http://localhost:1234` (see `web/vite.config.ts`)

## Architecture Notes

### Backend (Go/Gin)

**Entry point:** `server-go/cmd/server/main.go`

**Key conventions:**
- Routes registered via `gin.Group("/api/v1/...")` in main.go
- Request body binding: `c.ShouldBindJSON(&struct)` with struct tags
- JWT auth: middleware in `internal/middleware/auth.go` parses Bearer token, sets `userId` + `role` in Gin context
- Role guards: `RequireRole("company")` + `RequireVerifiedCompany(pool)` middleware
- Rate limiting on auth endpoints via `internal/middleware/ratelimit.go`
- Password hashing: bcrypt (cost 4 in dev) via `internal/lib/password.go`, with argon2id fallback for existing hashes
- **All JSON responses use camelCase** field names (struct tags: `json:"fieldName"`)
- Named response structs in `handlers/responses.go` — never return raw `gin.H` or go-jet `internal/gen` types from success paths
- Config from `internal/config/config.go` reads: `JWT_SECRET`, `DATABASE_URL`, `PORT`, `CORS_ORIGIN` from `.env` or env vars
- **Important:** Server looks for `.env` in `server-go/.env`, not root `.env`

**Database:**
- pgx pool setup in `internal/db/db.go`
- Raw SQL in `server-go/migrations/` (DDL files, 17 migrations)
- go-jet generated types in `server-go/.gen/`, re-exported via `server-go/internal/gen/`
- Codegen: `bun run db:generate` runs the `jet` CLI against the live DB
- DB-backed migration tracking (replaced brittle file markers)

**Handlers:** One handler type per domain:
- `handlers/auth.go` — registration, login, refresh tokens, email verification, password reset
- `handlers/jobs.go` — CRUD for job postings
- `handlers/profiles.go` — jobseeker profiles
- `handlers/companies.go` — company profiles
- `handlers/search.go` — candidate/company search
- `handlers/passport.go` — public passport + OG image
- `handlers/uploads.go` — file uploads
- `handlers/admin.go` — admin verification management
- `handlers/reference.go` — reference checks
- `handlers/health.go` — health check
- `evaluation/handler.go` — AI skill evaluation
- `application/handler.go` — job applications
- `matching/handler.go` — job-candidate matching
- `career/handler.go` — career path suggestions
- `feedback/handler.go` — structured feedback
- `companyreviews/handler.go` — company reviews
- `analytics/handler.go` — profile view analytics
- `notification/handler.go` — notifications
- `webhook/handler.go` — webhook management
- `resume/handler.go` — PDF resume generation
- `hris/employee/handler.go` — employee management
- `hris/org/handler.go` — org structure (branches, departments, positions, roles)
- `spdid/handler.go` — SP-DID records

### Frontend (React SPA)

**Entry point:** `web/src/main.tsx`

**Key conventions:**
- **All API calls must use `src/lib/api.ts`** — never raw `fetch()` to `/api/v1/...`
- `api()` wrapper auto-attaches Bearer token and auto-refreshes on 401
- TanStack Query v5 for server state — `useQuery`, `useMutation`, `queryClient`
- react-hook-form + Zod for form validation
- Auth state via `AuthProvider` in `hooks/useAuth.tsx` (reads/writes to localStorage)
- Tokens stored in localStorage: `accessToken` + `refreshToken`
- Route definitions in `src/App.tsx` inside `<QueryClientProvider>` + `<AuthProvider>` + `<ErrorBoundary>` + `<Suspense>`
- Lazy-loaded routes via `React.lazy` + `Suspense` with `LoadingFallback`
- Path alias: `@/*` → `src/*` (tsconfig.json paths)
- **Styling:** Tailwind v4 with no `tailwind.config.*` — config is in `web/src/styles/index.css` via `@import "tailwindcss"; @plugin "daisyui";`
- **Accessibility:** WCAG 2.1 AA — skip links, ARIA labels, focus management, menu semantics

**Folder structure:**
- `src/pages/` — Page components (folder per page with `index.tsx`)
  - `pages/jobseeker/` — EvaluationPage, ApplicationsPage, MatchesPage, FeedbackPage, CareerPage, AnalyticsPage
  - `pages/company/` — FeedbackHistoryPage, ReputationPage
  - `pages/hris/` — EmployeeList, EmployeeCreate, EmployeeDetail, BranchManagement, DepartmentManagement, PositionManagement, OrgChart, RoleManagement
- `src/components/` — Reusable UI
  - `components/layout/` — Navbar, RootLayout, NotificationBell, VerifyEmailBanner
  - `components/ui/` — FormInput, FormSelect, FormTextarea, FormNumberInput, FormField, ToggleButtonGroup, ErrorBoundary, ThemeToggle, LoadingFallback, ProtectedRoute
  - `components/jobseeker/` — ApplicationKanban, AvatarUploader, EvaluationScoreBadge, SkillScoresChart
  - `components/company/` — JobMatches
  - `components/hris/` — HRISLayout, HRISSidebar
  - `components/onboarding/` — ChecklistCard, CompanyOnboarding, JobseekerOnboarding
  - `components/passport/` — SharePassport (QR code)
- `src/hooks/` — Custom hooks (useAuth, useIndustries, usePermissions)
- `src/lib/` — Utilities
  - `lib/api.ts` — HTTP client with JWT handling
  - `lib/api-types.ts` — barrel over generated OpenAPI types
  - `lib/generated/` — auto-generated TypeScript types from Swagger
  - `lib/schemas/` — Zod validation schemas
  - Domain modules: `application.ts`, `evaluation.ts`, `matching.ts`, `career.ts`, `feedback.ts`, `company-reviews.ts`, `analytics.ts`, `notifications.ts`, `webhooks.ts`, `resume.ts`, `profile-views.ts`
  - `lib/hris/` — HRIS API modules (employees, org, calendar, rbac, spdid)
- `src/stories/` — Storybook stories (tokens, components, patterns)
- `src/test/` — Test utilities (renderWithClient, setup)

### Styling

**Design system in `DESIGN.md`** — read this for:
- Color tokens (all DaisyUI semantic, no hex values)
- Typography (Outfit + Fira Code)
- Spacing scale
- Container widths by page type
- Component patterns (card, button, form input, badge, etc.)
- Dark mode (data-theme attribute toggled in ThemeToggle.tsx)
- Accessibility patterns (skip links, ARIA, focus management)

**Key principle:** Zero custom CSS (except skip-link base styles). All utility classes from Tailwind + DaisyUI. This enables instant dark mode and theme switching.

## Git Workflow

**Git hooks via `lefthook.yml`:**
- **pre-commit:** `bun run format` (auto-fix code style, auto-stages)
- **pre-push:**
  - `bun run test:server` (Go tests, -p 1 serial)
  - Web tests (if any exist)
  - `govulncheck` (Go vulnerability scanner)
  - `bun audit` (npm audit)
  - `bun run api:check` (API drift check — regenerates and diffs)
  - gen-types annotation check (prevents go-jet types in swagger annotations)

**Commit before pushing** if you want to reformat locally first (hooks auto-stage fixed files).

# Git Commit Style

- Commit messages must be a single line only — no body, no trailers.
- Never add "Co-Authored-By" (or any other) trailers to commits.
- Keep messages simple but meaningful, following conventional commits (e.g. `fix(web): ...`, `feat(server): ...`).

## Agent Development Kit

**Located in `.agents/`:**
- `rules/` — Governance rules (architecture, code-style, testing-and-tdd, security, naming-and-structure, commands, database)
- `agents/` — Agent definitions (subagents for specialized tasks)
- `skills/` — Skill definitions (reusable prompt templates)

**Configuration:**
- `opencode.json` points to `AGENTS.md` and `.agents/rules/*.md` for instructions
- Skills auto-discovered from `.agents/skills/SKILL.md` files
- Subagents in `.opencode/agents/` (mode: subagent)

**Key rules files:**
- `code-style.md` — Biome + formatting conventions
- `testing-and-tdd.md` — Testing strategy for both Go and React
- `architecture.md` — Monorepo structure and boundaries
- `security.md` — Auth, validation, SQL injection prevention
- `naming-and-structure.md` — Variable/function/file naming conventions
- `commands.md` — All available commands
- `database.md` — PostgreSQL, go-jet, migrations

## Environment Setup

**Required environment variables** (in `server-go/.env`):
```env
DATABASE_URL=postgres://postgres:postgres@localhost:5432/skillpass?sslmode=disable
JWT_SECRET=your-secret-key-here
PORT=1234
CORS_ORIGIN=http://localhost:4200
```

**Docker:** `docker-compose.yml` defines PostgreSQL service. Start with `docker compose up db -d` before running migrations.

## Testing Strategy

**Go server:**
- Tests in `*_test.go` files alongside code
- Use Go's `testing` package + `httptest` for handler tests
- Tests require a live DB (`SKILLPASS_TEST_DATABASE_URL`) — they truncate tables for isolation
- Run with `-p 1` (serial) because packages share one DB
- Test utilities in `internal/testutil/` (auth helpers, DB setup, factories)

**React web:**
- Tests in `**/*.test.ts` or `**/*.test.tsx`
- Use vitest + @testing-library/react + happy-dom
- Test utilities in `src/test/` (renderWithClient for TanStack Query)
- Mock API calls via `src/lib/api.ts`

**Current state:** Go has handler tests for most domains. Web has tests for components and query client. See `.agents/rules/testing-and-tdd.md` for guidance.

## Common Patterns

### Adding a new API endpoint

1. Define handler in `server-go/internal/handlers/` (or create a new domain package)
2. Register route in `server-go/cmd/server/main.go` under `api := r.Group("/api/v1")`
3. Use middleware (e.g., `AuthRequired(cfg.JWTSecret)`) for auth-gated routes
4. Return JSON with camelCase field names using named response structs
5. Add swagger annotations for auto-generated docs
6. Call from frontend via `api()` wrapper in `src/lib/api.ts`
7. Run `bun run api:generate` to regenerate types

### Adding a new React page

1. Create folder `src/pages/PageName/index.tsx` (+ optional `type.tsx` for interfaces)
2. Add route in `src/App.tsx` (inside `<RootLayout>`, wrapped in `<ProtectedRoute>` if auth-gated)
3. Use `api()` from `src/lib/api.ts` for HTTP calls
4. Use TanStack Query (`useQuery`/`useMutation`) for server state
5. Use react-hook-form + Zod for forms
6. Style with Tailwind + DaisyUI classes (see DESIGN.md patterns)
7. Lazy-load with `React.lazy(() => import('./pages/PageName'))`

### Database schema changes

1. Write new SQL file in `server-go/migrations/` (naming: `000018_<kebab-name>.sql`)
2. Run `bun run db:migrate`
3. Run `bun run db:generate` to regenerate go-jet types
4. Types appear in `server-go/.gen/`, re-import via `server-go/internal/gen/`

### Changing an API request/response shape

1. Edit the named struct in `server-go/internal/handlers/` (or evaluation/application/matching) — never return raw `gin.H` or go-jet `internal/gen` types from success paths
2. Run `bun run api:generate` — regenerates `server-go/docs/` and `web/src/lib/generated/`
3. Commit BOTH the Go change and the regenerated files together
4. Web types come from `@/lib/api-types` (a barrel over `web/src/lib/generated/api.d.ts`) — never hand-write API response interfaces

## Documentation

- **Architecture:** AGENTS.md (agent-focused), this file (developer-focused)
- **Design System:** DESIGN.md (typography, colors, spacing, components, accessibility)
- **Specs & Plans:** `docs/specs/` and `docs/plans/`
