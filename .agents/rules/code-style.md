# Code Style
- Formatting: Biome (single binary, replaces ESLint + Prettier)
- Minimal comments — explain WHY, not WHAT
- Go: camelCase JSON responses, use gin.H for simple responses
- Frontend: functional components with hooks, use `api()` wrapper for all authenticated requests
- Import order: standard lib → third-party → internal (Go); react → third-party → @/ (frontend)
