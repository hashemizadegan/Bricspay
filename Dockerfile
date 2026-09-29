# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# ابتدا تمام سورس کد را کپی می‌کنیم
COPY . .

# با توجه به سورس کدها، go.sum را به صورت خودکار و دقیق تولید می‌کنیم
RUN go mod tidy

# کامپایل برنامه
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bricspay-api ./cmd/api

# Run stage
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/bricspay-api .
COPY --from=builder /app/internal/api/static ./internal/api/static

EXPOSE 8080

CMD ["./bricspay-api"]
