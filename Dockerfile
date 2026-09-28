# -------------------------------------------------------------
# Stage 1: Build the Go binary
# -------------------------------------------------------------
FROM golang:1.23-alpine AS builder

WORKDIR /app

# نصب ابزارهای مورد نیاز
RUN apk add --no-cache git ca-certificates

# کپی کل سورس کد به همراه ماژول‌ها
COPY . .

# ساخت و تکمیل خودکار go.sum و دانلود پکیج‌ها
RUN go mod tidy

# کامپایل مستقل و استاتیک باینری
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-w -s" -o bricspay-api ./cmd/api

# -------------------------------------------------------------
# Stage 2: Final minimal runtime image
# -------------------------------------------------------------
FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

# ساخت دایرکتوری ذخیره‌سازی مدارک KYC
RUN mkdir -p /data/uploads && chmod 700 /data/uploads

# کپی باینری کامپایل‌شده
COPY --from=builder /app/bricspay-api .

# کپی فایل‌های استاتیک فرانت‌اند (در صورت وجود)
COPY --from=builder /app/internal/api/static ./static

EXPOSE 8080

CMD ["./bricspay-api"]
