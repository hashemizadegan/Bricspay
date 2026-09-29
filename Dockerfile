# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# کپی سورس کد
COPY . .

# حذف go.sum قدیمی و ساخت مجدد و معتبر آن بر اساس کدهای موجود
RUN rm -f go.sum && go mod tidy

# کامپایل پروژه
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bricspay-api ./cmd/api

# Run stage
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/bricspay-api .
COPY --from=builder /app/internal/api/static ./internal/api/static

EXPOSE 8080

CMD ["./bricspay-api"]
