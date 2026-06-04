FROM golang:1.24-alpine AS builder
WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /skillpass-server ./cmd/server/

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /skillpass-server .
COPY migrations/ ./migrations/

EXPOSE 1234
CMD ["./skillpass-server"]
