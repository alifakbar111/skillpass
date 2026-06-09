---
name: db-migration
description: |-
  Use this agent when asked to create a SQL migration, add or modify a database table or column, or trigger go-jet codegen after schema changes. Examples:

  <example>
  Context: User wants to add a skills table to the database.
  user: "Add a skills table with name, category, and verified fields"
  assistant: "I'll use db-migration to create a timestamped migration file and remind you to run db:migrate && db:generate."
  <commentary>
  Schema additions require a timestamped migration file — this agent's specialty.
  </commentary>
  </example>

  <example>
  Context: User needs to add a column to the existing jobs table.
  user: "Add a salary_range column to the jobs table"
  assistant: "I'll dispatch db-migration to write the ALTER TABLE migration and outline the go-jet codegen steps."
  <commentary>
  Column additions to existing tables require a migration, which this agent handles.
  </commentary>
  </example>
model: sonnet
color: blue
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
