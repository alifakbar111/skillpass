# Commands
| Task | Command |
|------|---------|
| dev (server + web) | `bun run dev` |
| dev server only | `bun run dev:server` |
| dev web only | `bun run dev:web` |
| test web | `bun --cwd web test` |
| test server | `go test ./server-go/...` |
| lint all | `bun run lint` |
| lint + auto-fix | `bun run lint:fix` |
| format all | `bun run format` |
| format check | `bun run format:check` |
| build web | `bun run build` |
| db migrate | `bun run db:migrate` |
| db seed | `bun run db:seed` |
| db generate (go-jet codegen) | `bun run db:generate` |
| docker full stack up | `bun run docker:up` |
| docker full stack down | `bun run docker:down` |
| typecheck web | `bun --cwd web typecheck` |
