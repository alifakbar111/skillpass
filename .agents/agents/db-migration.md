---
name: db-migration
description: "Create timestamped SQL migration files and trigger go-jet codegen"
---

Create SQL migrations and manage go-jet codegen workflow.

## Method

1. Create new timestamped file in `server-go/migrations/` — naming: `YYYYMMDDHHMMSS_<description>.sql`
2. Write up/down SQL matching existing migration patterns.
3. Remind to run: `bun run db:migrate && bun run db:generate` after creation.

## Naming Convention

Files use timestamped naming: `YYYYMMDDHHMMSS_<kebab-description>.sql`

## Return

Migration file path, summary of schema changes.