# -------------------------------------------------------------
# Stage 1: Build the Go binary
# -------------------------------------------------------------
FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY . .

RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-w -s" -o bricspay-api ./cmd/api

# -------------------------------------------------------------
# Stage 2: Final minimal runtime image
# -------------------------------------------------------------
FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

RUN mkdir -p /data/uploads && chmod 700 /data/uploads

COPY --from=builder /app/bricspay-api .

COPY --from=builder /app/internal/api/static ./static

EXPOSE 8080

CMD ["./bricspay-api"]
