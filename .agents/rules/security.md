# Security
- JWT auth: Bearer token, parsed by `AuthRequired(jwtSecret)` middleware — sets `userId` + `role` in context
- Role guards: `RequireRole("company")` + `RequireVerifiedCompany(pool)` middleware
- Password hashing: bcrypt (default cost 4 for dev) + argon2id fallback via `internal/lib/password.go`
- Config from `.env`: `JWT_SECRET`, `DATABASE_URL`, `PORT`, `CORS_ORIGIN`
- All API responses use camelCase (no accidental schema leakage)
