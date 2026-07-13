# Stage 1: Build web assets
FROM oven/bun:1 AS web-builder
WORKDIR /app/web
COPY web/package.json web/bun.lock ./
RUN --mount=type=cache,target=/root/.bun/install/cache \
    bun install --frozen-lockfile
COPY web/ .
RUN bun run build

# Stage 2: Build all Go binaries (server, migrate, seed)
FROM golang:1.26.4-alpine AS go-builder
WORKDIR /app
COPY server-go/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /skillpass-server ./cmd/server/
RUN CGO_ENABLED=0 GOOS=linux go build -o /skillpass-migrate ./cmd/migrate/
RUN CGO_ENABLED=0 GOOS=linux go build -o /skillpass-seed ./cmd/seed/

# Stage 3: Build web assets
FROM oven/bun:1 AS web-builder
WORKDIR /app/web
COPY web/package.json web/bun.lock ./
RUN --mount=type=cache,target=/root/.bun/install/cache \
    bun install --frozen-lockfile
COPY web/ .
RUN bun run build

# Stage 4: Server runtime
FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=go-builder /skillpass-server .
COPY --from=go-builder /skillpass-migrate .
COPY --from=go-builder /skillpass-seed .
COPY --from=go-builder /app/migrations ./migrations/
COPY --from=web-builder /app/web/dist ./internal/static/dist
EXPOSE 1234
CMD ["./skillpass-server"]
