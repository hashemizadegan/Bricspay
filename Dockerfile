# Stage 1: Build
FROM golang:1.22-alpine AS builder

WORKDIR /app

# کپی کل پروژه
COPY . .

# دانلود و ساخت وابستگی‌ها بدون سخت‌گیری روی go.sum
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o bricspay-api ./cmd/api

# Stage 2: Run
FROM alpine:3.19

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/bricspay-api .

EXPOSE 8080

CMD ["./bricspay-api"]
