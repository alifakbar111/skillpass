# SkillPass — Agent Guide

## Stack
- **Runtime**: Go 1.26 (server), Bun (web tooling)
- **Server**: Gin (Go) — migrated from Elysia (Bun)
- **Frontend**: React 19 SPA (not Next.js), React Router v7, Vite 7
- **Styling**: Tailwind CSS v4 + DaisyUI 5 (no `tailwind.config.*` — uses `@import "tailwindcss"; @plugin "daisyui"` in CSS)
- **DB**: PostgreSQL + go-jet (codegen)
- **Linter**: Biome (single binary, replaces ESLint + Prettier)

## Agent Dev Kit

- Rules: `.agents/rules/`  · Skills: `.agents/skills/`  · Agents: `.agents/agents/`
- Deterministic checks: git hooks (`lefthook.yml`)
- Docs output: specs go to `docs/specs/`, plans go to `docs/plans/`

### MiMo Orchestrator

- Orchestrator skill: `.agents/skills/mimo-orchestrator/SKILL.md`
- Dispatch templates: `.agents/skills/mimo-orchestrator/dispatch-templates.md`
- Configuration: `.mimocode/config.json`
- Agent registry: `.agents/agents/*.md` (auto-discovered)
- Subagent types: `explore` (read-only), `general` (full capabilities)

### Per-Tool Enablement (opencode)

**Setup:**
1. `bun add -D lefthook && bun run lefthook install` — install git hooks
2. Skills auto-discovered from `.agents/skills/`
3. Subagents in `.opencode/agents/` (`mode: subagent`)

> `opencode.json` points to `AGENTS.md` and `.agents/rules/*.md` for instructions.

## Monorepo layout
```
skillpass/          — root: orchestration (concurrently runs both)
├── server-go/      — Go (Gin) API — entrypoint: cmd/server/main.go
│   ├── cmd/
│   │   ├── server/ — main.go
│   │   ├── migrate/ — SQL migration runner
│   │   └── seed/   — DB seeder
│   ├── internal/   — handlers/, middleware/, db/, config/, gen/, lib/,
│   │                 evaluation/, application/, matching/, resume/,
│   │                 email/, notification/, analytics/, authtoken/,
│   │                 storage/, webhook/, testutil/, static/,
│   │                 career/, companyreviews/, feedback/, hris/,
│   │                 rbac/, spdid/, profileviews/
│   ├── migrations/ — 17 SQL DDL files (000001-000017)
│   ├── .gen/       — go-jet generated types
│   └── docs/       — Swagger spec
├── web/            — React SPA — entrypoint: src/main.tsx
│   └── src/
│       ├── pages/          — page folders (index.tsx per page)
│       │   ├── jobseeker/  — EvaluationPage, ApplicationsPage, MatchesPage, etc.
│       │   ├── company/    — FeedbackHistoryPage, ReputationPage
│       │   └── hris/       — EmployeeList, EmployeeCreate, OrgChart, etc.
│       ├── components/     — layout/, ui/, jobseeker/, company/, hris/, onboarding/, passport/
│       ├── hooks/          — useAuth.tsx, useIndustries.ts, usePermissions.ts
│       ├── lib/            — api.ts, api-types.ts, domain modules, generated types, schemas/
│       └── stories/        — Storybook stories
├── .agents/        — Agent definitions, rules, skills
└── docs/           — specs, plans, migration docs
```

## Essential commands (run from root)

| Action | Command |
|---|---|
| Full setup (fresh clone) | `bun run setup` — starts DB, runs migrations & seed |
| Dev (server + web concurrently) | `bun run dev` |
| Dev server only | `bun run dev:server` |
| Dev web only | `bun run dev:web` |
| DB migrate | `bun run db:migrate` |
| DB seed | `bun run db:seed` |
| DB generate (go-jet codegen) | `bun run db:generate` |
| API generate (swag + openapi-typescript) | `bun run api:generate` |
| API drift check (pre-push gate) | `bun run api:check` |
| Start fresh | `bun run setup` |
| Typecheck web | `bun --cwd web typecheck` (tsc --noEmit) |
| Lint all | `bun run lint` (Biome check) |
| Lint + auto-fix | `bun run lint:fix` (Biome check --write) |
| Format all | `bun run format` (Biome format --write) |
| Format check | `bun run format:check` (Biome format, read-only) |
| Test web | `bun --cwd web test` (vitest) |
| Test server | `bun run test:server` (go test -p 1) |
| Build web (tsc + vite) | `bun run build` |
| Docker full stack | `bun run docker:up` / `bun run docker:down` |
| Storybook | `bun --cwd web storybook` (port 6006) |

**Local dev startup (non-Docker)** — the server connects to PostgreSQL on `localhost:5432` by default.
Before running `bun run dev`, you must:

1. `bun run setup`                      — starts DB, runs migrations & seed (does all 3 steps below)
2. `bun run dev`                        — now safe to start server + web

Or step by step:
1. `docker compose up db -d`           — start the database container
2. `bun run db:migrate`                — run SQL migrations
3. `bun run db:seed`                   — seed initial data
4. `bun run dev`                        — now safe to start server + web

