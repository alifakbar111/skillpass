# Stage 1: Build web assets
FROM oven/bun:1 AS web-builder
WORKDIR /app/web
COPY web/package.json web/bun.lock* ./
RUN bun install --frozen-lockfile
COPY web/ .
RUN bun run build

# Stage 2: Build Go server (embeds web/dist)
FROM golang:1.26.4-alpine AS server-builder
WORKDIR /app
COPY server-go/ .
COPY --from=web-builder /app/web/dist ./internal/static/dist
RUN CGO_ENABLED=0 GOOS=linux go build -o /skillpass-server ./cmd/server/

# Stage 3: Runtime
FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=server-builder /skillpass-server .
COPY --from=server-builder /app/migrations ./migrations/
EXPOSE 1234
CMD ["./skillpass-server"]
