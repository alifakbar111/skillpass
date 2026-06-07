# Database
- PostgreSQL, accessed via pgx pool + go-jet query builder
- Raw SQL migrations in `server-go/migrations/` (run via `bun run db:migrate`)
- go-jet codegen: `bun run db:generate` (requires live DB)
- Generated types in `server-go/.gen/`, re-exported via `server-go/internal/gen/`
- Always run `db:migrate` then `db:generate` after schema changes
