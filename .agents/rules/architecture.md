# Architecture
- Monorepo: root orchestration (`bun run dev` via concurrently), `server-go/` (Go), `web/` (React/Vite)
- API: Gin groups at `/api/v1/...`, JWT auth via `internal/middleware/auth.go`
- DB: pgx connection pool (`internal/db/db.go`), Bun ORM + raw SQL migrations
- Frontend: React 19 SPA (not Next.js), React Router v7 client-side routing
- Bun model structs in `server-go/internal/models/`
