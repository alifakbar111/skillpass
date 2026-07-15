# Database
- PostgreSQL, accessed via pgx pool + Bun ORM
- Raw SQL migrations in `server-go/migrations/` (run via `bun run db:migrate`)
- Bun codegen: `bun run db:generate` (requires live DB)
- Generated types in `server-go/internal/models/`
- Always run `db:migrate` then `db:generate` after schema changes
