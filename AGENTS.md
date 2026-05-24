# SkillPass — Agent Guide

## Stack
- **Runtime**: Bun (not Node.js — all commands use `bun`)
- **Server**: Elysia 1.x with `@elysiajs/jwt`, `@elysiajs/cors`, `@elysiajs/swagger`
- **Frontend**: React 19 SPA (not Next.js), React Router v7, Vite 7
- **Styling**: Tailwind CSS v4 + DaisyUI 5 (no `tailwind.config.*` — uses `@import "tailwindcss"; @plugin "daisyui"` in CSS)
- **DB**: PostgreSQL + Drizzle ORM (`drizzle-kit`)
- **Linter**: Biome (single binary, replaces ESLint + Prettier)

## Monorepo layout
```
skillpass/          — root: orchestration (concurrently runs both)
├── server/         — Elysia API entrypoint: src/index.ts
│   ├── src/db/     — Drizzle schema + client
│   ├── src/routes/ — One file per domain, export named *Routes
│   └── src/middleware/auth.ts
├── web/            — React SPA entrypoint: src/main.tsx
│   └── src/        — pages/, components/, hooks/useAuth.tsx, lib/api.ts
└── docs/superpowers/ — specs + plans
```

## Essential commands (run from root)

| Action | Command |
|---|---|
| Dev (server + web concurrently) | `bun run dev` |
| Dev server only | `bun run dev:server` |
| Dev web only | `bun run dev:web` |
| Push Drizzle schema | `bun run db:push` |
| Seed DB (industries) | `bun run db:seed` |
| Start fresh | `docker compose up db -d && bun run db:push && bun run db:seed` |
| Typecheck server | `bun --cwd server typecheck` (tsc --noEmit) |
| Typecheck web | `bun --cwd web typecheck` (tsc --noEmit) |
| Lint all | `bun run lint` (Biome check) |
| Lint + auto-fix | `bun run lint:fix` (Biome check --write --unsafe) |
| Format all | `bun run format` (Biome format --write) |
| Format check | `bun run format:check` (Biome format, read-only) |
| Test server | `bun --cwd server test` (Bun test runner) |
| Test web | `bun --cwd web test` (vitest) |
| Build web (tsc + vite) | `bun run build` |
| Docker full stack | `bun run docker:up` / `bun run docker:down` |

**Local dev startup (non-Docker)** — the server connects to PostgreSQL on `localhost:5432` by default.
Before running `bun run dev`, you must:

1. `docker compose up db -d`           — start the database container
2. `bun run db:push`                   — push schema to the DB
3. `bun run db:seed`                   — seed initial data
4. `bun run dev`                        — now safe to start server + web

**Docker full stack** — `bun run docker:up` / `bun run docker:down` runs everything in containers (DB, server, web).

> If you skip step 1, the server will fail to start with a clear error message telling you to run `docker compose up db -d`.

## Dev URLs
- Web: http://localhost:4200
- API: http://localhost:8800
- Swagger: http://localhost:8800/docs
- Vite proxies `/api` → `:8800` (see web/vite.config.ts)

## Server conventions
- Routes use `new Elysia({ prefix: '/api/v1/...' })` pattern
- Body validation via Elysia `t.*` (`t.String`, `t.Object`, etc.)
- JWT auth uses `@elysiajs/jwt` plugin + `.resolve()` hook in route files (see `src/routes/profiles.ts` for pattern)
- Password hashing uses `Bun.password.hash` (native bcrypt, no npm package)
- All routes registered explicitly in `src/index.ts`
- `process.env.JWT_SECRET` — defaults to `'skillpass-dev-secret-change-in-prod'`
- **Gotcha**: `server/src/middleware/auth.ts` is **dead code** — never imported. Each route file does its own JWT verification via `.resolve()` hook inline.
- **Gotcha**: `JWT_SECRET` is redundantly copy-pasted with fallback default in **6 route files**. Adding a new route requires the same pattern (`const JWT_SECRET = ...` + `.use(jwt(...))` + `.resolve(...)`).
- Seed (`server/seed.ts`) only inserts industry categories — extend it if adding look-up data.

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

## DB / Drizzle
- Schema: `server/src/db/schema.ts`
- Migrations: `drizzle-kit push` (no migration file workflow — direct push)
- Seed: `server/seed.ts` (only industry categories)
- No migration history committed; `drizzle/` is gitignored

## Testing
- **No tests written yet** in either package
- Server: `bun test` (Bun built-in)
- Web: `vitest` (happy-dom, @testing-library/react)
- Match the package's runner when adding tests
