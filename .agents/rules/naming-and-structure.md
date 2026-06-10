# Naming & Structure
- Go files: `snake_case.go` in feature packages under `server-go/internal/`
- React files: `PascalCase.tsx` for components, `camelCase.ts` for hooks/lib
- **Page folders**: each page is a folder under `web/src/pages/` with `index.tsx` (component) + optional `type.tsx` (interfaces)
- Go structs: PascalCase with JSON tags (`json:"camelCase"`)
- Frontend path alias: `@/*` → `web/src/*`
- DB tables: `snake_case`, generated as PascalCase by go-jet
