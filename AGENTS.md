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
│   └── internal/   — handlers/, middleware/, db/, config/, gen/, lib/
├── web/            — React SPA — entrypoint: src/main.tsx
│   └── src/        — pages/, components/, hooks/useAuth.tsx, lib/api.ts
└── docs/           — specs, plans, migration docs
```

## Essential commands (run from root)

| Action | Command |
|---|---|
| Dev (server + web concurrently) | `bun run dev` |
| Dev server only | `bun run dev:server` |
| Dev web only | `bun run dev:web` |
| DB migrate | `bun run db:migrate` |
| DB seed | `bun run db:seed` |
| DB generate (go-jet codegen) | `bun run db:generate` |
| Start fresh | `docker compose up db -d && bun run db:migrate && bun run db:seed` |
| Typecheck web | `bun --cwd web typecheck` (tsc --noEmit) |
| Lint all | `bun run lint` (Biome check) |
| Lint + auto-fix | `bun run lint:fix` (Biome check --write --unsafe) |
| Format all | `bun run format` (Biome format --write) |
| Format check | `bun run format:check` (Biome format, read-only) |
| Test web | `bun --cwd web test` (vitest) |
| Build web (tsc + vite) | `bun run build` |
| Docker full stack | `bun run docker:up` / `bun run docker:down` |

**Local dev startup (non-Docker)** — the server connects to PostgreSQL on `localhost:5432` by default.
Before running `bun run dev`, you must:

1. `docker compose up db -d`           — start the database container
2. `bun run db:migrate`                — run SQL migrations
3. `bun run db:seed`                   — seed initial data
4. `bun run dev`                        — now safe to start server + web

**Docker full stack** — `bun run docker:up` / `bun run docker:down` runs everything in containers (DB, server, web).

> The Go server reads `.env` from `server-go/.env` via godotenv. No env vars need to be set manually in dev.

## Dev URLs
- Web: http://localhost:4200
- API: http://localhost:1234
- Vite proxies `/api` → `:1234` (see web/vite.config.ts)

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

## Frontend conventions
- API calls go through `src/lib/api.ts` — auto-attaches Bearer token, auto-refreshes on 401
- Always use the `api()` wrapper from `lib/api.ts` for authenticated requests (never raw `fetch` to `/api/v1/...`)
- Path alias `@/*` → `src/*` (tsconfig paths)
- Auth state via `AuthProvider` in `hooks/useAuth.tsx` — reads tokens from localStorage
- Token storage: `accessToken` + `refreshToken` in localStorage
- Route definitions in `src/App.tsx`, inside `<AuthProvider>` + `<RootLayout>`

## Styling
- Tailwind v4: no `tailwind.config.*`. Config is in `web/src/styles/index.css` via `@import "tailwindcss"; @plugin "daisyui";`
- Uses `@tailwindcss/vite` plugin (not PostCSS)

## DB / go-jet
- go-jet code generator (database-first): `bun run db:generate` runs `jet` CLI against live DB
- Raw SQL migrations in `server-go/migrations/`
- Generated types in `server-go/.gen/`, re-exported through `server-go/internal/gen/`

## Testing
- **No tests written yet**
- Web: `vitest` (happy-dom, @testing-library/react)
- Go server: use Go's `testing` package with `httptest`
