# Go Scaffolder (Cline Workflow)

Invoke via: `/go-scaffolder <target>`

Scaffold new Go server files following project conventions.

## Method

1. Identify the target area: handler (`server-go/internal/handlers/`), middleware (`server-go/internal/middleware/`), migration (`server-go/migrations/`), or seeder.
2. Read existing files in the target area for pattern reference.
3. Create files with `snake_case.go` naming, proper JSON tags, Bun model struct usage, pgx pool injection.
4. Create corresponding `_test.go` with httptest setup.

## Return

Paths to created files, summary of what was scaffolded.
