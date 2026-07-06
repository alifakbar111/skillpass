# Testing & TDD
- Web: vitest with happy-dom + @testing-library/react
- Server: Go `testing` package with `httptest`
- No tests written yet — prefer TDD for new features
- Run tests: `bun --cwd web test` (web) or `bun run test:server` (Go, auto-creates `skillpass_test` DB)
