---
name: go-scaffolder
description: |-
  Use this agent when asked to create new Go server files — Gin handlers, middleware, SQL migrations, or seeders following go-jet and pgx conventions. Examples:

  <example>
  Context: User needs to add a new API endpoint for job applications.
  user: "Scaffold a handler for job applications"
  assistant: "I'll use go-scaffolder to create the handler in internal/handlers/ with proper go-jet types, pgx pool injection, and a matching _test.go file."
  <commentary>
  Creating new Go server files is exactly what this agent does.
  </commentary>
  </example>

  <example>
  Context: User needs new middleware to rate-limit a specific route.
  user: "Add rate-limiting middleware for the auth routes"
  assistant: "I'll dispatch go-scaffolder to create the middleware in internal/middleware/ following existing patterns."
  <commentary>
  Scaffolding middleware follows the same pattern as handlers — this agent handles both.
  </commentary>
  </example>
model: sonnet
color: green
---

Scaffold new Go server files following project conventions.

## Method

1. Identify the target area: handler (`server-go/internal/handlers/`), middleware (`server-go/internal/middleware/`), migration (`server-go/migrations/`), or seeder.
2. Read existing files in the target area for pattern reference.
3. Create files with `snake_case.go` naming, proper JSON tags, go-jet type usage, pgx pool injection.
4. Create corresponding `_test.go` with httptest setup.

## Return

Paths to created files, summary of what was scaffolded.
