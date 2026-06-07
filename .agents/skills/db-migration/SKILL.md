---
name: db-migration
description: Create timestamped SQL migration files for SkillPass PostgreSQL database. Use when adding or modifying database schema.
---

# DB Migration

Creates SQL migration files following SkillPass conventions.

## Method

1. Create new timestamped file in `server-go/migrations/` — naming: `YYYYMMDDHHMMSS_<description>.sql`
2. Write up/down SQL matching existing patterns.
3. Remind to run: `bun run db:migrate && bun run db:generate` after creation.

## Naming Convention

Files use timestamped naming: `YYYYMMDDHHMMSS_<kebab-description>.sql`

## Return

Migration file path, summary of schema changes.
