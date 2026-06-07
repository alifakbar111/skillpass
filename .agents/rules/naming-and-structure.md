# Naming & Structure
- Go files: `snake_case.go` in feature packages under `server-go/internal/`
- React files: `PascalCase.tsx` for components, `camelCase.ts` for hooks/lib
- Go structs: PascalCase with JSON tags (`json:"camelCase"`)
- Frontend path alias: `@/*` → `web/src/*`
- DB tables: `snake_case`, generated as PascalCase by go-jet