**Docker full stack** — `bun run docker:up` / `bun run docker:down` runs everything in containers (DB, server, web).

> The Go server reads `.env` from `server-go/.env` via godotenv. No env vars need to be set manually in dev.

## Dev URLs
- Web: http://localhost:4200
- API: http://localhost:1234
- Storybook: http://localhost:6006
- Vite proxies `/api` and `/uploads` → `:1234` (see web/vite.config.ts)

## Server conventions (Go / Gin)
- Routes registered in `cmd/server/main.go` using `gin.Group("/api/v1/...")`
- Body binding via `c.ShouldBindJSON(&struct)` with struct tags
- JWT auth via `internal/middleware/auth.go` — `AuthRequired(jwtSecret)` middleware parses Bearer token, sets `userId` + `role` in context
- Role guards: `RequireRole("company")` + `RequireVerifiedCompany(pool)` middleware
- Password hashing: `internal/lib/password.go` — bcrypt (default) + argon2id fallback for existing hashes. Cost = `BcryptCost` (default 4 for dev)
- Config from `internal/config/config.go` — reads `JWT_SECRET`, `DATABASE_URL`, `PORT`, `CORS_ORIGIN` from `.env` file or env vars
- DB: pgx pool (`internal/db/db.go`), raw SQL queries + go-jet query builder
- go-jet generated types in `.gen/` directory, re-exported via `internal/gen/`
- All responses use **camelCase** JSON field names

### API response shape (important gotcha)

When changing an API request/response shape:
1. Edit the **handler-level response struct** in `server-go/internal/handlers/` (or `evaluation/`, `application/`, `matching/`) — never return raw `gin.H` or go-jet `internal/gen/` types from success paths
2. Run `bun run api:generate` — regenerates `server-go/docs/` (swagger) and `web/src/lib/generated/` (TypeScript types)
3. Commit **both** the Go change and the regenerated files together
4. Web types come from `@/lib/api-types` (barrel over `web/src/lib/generated/api.d.ts`) — never hand-write API response interfaces

> **Pre-push hook enforces API drift check** (`bun run api:check`). If you change a response struct without running `api:generate`, the hook will fail.

## Frontend conventions
- API calls go through `src/lib/api.ts` — auto-attaches Bearer token, auto-refreshes on 401
- Always use the `api()` wrapper from `lib/api.ts` for authenticated requests (never raw `fetch` to `/api/v1/...`)
- TanStack Query v5 for server state — `useQuery`, `useMutation`, `queryClient` in `lib/queryClient.ts`
- react-hook-form + Zod for form validation — schemas in `lib/schemas/`
- Path alias `@/*` → `src/*` (tsconfig paths)
- Auth state via `AuthProvider` in `hooks/useAuth.tsx` — reads tokens from localStorage
- Token storage: `accessToken` + `refreshToken` in localStorage
- Route definitions in `src/App.tsx`, inside `<QueryClientProvider>` + `<AuthProvider>` + `<ErrorBoundary>` + `<Suspense>`
- Lazy-loaded routes via `React.lazy` + `Suspense`
- **Accessibility:** WCAG 2.1 AA — skip links, ARIA labels, focus management, menu semantics

## Styling
- Tailwind v4: no `tailwind.config.*`. Config is in `web/src/styles/index.css` via `@import "tailwindcss"; @plugin "daisyui";`
- Uses `@tailwindcss/vite` plugin (not PostCSS)
- Zero custom CSS — all utility classes from Tailwind + DaisyUI
- Read `DESIGN.md` for color tokens, typography, spacing, and component patterns

## DB / go-jet
- go-jet code generator (database-first): `bun run db:generate` runs `jet` CLI against live DB
- Raw SQL migrations in `server-go/migrations/` (numbered `000001_init.sql` through `000017_phase3_profile_views.sql`)
- Generated types in `server-go/.gen/`, re-exported through `server-go/internal/gen/`
- Migration naming: `000018_<kebab-name>.sql`

## Testing
- Go has handler tests for most domains (auth, jobs, profiles, companies, search, admin, etc.)
- Web: `vitest` (happy-dom, @testing-library/react) — tests in `src/**/*.test.{ts,tsx}`
- Go server: use Go's `testing` package with `httptest`
- Go tests require a live DB (`SKILLPASS_TEST_DATABASE_URL`) — they truncate tables for isolation
- Go tests run with `-p 1` (serial) because packages share one DB
- CI runs: Go tests, web typecheck, web tests, web build

## Git hooks (lefthook)
- **pre-commit**: `bun run format` (auto-fix code style, auto-stages)
- **pre-push**: Go tests, web tests (if any), govulncheck, `bun audit`, API drift check, gen-types annotation check
- The `no-gen-types-in-annotations` hook prevents go-jet `internal/gen` types from appearing in `@Success` swagger annotations — always wrap in a handler response struct

## Git commits
- Commit messages must be a single line only — no body, no trailers
- Never add "Co-Authored-By" (or any other) trailers
- Simple but meaningful, conventional commits style (e.g. `fix(web): ...`, `feat(server): ...`)
