---
name: go-scaffold
description: Scaffold Gin handlers, middleware, SQL migrations, and seeders following project conventions (Gin groups, go-jet, pgx pool, httptest). Use when creating new Go server files.
---

# Go Scaffold

Creates new Go server files following SkillPass conventions.

## Method

1. Identify the target area: handler (`server-go/internal/handlers/`), middleware (`server-go/internal/middleware/`), migration (`server-go/migrations/`), or seeder.
2. Read existing files for pattern reference.
3. Create files with `snake_case.go` naming, proper JSON tags, go-jet type usage, pgx pool injection.
4. Create corresponding `_test.go` with httptest setup.

## Return

Paths to created files, summary of what was scaffolded.
