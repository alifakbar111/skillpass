# Stage 1: Build all Go binaries (server, migrate, seed)
FROM golang:1.26.4-alpine AS go-builder
WORKDIR /app
COPY server-go/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /skillpass-server ./cmd/server/
RUN CGO_ENABLED=0 GOOS=linux go build -o /skillpass-migrate ./cmd/migrate/
RUN CGO_ENABLED=0 GOOS=linux go build -o /skillpass-seed ./cmd/seed/

# Stage 2: Build web assets
FROM oven/bun:1 AS web-builder
WORKDIR /app/web
COPY web/package.json web/bun.lock ./
RUN --mount=type=cache,target=/root/.bun/install/cache \
    bun install --frozen-lockfile
COPY web/ .
RUN bun run build

# Stage 3: Server runtime
FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user (uid 1000) — container escapes get a low-privilege user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup -u 1000

WORKDIR /app
COPY --from=go-builder /skillpass-server .
COPY --from=go-builder /skillpass-migrate .
COPY --from=go-builder /skillpass-seed .
COPY --from=go-builder /app/migrations ./migrations/
COPY --from=web-builder /app/web/dist ./internal/static/dist

# Ensure uploads directory exists and is writable by appuser
RUN mkdir -p /app/uploads && chown -R appuser:appgroup /app

EXPOSE 1234
USER appuser
CMD ["./skillpass-server"]
